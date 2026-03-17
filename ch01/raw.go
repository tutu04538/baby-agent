package ch01

import (
	"babyagent/shared"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

type OpenAIChatCompletionStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content,omitempty"`
		} `json:"delta"`
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

func StreamingRequestRawHTTP(ctx context.Context, modelConf *shared.ModelConfig, prompt string) {
	client := http.Client{}

	requestBody := OpenAIChatCompletionRequest{
		Model: modelConf.Model,
		Messages: []OpenAIChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("failed to marshal request body: %v", err)
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", modelConf.BaseURL+"/chat/completions", bytes.NewReader(bodyBytes))
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

	reader := bufio.NewScanner(resp.Body)
	for reader.Scan() {
		line := reader.Text()
		if line == "" {
			continue
		}
		if line == "data: [DONE]" {
			log.Println("streaming completed")
			break
		}
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
			var result OpenAIChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(line), &result); err != nil {
				log.Printf("failed to unmarshal stream response: %v", err)
				continue
			}
			log.Printf("streaming content: %s", result.Choices[0].Delta.Content)
			if result.Usage != nil {
				log.Printf("token usage: %+v", result.Usage)
			}
		} else {
			log.Printf("unexpected line format: %s", line)
			continue
		}

	}

	if err := reader.Err(); err != nil {
		log.Printf("error reading stream: %v", err)
	}

}
