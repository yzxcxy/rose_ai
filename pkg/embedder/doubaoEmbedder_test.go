package embedder

import (
	"context"
	"testing"
)

func TestDoubaoEmbedder_EmbedDocuments(t *testing.T) {
	ctx := context.Background()
	apiKey := "" // 替换为你的 API Key
	model := "doubao-embedding-large-text-250515"

	embedder := NewDoubaoEmbedder(model, apiKey)

	texts := []string{
		"Hello, world!",
		"这是一个测试文本。",
		"Doubao embedding is powerful.",
	}

	vecs, err := embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		t.Fatalf("EmbedDocuments failed: %v", err)
	}
	if len(vecs) != len(texts) {
		t.Fatalf("Expected %d vectors, got %d", len(texts), len(vecs))
	}

	for i, vec := range vecs {
		if len(vec) == 0 {
			t.Errorf("Vector %d is empty", i)
		} else {
			t.Logf("Vector %d: %v", i, vec)
		}
	}
}
