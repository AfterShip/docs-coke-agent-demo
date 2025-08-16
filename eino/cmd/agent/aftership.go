package main

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
	"github.com/go-resty/resty/v2"
	"log"
	"os"
)

func newAfterShipChatModel(ctx context.Context, cfg Config) (model.ToolCallingChatModel, error) {
	client := resty.New().
		SetHeader("am-api-key", cfg.APIKey)

	claudeConfig := &claude.Config{
		APIKey:     cfg.APIKey,
		BaseURL:    &cfg.AIGCHubBaseURL,
		Model:      cfg.LLMModel,
		HTTPClient: client.GetClient(),
		MaxTokens:  2048,
	}

	chatModel, err := claude.NewChatModel(ctx, claudeConfig)
	if err != nil {
		return nil, err
	}
	return chatModel, nil
}

func createClaudeChatModel(ctx context.Context) model.ToolCallingChatModel {
	cfg := Config{
		AIGCHubBaseURL: "http://data-aigc.as-in.io/v1/stub/vendors/AWS",
		LLMModel:       "anthropic.claude-3-5-sonnet-20241022-v2:0",
		APIKey:         os.Getenv("AM_API_KEY"),
	}

	if cfg.APIKey == "" {
		log.Fatal("AM_API_KEY environment variable is not set")
	}

	chatModel, err := newAfterShipChatModel(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	return chatModel
}
