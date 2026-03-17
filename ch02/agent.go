package ch02

import (
	"babyagent/ch02/tools"
	"babyagent/shared"
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Agent struct {
	systemPrompt string
	model        string
	client       openai.Client
	messages     []openai.ChatCompletionMessageParamUnion
	tools        map[tools.AgentTool]tools.Tool
}

func NewAgent(modelConf *shared.ModelConfig, systemPrompt string, tool []tools.Tool) *Agent {
	a := &Agent{
		systemPrompt: systemPrompt,
		model:        modelConf.Model,
		client:       openai.NewClient(option.WithBaseURL(modelConf.BaseURL), option.WithAPIKey(modelConf.APIKey)),
		messages:     []openai.ChatCompletionMessageParamUnion{openai.SystemMessage(systemPrompt)},
		tools:        make(map[tools.AgentTool]tools.Tool),
	}

	for _, t := range tool {
		a.tools[t.ToolName()] = t
	}

	return a
}

func (a *Agent) execute(ctx context.Context, toolName string, argumentsInJSON string) (string, error) {
	t, ok := a.tools[tools.AgentTool(toolName)]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}
	return t.Execute(ctx, argumentsInJSON)
}

// 对话栈：[System, User, Assistant(ToolCalls), Tool(Result), Assistant(Text)]
func (a *Agent) Run(ctx context.Context, userPrompt string) (string, error) {
	a.messages = append(a.messages, openai.UserMessage(userPrompt))

	var result string
	for {
		params := openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: a.messages,
			Tools: func() []openai.ChatCompletionToolUnionParam {
				toolParams := make([]openai.ChatCompletionToolUnionParam, 0, len(a.tools))
				for _, t := range a.tools {
					toolParams = append(toolParams, t.Info())
				}
				return toolParams
			}(),
		}

		log.Printf("calling llm model %s...", a.model)
		resp, err := a.client.Chat.Completions.New(ctx, params)
		if err != nil {
			log.Fatalf("failed to send a new completion request: %v", err)
			return "", fmt.Errorf("failed to get completion: %w", err)
		}

		if len(resp.Choices) == 0 {
			log.Printf("no choices returned, resp: %v", resp)
			return "", fmt.Errorf("no choices in response")
		}

		messages := resp.Choices[0].Message
		// fmt.Printf("model response: %s\n", messages.Content)
		a.messages = append(a.messages, messages.ToParam())

		if len(messages.ToolCalls) == 0 {
			result = messages.Content
			break
		}

		for _, toolCall := range messages.ToolCalls {
			log.Printf("tool call: %s, arguments: %s", toolCall.Function.Name, toolCall.Function.Arguments)
			toolResult, err := a.execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			// log.Printf("tool result: %s", toolResult)
			if err != nil {
				log.Printf("tool execution error: %v", err)
				toolResult = fmt.Sprintf("error executing tool %s: %v", toolCall.Function.Name, err)
			}
			a.messages = append(a.messages, openai.ToolMessage(toolResult, toolCall.ID))
		}
	}
	return result, nil
}
