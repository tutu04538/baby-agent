package ch01

import (
	"babyagent/shared"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type OpenAIChatCompletionRequest struct {
	Model    string              `json:"model"`
	Messages []OpenAIChatMessage `json:"messages"`
	Stream   bool                `json:"stream,omitempty"`
}

type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIChatCompletionResponse struct {
	Choices []struct {
		Message OpenAIChatMessage `json:"message"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
}

func NonStreamingRequestRawHTTP(ctx context.Context, modelConf *shared.ModelConfig, prompt string) {
	client := http.Client{}

	requestBody := OpenAIChatCompletionRequest{
		Model: modelConf.Model,
		Messages: []OpenAIChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("failed to marshal request body: %v", err)
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", modelConf.BaseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Fatalf("failed to create HTTP request: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+modelConf.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("HTTP request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP request failed with status: %d", resp.StatusCode)
		return
	}

	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read Body: %v\n", err)
		return
	}

	var result OpenAIChatCompletionResponse
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		log.Fatalf("failed to decode response: %v", err)
		return
	}

	log.Printf("resp content: %s", result.Choices[0].Message.Content)
	log.Printf("token usage: %+v", result.Usage)

}
