package main

import (
	"context"
	"fmt"
	"log"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/api/resources/commons/types"
	"eino/pkg/langfuse/internal/utils"
)

func main() {
	// Initialize client
	langfuseClient, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
	)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	// Create a trace with a generation
	trace := langfuseClient.Trace("qa-evaluation").
		WithUserID("evaluator-1").
		WithInput("What is machine learning?")

	generation := trace.Generation("answer-generation").
		WithModel("gpt-4", map[string]interface{}{
			"temperature": 0.3,
		}).
		WithInput("What is machine learning?").
		WithOutput("Machine learning is a subset of artificial intelligence that enables computers to learn and make decisions from data without being explicitly programmed for every scenario.")

	if err := generation.End(); err != nil {
		log.Fatal("Failed to submit generation:", err)
	}

	traceID := trace.GetID()

	if err := trace.End(); err != nil {
		log.Fatal("Failed to submit trace:", err)
	}

	// Create scores for the generation
	scores := []*types.Score{
		{
			ID:        utils.GenerateID(),
			TraceID:   traceID,
			Name:      "accuracy",
			Value:     0.95,
			DataType:  types.ScoreDataTypeNumeric,
			Comment:   stringPtr("High accuracy answer"),
		},
		{
			ID:        utils.GenerateID(),
			TraceID:   traceID,
			Name:      "relevance",
			Value:     0.90,
			DataType:  types.ScoreDataTypeNumeric,
			Comment:   stringPtr("Very relevant to the question"),
		},
		{
			ID:        utils.GenerateID(),
			TraceID:   traceID,
			Name:      "helpfulness",
			Value:     "helpful",
			DataType:  types.ScoreDataTypeCategorical,
			Comment:   stringPtr("Provides clear explanation"),
		},
		{
			ID:        utils.GenerateID(),
			TraceID:   traceID,
			Name:      "contains_code",
			Value:     false,
			DataType:  types.ScoreDataTypeBoolean,
			Comment:   stringPtr("No code examples in answer"),
		},
	}

	// Submit scores
	for i, score := range scores {
		if err := langfuseClient.Score(score); err != nil {
			log.Printf("Failed to submit score %d (%s): %v", i+1, score.Name, err)
		} else {
			fmt.Printf("Score '%s' submitted: %v\n", score.Name, score.Value)
		}
	}

	fmt.Println("Scoring example completed successfully!")
}

func stringPtr(s string) *string {
	return &s
}