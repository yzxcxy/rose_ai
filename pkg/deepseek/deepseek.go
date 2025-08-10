package deepseek

import (
	"github.com/tmc/langchaingo/llms/openai"
	"rose/internal/config"
)

func New(config config.Config) (llm *openai.LLM, err error) {
	llm, err = openai.New(
		openai.WithModel(config.DeepSeek.Model),
		openai.WithBaseURL(config.DeepSeek.BaseURL),
		openai.WithToken(config.DeepSeek.Token),
	)
	return
}
