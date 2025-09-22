package custom_retriver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/agent/embedder"
	"rose/internal/config"
	"rose/internal/utils"
	"strconv"
)

type MilvusRetriever struct {
	Client   *milvusclient.Client
	Conf     *config.Config
	EmbModel embedding.Embedder
}

func NewMilvusRetriever(conf *config.Config) (*MilvusRetriever, error) {
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
	return &MilvusRetriever{
		Client:   client,
		Conf:     conf,
		EmbModel: embedder.GetEmbedder(conf),
	}, nil
}

func (this *MilvusRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	//collectionName, err := getCollectionName(ctx)
	//if err != nil {
	//	return nil, err
	//}
	collectionName := "user_77674160750333952"
	vectors, err := this.EmbModel.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	if len(vectors) == 0 {
		return nil, errors.New("empty query")
	}

	var queryVector = make([]float32, len(vectors[0]))
	// 将[]float64转为[]float32
	for i := range vectors[0] {
		queryVector[i] = float32(vectors[0][i])
	}

	annParam := index.NewCustomAnnParam()
	annParam.WithRadius(0.7)
	annParam.WithRangeFilter(1)
	// 以下是新增内容
	annRequest1 := milvusclient.NewAnnRequest("vector", 12, entity.FloatVector(queryVector)).
		WithAnnParam(annParam)

	annParam2 := index.NewSparseAnnParam()
	annParam2.WithDropRatio(0.2)
	// entity.Text(query)实现了entity.Vector接口
	annRequest2 := milvusclient.NewAnnRequest("sparse", 12, entity.Text(query)).
		WithAnnParam(annParam2)

	//reranker := milvusclient.NewRRFReranker().WithK(100)
	reranker := milvusclient.NewWeightedReranker([]float64{0.8, 0.2})
	hybridOpt := milvusclient.NewHybridSearchOption(collectionName, 12,
		annRequest1,
		annRequest2).
		WithReranker(reranker).
		WithOutputFields([]string{"id", "content", "meta_data"}...)

	resultSets, err := this.Client.HybridSearch(ctx, hybridOpt)

	//resultSets, err := this.Client.Search(ctx, milvusclient.NewSearchOption(
	//	collectionName,
	//	10,
	//	[]entity.Vector{entity.FloatVector(queryVector)}).
	//	WithOutputFields([]string{"id", "content", "meta_data"}...).
	//	WithConsistencyLevel(entity.ClStrong).
	//	WithANNSField("vector").
	//	WithAnnParam(annParam),
	//)

	if err != nil {
		return nil, err
	}

	// 解析结果
	var docs []*schema.Document
	// 不能直接range
	for i := range resultSets {
		if resultSets[i].ResultCount == 0 {
			continue
		}
		idScalars := resultSets[i].IDs.FieldData().GetScalars().GetStringData().GetData()
		contentScalars := resultSets[i].GetColumn("content").FieldData().GetScalars().GetStringData().GetData()
		metaDataScalars := resultSets[i].GetColumn("meta_data").FieldData().GetScalars().GetJsonData().GetData()
		scoreScalars := resultSets[i].Scores
		for idx := 0; idx < resultSets[i].ResultCount; idx++ {
			var doc = &schema.Document{}
			doc.ID = idScalars[idx]
			doc.Content = contentScalars[idx]
			err := json.Unmarshal(metaDataScalars[idx], &doc.MetaData)
			if err != nil {
				return nil, err
			}
			docs = append(docs, doc)
			// 出现召回分数为0的情况。
			logx.Infof("召回分数为: %f", scoreScalars[idx])
		}
	}

	return docs, nil
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
