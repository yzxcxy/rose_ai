package enio

import (
	"context"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"testing"
)

func TestEmbeddingForARK(t *testing.T) {
	ctx := context.Background()
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: "", // 使用 API Key 认证
		Model:  "", // Ark 平台的端点 ID
	})

	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	texts := []string{
		"这是第一段示例文本",
		"这是第二段示例文本",
	}

	embeddings, err := embedder.EmbedStrings(ctx, texts)

	if err != nil {
		t.Fatalf("EmbedStrings failed: %v", err)
	}

	for i, vec := range embeddings {
		t.Logf("Embedding %d len is %d", i, len(vec))
	}
}
