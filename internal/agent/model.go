package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/config"
)

// NewDeepSeekModel creates a new DeepSeek ChatModel instance.
func NewDeepSeekModel(conf *config.Config) (model.ToolCallingChatModel, error) {
	logx.Info("NewDeepSeekModel")
	ctx := context.Background()
	cm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:    conf.DeepSeek.Token,
		Model:     conf.DeepSeek.Model,
		MaxTokens: conf.DeepSeek.MaxTokens,
		BaseURL:   conf.DeepSeek.BaseURL,
	})

	if err != nil {
		logx.Error("Failed to create DeepSeek ChatModel:", err)
		return nil, err
	}

	logx.Info("DeepSeek ChatModel created successfully")
	return cm, nil
}
