package agent

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/config"
)

type VectorStore struct {
	Conf *config.Config
}

func NewVectorStore(conf *config.Config) *VectorStore {
	return &VectorStore{
		Conf: conf,
	}
}

func (vs *VectorStore) Store(ctx context.Context, collectionName string, src []*schema.Document, opts ...document.TransformerOption) ([]string, error) {
	cli, err := client.NewClient(ctx, client.Config{
		Address: "localhost:19530",
		DBName:  "rose",
	})
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	defer func(cli client.Client) {
		err := cli.Close()
		if err != nil {
			logx.Error(err)
		}
	}(cli)

	embedder = GetEmbedder(vs.Conf)

	fields := []*entity.Field{
		{
			Name:       "id",
			DataType:   entity.FieldTypeVarChar,
			PrimaryKey: true,
			TypeParams: map[string]string{"max_length": "64"},
		},
		{
			Name:       "vector", // 向量字段
			DataType:   entity.FieldTypeFloatVector,
			TypeParams: map[string]string{"dim": "2048"},
		},
		{
			Name:       "content", // 原文本
			DataType:   entity.FieldTypeVarChar,
			TypeParams: map[string]string{"max_length": "65535"},
		},
		{
			Name:       "metadata", // 元数据字段
			DataType:   entity.FieldTypeJSON,
			TypeParams: map[string]string{"max_length": "65535"},
		},
	}

	// 替换默认的文档转换函数，因为其实现是基于byte进行存储的，会导致无法直接使用[][]float64类型的向量数据
	converterFunc := func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
		em := make([]MilusSchema, 0, len(docs))
		texts := make([]string, 0, len(docs))
		rows := make([]interface{}, 0, len(docs))

		for _, doc := range docs {
			metadata, err := sonic.Marshal(doc.MetaData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}
			em = append(em, MilusSchema{
				ID:       doc.ID,
				Content:  doc.Content,
				Vector:   nil,
				Metadata: metadata,
			})
			texts = append(texts, doc.Content)
		}

		// build embedding documents for storing
		for idx, vec := range vectors {
			em[idx].Vector = Float64ToFloat32(vec)
			rows = append(rows, &em[idx])
		}
		return rows, nil
	}

	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Collection:        collectionName, // 集合名称
		Client:            cli,
		Fields:            fields,
		Embedding:         embedder,
		DocumentConverter: converterFunc,
		MetricType:        milvus.IP,
	})
	if err != nil {
		logx.Error(err)
		panic(err)
	}
	ids, err := indexer.Store(ctx, src)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

type MilusSchema struct {
	ID       string    `json:"id" milvus:"name:id"`
	Content  string    `json:"content" milvus:"name:content"`
	Vector   []float32 `json:"vector" milvus:"name:vector"`
	Metadata []byte    `json:"metadata" milvus:"name:metadata"`
}

func Float64ToFloat32(v []float64) []float32 {
	if v == nil {
		return nil
	}
	res := make([]float32, len(v))
	for i, val := range v {
		res[i] = float32(val)
	}
	return res
}
