// Package client provides the high-level Langfuse SDK client implementation.
//
// This package offers a fluent, builder-pattern API for creating traces, spans, and generations,
// while also providing direct access to the underlying REST API for advanced use cases.
//
// Basic usage:
//
//	config, err := client.LoadConfig()
//	if err != nil {
//		log.Fatal(err)
//	}
//
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
// For more examples, see the examples directory.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"eino/pkg/langfuse/api"
	"eino/pkg/langfuse/api/resources/commons/types"
	scoreTypes "eino/pkg/langfuse/api/resources/scores/types"
	"eino/pkg/langfuse/config"
	"eino/pkg/langfuse/internal/queue"
	"eino/pkg/langfuse/internal/utils"
)

// Langfuse is the main SDK client providing high-level builder APIs and direct API access.
//
// The client manages traces, spans, generations, and scores through a fluent builder pattern
// while handling async processing, retry logic, and connection management transparently.
//
// The client is thread-safe and can be used concurrently from multiple goroutines.
// It automatically batches and submits events to the Langfuse API in the background
// for optimal performance.
//
// Key features:
//   - Fluent builder pattern for traces, spans, and generations
//   - Automatic async processing with configurable batching
//   - Built-in retry logic and error handling
//   - Direct API access for advanced operations
//   - Thread-safe concurrent usage
//   - Comprehensive statistics and health monitoring
type Langfuse struct {
	// Core components
	config    *config.Config
	apiClient *api.APIClient
	queue     *queue.IngestionQueue

	// State management
	mu     sync.RWMutex
	closed bool

	// Statistics
	stats   *ClientStats
	statsMu sync.RWMutex
}

// ClientStats represents comprehensive usage statistics for the Langfuse client.
//
// These statistics help monitor SDK performance, track usage patterns,
// and debug potential issues with event submission.
type ClientStats struct {
	// TracesCreated is the total number of traces created since client initialization
	TracesCreated int64 `json:"tracesCreated"`

	// SpansCreated is the total number of spans created since client initialization
	SpansCreated int64 `json:"spansCreated"`

	// GenerationsCreated is the total number of generations created since client initialization
	GenerationsCreated int64 `json:"generationsCreated"`

	// EventsEnqueued is the total number of events added to the queue
	EventsEnqueued int64 `json:"eventsEnqueued"`

	// EventsSubmitted is the total number of events successfully submitted to Langfuse
	EventsSubmitted int64 `json:"eventsSubmitted"`

	// EventsFailed is the total number of events that failed to submit after all retries
	EventsFailed int64 `json:"eventsFailed"`

	// LastActivity is the timestamp of the last SDK activity (creation or submission)
	LastActivity time.Time `json:"lastActivity"`

	// CreatedAt is the timestamp when the client was created
	CreatedAt time.Time `json:"createdAt"`
}

// New creates a new Langfuse client instance with the provided configuration.
//
// The configuration should be created using LoadConfig() or NewConfig() with appropriate
// options. The client will validate the configuration and establish connections to the
// Langfuse API.
//
// If the configuration has Enabled=false, a disabled client is returned that accepts
// all operations but performs no actual work. This is useful for testing or when
// tracing should be conditionally disabled.
//
// Example:
//
//	config, err := client.LoadConfig()
//	if err != nil {
//		return fmt.Errorf("failed to load config: %w", err)
//	}
//
//	langfuse, err := client.New(config)
//	if err != nil {
//		return fmt.Errorf("failed to create client: %w", err)
//	}
//	defer langfuse.Shutdown(context.Background())
//
// Returns an error if:
//   - config is nil
//   - config validation fails
//   - API client creation fails
//   - queue initialization fails
func New(config *config.Config) (*Langfuse, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Skip if SDK is disabled
	if !config.Enabled {
		return newDisabledClient(config), nil
	}

	// Create API client
	apiClient, err := api.NewAPIClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Create client instance first so we can reference it in hooks
	client := &Langfuse{
		config:    config,
		apiClient: apiClient,
		closed:    false,
		stats: &ClientStats{
			CreatedAt: time.Now(),
		},
	}

	// Create ingestion queue with proper configuration and event hooks
	queueConfig := &queue.QueueConfig{
		FlushAt:       config.FlushAt,
		FlushInterval: config.FlushInterval,
		MaxRetries:    config.RetryCount,
		RetryBackoff:  config.RetryWaitTime,
		MaxQueueSize:  config.QueueSize,
		OnFlushEnd: func(batchSize int, success bool, err error) {
			client.statsMu.Lock()
			client.stats.LastActivity = time.Now()
			if success {
				client.stats.EventsSubmitted += int64(batchSize)
			} else {
				client.stats.EventsFailed += int64(batchSize)
			}
			client.statsMu.Unlock()
		},
	}

	client.queue = queue.NewIngestionQueue(apiClient.Ingestion, queueConfig)

	return client, nil
}

// NewWithOptions creates a new Langfuse client with configuration options.
//
// This is a convenience method that combines configuration creation and client
// initialization in a single call. It's equivalent to calling NewConfig() followed
// by New(), but provides a more fluent API.
//
// Example:
//
//	langfuse, err := client.NewWithOptions(
//		client.WithHost("https://cloud.langfuse.com"),
//		client.WithCredentials("pk_...", "sk_..."),
//		client.WithDebug(true),
//		client.WithEnvironment("development"),
//		client.WithFlushSettings(20, 30*time.Second),
//	)
//	if err != nil {
//		return fmt.Errorf("failed to create client: %w", err)
//	}
//	defer langfuse.Shutdown(context.Background())
//
// Returns an error if configuration creation or client initialization fails.
func NewWithOptions(options ...ConfigOption) (*Langfuse, error) {
	config, err := NewConfig(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return New(config)
}

// newDisabledClient creates a disabled client that accepts all operations but performs no work.
//
// This is used internally when the SDK is disabled via configuration. The disabled client
// maintains the same API surface but all operations become no-ops, allowing applications
// to conditionally disable tracing without code changes.
func newDisabledClient(config *Config) *Langfuse {
	return &Langfuse{
		config: config,
		closed: true, // Mark as closed to prevent operations
		stats: &ClientStats{
			CreatedAt: time.Now(),
		},
	}
}

// Trace creates a new trace builder for the given operation name.
//
// Traces represent complete execution flows, typically corresponding to user requests
// or top-level operations. Each trace can contain multiple spans and generations to
// provide detailed observability into complex workflows.
//
// The trace builder provides a fluent API for configuring the trace before submission:
//
//	trace := client.Trace("user-authentication").
//		WithUserID("user-123").
//		WithSessionID("session-456").
//		WithInput(loginRequest).
//		WithMetadata(map[string]interface{}{
//			"client_ip": "192.168.1.1",
//			"user_agent": "MyApp/1.0",
//		})
//
//	// Add spans and generations to the trace
//	authSpan := trace.Span("database-lookup")
//	// ... configure span ...
//	authSpan.End()
//
//	// Complete the trace
//	trace.WithOutput(authResult)
//	if err := trace.End(); err != nil {
//		log.Printf("Failed to submit trace: %v", err)
//	}
//
// The name should be descriptive and consistent across similar operations to enable
// effective grouping and analysis in the Langfuse UI.
//
// If the client is disabled, returns a no-op trace builder that accepts all operations
// but performs no actual work.
func (lf *Langfuse) Trace(name string) *TraceBuilder {
	if lf.isDisabled() {
		return newDisabledTraceBuilder(name)
	}

	lf.statsMu.Lock()
	lf.stats.TracesCreated++
	lf.stats.LastActivity = time.Now()
	lf.statsMu.Unlock()

	builder := NewTraceBuilder(lf)
	builder.Name(name)

	return builder
}

// Span creates a standalone span with an automatically generated parent trace.
//
// Spans represent units of work within a trace, such as database operations, API calls,
// or processing steps. When created as standalone spans (not within an existing trace),
// a parent trace is automatically generated.
//
// For most use cases, spans should be created within existing traces using trace.Span()
// rather than as standalone spans. Standalone spans are useful for simple operations
// that don't require complex trace hierarchies.
//
// Example:
//
//	span := client.Span("database-query").
//		WithInput(map[string]interface{}{
//			"query": "SELECT * FROM users WHERE id = ?",
//			"params": []interface{}{userID},
//		}).
//		WithStartTime(time.Now())
//
//	// Perform database operation
//	result, err := db.Query(query, userID)
//
//	span.WithOutput(map[string]interface{}{
//		"rows_returned": len(result),
//		"error": err,
//	}).WithEndTime(time.Now())
//
//	if err := span.End(); err != nil {
//		log.Printf("Failed to submit span: %v", err)
//	}
//
// If the client is disabled, returns a no-op span builder.
func (lf *Langfuse) Span(name string) *SpanBuilder {
	if lf.isDisabled() {
		return newDisabledSpanBuilder(name)
	}

	// Create a trace automatically for standalone spans
	traceID := utils.GenerateTraceID()

	lf.statsMu.Lock()
	lf.stats.SpansCreated++
	lf.stats.LastActivity = time.Now()
	lf.statsMu.Unlock()

	builder := NewSpanBuilder(lf, traceID)
	builder.Name(name)

	return builder
}

// Generation creates a standalone generation with an automatically generated parent trace.
//
// Generations are specialized observations for LLM inference calls, capturing model
// parameters, inputs, outputs, token usage, and costs. When created as standalone
// generations (not within an existing trace), a parent trace is automatically generated.
//
// Generations should be used to track all interactions with language models, whether
// they're part of complex AI workflows or simple standalone calls.
//
// Example:
//
//	generation := client.Generation("openai-chat-completion").
//		WithModel("gpt-4", map[string]interface{}{
//			"temperature": 0.7,
//			"max_tokens": 1000,
//		}).
//		WithInput([]map[string]string{
//			{"role": "user", "content": "Explain quantum computing"},
//		}).
//		WithStartTime(time.Now())
//
//	// Make LLM API call
//	response, usage, err := openaiClient.CreateChatCompletion(ctx, request)
//
//	generation.WithOutput(response.Choices[0].Message).
//		WithUsage(&types.Usage{
//			Input: &usage.PromptTokens,
//			Output: &usage.CompletionTokens,
//			Total: &usage.TotalTokens,
//		}).
//		WithEndTime(time.Now())
//
//	if err := generation.End(); err != nil {
//		log.Printf("Failed to submit generation: %v", err)
//	}
//
// If the client is disabled, returns a no-op generation builder.
func (lf *Langfuse) Generation(name string) *GenerationBuilder {
	if lf.isDisabled() {
		return newDisabledGenerationBuilder(name)
	}

	// Create a trace automatically for standalone generations
	traceID := utils.GenerateTraceID()

	lf.statsMu.Lock()
	lf.stats.GenerationsCreated++
	lf.stats.LastActivity = time.Now()
	lf.statsMu.Unlock()

	builder := NewGenerationBuilder(lf, traceID)
	builder.Name(name)

	return builder
}

// Score creates and submits a score directly to the Langfuse API.
//
// Scores are used to evaluate and rate traces, spans, or generations. They can be
// numeric (0.0-1.0 quality scores), categorical (pass/fail, good/bad/excellent),
// or boolean (true/false flags).
//
// Unlike traces and generations which use builders, scores are submitted immediately
// and synchronously. This method is useful for programmatic evaluation and automated
// scoring workflows.
//
// Example:
//
//	score := &types.Score{
//		TraceID: traceID,
//		Name: "response_quality",
//		Value: 0.85,
//		DataType: types.ScoreDataTypeNumeric,
//		Comment: stringPtr("High quality response with good accuracy"),
//	}
//
//	if err := client.Score(score); err != nil {
//		log.Printf("Failed to submit score: %v", err)
//	}
//
// The score must have a valid TraceID referencing an existing trace. The Name should
// be consistent across similar evaluations to enable analysis and aggregation.
//
// Returns an error if the score is invalid, the trace doesn't exist, or submission fails.
// If the client is disabled, this method returns nil without error.
func (lf *Langfuse) Score(score *types.Score) error {
	if lf.isDisabled() {
		return nil
	}

	if err := lf.validateScore(score); err != nil {
		return fmt.Errorf("score validation failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), lf.config.RequestTimeout)
	defer cancel()

	// Convert Score to CreateScoreRequest
	var value interface{}
	if score.Value != nil {
		if err := json.Unmarshal(score.Value, &value); err != nil {
			return fmt.Errorf("failed to unmarshal score value: %w", err)
		}
	}

	req := &scoreTypes.CreateScoreRequest{
		TraceID:       score.TraceID,
		ObservationID: score.ObservationID,
		Name:          score.Name,
		Value:         value,
		DataType:      score.DataType,
		Comment:       score.Comment,
		ConfigID:      score.ConfigID,
	}

	if score.ID != "" {
		req.ID = &score.ID
	}

	_, err := lf.apiClient.Scores.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}

	lf.statsMu.Lock()
	lf.stats.LastActivity = time.Now()
	lf.statsMu.Unlock()

	return nil
}

// API returns the underlying API client for direct API access
func (lf *Langfuse) API() *api.APIClient {
	if lf.isDisabled() {
		return nil
	}
	return lf.apiClient
}

// GetConfig returns a copy of the client configuration
func (lf *Langfuse) GetConfig() *Config {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	// Return a copy to prevent modification
	configCopy := *lf.config
	return &configCopy
}

// GetStats returns current client statistics
func (lf *Langfuse) GetStats() *ClientStats {
	lf.statsMu.RLock()
	defer lf.statsMu.RUnlock()

	// Return a copy to prevent modification
	statsCopy := *lf.stats
	return &statsCopy
}

// IsEnabled returns whether the client is enabled and operational
func (lf *Langfuse) IsEnabled() bool {
	return lf.config.Enabled && !lf.closed
}

// IsHealthy returns whether the underlying API client is healthy
func (lf *Langfuse) IsHealthy() bool {
	if lf.isDisabled() {
		return false
	}
	return lf.apiClient.IsHealthy()
}

// Flush forces immediate submission of all queued events
func (lf *Langfuse) Flush(ctx context.Context) error {
	if lf.isDisabled() {
		return nil
	}

	if lf.queue == nil {
		return nil
	}

	return lf.queue.Flush()
}

// Shutdown gracefully shuts down the client, flushing pending events
func (lf *Langfuse) Shutdown(ctx context.Context) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.closed {
		return nil
	}

	var shutdownError error

	// Flush pending events first
	if lf.queue != nil {
		if err := lf.queue.Flush(); err != nil {
			shutdownError = fmt.Errorf("failed to flush queue during shutdown: %w", err)
		}

		// Shutdown the queue
		if err := lf.queue.Shutdown(ctx); err != nil {
			if shutdownError != nil {
				shutdownError = fmt.Errorf("%w; queue shutdown error: %v", shutdownError, err)
			} else {
				shutdownError = fmt.Errorf("failed to shutdown queue: %w", err)
			}
		}
	}

	// Close API client
	if lf.apiClient != nil {
		if err := lf.apiClient.Close(); err != nil {
			if shutdownError != nil {
				shutdownError = fmt.Errorf("%w; API client close error: %v", shutdownError, err)
			} else {
				shutdownError = fmt.Errorf("failed to close API client: %w", err)
			}
		}
	}

	lf.closed = true
	return shutdownError
}

// HealthCheck performs a health check against the Langfuse API
func (lf *Langfuse) HealthCheck(ctx context.Context) error {
	if lf.isDisabled() {
		return fmt.Errorf("client is disabled")
	}

	return lf.apiClient.TestConnection(ctx)
}

// WaitForHealthy waits for the service to become healthy within the given timeout
func (lf *Langfuse) WaitForHealthy(ctx context.Context, checkInterval time.Duration) error {
	if lf.isDisabled() {
		return fmt.Errorf("client is disabled")
	}

	return lf.apiClient.WaitForHealthy(ctx, checkInterval)
}

// GetVersion returns the SDK version
func (lf *Langfuse) GetVersion() string {
	return lf.config.Version
}

// GetEnvironment returns the configured environment
func (lf *Langfuse) GetEnvironment() string {
	return lf.config.Environment
}

// isDisabled checks if the client is disabled or closed
func (lf *Langfuse) isDisabled() bool {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	return !lf.config.Enabled || lf.closed
}

// validateScore performs basic validation on a score
func (lf *Langfuse) validateScore(score *types.Score) error {
	if score == nil {
		return fmt.Errorf("score cannot be nil")
	}

	if score.Name == "" {
		return fmt.Errorf("score name is required")
	}

	if score.TraceID == "" {
		return fmt.Errorf("score trace ID is required")
	}

	if score.Value == nil {
		return fmt.Errorf("score value is required")
	}

	return nil
}

// Disabled builder constructors that return no-op builders
func newDisabledTraceBuilder(name string) *TraceBuilder {
	return &TraceBuilder{
		name:      name,
		submitted: true, // Mark as submitted to prevent operations
	}
}

func newDisabledSpanBuilder(name string) *SpanBuilder {
	return &SpanBuilder{
		name:      name,
		submitted: true, // Mark as submitted to prevent operations
	}
}

func newDisabledGenerationBuilder(name string) *GenerationBuilder {
	return &GenerationBuilder{
		name:      name,
		submitted: true, // Mark as submitted to prevent operations
	}
}

// Context-aware operations

// WithTimeout returns a new client instance that uses the specified timeout for operations
func (lf *Langfuse) WithTimeout(timeout time.Duration) *Langfuse {
	if lf.isDisabled() {
		return lf
	}

	// Create a copy of the config with the new timeout
	newConfig := *lf.config
	newConfig.RequestTimeout = timeout

	// Create a new client instance with the updated config
	newClient := *lf
	newClient.config = &newConfig

	return &newClient
}

// WithContext returns operations that can be performed with a specific context
// This is a convenience method for context-aware operations
func (lf *Langfuse) WithContext(ctx context.Context) *ContextualOperations {
	return &ContextualOperations{
		ctx:    ctx,
		client: lf,
	}
}

// ContextualOperations provides context-aware operations
type ContextualOperations struct {
	ctx    context.Context
	client *Langfuse
}

// Flush flushes queued events using the provided context
func (co *ContextualOperations) Flush() error {
	return co.client.Flush(co.ctx)
}

// Shutdown shuts down the client using the provided context
func (co *ContextualOperations) Shutdown() error {
	return co.client.Shutdown(co.ctx)
}

// HealthCheck performs a health check using the provided context
func (co *ContextualOperations) HealthCheck() error {
	return co.client.HealthCheck(co.ctx)
}
