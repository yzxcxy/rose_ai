package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
	"rose/internal/config"
	"sync"
)

var embedder embedding.Embedder
var once sync.Once

func GetEmbedder(conf *config.Config) embedding.Embedder {
	once.Do(func() {
		newEmbedderOnce(conf)
	})
	return embedder
}

func newEmbedderOnce(conf *config.Config) {
	ctx := context.Background()
	embedder, _ = ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: conf.Doubao.APIKey, // 使用 API Key 认证
		Model:  conf.Doubao.Model,  // Ark 平台的端点 ID
	})
}
