// Package types defines the core data structures used throughout the Langfuse SDK.
//
// These types represent the fundamental concepts in Langfuse observability:
//   - Traces: Complete execution flows
//   - Observations: Individual operations (spans, generations, events)
//   - Scores: Evaluation metrics
//   - Usage: Token and cost tracking
//   - Sessions: Grouped conversations or workflows
//   - Datasets: Collections of test cases
//
// All types are designed to be JSON-serializable for API communication
// and include proper validation and constraints.
package types

import (
	"encoding/json"
	"time"
)

// Trace represents a complete execution flow in Langfuse.
//
// Traces are the top-level container for observability data, typically corresponding
// to user requests, API calls, or complete workflows. Each trace can contain multiple
// observations (spans, generations, events) that represent individual operations
// within the overall execution.
//
// Traces enable you to:
//   - Group related operations for analysis
//   - Track end-to-end performance and behavior
//   - Associate operations with users and sessions
//   - Add metadata and tags for filtering and search
//   - Measure costs and resource usage across workflows
//
// Example trace structure:
//   Trace: "user-authentication"
//   ├── Span: "database-lookup" 
//   ├── Generation: "welcome-message"
//   └── Event: "login-success"
type Trace struct {
	// ID is the unique identifier for the trace.
	// Must be unique within the Langfuse project.
	ID string `json:"id"`

	// ExternalID is an optional external identifier for linking with external systems.
	// Useful for correlating traces with logs, metrics, or other monitoring systems.
	ExternalID *string `json:"externalId,omitempty"`

	// Timestamp indicates when the trace was created or started.
	// Should be in UTC for consistency across distributed systems.
	Timestamp time.Time `json:"timestamp"`

	// Name is a human-readable identifier for the trace.
	// Should be descriptive and consistent across similar operations.
	// Examples: "user-authentication", "document-processing", "api-/users/{id}"
	Name *string `json:"name,omitempty"`

	// Input contains the input data or parameters for the traced operation.
	// Can be any JSON-serializable data structure.
	Input json.RawMessage `json:"input,omitempty"`

	// Output contains the output data or results from the traced operation.
	// Can be any JSON-serializable data structure.
	Output json.RawMessage `json:"output,omitempty"`

	// SessionID groups traces that belong to the same user session or conversation.
	// Enables analysis of multi-turn interactions and conversation flows.
	SessionID *string `json:"sessionId,omitempty"`

	// User ID associated with this trace
	UserID *string `json:"userId,omitempty"`

	// Metadata associated with the trace
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Tags associated with the trace
	Tags []string `json:"tags,omitempty"`

	// Version of the trace
	Version *string `json:"version,omitempty"`

	// Release version
	Release *string `json:"release,omitempty"`

	// Whether the trace is public
	Public *bool `json:"public,omitempty"`
}

// TraceCreateRequest represents a request to create a new trace
type TraceCreateRequest struct {
	// Unique identifier for the trace
	ID *string `json:"id,omitempty"`

	// Optional external ID for linking with external systems
	ExternalID *string `json:"externalId,omitempty"`

	// Timestamp when the trace was created
	Timestamp *time.Time `json:"timestamp,omitempty"`

	// Name/title of the trace
	Name *string `json:"name,omitempty"`

	// Input data for the trace
	Input interface{} `json:"input,omitempty"`

	// Output data from the trace
	Output interface{} `json:"output,omitempty"`

	// Session ID this trace belongs to
	SessionID *string `json:"sessionId,omitempty"`

	// User ID associated with this trace
	UserID *string `json:"userId,omitempty"`

	// Metadata associated with the trace
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Tags associated with the trace
	Tags []string `json:"tags,omitempty"`

	// Version of the trace
	Version *string `json:"version,omitempty"`

	// Release version
	Release *string `json:"release,omitempty"`

	// Whether the trace is public
	Public *bool `json:"public,omitempty"`
}

// TraceUpdateRequest represents a request to update an existing trace
type TraceUpdateRequest struct {
	// Name/title of the trace
	Name *string `json:"name,omitempty"`

	// Input data for the trace
	Input interface{} `json:"input,omitempty"`

	// Output data from the trace
	Output interface{} `json:"output,omitempty"`

	// Session ID this trace belongs to
	SessionID *string `json:"sessionId,omitempty"`

	// User ID associated with this trace
	UserID *string `json:"userId,omitempty"`

	// Metadata associated with the trace
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Tags associated with the trace
	Tags []string `json:"tags,omitempty"`

	// Version of the trace
	Version *string `json:"version,omitempty"`

	// Release version
	Release *string `json:"release,omitempty"`

	// Whether the trace is public
	Public *bool `json:"public,omitempty"`
}

// TraceListResponse represents the response from listing traces
type TraceListResponse struct {
	Data []Trace `json:"data"`
	Meta struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		TotalItems int `json:"totalItems"`
		TotalPages int `json:"totalPages"`
	} `json:"meta"`
}
