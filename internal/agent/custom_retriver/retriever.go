package custom_retriver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/volcengine/volc-sdk-golang/service/vikingdb"
	"rose/internal/agent/embedder"
	"rose/internal/config"
	"rose/internal/utils"
	"strconv"
	"sync"
)

type VikingDBRetriever struct {
	embModel    embedding.Embedder
	indexClient *vikingdb.IndexClient
	service     *vikingdb.VikingDBService
}

var vikingDBRetriever *VikingDBRetriever
var onceForRetriever sync.Once

func GetRetriever(conf *config.Config) *VikingDBRetriever {
	onceForRetriever.Do(func() {
		vikingDBRetriever = &VikingDBRetriever{}
		vikingDBRetriever.embModel = embedder.GetEmbedder(conf)
		vikingDBRetriever.createService(conf)
		_ = vikingDBRetriever.createIndex(conf)
	})
	return vikingDBRetriever
}

// 创建service
func (vr *VikingDBRetriever) createService(conf *config.Config) {
	vr.service = vikingdb.NewVikingDBService(conf.VikingDB.Host, conf.VikingDB.Region, conf.VikingDB.AK, conf.VikingDB.SK, conf.VikingDB.Scheme)
}

// 创建索引
func (vr *VikingDBRetriever) createIndex(conf *config.Config) (err error) {
	indexName := "vector_index"
	_, err = vr.service.GetIndex(conf.VikingDB.Collection, indexName)
	if err != nil {
		//TODO :后续可以根据数据量的不同灵活使用不同的索引方式，小数据量使用FLAT可以提高召回率
		vectorIndex := &vikingdb.VectorIndexParams{
			Distance:  vikingdb.COSINE,
			IndexType: vikingdb.FLAT,
			Quant:     vikingdb.Float,
		}
		indexOptions := vikingdb.NewIndexOptions().
			SetVectorIndex(vectorIndex).
			SetCpuQuota(1).
			SetDescription("this is an index").
			SetScalarIndex([]string{"user"})
		_, err := vr.service.CreateIndex(conf.VikingDB.Collection, indexName, indexOptions)
		if err != nil {
			fmt.Println(err)
		}
	}
	vr.indexClient = vikingdb.NewIndexClient(conf.VikingDB.Collection, indexName, conf.VikingDB.Host, conf.VikingDB.Region,
		conf.VikingDB.AK, conf.VikingDB.SK, conf.VikingDB.Scheme,
	)
	return nil
}

// Retrieve 实现Retriever接口
func (vr *VikingDBRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	uid, _, err := utils.GetUserIdAndUserNameFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user := "user" + strconv.Itoa(int(uid))
	//user := "user77674160750333952" //测试的时候写死的，不测试就不用了
	searchOptions := vikingdb.NewSearchOptions().
		SetFilter(map[string]interface{}{"op": "must", "field": "user", "conds": []string{user}}).
		SetLimit(5).
		SetOutputFields([]string{"id", "content", "metadata"})

	embRes, err := vr.embModel.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	var dense []float64 = embRes[0]
	res, err := vr.indexClient.SearchByVector(dense, searchOptions)
	if err != nil {
		return nil, err
	}
	// 将结果转换为[]*schema.Document，然后返回,并注意筛选返回分数
	scoreThreshold := 0.5

	var data []*schema.Document
	for idx, _ := range res {
		if res[idx].Score < scoreThreshold {
			continue
		}
		var m map[string]any
		if err = json.Unmarshal([]byte(res[idx].Fields["metadata"].(string)), &m); err != nil {
			return nil, err
		}
		data = append(data, &schema.Document{
			ID:       res[idx].Fields["id"].(string),
			Content:  res[idx].Fields["content"].(string),
			MetaData: m,
		})
	}

	return data, nil
}

//var retriever func(username string) retriever2.Retriever
//var onceForRetriever sync.Once
//
//func GetRetriever(conf *config.Config) func(collectionName string) retriever2.Retriever {
//	onceForRetriever.Do(func() {
//		retriever = newRetrieverOnce(conf)
//	})
//
//	return retriever
//}
//
//// 这个是用于IP度量的时候的转换函数，如果使用默认的汉明码则不需要这个
//var convFunc = func(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
//	result := make([]entity.Vector, len(vectors))
//	for i, vec := range vectors {
//		// Convert []float64 to []float32
//		float32Vec := make([]float32, len(vec))
//		for j, val := range vec {
//			float32Vec[j] = float32(val)
//		}
//		// Create FloatVector for Milvus
//		result[i] = entity.FloatVector(float32Vec)
//	}
//	return result, nil
//}
//
//func newRetrieverOnce(conf *config.Config) func(collectionName string) retriever2.Retriever {
//	ctx := context.Background()
//	cli, err := client.NewClient(ctx, client.Config{
//		Address: "localhost:19530",
//		DBName:  "rose_test",
//	})
//	if err != nil {
//		panic(err)
//	}
//	var emb = embedder.GetEmbedder(conf)
//	var cacheForRetriever sync.Map
//	return func(collectionName string) retriever2.Retriever {
//		if r, ok := cacheForRetriever.Load(collectionName); ok {
//			return r.(retriever2.Retriever)
//		}
//		searchParam, _ := entity.NewIndexAUTOINDEXSearchParam(1)
//		searchParam.AddRadius(10000)
//		searchParam.AddRangeFilter(0)
//
//		retriever, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
//			Client:      cli,
//			Collection:  collectionName,
//			Partition:   nil,
//			VectorField: "vector",
//			OutputFields: []string{
//				"id",
//				"content",
//				"metadata",
//			},
//			DocumentConverter: nil,
//			MetricType:        entity.HAMMING,
//			TopK:              3,
//			ScoreThreshold:    10000,
//			Sp:                searchParam,
//			Embedding:         emb,
//		})
//		if err != nil {
//			panic(err)
//		}
//		cacheForRetriever.Store(collectionName, retriever)
//		return retriever
//	}
//}
