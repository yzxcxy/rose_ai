package custom_splitter

import (
	"context"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"rose/internal/config"
	"sync"
)

var splitter document.Transformer
var onceForSplitter sync.Once

func GetSplitter(embedder embedding.Embedder, conf *config.Config) document.Transformer {
	onceForSplitter.Do(func() {
		newSplitterOnce(embedder, conf)
	})
	return splitter
}

func newSplitterOnce(embedder embedding.Embedder, conf *config.Config) {
	ctx := context.Background()
	splitter, _ = semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,                     // 必需：用于生成文本向量的嵌入器
		BufferSize:   2,                            // 可选：上下文缓冲区大小
		MinChunkSize: 50,                           // 可选：最小片段大小
		Separators:   []string{"。", ".", "?", "!"}, // 可选：分隔符列表
		Percentile:   0.6,                          // 可选：分割阈值百分位数
		LenFunc:      nil,                          // 可选：自定义长度计算函数
	})
}
