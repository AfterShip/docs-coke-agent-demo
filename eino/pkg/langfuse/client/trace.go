package client

import (
	"context"
	"time"

	"eino/pkg/langfuse/api/resources/ingestion/types"
	"eino/pkg/langfuse/internal/utils"
)

// TraceBuilder provides a fluent API for building and configuring trace events.
//
// Traces represent complete execution flows, typically corresponding to user requests
// or high-level operations. The builder pattern allows for easy configuration of
// trace properties before submission to Langfuse.
//
// The builder is NOT thread-safe and should not be shared across goroutines.
// Each trace should be built and submitted from a single goroutine.
//
// Example usage:
//
//	trace := client.Trace("user-authentication").
//		WithUserID("user-123").
//		WithSessionID("session-456").
//		WithInput(loginRequest).
//		WithMetadata(map[string]interface{}{
//			"ip_address": "192.168.1.1",
//			"user_agent": "MyApp/1.0",
//		}).
//		WithTags("authentication", "login")
//
//	// Add spans and generations
//	span := trace.Span("database-lookup")
//	// ... configure and end span
//
//	// Complete the trace
//	trace.WithOutput(authResult)
//	if err := trace.End(); err != nil {
//		log.Printf("Failed to submit trace: %v", err)
//	}
type TraceBuilder struct {
	id          string                    // Unique identifier for the trace
	name        string                    // Human-readable name describing the operation
	userID      *string                  // Optional user identifier
	sessionID   *string                  // Optional session identifier for grouping
	input       interface{}              // Input data or parameters
	output      interface{}              // Output data or results
	metadata    map[string]interface{}   // Additional key-value metadata
	tags        []string                 // List of tags for categorization
	version     *string                  // Optional version identifier
	release     *string                  // Optional release identifier
	public      *bool                    // Whether the trace should be publicly visible
	timestamp   time.Time                // When the trace was created
	client      *Langfuse               // Reference to parent client
	submitted   bool                     // Whether this trace has been submitted
}

// NewTraceBuilder creates a new TraceBuilder instance with default settings.
//
// The builder is initialized with:
//   - A generated unique trace ID
//   - Current UTC timestamp
//   - Empty metadata and tags collections
//   - Reference to the parent Langfuse client
//
// This function is typically called internally by Langfuse.Trace() rather
// than directly by application code.
func NewTraceBuilder(client *Langfuse) *TraceBuilder {
	return &TraceBuilder{
		id:        utils.GenerateTraceID(),
		timestamp: time.Now().UTC(),
		client:    client,
		metadata:  make(map[string]interface{}),
		tags:      make([]string, 0),
	}
}

// ID sets a custom trace ID, overriding the auto-generated one.
//
// This method is useful when you need to correlate traces with external systems
// or when you have your own ID generation scheme. The ID must be unique within
// your Langfuse project.
//
// Example:
//
//	trace := client.Trace("operation").
//		ID("my-custom-trace-id-12345")
//
// If the trace has already been submitted, this method has no effect and returns
// the builder unchanged to maintain the fluent interface.
func (tb *TraceBuilder) ID(id string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.id = id
	return tb
}

// Name sets the human-readable name for the trace.
//
// The name should be descriptive and consistent across similar operations to enable
// effective grouping and analysis in the Langfuse UI. Good names are concise but
// specific enough to understand the operation being traced.
//
// Examples of good trace names:
//   - "user-authentication"
//   - "document-processing"
//   - "payment-processing"
//   - "api-request-/users/{id}"
//
// Example:
//
//	trace := client.Trace("user-registration").
//		Name("user-email-verification")  // Override the initial name
//
// If the trace has already been submitted, this method has no effect.
func (tb *TraceBuilder) Name(name string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.name = name
	return tb
}

// UserID sets the user ID
func (tb *TraceBuilder) UserID(userID string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.userID = &userID
	return tb
}

// SessionID sets the session ID
func (tb *TraceBuilder) SessionID(sessionID string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.sessionID = &sessionID
	return tb
}

// Input sets the input data
func (tb *TraceBuilder) Input(input interface{}) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.input = input
	return tb
}

// Output sets the output data
func (tb *TraceBuilder) Output(output interface{}) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.output = output
	return tb
}

// Metadata sets the metadata map
func (tb *TraceBuilder) Metadata(metadata map[string]interface{}) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.metadata = metadata
	return tb
}

// AddMetadata adds a single metadata key-value pair
func (tb *TraceBuilder) AddMetadata(key string, value interface{}) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	if tb.metadata == nil {
		tb.metadata = make(map[string]interface{})
	}
	tb.metadata[key] = value
	return tb
}

// Tags sets the tags
func (tb *TraceBuilder) Tags(tags ...string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.tags = tags
	return tb
}

// AddTag adds a single tag
func (tb *TraceBuilder) AddTag(tag string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.tags = append(tb.tags, tag)
	return tb
}

// Version sets the version
func (tb *TraceBuilder) Version(version string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.version = &version
	return tb
}

// Release sets the release
func (tb *TraceBuilder) Release(release string) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.release = &release
	return tb
}

// Public sets the public flag
func (tb *TraceBuilder) Public(public bool) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.public = &public
	return tb
}

// Timestamp sets the timestamp
func (tb *TraceBuilder) Timestamp(timestamp time.Time) *TraceBuilder {
	if tb.submitted {
		return tb
	}
	tb.timestamp = timestamp.UTC()
	return tb
}

// GetID returns the trace ID
func (tb *TraceBuilder) GetID() string {
	return tb.id
}

// GetName returns the trace name
func (tb *TraceBuilder) GetName() string {
	return tb.name
}

// GetUserID returns the user ID
func (tb *TraceBuilder) GetUserID() string {
	if tb.userID == nil {
		return ""
	}
	return *tb.userID
}

// GetSessionID returns the session ID
func (tb *TraceBuilder) GetSessionID() string {
	if tb.sessionID == nil {
		return ""
	}
	return *tb.sessionID
}

// WithUser is an alias for UserID for fluent API
func (tb *TraceBuilder) WithUser(userID string) *TraceBuilder {
	return tb.UserID(userID)
}

// WithSession is an alias for SessionID for fluent API
func (tb *TraceBuilder) WithSession(sessionID string) *TraceBuilder {
	return tb.SessionID(sessionID)
}

// WithInput is an alias for Input for fluent API
func (tb *TraceBuilder) WithInput(input interface{}) *TraceBuilder {
	return tb.Input(input)
}

// WithOutput is an alias for Output for fluent API
func (tb *TraceBuilder) WithOutput(output interface{}) *TraceBuilder {
	return tb.Output(output)
}

// WithMetadata is an alias for Metadata for fluent API
func (tb *TraceBuilder) WithMetadata(metadata map[string]interface{}) *TraceBuilder {
	return tb.Metadata(metadata)
}

// WithTags is an alias for Tags for fluent API
func (tb *TraceBuilder) WithTags(tags ...string) *TraceBuilder {
	return tb.Tags(tags...)
}

// Span creates a new span within this trace
func (tb *TraceBuilder) Span(name string) *SpanBuilder {
	span := NewSpanBuilder(tb.client, tb.id)
	return span.Name(name)
}

// validate performs validation on the trace builder
func (tb *TraceBuilder) validate() error {
	if tb.id == "" {
		return &ValidationError{Field: "id", Message: "trace id is required"}
	}
	
	if tb.name == "" {
		return &ValidationError{Field: "name", Message: "trace name is required"}
	}
	
	if tb.timestamp.IsZero() {
		return &ValidationError{Field: "timestamp", Message: "trace timestamp is required"}
	}
	
	return nil
}

// toTraceEvent converts the builder to a TraceEvent
func (tb *TraceBuilder) toTraceEvent() *types.TraceEvent {
	return &types.TraceEvent{
		ID:        tb.id,
		Name:      tb.name,
		UserID:    tb.userID,
		SessionID: tb.sessionID,
		Input:     tb.input,
		Output:    tb.output,
		Metadata:  tb.metadata,
		Tags:      tb.tags,
		Version:   tb.version,
		Release:   tb.release,
		Public:    tb.public,
		Timestamp: tb.timestamp,
	}
}

// toTraceCreateEvent converts the builder to a TraceCreateEvent
func (tb *TraceBuilder) toTraceCreateEvent() *types.TraceCreateEvent {
	return &types.TraceCreateEvent{
		TraceEvent: *tb.toTraceEvent(),
		Type:       "trace-create",
	}
}

// Submit submits the trace to the ingestion queue
func (tb *TraceBuilder) Submit(ctx context.Context) error {
	if tb.submitted {
		return &ValidationError{Field: "state", Message: "trace already submitted"}
	}
	
	if err := tb.validate(); err != nil {
		return err
	}
	
	event := tb.toTraceCreateEvent()
	ingestionEvent := event.ToIngestionEvent()
	
	if err := tb.client.queue.Enqueue(ingestionEvent); err != nil {
		return err
	}
	
	tb.submitted = true
	return nil
}

// Update updates an existing trace
func (tb *TraceBuilder) Update(ctx context.Context) error {
	if tb.submitted {
		return &ValidationError{Field: "state", Message: "trace already submitted"}
	}
	
	if err := tb.validate(); err != nil {
		return err
	}
	
	traceEvent := tb.toTraceEvent()
	updateEvent := &types.TraceUpdateEvent{
		TraceEvent: *traceEvent,
		Type:       "trace-update",
	}
	
	ingestionEvent := updateEvent.ToIngestionEvent()
	
	if err := tb.client.queue.Enqueue(ingestionEvent); err != nil {
		return err
	}
	
	tb.submitted = true
	return nil
}

// End marks the trace as ended with the current timestamp
func (tb *TraceBuilder) End(ctx context.Context) error {
	return tb.EndAt(ctx, time.Now().UTC())
}

// EndAt marks the trace as ended with a specific timestamp
func (tb *TraceBuilder) EndAt(ctx context.Context, endTime time.Time) error {
	if tb.submitted {
		return &ValidationError{Field: "state", Message: "trace already submitted"}
	}
	
	if err := tb.validate(); err != nil {
		return err
	}
	
	traceEvent := tb.toTraceEvent()
	updateEvent := &types.TraceUpdateEvent{
		TraceEvent: *traceEvent,
		Type:       "trace-update",
		EndTime:    &endTime,
	}
	
	ingestionEvent := updateEvent.ToIngestionEvent()
	
	if err := tb.client.queue.Enqueue(ingestionEvent); err != nil {
		return err
	}
	
	tb.submitted = true
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}