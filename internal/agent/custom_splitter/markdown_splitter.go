package custom_splitter

import (
	"context"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/components/document"
	"rose/internal/config"
)

func GetMarkdownSplitter(conf *config.Config) (document.Transformer, error) {
	return markdown.NewHeaderSplitter(context.Background(), &markdown.HeaderConfig{
		Headers: map[string]string{
			"#":   "h1", // 一级标题
			"##":  "h2", // 二级标题
			"###": "h3", // 三级标题
		},
		TrimHeaders: true, // 是否在输出中保留标题行
	})

}
