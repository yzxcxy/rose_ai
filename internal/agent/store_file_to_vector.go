package agent

import (
	"context"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/agent/custom_splitter"
	"rose/internal/agent/document_loader"
	"rose/internal/agent/indexer"
	"rose/internal/config"
)

type StoreFileToVector struct {
	Splitter document.Transformer
	// Indexer 隐含了Embedder
	VectorStore *indexer.MilvusIndexer
	Loader      *document_loader.DocumentLoader
}

func NewStoreFileToVector(conf *config.Config) *StoreFileToVector {
	milvusIndexer, err := indexer.NewMilvusIndexer(conf)
	if err != nil {
		logx.Errorf("NewMilvusIndexer err: %v", err)
		return nil
	}
	return &StoreFileToVector{
		Splitter:    custom_splitter.NewSplitterProxy(conf),
		VectorStore: milvusIndexer,
		Loader:      document_loader.GetDocumentLoader(conf),
	}
}

// Store 读取文件，分块文本，存储向量到向量数据库
// fileName 支持多个文件
// collectionName 向量数据库集合名称, 在这里指用户 “user”+userId
func (s *StoreFileToVector) Store(ctx context.Context, fileName []string, user string) (ids []string, err error) {
	// TODO 注意这里可以使用go routine 进行优化
	// 1. 读取文件内容
	var docs []*schema.Document
	for _, file := range fileName {
		returnDocs, err := s.Loader.Loader.Load(ctx, document.Source{
			URI: file,
		})
		if err != nil {
			logx.Errorf("Loader.Load err: %v", err)
			return nil, err
		}
		docs = append(docs, returnDocs...)
	}
	// 2. 文本分块
	splitterRes, err := s.Splitter.Transform(ctx, docs)
	if err != nil {
		logx.Error(err)
		return nil, err
	}

	for i := 0; i < len(splitterRes); i++ {
		// TODO 这是为了解决 splitter 出现空内容的bug，后续需要排查原因
		if !isValid(splitterRes[i].Content) {
			// 不需要关心顺序，直接用最后一个元素覆盖
			splitterRes[i] = splitterRes[len(splitterRes)-1]
			splitterRes = splitterRes[:len(splitterRes)-1]
			continue
		}
		u := uuid.New()
		splitterRes[i].ID = u.String()
		// 将文件名放到 content field
		if fileFiled, ok := splitterRes[i].MetaData["_file_name"]; ok {
			splitterRes[i].Content = fileFiled.(string) + ":\n" + splitterRes[i].Content
		}
	}
	// 3. 存储向量到向量数据库
	ids, err = s.VectorStore.Store(ctx, splitterRes)
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	logx.Infof("StoreFileToVector Store success, count: %d", len(ids))
	return ids, nil
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
