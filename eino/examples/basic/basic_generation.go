package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/api/resources/commons/types"
)

func main() {
	// Create client with environment variables (recommended for production)
	// Set these environment variables:
	// LANGFUSE_PUBLIC_KEY=your-public-key
	// LANGFUSE_SECRET_KEY=your-secret-key
	// LANGFUSE_HOST=https://cloud.langfuse.com
	config, err := client.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	// Override with code if needed
	config.Debug = true
	config.Environment = "development"
	
	langfuseClient, err := client.New(config)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	// Create a generation for an LLM call
	generation := langfuseClient.Generation("chat-completion").
		WithModel("gpt-4o-mini", map[string]interface{}{
			"temperature":   0.7,
			"max_tokens":    1000,
			"top_p":        0.9,
		}).
		WithInput([]map[string]string{
			{"role": "user", "content": "What is the capital of France?"},
		}).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"api_version": "v1",
			"provider":    "openai",
		})

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Complete the generation with output and usage
	generation.WithOutput([]map[string]string{
		{"role": "assistant", "content": "The capital of France is Paris."},
	}).
		WithEndTime(time.Now()).
		WithUsage(&types.Usage{
			Input:     intPtr(15),
			Output:    intPtr(8),
			Total:     intPtr(23),
			Unit:      stringPtr("TOKENS"),
			InputCost: floatPtr(0.00003),
			OutputCost: floatPtr(0.00006),
			TotalCost: floatPtr(0.00009),
		})

	// Submit the generation
	if err := generation.End(); err != nil {
		log.Fatal("Failed to submit generation:", err)
	}

	fmt.Println("Basic generation created and submitted successfully!")

	// Flush to ensure immediate submission
	if err := langfuseClient.Flush(context.Background()); err != nil {
		log.Printf("Warning: Failed to flush: %v", err)
	}
}

// Helper functions for pointers
func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}