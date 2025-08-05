package embedder

import (
	"context"
	_ "github.com/tmc/langchaingo/embeddings"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

// 参考文档：https://www.volcengine.com/docs/82379/1521766

// DoubaoEmbedder 满足 Embedder 接口
type DoubaoEmbedder struct {
	client *arkruntime.Client
	model  string // 如 "doubao-embedding-text-240715"
}

// NewDoubaoEmbedder 新建实例
func NewDoubaoEmbedder(model string, apiKey string) *DoubaoEmbedder {
	return &DoubaoEmbedder{
		client: arkruntime.NewClientWithApiKey(apiKey),
		model:  model,
	}
}

// EmbedDocuments 将一组文本批量向量化。
func (d *DoubaoEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	req := model.EmbeddingRequestStrings{
		Input:          texts,
		Model:          d.model,
		EncodingFormat: "float",
	}
	resp, err := d.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	vecs := make([][]float32, len(resp.Data))
	for i, e := range resp.Data {
		vecs[i] = e.Embedding
	}
	return vecs, nil
}

// EmbedQuery 将单个文本向量化。
func (d *DoubaoEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	// 直接复用 EmbedDocuments 的批量实现，但只传一个文本
	vecs, err := d.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return vecs[0], nil
}
