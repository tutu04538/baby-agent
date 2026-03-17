package main

import (
	"context"
	"flag"
	"log"

	"github.com/joho/godotenv"

	"babyagent/ch02"
	"babyagent/ch02/tools"
	"babyagent/shared"
)

func main() {
	_ = godotenv.Load()

	query := flag.String("q", "hello", "prompt text")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	modelConf := shared.NewModelConfig()

	agent := ch02.NewAgent(modelConf, ch02.CodingAgentSystemPrompt, []tools.Tool{
		tools.NewReadTool(),
		tools.NewEditTool(),
		tools.NewWriteTool(),
		tools.NewBashTool(),
	})
	result, err := agent.Run(ctx, *query)
	if err != nil {
		log.Printf("agent run error: %v", err)
		return
	}

	log.Printf("agent result: %s", result)
}
