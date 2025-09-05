package document_loader

import (
	"context"
	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"rose/internal/config"
	"sync"
)

type DocumentLoader struct {
	Loader document.Loader
}

var onceForParserCollection sync.Once
var documentLoader *DocumentLoader

func GetDocumentLoader(conf *config.Config) *DocumentLoader {
	onceForParserCollection.Do(func() {
		newDocumentLoader(conf)
	})
	return documentLoader
}

func newDocumentLoader(conf *config.Config) {
	documentLoader = &DocumentLoader{}
	ctx := context.Background()
	documentLoader.Loader, _ = file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: false,                // 是否使用文件名作为文档ID
		Parser:      &parser.TextParser{}, // 可选：指定自定义解析器
	})
}
