package indexer

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/agent/embedder"
	"rose/internal/config"
	"rose/internal/utils"
	"strconv"
)

type MilvusIndexer struct {
	Client   *milvusclient.Client
	Conf     *config.Config
	EebModel embedding.Embedder
}

func NewMilvusIndexer(conf *config.Config) (*MilvusIndexer, error) {

	exists := checkDB(context.Background(), conf)
	if !exists {
		err := createDB(context.Background(), conf)
		if err != nil {
			return nil, err
		}
		logx.Infof("created milvus database: %s", conf.Milvus.DBName)
	}

	client, err := milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address: conf.Milvus.Host,
		DBName:  conf.Milvus.DBName,
	})
	if err != nil {
		return nil, err
	}
	return &MilvusIndexer{
		Client:   client,
		Conf:     conf,
		EebModel: embedder.GetEmbedder(conf),
	}, nil
}

func (this *MilvusIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	collectionName, err := getCollectionName(ctx)
	if err != nil {
		return nil, err
	}

	// 查看集合是否存在，如果不存在则创建
	_, err = this.Client.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption(collectionName))
	if err != nil {
		logx.Infof("find collection error:%v", err)
		err = this.CreateCollectionAndIndex(ctx, collectionName)
		if err != nil {
			return nil, err
		}
	}

	// 插入数据
	// 1. 准备数据
	var idsCol []string
	var vectorsCol [][]float32
	var contentsCol []string
	var metaDataCol [][]byte
	var fileNameCol []string

	for i := range docs {
		idsCol = append(idsCol, docs[i].ID)
		contentsCol = append(contentsCol, docs[i].Content)
	}

	vectorFloat64, err := this.EebModel.EmbedStrings(ctx, contentsCol)
	if err != nil {
		return nil, err
	}

	vectorsCol = convertFloat64ToFloat32(vectorFloat64)

	for i := range docs {
		jsonData, err := json.Marshal(docs[i].MetaData)
		if err != nil {
			return nil, err
		}
		metaDataCol = append(metaDataCol, jsonData)
	}

	// 提取文件名
	for i := range docs {
		if name, ok := docs[i].MetaData["_file_name"]; ok {
			fileNameCol = append(fileNameCol, name.(string))
		} else {
			fileNameCol = append(fileNameCol, "unknown")
		}
	}

	// 插入数据
	insertResult, err := this.Client.Insert(ctx, milvusclient.NewColumnBasedInsertOption(collectionName).
		WithVarcharColumn("id", idsCol).
		WithFloatVectorColumn("vector", 2048, vectorsCol).
		WithVarcharColumn("content", contentsCol).
		WithColumns(column.NewColumnJSONBytes("meta_data", metaDataCol)).
		WithVarcharColumn("file_name", fileNameCol))

	if err != nil {
		return nil, err
	}

	logx.Infof("inserted %d rows", insertResult.InsertCount)
	return idsCol, nil
}

// CreateCollectionAndIndex creates a collection and index in Milvus for the given user.
func (this *MilvusIndexer) CreateCollectionAndIndex(ctx context.Context, collectionName string) (err error) {
	schema := entity.NewSchema().WithDynamicFieldEnabled(false).
		WithField(entity.NewField().WithName("id").WithIsAutoID(false).WithDataType(entity.FieldTypeVarChar).WithIsPrimaryKey(true).WithMaxLength(258)).
		WithField(entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(2048)).
		WithField(entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(1024)).
		WithField(entity.NewField().WithName("meta_data").WithDataType(entity.FieldTypeJSON).WithMaxLength(1024)).
		WithField(entity.NewField().WithName("file_name").WithDataType(entity.FieldTypeVarChar).WithMaxLength(1024).WithEnableAnalyzer(true)).
		WithField(entity.NewField().WithName("sparse").WithDataType(entity.FieldTypeSparseVector))

	// 将文本转换为稀疏向量
	function := entity.NewFunction().
		WithName("text_bm25_emb").
		WithInputFields("file_name").
		WithOutputFields("sparse").
		WithType(entity.FunctionTypeBM25)
	schema.WithFunction(function)

	err = this.Client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collectionName, schema))
	if err != nil {
		return err
	}

	// 创建索引（稠密向量的索引）
	hnswIndex := index.NewHNSWIndex(entity.COSINE, 16, 200)
	loadTask, err := this.Client.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collectionName, "vector", hnswIndex))
	if err != nil {
		return err
	}

	// 创建稀疏向量的索引
	indexOption := milvusclient.NewCreateIndexOption(collectionName, "sparse",
		index.NewSparseInvertedIndex(entity.BM25, 0.2))

	indexOption.WithExtraParam("inverted_index_algo", "DAAT_MAXSCORE")
	indexOption.WithExtraParam("bm25_k1", 1.2)
	indexOption.WithExtraParam("bm25_b", 0.75)

	loadTask2, err := this.Client.CreateIndex(ctx, indexOption)
	if err != nil {
		return err
	}

	// 等待装载
	_, err = this.Client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))

	// sync wait collection to be loaded
	err = loadTask.Await(ctx)
	if err != nil {
		return err
	}

	err = loadTask2.Await(ctx)
	if err != nil {
		return err
	}

	_, err = this.Client.GetLoadState(ctx, milvusclient.NewGetLoadStateOption(collectionName))
	if err != nil {
		return err
	}

	return nil
}

func getCollectionName(ctx context.Context) (string, error) {
	uid, _, err := utils.GetUserIdAndUserNameFromContext(ctx)
	if err != nil {
		return "", err
	}

	// 转换uid为字符串
	collectionName := "user_" + strconv.FormatInt(uid, 10)
	return collectionName, nil
}

func convertFloat64ToFloat32(input [][]float64) [][]float32 {
	// 创建目标切片，预分配空间
	result := make([][]float32, len(input))
	for i := range input {
		result[i] = make([]float32, len(input[i]))
		for j := range input[i] {
			result[i][j] = float32(input[i][j])
		}
	}
	return result
}

func checkDB(ctx context.Context, conf *config.Config) bool {
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: conf.Milvus.Host,
	})
	if err != nil {
		return false
	}
	defer client.Close(ctx)

	_, err = client.DescribeDatabase(ctx, milvusclient.NewDescribeDatabaseOption(conf.Milvus.DBName))
	if err != nil {
		// 返回false
		return false
	}
	return true
}

func createDB(ctx context.Context, conf *config.Config) error {
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: conf.Milvus.Host,
	})
	if err != nil {
		return err
	}
	defer client.Close(ctx)

	err = client.CreateDatabase(ctx, milvusclient.NewCreateDatabaseOption(conf.Milvus.DBName))
	if err != nil {
		return err
	}
	return nil
}
