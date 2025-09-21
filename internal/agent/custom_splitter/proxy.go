package custom_splitter

import (
	"context"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/agent/embedder"
	"rose/internal/config"
)

type Proxy struct {
	HtmlSplitter     document.Transformer
	MarkDownSplitter document.Transformer
	SemanticSplitter document.Transformer
}

func NewSplitterProxy(config *config.Config) *Proxy {
	htmlSplitter, err := GetHtmlSplitter(config)
	if err != nil {
		logx.Error("get html splitter err", err)
		panic(err)
	}

	markdownSplitter, err := GetMarkdownSplitter(config)
	if err != nil {
		logx.Error("get markdown splitter err", err)
		panic(err)
	}
	return &Proxy{
		SemanticSplitter: GetSemanticSplitter(embedder.GetEmbedder(config), config),
		HtmlSplitter:     htmlSplitter,
		MarkDownSplitter: markdownSplitter,
	}
}

func (proxy *Proxy) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) (results []*schema.Document, err error) {
	for idx := range src {
		var res []*schema.Document
		if extension, ok := src[idx].MetaData["_extension"]; ok {
			switch extension {
			case ".html", ".htm":
				logx.Info("use html splitter")
				res, err = proxy.HtmlSplitter.Transform(ctx, src, opts...)
				if err != nil {
					logx.Error("transform html err", err)
					return nil, err
				}
			case ".md", ".markdown":
				logx.Infof("match markdown file, use markdown splitter")
				res, err = proxy.MarkDownSplitter.Transform(ctx, src, opts...)
				if err != nil {
					logx.Error("transform markdown err", err)
					return nil, err
				}
				contextMarkdown(res)
			default:
				logx.Info("use semantic splitter")
				res, err = proxy.SemanticSplitter.Transform(ctx, src, opts...)
				if err != nil {
					logx.Error("transform semantic err", err)
					return nil, err
				}
			}
		} else {
			logx.Info("use semantic splitter")
			res, err = proxy.SemanticSplitter.Transform(ctx, src, opts...)
			if err != nil {
				logx.Error("transform semantic err", err)
				return nil, err
			}
		}
		results = append(results, res...)
	}

	return results, nil
}

func contextMarkdown(docs []*schema.Document) {
	headerKeys := []string{"h3", "h2", "h1"}
	for i := 0; i < len(docs); i++ {
		if !isValid(docs[i].Content) {
			docs[i] = docs[len(docs)-1]
			docs = docs[:len(docs)-1]
			continue
		}
		for _, key := range headerKeys {
			if val, ok := docs[i].MetaData[key]; ok {
				docs[i].Content = val.(string) + "\n" + docs[i].Content
			}
		}
	}
}

func isValid(str string) bool {
	checks := map[string]interface{}{
		"":   nil,
		" ":  nil,
		"\n": nil,
	}

	if _, exists := checks[str]; exists {
		return false
	}
	return true
}
