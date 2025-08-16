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
	// Create a new Langfuse client with basic configuration
	config, err := client.NewConfig(
		client.WithHost("https://cloud.langfuse.com"),
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
	)
	if err != nil {
		log.Fatal("Failed to create config:", err)
	}

	// Initialize the client
	langfuseClient, err := client.New(config)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer func() {
		if err := langfuseClient.Shutdown(context.Background()); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Create a simple trace
	trace := langfuseClient.Trace("basic-example-trace").
		WithUserID("user-123").
		WithSessionID("session-456").
		WithInput("Hello, world!").
		WithMetadata(map[string]interface{}{
			"example":     true,
			"version":     "1.0",
			"description": "Basic trace example",
		})

	// Submit the trace
	if err := trace.End(); err != nil {
		log.Fatal("Failed to submit trace:", err)
	}

	fmt.Println("Simple trace created and submitted successfully!")

	// Wait a moment for the trace to be processed
	time.Sleep(1 * time.Second)

	// Check client statistics
	stats := langfuseClient.GetStats()
	fmt.Printf("Client stats: Traces created: %d, Events submitted: %d\n", 
		stats.TracesCreated, stats.EventsSubmitted)
}