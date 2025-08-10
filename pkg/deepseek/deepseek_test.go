package deepseek

import (
	"context"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"log"
	"testing"
)

func TestDeepseek(t *testing.T) {
	// Initialize the OpenAI client with Deepseek model
	llm, err := openai.New(
		openai.WithModel("deepseek-reasoner"),
		openai.WithBaseURL("https://api.deepseek.com/v1"),
		openai.WithToken("sk-065e80673c644ff58fc582a9e5f229b4"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create messages for the chat
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant that explains complex topics step by step"),
		llms.TextParts(llms.ChatMessageTypeHuman, "Explain how quantum entanglement works and why it's important for quantum computing"),
	}

	// Generate content with streaming to see both reasoning and final answer in real-time
	completion, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(2000),
		llms.WithTemperature(0.7),
	)
	if err != nil {
		log.Fatal(err)
	}

	t.Logf("Completion len: %d\n", len(completion.Choices))

	// Access the reasoning content and final answer separately
	if len(completion.Choices) > 0 {
		choice := completion.Choices[0]
		t.Logf("\nFinal Answer:\n%s\n", choice.Content)
	}
}
