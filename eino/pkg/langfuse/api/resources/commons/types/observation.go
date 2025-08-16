package types

import (
	"encoding/json"
	"time"
)

// ObservationType represents the type of observation
type ObservationType string

const (
	// ObservationTypeSpan represents a general span/event
	ObservationTypeSpan ObservationType = "SPAN"

	// ObservationTypeGeneration represents an LLM generation
	ObservationTypeGeneration ObservationType = "GENERATION"

	// ObservationTypeEvent represents a discrete event
	ObservationTypeEvent ObservationType = "EVENT"
)

// ObservationLevel represents the severity/importance level of an observation
type ObservationLevel string

const (
	// ObservationLevelDebug for debug-level observations
	ObservationLevelDebug ObservationLevel = "DEBUG"

	// ObservationLevelDefault for normal observations
	ObservationLevelDefault ObservationLevel = "DEFAULT"

	// ObservationLevelWarning for warning-level observations
	ObservationLevelWarning ObservationLevel = "WARNING"

	// ObservationLevelError for error-level observations
	ObservationLevelError ObservationLevel = "ERROR"
)

// Observation represents a single observation within a trace
type Observation struct {
	// Unique identifier for the observation
	ID string `json:"id"`

	// ID of the trace this observation belongs to
	TraceID string `json:"traceId"`

	// Type of observation (SPAN, GENERATION, EVENT)
	Type ObservationType `json:"type"`

	// Optional external ID for linking with external systems
	ExternalID *string `json:"externalId,omitempty"`

	// Name/title of the observation
	Name *string `json:"name,omitempty"`

	// Start timestamp of the observation
	StartTime time.Time `json:"startTime"`

	// End timestamp of the observation (optional, for ongoing operations)
	EndTime *time.Time `json:"endTime,omitempty"`

	// Completion timestamp
	CompletionStartTime *time.Time `json:"completionStartTime,omitempty"`

	// Model name (for generations)
	Model *string `json:"model,omitempty"`

	// Model parameters (for generations)
	ModelParameters map[string]interface{} `json:"modelParameters,omitempty"`

	// Input data for the observation
	Input json.RawMessage `json:"input,omitempty"`

	// Output data from the observation
	Output json.RawMessage `json:"output,omitempty"`

	// Usage statistics (tokens, costs, etc.)
	Usage *Usage `json:"usage,omitempty"`

	// Metadata associated with the observation
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Parent observation ID for nested observations
	ParentObservationID *string `json:"parentObservationId,omitempty"`

	// Level/severity of the observation
	Level *ObservationLevel `json:"level,omitempty"`

	// Status message
	StatusMessage *string `json:"statusMessage,omitempty"`

	// Version of the observation
	Version *string `json:"version,omitempty"`
}

// ObservationCreateRequest represents a request to create a new observation
type ObservationCreateRequest struct {
	// Unique identifier for the observation
	ID *string `json:"id,omitempty"`

	// ID of the trace this observation belongs to
	TraceID string `json:"traceId"`

	// Type of observation (SPAN, GENERATION, EVENT)
	Type ObservationType `json:"type"`

	// Optional external ID for linking with external systems
	ExternalID *string `json:"externalId,omitempty"`

	// Name/title of the observation
	Name *string `json:"name,omitempty"`

	// Start timestamp of the observation
	StartTime *time.Time `json:"startTime,omitempty"`

	// End timestamp of the observation
	EndTime *time.Time `json:"endTime,omitempty"`

	// Completion timestamp
	CompletionStartTime *time.Time `json:"completionStartTime,omitempty"`

	// Model name (for generations)
	Model *string `json:"model,omitempty"`

	// Model parameters (for generations)
	ModelParameters map[string]interface{} `json:"modelParameters,omitempty"`

	// Input data for the observation
	Input interface{} `json:"input,omitempty"`

	// Output data from the observation
	Output interface{} `json:"output,omitempty"`

	// Usage statistics (tokens, costs, etc.)
	Usage *Usage `json:"usage,omitempty"`

	// Metadata associated with the observation
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Parent observation ID for nested observations
	ParentObservationID *string `json:"parentObservationId,omitempty"`

	// Level/severity of the observation
	Level *ObservationLevel `json:"level,omitempty"`

	// Status message
	StatusMessage *string `json:"statusMessage,omitempty"`

	// Version of the observation
	Version *string `json:"version,omitempty"`
}

// ObservationUpdateRequest represents a request to update an existing observation
type ObservationUpdateRequest struct {
	// Name/title of the observation
	Name *string `json:"name,omitempty"`

	// End timestamp of the observation
	EndTime *time.Time `json:"endTime,omitempty"`

	// Completion timestamp
	CompletionStartTime *time.Time `json:"completionStartTime,omitempty"`

	// Model name (for generations)
	Model *string `json:"model,omitempty"`

	// Model parameters (for generations)
	ModelParameters map[string]interface{} `json:"modelParameters,omitempty"`

	// Input data for the observation
	Input interface{} `json:"input,omitempty"`

	// Output data from the observation
	Output interface{} `json:"output,omitempty"`

	// Usage statistics (tokens, costs, etc.)
	Usage *Usage `json:"usage,omitempty"`

	// Metadata associated with the observation
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Level/severity of the observation
	Level *ObservationLevel `json:"level,omitempty"`

	// Status message
	StatusMessage *string `json:"statusMessage,omitempty"`

	// Version of the observation
	Version *string `json:"version,omitempty"`
}
