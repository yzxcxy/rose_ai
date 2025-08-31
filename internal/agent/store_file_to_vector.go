package agent

import (
	"context"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/config"
)

type StoreFileToVector struct {
	Splitter document.Transformer
	// Indexer 隐含了Embedder
	VectorStore *VectorStore
	Loader      *DocumentLoader
}

func NewStoreFileToVector(conf *config.Config) *StoreFileToVector {
	return &StoreFileToVector{
		Splitter:    GetSplitter(GetEmbedder(conf), conf),
		VectorStore: NewVectorStore(conf),
		Loader:      GetDocumentLoader(conf),
	}
}

// Store 读取文件，分块文本，存储向量到向量数据库
// fileName 支持多个文件
// collectionName 向量数据库集合名称, 在这里指用户 “user”+userId
func (s *StoreFileToVector) Store(ctx context.Context, fileName []string, collectionName string) (ids []string, err error) {
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
	ids, err = s.VectorStore.Store(ctx, collectionName, splitterRes)
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	logx.Infof("StoreFileToVector Store success, collectionName: %s, count: %d", collectionName, len(ids))
	return ids, nil
}
