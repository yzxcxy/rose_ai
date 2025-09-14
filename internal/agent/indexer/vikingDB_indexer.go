package indexer

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/volcengine/volc-sdk-golang/service/vikingdb"
	"rose/internal/agent/embedder"
	"rose/internal/config"
	"rose/internal/utils"
	"strconv"
)

type VectorStoreForVikingDB struct {
	Conf            *config.Config
	Service         *vikingdb.VikingDBService
	Collection      *vikingdb.Collection
	BuiltInEmbModel *vikingdb.EmbModel
	EebModel        embedding.Embedder
}

var fields = []vikingdb.Field{
	{
		FieldName:    "id",
		FieldType:    vikingdb.String,
		IsPrimaryKey: true,
	},
	{
		FieldName: "vector",
		FieldType: vikingdb.Vector,
		Dim:       2048,
	},
	{
		FieldName:  "user",
		FieldType:  vikingdb.String,
		DefaultVal: "unknown user",
	},
	{
		FieldName:  "content",
		FieldType:  vikingdb.String,
		DefaultVal: "",
	},
	{
		FieldName:  "metadata",
		FieldType:  vikingdb.String,
		DefaultVal: "",
	},
}

func NewVectorStoreForVikingDB(conf *config.Config) *VectorStoreForVikingDB {
	service := vikingdb.NewVikingDBService(conf.VikingDB.Host, conf.VikingDB.Region, conf.VikingDB.AK, conf.VikingDB.SK, conf.VikingDB.Scheme)
	if conf.VikingDB.ConnectionTimeout != 0 {
		service.SetConnectionTimeout(conf.VikingDB.ConnectionTimeout)
	}

	collection, err := service.GetCollection(conf.VikingDB.Collection)
	if err != nil {
		// 如果不存在则创建
		collection, err = service.CreateCollection(conf.VikingDB.Collection, fields, "rose ai collection")
		if err != nil {
			panic(err)
		}
	}

	return &VectorStoreForVikingDB{
		Conf:       conf,
		Service:    service,
		Collection: collection,
		EebModel:   embedder.GetEmbedder(conf),
	}
}

func (vs *VectorStoreForVikingDB) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	// TODO 添加callbacks的切入点
	data, err := vs.convertDocuments(ctx, docs)
	if err != nil {
		return nil, err
	}
	for i := 0; i*vs.Conf.VikingDB.Chunk < len(data); i++ {
		batchSize := vs.Conf.VikingDB.Chunk
		if (i+1)*vs.Conf.VikingDB.Chunk > len(data) {
			batchSize = len(data) - i*vs.Conf.VikingDB.Chunk
		}

		// 获得待插入的切片
		d := data[i*vs.Conf.VikingDB.Chunk : i*vs.Conf.VikingDB.Chunk+batchSize]
		// 插入数据
		err := vs.Collection.UpsertData(d)
		if err != nil {
			return nil, err
		}
	}

	ids = make([]string, len(docs))
	for idx, _ := range docs {
		ids[idx] = docs[idx].ID
	}
	return ids, nil
}

func (i *VectorStoreForVikingDB) convertDocuments(ctx context.Context, docs []*schema.Document) (data []vikingdb.Data, err error) {

	uid, _, err := utils.GetUserIdAndUserNameFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user := "user" + strconv.Itoa(int(uid))
	var dense [][]float64

	var texts = make([]string, len(docs))
	for idx, _ := range docs {
		texts[idx] = docs[idx].Content
	}

	// 暂时不支持内置的EmbModel
	if i.Conf.VikingDB.UseBuiltin != true {
		dense, err = i.EebModel.EmbedStrings(ctx, texts)
		if err != nil {
			return nil, err
		}
	}

	data = make([]vikingdb.Data, len(docs))

	for idx, _ := range docs {
		jsonData, err := json.Marshal(docs[idx].MetaData)
		if err != nil {
			return nil, err
		}

		field := map[string]interface{}{
			"id":       docs[idx].ID,
			"vector":   dense[idx],
			"user":     user,
			"content":  docs[idx].Content,
			"metadata": string(jsonData),
		}
		data[idx] = vikingdb.Data{
			Fields: field,
			TTL:    0, // 暂时不过期
		}
	}

	return data, nil
}

//type VectorStoreForMilvus struct {
//	Conf *config.Config
//}
//
//func NewVectorStoreForMilvus(conf *config.Config) *VectorStoreForMilvus {
//	return &VectorStoreForMilvus{
//		Conf: conf,
//	}
//}
//
//const (
//	defaultCollectionID           = "id"
//	defaultCollectionIDDesc       = "the unique id of the document"
//	defaultCollectionVector       = "vector"
//	defaultCollectionVectorDesc   = "the vector of the document"
//	defaultCollectionContent      = "content"
//	defaultCollectionContentDesc  = "the content of the document"
//	defaultCollectionMetadata     = "metadata"
//	defaultCollectionMetadataDesc = "the metadata of the document"
//
//	defaultDim = 65536
//)
//
//func getDefaultFields() []*entity.Field {
//	return []*entity.Field{
//		entity.NewField().
//			WithName(defaultCollectionID).
//			WithDescription(defaultCollectionIDDesc).
//			WithIsPrimaryKey(true).
//			WithDataType(entity.FieldTypeVarChar).
//			WithMaxLength(255),
//		entity.NewField().
//			WithName(defaultCollectionVector).
//			WithDescription(defaultCollectionVectorDesc).
//			WithIsPrimaryKey(false).
//			WithDataType(entity.FieldTypeBinaryVector).
//			WithDim(defaultDim),
//		entity.NewField().
//			WithName(defaultCollectionContent).
//			WithDescription(defaultCollectionContentDesc).
//			WithIsPrimaryKey(false).
//			WithDataType(entity.FieldTypeVarChar).
//			WithMaxLength(2048),
//		entity.NewField().
//			WithName(defaultCollectionMetadata).
//			WithDescription(defaultCollectionMetadataDesc).
//			WithIsPrimaryKey(false).
//			WithDataType(entity.FieldTypeJSON),
//	}
//}
//
//func (vs *VectorStoreForMilvus) Store(ctx context.Context, collectionName string, src []*schema.Document, opts ...document.TransformerOption) ([]string, error) {
//	cli, err := client.NewClient(ctx, client.Config{
//		Address: "localhost:19530",
//		DBName:  "rose_test",
//	})
//	if err != nil {
//		logx.Error(err)
//		return nil, err
//	}
//	defer func(cli client.Client) {
//		err := cli.Close()
//		if err != nil {
//			logx.Error(err)
//		}
//	}(cli)
//
//	agent.embedder = embedder.GetEmbedder(vs.Conf)
//
//	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
//		Collection: collectionName,
//		Client:     cli,
//		Embedding:  agent.embedder,
//		MetricType: milvus.HAMMING,
//		Fields:     getDefaultFields(),
//	})
//	if err != nil {
//		logx.Error(err)
//		panic(err)
//	}
//	ids, err := indexer.Store(ctx, src)
//	if err != nil {
//		return nil, err
//	}
//	return ids, nil
//}
//
//func (vs *VectorStoreForMilvus) Store2(ctx context.Context, collectionName string, src []*schema.Document, opts ...document.TransformerOption) ([]string, error) {
//	cli, err := client.NewClient(ctx, client.Config{
//		Address: "localhost:19530",
//		DBName:  "rose",
//	})
//	if err != nil {
//		logx.Error(err)
//		return nil, err
//	}
//	defer func(cli client.Client) {
//		err := cli.Close()
//		if err != nil {
//			logx.Error(err)
//		}
//	}(cli)
//
//	agent.embedder = embedder.GetEmbedder(vs.Conf)
//
//	fields := []*entity.Field{
//		{
//			Name:       "id",
//			DataType:   entity.FieldTypeVarChar,
//			PrimaryKey: true,
//			TypeParams: map[string]string{
//				"max_length": "64",
//			},
//		},
//		{
//			Name:       "vector", // 向量字段
//			DataType:   entity.FieldTypeFloatVector,
//			TypeParams: map[string]string{"dim": "2048"},
//		},
//		{
//			Name:     "content", // 原文本
//			DataType: entity.FieldTypeVarChar,
//			TypeParams: map[string]string{
//				"max_length": "2048",
//			},
//		},
//		{
//			Name:     "metadata", // 元数据字段
//			DataType: entity.FieldTypeVarChar,
//			TypeParams: map[string]string{
//				"max_length": "2048",
//			},
//		},
//	}
//
//	// 替换默认的文档转换函数，因为其实现是基于byte进行存储的，会导致无法直接使用[][]float64类型的向量数据
//	converterFunc := func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
//		em := make([]MilusSchema, 0, len(docs))
//		texts := make([]string, 0, len(docs))
//		rows := make([]interface{}, 0, len(docs))
//
//		for _, doc := range docs {
//			metadataForBytes, err := sonic.Marshal(doc.MetaData)
//			metadata := string(metadataForBytes)
//			if err != nil {
//				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
//			}
//			em = append(em, MilusSchema{
//				ID:       doc.ID,
//				Content:  doc.Content,
//				Vector:   nil,
//				Metadata: metadata,
//			})
//			texts = append(texts, doc.Content)
//		}
//
//		// build embedder documents for storing
//		for idx, vec := range vectors {
//			em[idx].Vector = Float64ToFloat32(vec)
//			rows = append(rows, &em[idx])
//		}
//		return rows, nil
//	}
//
//	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
//		Collection:        collectionName, // 集合名称
//		Client:            cli,
//		Fields:            fields,
//		Embedding:         agent.embedder,
//		DocumentConverter: converterFunc,
//		MetricType:        milvus.IP,
//	})
//	if err != nil {
//		logx.Error(err)
//		panic(err)
//	}
//	ids, err := indexer.Store(ctx, src)
//	if err != nil {
//		return nil, err
//	}
//	return ids, nil
//}
//
//type MilusSchema struct {
//	ID       string    `json:"id" milvus:"name:id"`
//	Content  string    `json:"content" milvus:"name:content"`
//	Vector   []float32 `json:"vector" milvus:"name:vector"`
//	Metadata string    `json:"metadata" milvus:"name:metadata"`
//}
//
//func Float64ToFloat32(v []float64) []float32 {
//	if v == nil {
//		return nil
//	}
//	res := make([]float32, len(v))
//	for i, val := range v {
//		res[i] = float32(val)
//	}
//	return res
//}
