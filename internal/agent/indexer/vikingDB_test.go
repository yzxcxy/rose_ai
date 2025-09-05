package indexer

import (
	"context"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/conf"
	"os"
	"path/filepath"
	"rose/internal/config"
	"testing"
)

func TestVectorStoreForVikingDB(t *testing.T) {
	var base, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var path = filepath.Join(base, "..", "..", "..", "etc/rose-api.yaml")
	var c config.Config
	conf.MustLoad(path, &c)
	store := NewVectorStoreForVikingDB(&c)

	documents := []*schema.Document{
		{
			ID:      "123",
			Content: "golang是最伟大的语言",
			MetaData: map[string]interface{}{
				"author": "fqc",
				"book":   "golang",
			},
		},
		{
			ID:      "456",
			Content: "golang range关键字用于遍历可迭代对象",
			MetaData: map[string]interface{}{
				"author": "fqc",
				"book":   "golang",
			},
		},
	}

	ids, err := store.Store(context.Background(), documents)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("ids: %v", ids)
}
