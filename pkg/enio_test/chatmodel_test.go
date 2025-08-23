package enio

import (
	"context"
	"encoding/json"
	"testing"
)
import "github.com/cloudwego/eino/schema"

import "github.com/cloudwego/eino-ext/components/model/deepseek"

func TestChatModel(t *testing.T) {
	t.Log("Testing ChatModel")
	ctx := context.Background()
	cm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:    "",
		Model:     "deepseek-chat",
		MaxTokens: 2000,
	})
	if err != nil {
		t.Error(err)
		return
	}

	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "You are a helpful AI assistant. Be concise in your responses.",
		},
		{
			Role:    schema.User,
			Content: "What is the capital of France?",
		},
	}

	resp, err := cm.Generate(ctx, messages)
	if err != nil {
		t.Error(err)
		return
	}

	// 检查响应的结果
	t.Logf("Assistant Response: %s", resp.Content)
	respJson, _ := json.Marshal(resp)
	t.Logf("Response JSON: %s", respJson)
}
