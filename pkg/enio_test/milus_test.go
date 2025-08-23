package enio_test

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"testing"
)

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

func TestMilus(t *testing.T) {
	t.Log("Testing Milus")
	ctx := context.Background()
	cli, err := client.NewClient(ctx, client.Config{
		Address: "localhost:19530",
		DBName:  "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func(cli client.Client) {
		err := cli.Close()
		if err != nil {

		}
	}(cli)

	embedder, _ := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: "",                                   // 使用 API Key 认证
		Model:  "doubao-embedding-large-text-250515", // Ark 平台的端点 ID
	})

	fields := []*entity.Field{
		{
			Name:       "id", // 主键
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
		Collection:        "test_collection", // 集合名称
		Client:            cli,
		Fields:            fields,
		Embedding:         embedder,
		DocumentConverter: converterFunc,
		MetricType:        milvus.IP,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Store documents
	docs := []*schema.Document{
		{
			ID:      "milvus-1",
			Content: "milvus is an open-source vector database",
			MetaData: map[string]any{
				"h1": "milvus",
				"h2": "open-source",
				"h3": "vector database",
			},
		},
		{
			ID:      "milvus-2",
			Content: "milvus is a distributed vector database",
		},
	}

	ids, err := indexer.Store(ctx, docs)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("ids: %v", ids)
}
