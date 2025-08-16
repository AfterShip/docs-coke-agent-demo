// Package langfuse provides a comprehensive Go SDK for Langfuse observability and tracing.
//
// Langfuse is an observability platform for AI applications that helps you trace,
// debug, and monitor LLM-based applications, agents, and complex AI workflows.
// This SDK provides both high-level builder patterns and direct API access for
// maximum flexibility.
//
// # Quick Start
//
// Basic usage with environment variables:
//
//	import "eino/pkg/langfuse/client"
//
//	// Load configuration from environment
//	config, err := client.LoadConfig()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create client
//	langfuse, err := client.New(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer langfuse.Shutdown(context.Background())
//
//	// Create a simple trace
//	trace := langfuse.Trace("my-operation").
//		WithUserID("user-123").
//		WithInput("Hello, world!")
//
//	if err := trace.End(); err != nil {
//		log.Printf("Failed to submit trace: %v", err)
//	}
//
// # Environment Variables
//
// The SDK supports configuration via environment variables:
//
//	LANGFUSE_PUBLIC_KEY    - Your Langfuse public API key (required)
//	LANGFUSE_SECRET_KEY    - Your Langfuse secret API key (required)
//	LANGFUSE_HOST          - Langfuse API endpoint (default: https://cloud.langfuse.com)
//	LANGFUSE_DEBUG         - Enable debug logging (default: false)
//	LANGFUSE_ENABLED       - Enable/disable SDK (default: true)
//	LANGFUSE_FLUSH_AT      - Batch size for auto-flush (default: 15)
//	LANGFUSE_FLUSH_INTERVAL - Time interval for auto-flush (default: 10s)
//	LANGFUSE_TIMEOUT       - Request timeout (default: 10s)
//	LANGFUSE_ENVIRONMENT   - Environment name for traces
//	LANGFUSE_RELEASE       - Release version for traces
//
// # Programmatic Configuration
//
// For more control, use configuration options:
//
//	langfuse, err := client.NewWithOptions(
//		client.WithHost("https://cloud.langfuse.com"),
//		client.WithCredentials("pk_...", "sk_..."),
//		client.WithDebug(true),
//		client.WithEnvironment("production"),
//		client.WithFlushSettings(50, 30*time.Second),
//		client.WithRetrySettings(5, 2*time.Second),
//	)
//
// # Core Concepts
//
// ## Traces
//
// Traces represent complete execution flows, typically user requests or high-level operations:
//
//	trace := langfuse.Trace("user-authentication").
//		WithUserID("user-123").
//		WithSessionID("session-456").
//		WithInput(loginRequest).
//		WithMetadata(map[string]interface{}{
//			"ip_address": clientIP,
//			"user_agent": userAgent,
//		})
//
// ## Spans
//
// Spans represent individual operations within a trace:
//
//	span := trace.Span("database-lookup").
//		WithInput(map[string]interface{}{
//			"query": "SELECT * FROM users WHERE id = ?",
//			"params": []interface{}{userID},
//		}).
//		WithStartTime(time.Now())
//
//	// Perform operation
//	result, err := db.Query(query, userID)
//
//	span.WithOutput(map[string]interface{}{
//		"rows_found": len(result),
//		"error": err,
//	}).WithEndTime(time.Now())
//
//	span.End()
//
// ## Generations
//
// Generations are specialized for LLM inference calls:
//
//	generation := trace.Generation("openai-completion").
//		WithModel("gpt-4", map[string]interface{}{
//			"temperature": 0.7,
//			"max_tokens": 1000,
//		}).
//		WithInput(prompt).
//		WithStartTime(time.Now())
//
//	// Make LLM call
//	response, usage, err := llmClient.Complete(ctx, prompt)
//
//	generation.WithOutput(response).
//		WithUsage(&types.Usage{
//			Input: &usage.PromptTokens,
//			Output: &usage.CompletionTokens,
//			Total: &usage.TotalTokens,
//			InputCost: &usage.PromptCost,
//			OutputCost: &usage.CompletionCost,
//			TotalCost: &usage.TotalCost,
//		}).
//		WithEndTime(time.Now())
//
//	generation.End()
//
// ## Scores
//
// Scores are used for evaluation and rating:
//
//	// Numeric score (0.0-1.0)
//	score := &types.Score{
//		TraceID: traceID,
//		Name: "response_quality",
//		Value: 0.85,
//		DataType: types.ScoreDataTypeNumeric,
//		Comment: stringPtr("High quality response"),
//	}
//	langfuse.Score(score)
//
//	// Categorical score
//	score := &types.Score{
//		TraceID: traceID,
//		Name: "sentiment",
//		Value: "positive",
//		DataType: types.ScoreDataTypeCategorical,
//	}
//	langfuse.Score(score)
//
//	// Boolean score
//	score := &types.Score{
//		TraceID: traceID,
//		Name: "contains_pii",
//		Value: false,
//		DataType: types.ScoreDataTypeBoolean,
//	}
//	langfuse.Score(score)
//
// # Advanced Usage
//
// ## Direct API Access
//
// For advanced use cases, access the underlying API client:
//
//	api := langfuse.API()
//	
//	// List traces
//	traces, err := api.Traces.List(ctx)
//	
//	// Get specific trace
//	trace, err := api.Traces.Get(ctx, traceID)
//	
//	// Manage datasets
//	datasets, err := api.Datasets.List(ctx)
//
// ## Concurrent Usage
//
// The client is thread-safe and supports concurrent usage:
//
//	var wg sync.WaitGroup
//	for i := 0; i < 100; i++ {
//		wg.Add(1)
//		go func(id int) {
//			defer wg.Done()
//			trace := langfuse.Trace(fmt.Sprintf("operation-%d", id))
//			// Configure and end trace...
//		}(i)
//	}
//	wg.Wait()
//
// ## Error Handling
//
// The SDK uses graceful error handling and will not block your application:
//
//	// Trace creation never fails
//	trace := langfuse.Trace("operation")
//
//	// Only trace submission can fail
//	if err := trace.End(); err != nil {
//		// Log error but continue application flow
//		log.Printf("Failed to submit trace: %v", err)
//	}
//
//	// Check if client is healthy
//	if err := langfuse.HealthCheck(ctx); err != nil {
//		log.Printf("Langfuse unhealthy: %v", err)
//	}
//
// ## Performance Monitoring
//
// Monitor SDK performance with built-in statistics:
//
//	stats := langfuse.GetStats()
//	fmt.Printf("Traces created: %d\n", stats.TracesCreated)
//	fmt.Printf("Events submitted: %d\n", stats.EventsSubmitted)
//	fmt.Printf("Events failed: %d\n", stats.EventsFailed)
//	fmt.Printf("Success rate: %.2f%%\n", 
//		float64(stats.EventsSubmitted)/float64(stats.EventsSubmitted+stats.EventsFailed)*100)
//
// # Middleware Integration
//
// ## HTTP Middleware
//
//	import "eino/pkg/langfuse/middleware"
//
//	handler := middleware.HTTP(langfuse)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// Extract trace from context
//		if trace := middleware.TraceFromContext(r.Context()); trace != nil {
//			trace.WithInput(r.URL.Path).WithMetadata(map[string]interface{}{
//				"method": r.Method,
//				"user_agent": r.UserAgent(),
//			})
//		}
//		// Handle request...
//	}))
//
// ## gRPC Interceptors
//
//	server := grpc.NewServer(
//		grpc.UnaryInterceptor(middleware.UnaryServerInterceptor(langfuse)),
//	)
//
// # Best Practices
//
// ## Configuration
//   - Use environment variables for credentials in production
//   - Set appropriate flush intervals based on your throughput
//   - Enable debug mode for development and troubleshooting
//   - Configure timeouts appropriate for your infrastructure
//
// ## Tracing
//   - Use consistent, descriptive names for traces and spans
//   - Include relevant metadata for filtering and analysis
//   - Don't include sensitive data in inputs/outputs
//   - Associate traces with users and sessions when possible
//
// ## Performance
//   - The SDK is async by default for minimal performance impact
//   - Adjust batch sizes and flush intervals for your workload
//   - Use sampling for high-volume applications if needed
//   - Monitor queue depth and submission rates
//
// ## Error Handling
//   - Always call Shutdown() for graceful cleanup
//   - Log but don't fail on tracing errors
//   - Implement circuit breakers for production resilience
//   - Monitor SDK health and statistics
//
// # Examples
//
// See the examples directory for comprehensive usage examples:
//   - examples/basic/ - Simple usage patterns
//   - examples/advanced/ - Complex workflows and patterns
//   - examples/middleware/ - HTTP/gRPC integration
//
// For more information, visit: https://langfuse.com/docs
package langfuse