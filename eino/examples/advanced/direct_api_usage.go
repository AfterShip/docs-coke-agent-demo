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
	// Initialize Langfuse client
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

	fmt.Println("=== Direct API Usage Examples ===\n")

	// Get direct API access
	api := langfuseClient.API()
	if api == nil {
		log.Fatal("API client is not available")
	}

	ctx := context.Background()

	// 1. Health Check
	fmt.Println("1. Health Check:")
	if err := langfuseClient.HealthCheck(ctx); err != nil {
		fmt.Printf("   ❌ Health check failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Service is healthy")
	}
	fmt.Println()

	// 2. List Traces (if any exist)
	fmt.Println("2. List Traces:")
	traces, err := api.Traces.List(ctx)
	if err != nil {
		fmt.Printf("   ❌ Failed to list traces: %v\n", err)
	} else {
		fmt.Printf("   ✅ Found %d traces in the system\n", len(traces))
		for i, trace := range traces {
			if i < 3 { // Show first 3 traces
				fmt.Printf("      - Trace ID: %s, Name: %s\n", trace.ID, trace.Name)
			}
		}
		if len(traces) > 3 {
			fmt.Printf("      ... and %d more traces\n", len(traces)-3)
		}
	}
	fmt.Println()

	// 3. Create Score Directly
	fmt.Println("3. Create Score Directly:")
	score := &types.Score{
		Name:     "api_example_score",
		Value:    0.95,
		DataType: types.ScoreDataTypeNumeric,
		Comment:  stringPtr("Score created via direct API"),
		TraceID:  "example-trace-id", // This should be a real trace ID in practice
	}

	createdScore, err := api.Scores.Create(ctx, score)
	if err != nil {
		fmt.Printf("   ❌ Failed to create score: %v\n", err)
	} else {
		fmt.Printf("   ✅ Score created with ID: %s\n", createdScore.ID)
	}
	fmt.Println()

	// 4. List Sessions
	fmt.Println("4. List Sessions:")
	sessions, err := api.Sessions.List(ctx)
	if err != nil {
		fmt.Printf("   ❌ Failed to list sessions: %v\n", err)
	} else {
		fmt.Printf("   ✅ Found %d sessions\n", len(sessions))
		for i, session := range sessions {
			if i < 3 {
				fmt.Printf("      - Session ID: %s\n", session.ID)
			}
		}
	}
	fmt.Println()

	// 5. Create and Manage Datasets
	fmt.Println("5. Dataset Management:")
	
	// Create a dataset
	dataset := &types.Dataset{
		Name:        "api-example-dataset",
		Description: stringPtr("Dataset created via direct API"),
		Metadata: map[string]interface{}{
			"created_by": "direct_api_example",
			"version":    "1.0",
		},
	}

	createdDataset, err := api.Datasets.Create(ctx, dataset)
	if err != nil {
		fmt.Printf("   ❌ Failed to create dataset: %v\n", err)
	} else {
		fmt.Printf("   ✅ Dataset created: %s (ID: %s)\n", createdDataset.Name, createdDataset.ID)
		
		// List all datasets
		datasets, err := api.Datasets.List(ctx)
		if err != nil {
			fmt.Printf("   ❌ Failed to list datasets: %v\n", err)
		} else {
			fmt.Printf("   ✅ Total datasets in system: %d\n", len(datasets))
		}
	}
	fmt.Println()

	// 6. Advanced Trace Operations
	fmt.Println("6. Advanced Trace Operations:")
	
	// Create a complex trace using the builder pattern first
	trace := langfuseClient.Trace("direct-api-example").
		WithUserID("api-user-123").
		WithSessionID("api-session-456").
		WithInput(map[string]interface{}{
			"operation": "direct_api_demo",
			"timestamp": time.Now(),
		}).
		WithMetadata(map[string]interface{}{
			"api_version": "v1",
			"client_type": "direct",
		})

	// Add a generation
	generation := trace.Generation("api-generation").
		WithModel("gpt-4", map[string]interface{}{
			"temperature": 0.8,
		}).
		WithInput("Demonstrate direct API usage").
		WithOutput("This is a demonstration of direct API usage with Langfuse Go SDK")

	if err := generation.End(); err != nil {
		fmt.Printf("   ❌ Failed to submit generation: %v\n", err)
	} else {
		fmt.Println("   ✅ Generation submitted successfully")
	}

	traceID := trace.GetID()

	if err := trace.End(); err != nil {
		fmt.Printf("   ❌ Failed to submit trace: %v\n", err)
	} else {
		fmt.Printf("   ✅ Trace submitted with ID: %s\n", traceID)
	}

	// Wait a moment for processing
	time.Sleep(1 * time.Second)

	// Try to retrieve the trace we just created
	retrievedTrace, err := api.Traces.Get(ctx, traceID)
	if err != nil {
		fmt.Printf("   ❌ Failed to retrieve trace: %v\n", err)
	} else {
		fmt.Printf("   ✅ Retrieved trace: %s\n", retrievedTrace.Name)
	}
	fmt.Println()

	// 7. Observations Management
	fmt.Println("7. Observations Management:")
	
	observations, err := api.Observations.List(ctx)
	if err != nil {
		fmt.Printf("   ❌ Failed to list observations: %v\n", err)
	} else {
		fmt.Printf("   ✅ Found %d observations\n", len(observations))
		
		// Show breakdown by type
		spanCount := 0
		generationCount := 0
		eventCount := 0
		
		for _, obs := range observations {
			switch obs.Type {
			case types.ObservationTypeSpan:
				spanCount++
			case types.ObservationTypeGeneration:
				generationCount++
			case types.ObservationTypeEvent:
				eventCount++
			}
		}
		
		fmt.Printf("      - Spans: %d\n", spanCount)
		fmt.Printf("      - Generations: %d\n", generationCount)
		fmt.Printf("      - Events: %d\n", eventCount)
	}
	fmt.Println()

	// 8. Models Information
	fmt.Println("8. Models Information:")
	
	models, err := api.Models.List(ctx)
	if err != nil {
		fmt.Printf("   ❌ Failed to list models: %v\n", err)
	} else {
		fmt.Printf("   ✅ Found %d models configured\n", len(models))
		for i, model := range models {
			if i < 3 {
				fmt.Printf("      - Model: %s\n", model.ModelName)
			}
		}
	}
	fmt.Println()

	// 9. Organization and Project Information
	fmt.Println("9. Organization & Project Info:")
	
	// Get project information (if available)
	projects, err := api.Projects.List(ctx)
	if err != nil {
		fmt.Printf("   ❌ Failed to list projects: %v\n", err)
	} else {
		fmt.Printf("   ✅ Found %d projects\n", len(projects))
		for i, project := range projects {
			if i < 2 {
				fmt.Printf("      - Project: %s (ID: %s)\n", project.Name, project.ID)
			}
		}
	}

	// Get organization information (if available)
	orgs, err := api.Organizations.List(ctx)
	if err != nil {
		fmt.Printf("   ❌ Failed to list organizations: %v\n", err)
	} else {
		fmt.Printf("   ✅ Found %d organizations\n", len(orgs))
	}
	fmt.Println()

	// 10. Final Statistics
	fmt.Println("10. Client Statistics:")
	stats := langfuseClient.GetStats()
	fmt.Printf("    - Traces Created: %d\n", stats.TracesCreated)
	fmt.Printf("    - Generations Created: %d\n", stats.GenerationsCreated)
	fmt.Printf("    - Events Submitted: %d\n", stats.EventsSubmitted)
	fmt.Printf("    - Events Failed: %d\n", stats.EventsFailed)
	fmt.Printf("    - Last Activity: %s\n", stats.LastActivity.Format(time.RFC3339))

	fmt.Println("\n=== Direct API Usage Complete ===")
}

func stringPtr(s string) *string {
	return &s
}