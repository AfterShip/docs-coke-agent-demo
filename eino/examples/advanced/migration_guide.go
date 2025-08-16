package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/api/resources/commons/types"
)

// This file demonstrates migration patterns from basic to advanced usage
// of the Langfuse Go SDK. Run with: go run examples/advanced/migration_guide.go

func main() {
	fmt.Println("=== Langfuse SDK Migration Guide ===")
	fmt.Println("Demonstrating progression from basic to advanced usage patterns\n")

	// Initialize client (same for all levels)
	langfuse := initializeClient()
	defer langfuse.Shutdown(context.Background())

	fmt.Println("1. Basic Usage → Structured Tracing")
	basicToStructured(langfuse)
	
	fmt.Println("\n2. Simple Operations → Complex Workflows")
	simpleToComplex(langfuse)
	
	fmt.Println("\n3. Manual Submission → Automated Patterns")
	manualToAutomated(langfuse)
	
	fmt.Println("\n4. Basic Configuration → Production Settings")
	basicToProductionConfig()
	
	fmt.Println("\n5. Direct Usage → Middleware Integration")
	directToMiddleware()
	
	fmt.Println("\n=== Migration Complete ===")
}

func initializeClient() *client.Langfuse {
	langfuse, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
	)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	return langfuse
}

// Migration 1: From basic trace creation to structured, hierarchical tracing
func basicToStructured(langfuse *client.Langfuse) {
	fmt.Println("BEFORE: Basic flat tracing")
	// Old way - simple, flat traces
	basicTrace := langfuse.Trace("basic-operation").
		WithInput("simple input")
	
	basicTrace.End()
	fmt.Println("  ✓ Created basic flat trace")

	fmt.Println("AFTER: Structured hierarchical tracing")
	// New way - structured with hierarchy and relationships
	structuredTrace := langfuse.Trace("structured-operation").
		WithUserID("user-123").
		WithSessionID("session-456").
		WithInput(map[string]interface{}{
			"request_id": "req-789",
			"operation": "complex-workflow",
			"parameters": map[string]interface{}{
				"timeout": 30,
				"retries": 3,
			},
		}).
		WithMetadata(map[string]interface{}{
			"version": "2.0",
			"environment": "production",
			"feature_flags": []string{"new-algorithm", "enhanced-logging"},
		}).
		WithTags("workflow", "production", "v2")

	// Add structured spans
	validationSpan := structuredTrace.Span("input-validation").
		WithInput("validation parameters").
		WithStartTime(time.Now())
	
	time.Sleep(10 * time.Millisecond) // Simulate work
	
	validationSpan.WithOutput("validation passed").
		WithEndTime(time.Now())
	validationSpan.End()

	// Add generation with detailed model info
	aiSpan := structuredTrace.Generation("ai-processing").
		WithModel("gpt-4", map[string]interface{}{
			"temperature": 0.7,
			"max_tokens": 500,
			"top_p": 0.9,
		}).
		WithInput("process this complex request").
		WithStartTime(time.Now())
	
	time.Sleep(50 * time.Millisecond) // Simulate AI processing
	
	aiSpan.WithOutput("AI processing complete").
		WithUsage(&types.Usage{
			Input:     intPtr(25),
			Output:    intPtr(150),
			Total:     intPtr(175),
			Unit:      stringPtr("TOKENS"),
			InputCost: floatPtr(0.00075),
			OutputCost: floatPtr(0.003),
			TotalCost: floatPtr(0.00375),
		}).
		WithEndTime(time.Now())
	aiSpan.End()

	structuredTrace.WithOutput(map[string]interface{}{
		"status": "success",
		"processing_time": "60ms",
		"tokens_used": 175,
		"cost": 0.00375,
	})
	structuredTrace.End()
	
	fmt.Println("  ✓ Created structured trace with spans, generations, and detailed metadata")
}

// Migration 2: From simple operations to complex workflow patterns
func simpleToComplex(langfuse *client.Langfuse) {
	fmt.Println("BEFORE: Simple single-step operations")
	// Old way - single operation
	simple := langfuse.Generation("simple-llm-call").
		WithModel("gpt-3.5-turbo", map[string]interface{}{
			"temperature": 0.5,
		}).
		WithInput("What is AI?")
	
	simple.WithOutput("AI is artificial intelligence...")
	simple.End()
	fmt.Println("  ✓ Simple single LLM call")

	fmt.Println("AFTER: Complex multi-step workflows")
	// New way - complex workflow with error handling, retries, and conditional logic
	workflowTrace := langfuse.Trace("complex-ai-workflow").
		WithInput(map[string]interface{}{
			"task": "multi-step-analysis",
			"complexity": "high",
		})

	// Step 1: Data preparation
	prepSpan := workflowTrace.Span("data-preparation").
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"step": 1,
			"type": "preprocessing",
		})
	
	time.Sleep(20 * time.Millisecond)
	prepResult := map[string]interface{}{
		"records_processed": 1000,
		"cleaning_applied": []string{"normalize", "validate", "enrich"},
	}
	prepSpan.WithOutput(prepResult).WithEndTime(time.Now())
	prepSpan.End()

	// Step 2: Analysis with retry logic
	for attempt := 1; attempt <= 3; attempt++ {
		analysisGen := workflowTrace.Generation(fmt.Sprintf("analysis-attempt-%d", attempt)).
			WithModel("gpt-4", map[string]interface{}{
				"temperature": 0.3,
				"attempt": attempt,
			}).
			WithInput(prepResult).
			WithStartTime(time.Now()).
			WithMetadata(map[string]interface{}{
				"step": 2,
				"retry_attempt": attempt,
			})

		// Simulate occasional failures
		if attempt < 3 {
			analysisGen.WithOutput(map[string]interface{}{
				"error": "temporary failure",
				"retry_in": "1s",
			}).WithLevel("ERROR")
		} else {
			analysisGen.WithOutput(map[string]interface{}{
				"analysis": "detailed insights...",
				"confidence": 0.92,
				"key_findings": []string{"insight1", "insight2", "insight3"},
			}).WithUsage(&types.Usage{
				Input:  intPtr(200),
				Output: intPtr(150),
				Total:  intPtr(350),
			})
		}
		
		analysisGen.WithEndTime(time.Now())
		analysisGen.End()

		if attempt == 3 {
			break // Success on third attempt
		}
		time.Sleep(100 * time.Millisecond) // Retry delay
	}

	// Step 3: Validation and scoring
	validationGen := workflowTrace.Generation("result-validation").
		WithModel("gpt-3.5-turbo", map[string]interface{}{
			"temperature": 0.1, // Low temperature for validation
		}).
		WithInput("validate analysis results").
		WithStartTime(time.Now())
	
	time.Sleep(30 * time.Millisecond)
	
	validationGen.WithOutput(map[string]interface{}{
		"validation": "passed",
		"quality_score": 0.95,
		"issues_found": 0,
	}).WithEndTime(time.Now())
	validationGen.End()

	workflowTrace.WithOutput(map[string]interface{}{
		"workflow_status": "completed",
		"total_steps": 3,
		"success_rate": "100%",
		"total_cost": 0.0125,
	})
	workflowTrace.End()

	fmt.Println("  ✓ Complex workflow with multiple steps, retries, and validation")
}

// Migration 3: From manual submission to automated patterns
func manualToAutomated(langfuse *client.Langfuse) {
	fmt.Println("BEFORE: Manual trace management")
	// Old way - manual everything
	manualTrace := langfuse.Trace("manual-operation")
	manualTrace.WithInput("manual input")
	// ... do work manually ...
	manualTrace.WithOutput("manual output")
	if err := manualTrace.End(); err != nil {
		fmt.Printf("  ! Manual submission error: %v\n", err)
	}
	fmt.Println("  ✓ Manual trace created and submitted")

	fmt.Println("AFTER: Automated patterns with helpers")
	// New way - automated patterns and helpers
	
	// Pattern 1: Automated timing
	timedTrace := langfuse.Trace("auto-timed-operation")
	startTime := time.Now()
	
	// Simulate work
	time.Sleep(25 * time.Millisecond)
	
	duration := time.Since(startTime)
	timedTrace.WithInput("auto-timed input").
		WithOutput(map[string]interface{}{
			"result": "success",
			"duration_ms": duration.Milliseconds(),
		}).
		WithMetadata(map[string]interface{}{
			"auto_timed": true,
			"start_time": startTime,
			"end_time": time.Now(),
		})
	timedTrace.End()
	
	// Pattern 2: Error handling wrapper
	func() {
		errorTrace := langfuse.Trace("error-handled-operation")
		defer func() {
			if r := recover(); r != nil {
				errorTrace.WithOutput(map[string]interface{}{
					"error": fmt.Sprintf("panic: %v", r),
					"recovered": true,
				}).WithLevel("ERROR")
				errorTrace.End()
			}
		}()
		
		errorTrace.WithInput("risky operation")
		// Simulate work that might panic
		if time.Now().UnixNano()%2 == 0 {
			// Sometimes succeed
			errorTrace.WithOutput("success")
		} else {
			// Sometimes fail gracefully
			errorTrace.WithOutput(map[string]interface{}{
				"error": "controlled failure",
				"handled": true,
			}).WithLevel("WARNING")
		}
		errorTrace.End()
	}()

	fmt.Println("  ✓ Automated patterns: timing, error handling, recovery")
}

// Migration 4: From basic to production configuration
func basicToProductionConfig() {
	fmt.Println("BEFORE: Basic configuration")
	// Old way - minimal configuration
	basicClient, err := client.NewWithOptions(
		client.WithCredentials("pk_basic", "sk_basic"),
		client.WithDebug(true),
	)
	if err == nil {
		fmt.Println("  ✓ Basic client with minimal settings")
		basicClient.Shutdown(context.Background())
	}

	fmt.Println("AFTER: Production-ready configuration")
	// New way - comprehensive production configuration
	prodClient, err := client.NewWithOptions(
		// Authentication
		client.WithCredentials("pk_prod", "sk_prod"),
		client.WithHost("https://cloud.langfuse.com"),
		
		// Environment identification
		client.WithEnvironment("production"),
		client.WithRelease("v2.1.0"),
		
		// Performance optimization
		client.WithFlushSettings(50, 30*time.Second),  // Larger batches, less frequent
		client.WithQueueSize(5000),                     // Larger queue for high throughput
		client.WithTimeout(30*time.Second),             // Longer timeout for reliability
		
		// Reliability settings
		client.WithRetrySettings(5, 2*time.Second),    // More aggressive retries
		
		// Production features
		client.WithDebug(false),                        // Disable debug in production
		client.WithEnabled(true),                       // Explicit enable
	)
	if err == nil {
		fmt.Println("  ✓ Production client with comprehensive settings:")
		fmt.Println("    - Optimized batching and queue sizes")
		fmt.Println("    - Aggressive retry and timeout settings")
		fmt.Println("    - Environment and release tracking")
		fmt.Println("    - Performance monitoring ready")
		prodClient.Shutdown(context.Background())
	}
}

// Migration 5: From direct usage to middleware integration
func directToMiddleware() {
	fmt.Println("BEFORE: Direct manual instrumentation")
	fmt.Println("  // Manual trace creation in every handler")
	fmt.Println("  func handler(w http.ResponseWriter, r *http.Request) {")
	fmt.Println("    trace := langfuse.Trace(r.URL.Path)")
	fmt.Println("    defer trace.End()")
	fmt.Println("    // ... handle request manually ...")
	fmt.Println("  }")

	fmt.Println("AFTER: Automatic middleware integration")
	fmt.Println("  // Automatic tracing for all requests")
	fmt.Println("  import \"eino/pkg/langfuse/middleware\"")
	fmt.Println("  ")
	fmt.Println("  // HTTP middleware")
	fmt.Println("  handler := middleware.HTTP(langfuse)(yourHandler)")
	fmt.Println("  ")
	fmt.Println("  // gRPC interceptors")
	fmt.Println("  server := grpc.NewServer(")
	fmt.Println("    grpc.UnaryInterceptor(middleware.UnaryServerInterceptor(langfuse)),")
	fmt.Println("  )")
	fmt.Println("  ")
	fmt.Println("  // Extract traces from context automatically")
	fmt.Println("  func handler(w http.ResponseWriter, r *http.Request) {")
	fmt.Println("    if trace := middleware.TraceFromContext(r.Context()); trace != nil {")
	fmt.Println("      trace.WithMetadata(customData)")
	fmt.Println("    }")
	fmt.Println("    // ... focus on business logic ...")
	fmt.Println("  }")
	fmt.Println("  ✓ Automatic instrumentation with context propagation")
}

// Helper functions
func intPtr(v int) *int { return &v }
func stringPtr(s string) *string { return &s }
func floatPtr(v float64) *float64 { return &v }