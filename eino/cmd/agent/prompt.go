package main

import (
	"context"
	"eino/pkg/langfuse/api"
	"eino/pkg/langfuse/api/resources/prompts/types"
	"eino/pkg/langfuse/config"
	"os"
	"path/filepath"
	"time"
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
	// Demo 演示需要，这里如果发现本地 langfuse 没起服务，就跳过；直接读本地配置的 Prompt；
	// 这个仅仅只是 demo 方便同学快速起服务，正式服务不这样处理。
	if client != nil {
		prompt, err := client.Prompts.Get(ctx, name, nil)
		if err == nil && prompt != nil {
			return prompt, nil
		}
	}

	// If not found in API, try to read from local file
	content, err := readAgentPrompt()
	if err != nil {
		return nil, err
	}

	// Create a prompt object with the file content as a system message
	localPrompt := &types.Prompt{
		ID:        "local-" + name,
		Name:      name,
		Version:   1,
		Type:      types.PromptTypeChatMessage,
		Prompt:    []types.ChatMessage{{Role: "system", Content: content}},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return localPrompt, nil
}

func readAgentPrompt() (string, error) {
	promptPath := filepath.Join("prompts", "agent_prompt.md")
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
