package custom_retriver

import (
	"context"
	"github.com/zeromicro/go-zero/core/conf"
	"os"
	"path/filepath"
	"rose/internal/config"
	"testing"
)

// TODO: 测试的时候会报两次错，如果index还没有创建的时候，等两次报错之后就会正常了，注意一定要等index初始化完成
func TestRetriever(t *testing.T) {
	var base, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var path = filepath.Join(base, "..", "..", "..", "etc/rose-api.yaml")
	var c config.Config
	conf.MustLoad(path, &c)

	var r = GetRetriever(&c)

	documents, err := r.Retrieve(context.Background(), "解释golang中的range关键字")
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Logf("retrieved %d documents \n", len(documents))
	for _, doc := range documents {
		t.Logf("id: %v, content: %v, metadata: %v \n", doc.ID, doc.Content, doc.MetaData)
	}
}
