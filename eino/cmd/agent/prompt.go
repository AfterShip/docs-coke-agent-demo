package main

import (
	"context"
	"eino/pkg/langfuse/api"
	"eino/pkg/langfuse/api/resources/prompts/types"
	"eino/pkg/langfuse/config"
	"os"
)

func initLangfuseAPIClient() (*api.APIClient, error) {
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")

	cfg, err := config.NewConfig(
		config.WithHost("http://127.0.0.1:3000"),
		config.WithPublicKey(publicKey),
		config.WithSecretKey(secretKey),
		config.WithDebug(true),
	)
	if err != nil {
		return nil, err
	}
	return api.NewAPIClient(cfg)
}

func getPrompts(ctx context.Context, client *api.APIClient) ([]types.Prompt, error) {
	limit := 10
	page := 1
	req := &types.GetPromptsRequest{
		Page:  &page,  // Adjust page as needed
		Limit: &limit, // Adjust limit as needed
	}
	promptsList, err := client.Prompts.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return promptsList.Data, nil
}

func getPromptByName(ctx context.Context, client *api.APIClient, name string) (*types.Prompt, error) {
	prompt, err := client.Prompts.Get(ctx, name, nil)
	if err != nil {
		return nil, err
	}
	return prompt, nil
}
