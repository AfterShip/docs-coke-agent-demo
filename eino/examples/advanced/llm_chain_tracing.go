package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/api/resources/commons/types"
)

// Simulated LLM service
type LLMService struct {
	client *client.Langfuse
}

func (llm *LLMService) GenerateResponse(ctx context.Context, prompt string, traceID string, parentSpanID *string) (string, error) {
	// Create generation within the existing trace
	trace := llm.client.Trace("llm-chain-" + traceID) // Reuse existing trace
	
	generation := trace.Generation("llm-response").
		WithModel("gpt-4o-mini", map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  500,
		}).
		WithInput(prompt).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"parent_span": parentSpanID,
			"chain_step":  "generation",
		})

	// Simulate API call delay
	time.Sleep(time.Duration(100+len(prompt)) * time.Millisecond)

	response := fmt.Sprintf("Generated response for: %s", prompt)
	
	generation.WithOutput(response).
		WithEndTime(time.Now()).
		WithUsage(&types.Usage{
			Input:     intPtr(len(prompt) / 4),
			Output:    intPtr(len(response) / 4),
			Total:     intPtr((len(prompt) + len(response)) / 4),
			Unit:      stringPtr("TOKENS"),
			InputCost: floatPtr(0.00003),
			OutputCost: floatPtr(0.00006),
			TotalCost: floatPtr(0.00009),
		})

	if err := generation.End(); err != nil {
		return "", fmt.Errorf("failed to submit generation: %w", err)
	}

	return response, nil
}

// Chain processor that handles multi-step operations
type ChainProcessor struct {
	langfuse *client.Langfuse
	llm      *LLMService
}

func NewChainProcessor(langfuseClient *client.Langfuse) *ChainProcessor {
	return &ChainProcessor{
		langfuse: langfuseClient,
		llm:      &LLMService{client: langfuseClient},
	}
}

func (cp *ChainProcessor) ProcessChain(ctx context.Context, userQuery string) error {
	// Main trace for the entire chain
	mainTrace := cp.langfuse.Trace("llm-chain-processing").
		WithUserID("user-456").
		WithSessionID("session-789").
		WithInput(userQuery).
		WithMetadata(map[string]interface{}{
			"chain_type":    "sequential",
			"steps":         3,
			"max_retries":   2,
		})

	traceID := mainTrace.GetID()

	// Step 1: Query Analysis
	analysisSpan := mainTrace.Span("query-analysis").
		WithInput(userQuery).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"step": 1,
			"operation": "analyze",
		})

	analysisPrompt := fmt.Sprintf("Analyze this query and extract key intents: %s", userQuery)
	analysisResult, err := cp.llm.GenerateResponse(ctx, analysisPrompt, traceID, stringPtr(analysisSpan.GetID()))
	if err != nil {
		return fmt.Errorf("analysis step failed: %w", err)
	}

	analysisSpan.WithOutput(analysisResult).
		WithEndTime(time.Now()).
		WithLevel("DEFAULT")

	if err := analysisSpan.End(); err != nil {
		log.Printf("Failed to submit analysis span: %v", err)
	}

	// Step 2: Content Generation
	generationSpan := mainTrace.Span("content-generation").
		WithInput(analysisResult).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"step": 2,
			"operation": "generate",
			"based_on": "analysis_result",
		})

	generationPrompt := fmt.Sprintf("Based on this analysis, generate a comprehensive response: %s", analysisResult)
	contentResult, err := cp.llm.GenerateResponse(ctx, generationPrompt, traceID, stringPtr(generationSpan.GetID()))
	if err != nil {
		return fmt.Errorf("generation step failed: %w", err)
	}

	generationSpan.WithOutput(contentResult).
		WithEndTime(time.Now()).
		WithLevel("DEFAULT")

	if err := generationSpan.End(); err != nil {
		log.Printf("Failed to submit generation span: %v", err)
	}

	// Step 3: Quality Check
	qualitySpan := mainTrace.Span("quality-check").
		WithInput(contentResult).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"step": 3,
			"operation": "validate",
		})

	qualityPrompt := fmt.Sprintf("Review and rate the quality of this response (1-10): %s", contentResult)
	qualityResult, err := cp.llm.GenerateResponse(ctx, qualityPrompt, traceID, stringPtr(qualitySpan.GetID()))
	if err != nil {
		return fmt.Errorf("quality check step failed: %w", err)
	}

	qualitySpan.WithOutput(qualityResult).
		WithEndTime(time.Now()).
		WithLevel("DEFAULT")

	if err := qualitySpan.End(); err != nil {
		log.Printf("Failed to submit quality span: %v", err)
	}

	// Complete main trace
	finalResult := map[string]interface{}{
		"original_query": userQuery,
		"analysis":       analysisResult,
		"content":        contentResult,
		"quality_score":  qualityResult,
		"steps_completed": 3,
		"status":         "success",
	}

	mainTrace.WithOutput(finalResult)

	if err := mainTrace.End(); err != nil {
		return fmt.Errorf("failed to submit main trace: %w", err)
	}

	// Add evaluation scores
	scores := []*types.Score{
		{
			TraceID:  traceID,
			Name:     "chain_completeness",
			Value:    1.0,
			DataType: types.ScoreDataTypeNumeric,
			Comment:  stringPtr("All chain steps completed successfully"),
		},
		{
			TraceID:  traceID,
			Name:     "response_quality",
			Value:    "high",
			DataType: types.ScoreDataTypeCategorical,
			Comment:  stringPtr("Generated high-quality responses"),
		},
	}

	for _, score := range scores {
		if err := cp.langfuse.Score(score); err != nil {
			log.Printf("Failed to submit score %s: %v", score.Name, err)
		}
	}

	return nil
}

func main() {
	// Initialize Langfuse client
	langfuseClient, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
		client.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	// Create chain processor
	processor := NewChainProcessor(langfuseClient)

	// Process a complex query through the chain
	userQuery := "How can I optimize my machine learning model's performance while reducing computational costs?"

	fmt.Printf("Processing query: %s\n", userQuery)

	if err := processor.ProcessChain(context.Background(), userQuery); err != nil {
		log.Fatal("Chain processing failed:", err)
	}

	fmt.Println("LLM chain processing completed successfully!")

	// Wait for async operations and show stats
	time.Sleep(2 * time.Second)
	
	stats := langfuseClient.GetStats()
	fmt.Printf("Final stats - Traces: %d, Generations: %d, Events: %d\n",
		stats.TracesCreated, stats.GenerationsCreated, stats.EventsSubmitted)
}

// Helper functions
func intPtr(v int) *int { return &v }
func stringPtr(v string) *string { return &v }
func floatPtr(v float64) *float64 { return &v }