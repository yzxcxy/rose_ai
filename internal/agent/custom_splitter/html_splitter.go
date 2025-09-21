package custom_splitter

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/html"
	"github.com/cloudwego/eino/components/document"
	"rose/internal/config"
)

func GetHtmlSplitter(conf *config.Config) (document.Transformer, error) {
	htmlSplitterConfig := &html.HeaderConfig{
		Headers: map[string]string{
			"h1": "Header1",
			"h2": "Header2",
			"h3": "Header3",
		},
		IDGenerator: func(ctx context.Context, originalID string, splitIndex int) string {
			return fmt.Sprintf("%s_part%d", originalID, splitIndex)
		},
	}

	return html.NewHeaderSplitter(context.Background(), htmlSplitterConfig)
}
