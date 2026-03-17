package ch01

import (
	"babyagent/shared"
	"context"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func NonStreamingRequestSDK(ctx context.Context, modelConf *shared.ModelConfig, prompt string) {
	client := openai.NewClient(option.WithBaseURL(modelConf.BaseURL), option.WithAPIKey(modelConf.APIKey))

	req := openai.ChatCompletionNewParams{
		Model: modelConf.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}

	resp, err := client.Chat.Completions.New(ctx, req)
	if err != nil {
		log.Fatalf("failed to send a new completion request: %v", err)
	}

	if len(resp.Choices) > 0 {
		log.Printf("Response: %s", resp.Choices[0].Message.Content)
		log.Printf("token usage: %s", resp.Usage.RawJSON())
	}
}

func StreamingRequestSDK(ctx context.Context, modelConf *shared.ModelConfig, prompt string) {
	client := openai.NewClient(option.WithBaseURL(modelConf.BaseURL), option.WithAPIKey(modelConf.APIKey))

	req := openai.ChatCompletionNewParams{
		Model: modelConf.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}

	stream := client.Chat.Completions.NewStreaming(ctx, req)

	for stream.Next() {
		chunk := stream.Current()
		log.Printf("stream chunk: %s", chunk.RawJSON())
		if chunk.Usage.TotalTokens != 0 {
			log.Printf("token usage so far: %s", chunk.Usage.RawJSON())
		}
	}

	if stream.Err() != nil {
		log.Fatalf("streaming error: %v", stream.Err())
	}
}
