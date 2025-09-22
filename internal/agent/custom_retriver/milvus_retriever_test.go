package custom_retriver

import (
	"context"
	"github.com/zeromicro/go-zero/core/conf"
	"os"
	"path/filepath"
	"rose/internal/config"
	"testing"
)

func TestMilvusRetriever(t *testing.T) {
	var base, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var path = filepath.Join(base, "..", "..", "..", "etc/rose-api.yaml")
	var c config.Config
	conf.MustLoad(path, &c)

	r, err := NewMilvusRetriever(&c)
	if err != nil {
		t.Fatal(err)
		return
	}

	documents, err := r.Retrieve(context.Background(), "解释一下golang的sync包")
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Logf("retrieved %d documents \n", len(documents))
	for _, doc := range documents {
		t.Logf("id: %v, \n", doc.ID)
	}
}
