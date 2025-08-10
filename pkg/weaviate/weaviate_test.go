package weaviate

import (
	"context"
	_ "context"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
	"rose/pkg/embedder"
	"testing"
)

func TestStoreWeaviate(t *testing.T) {
	apiKey := "9f1c2796-55d4-416e-ab08-46c711294934" // 替换为你的 API Key
	model := "doubao-embedding-large-text-250515"
	doubaoEmbedder := embedder.NewDoubaoEmbedder(model, apiKey)
	store, err := weaviate.New(
		weaviate.WithEmbedder(doubaoEmbedder),
		weaviate.WithIndexName("TestWeaviate"),
		weaviate.WithHost("localhost:8080"),
		weaviate.WithScheme("http"),
	)
	if err != nil {
		return
	}

	docs := []schema.Document{
		{
			PageContent: "Go is an open source programming language.",
			Metadata:    map[string]any{"author": "Google"},
		},
		{
			PageContent: "Weaviate is a vector search engine.",
			Metadata:    map[string]any{"author": "Weaviate"},
		},
	}

	ctx := context.Background()
	ids, err := store.AddDocuments(ctx, docs)
	if err != nil {
		return
	}

	t.Logf("Added documents with IDs: %v", ids)
}
