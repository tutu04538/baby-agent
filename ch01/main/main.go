package main

import (
	"babyagent/ch01"
	"babyagent/shared"
	"context"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	useRaw := flag.Bool("raw", false, "use raw http implementation")
	useStream := flag.Bool("stream", false, "use stream api")
	query := flag.String("q", "hello", "prompt text")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	modelConf := shared.NewModelConfig()
	switch {
	case *useRaw && *useStream:
		ch01.StreamingRequestRawHTTP(ctx, modelConf, *query)
	case *useRaw:
		ch01.NonStreamingRequestRawHTTP(ctx, modelConf, *query)
	case *useStream:
		ch01.StreamingRequestSDK(ctx, modelConf, *query)
	default:
		ch01.NonStreamingRequestSDK(ctx, modelConf, *query)
	}
}
