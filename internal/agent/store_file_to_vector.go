package agent

import (
	"context"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/agent/custom_splitter"
	"rose/internal/agent/document_loader"
	"rose/internal/agent/embedder"
	"rose/internal/agent/indexer"
	"rose/internal/config"
)

type StoreFileToVector struct {
	Splitter document.Transformer
	// Indexer 隐含了Embedder
	VectorStore *indexer.VectorStoreForVikingDB
	Loader      *document_loader.DocumentLoader
}

func NewStoreFileToVector(conf *config.Config) *StoreFileToVector {
	return &StoreFileToVector{
		Splitter:    custom_splitter.GetSplitter(embedder.GetEmbedder(conf), conf),
		VectorStore: indexer.NewVectorStoreForVikingDB(conf),
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

	for i, _ := range splitterRes {
		u := uuid.New()
		splitterRes[i].ID = u.String()
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
