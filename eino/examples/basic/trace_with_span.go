package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"eino/pkg/langfuse/client"
)

func main() {
	// Initialize client
	langfuseClient, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
	)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	// Create a trace for a complex operation
	trace := langfuseClient.Trace("data-processing-pipeline").
		WithUserID("user-789").
		WithSessionID("session-123").
		WithInput(map[string]interface{}{
			"data_source": "users.csv",
			"batch_size":  1000,
		}).
		WithMetadata(map[string]interface{}{
			"pipeline_version": "v2.1",
			"environment":      "staging",
		})

	// Add a span for data loading phase
	span1 := trace.Span("load-data").
		WithInput(map[string]interface{}{
			"file_path": "data/users.csv",
			"format":    "csv",
		}).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"phase": "data-loading",
		})

	// Simulate data loading work
	time.Sleep(50 * time.Millisecond)

	span1.WithOutput(map[string]interface{}{
		"rows_loaded": 1000,
		"columns":     ["id", "name", "email"],
		"status":      "success",
	}).WithEndTime(time.Now())

	if err := span1.End(); err != nil {
		log.Printf("Failed to submit span1: %v", err)
	}

	// Add another span for data processing
	span2 := trace.Span("process-data").
		WithInput(map[string]interface{}{
			"rows_count": 1000,
			"operations": ["validate", "clean", "enrich"],
		}).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"phase": "data-processing",
		})

	// Simulate data processing work
	time.Sleep(200 * time.Millisecond)

	span2.WithOutput(map[string]interface{}{
		"processed_rows":  950,
		"invalid_rows":    50,
		"processing_time": "200ms",
	}).
		WithEndTime(time.Now()).
		WithLevel("DEFAULT")

	if err := span2.End(); err != nil {
		log.Printf("Failed to submit span2: %v", err)
	}

	// Complete the main trace
	trace.WithOutput(map[string]interface{}{
		"pipeline_status":  "completed",
		"total_processed":  950,
		"total_errors":     50,
		"execution_time":   "250ms",
	})

	if err := trace.End(); err != nil {
		log.Fatal("Failed to submit trace:", err)
	}

	fmt.Println("Trace with spans created and submitted successfully!")

	// Print client statistics
	stats := langfuseClient.GetStats()
	fmt.Printf("Traces: %d, Spans: %d, Events submitted: %d\n",
		stats.TracesCreated, stats.SpansCreated, stats.EventsSubmitted)
}