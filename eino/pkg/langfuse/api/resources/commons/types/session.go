package types

import (
	"time"
)

// Session represents a session for grouping related traces
type Session struct {
	// Unique identifier for the session
	ID string `json:"id"`

	// Timestamp when the session was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the session was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// Project ID this session belongs to
	ProjectID string `json:"projectId"`

	// Optional external ID for linking with external systems
	ExternalID *string `json:"externalId,omitempty"`

	// User ID associated with this session
	UserID *string `json:"userId,omitempty"`

	// Metadata associated with the session
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Whether the session is public
	Public *bool `json:"public,omitempty"`

	// Number of traces in this session
	TraceCount *int `json:"traceCount,omitempty"`

	// Aggregated usage across all traces in the session
	Usage *SessionUsage `json:"usage,omitempty"`
}

// SessionUsage represents aggregated usage statistics for a session
type SessionUsage struct {
	// Total input tokens across all traces
	InputTokens *int `json:"inputTokens,omitempty"`

	// Total output tokens across all traces
	OutputTokens *int `json:"outputTokens,omitempty"`

	// Total tokens across all traces
	TotalTokens *int `json:"totalTokens,omitempty"`

	// Total input cost across all traces
	InputCost *float64 `json:"inputCost,omitempty"`

	// Total output cost across all traces
	OutputCost *float64 `json:"outputCost,omitempty"`

	// Total cost across all traces
	TotalCost *float64 `json:"totalCost,omitempty"`
}

// SessionCreateRequest represents a request to create a new session
type SessionCreateRequest struct {
	// Unique identifier for the session
	ID *string `json:"id,omitempty"`

	// Optional external ID for linking with external systems
	ExternalID *string `json:"externalId,omitempty"`

	// User ID associated with this session
	UserID *string `json:"userId,omitempty"`

	// Metadata associated with the session
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Whether the session is public
	Public *bool `json:"public,omitempty"`
}

// SessionUpdateRequest represents a request to update an existing session
type SessionUpdateRequest struct {
	// User ID associated with this session
	UserID *string `json:"userId,omitempty"`

	// Metadata associated with the session
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Whether the session is public
	Public *bool `json:"public,omitempty"`
}

// SessionListResponse represents the response from listing sessions
type SessionListResponse struct {
	Data []Session `json:"data"`
	Meta struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		TotalItems int `json:"totalItems"`
		TotalPages int `json:"totalPages"`
	} `json:"meta"`
}

// SessionStats represents statistical information about a session
type SessionStats struct {
	// Session ID
	SessionID string `json:"sessionId"`

	// Number of traces in the session
	TraceCount int `json:"traceCount"`

	// Number of observations in the session
	ObservationCount int `json:"observationCount"`

	// Total duration of all traces in the session (in milliseconds)
	TotalDuration *int64 `json:"totalDuration,omitempty"`

	// Average duration of traces in the session (in milliseconds)
	AverageDuration *float64 `json:"averageDuration,omitempty"`

	// Aggregated usage statistics
	Usage *SessionUsage `json:"usage,omitempty"`
}
