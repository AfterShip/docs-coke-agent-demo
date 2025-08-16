package types

import (
	"encoding/json"
	"time"
)

// IngestionEvent represents a generic event in the Langfuse ingestion system
type IngestionEvent struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Body      interface{} `json:"body"`
}

// EventType represents the type of ingestion event
type EventType string

const (
	// EventTypeTraceCreate represents a trace creation event
	EventTypeTraceCreate EventType = "trace-create"
	
	// EventTypeTraceUpdate represents a trace update event
	EventTypeTraceUpdate EventType = "trace-update"
	
	// EventTypeObservationCreate represents an observation creation event
	EventTypeObservationCreate EventType = "observation-create"
	
	// EventTypeObservationUpdate represents an observation update event
	EventTypeObservationUpdate EventType = "observation-update"
	
	// EventTypeSpanCreate represents a span creation event
	EventTypeSpanCreate EventType = "span-create"
	
	// EventTypeSpanUpdate represents a span update event
	EventTypeSpanUpdate EventType = "span-update"
	
	// EventTypeGenerationCreate represents a generation creation event
	EventTypeGenerationCreate EventType = "generation-create"
	
	// EventTypeGenerationUpdate represents a generation update event
	EventTypeGenerationUpdate EventType = "generation-update"
	
	// EventTypeEventCreate represents an event creation
	EventTypeEventCreate EventType = "event-create"
	
	// EventTypeScoreCreate represents a score creation event
	EventTypeScoreCreate EventType = "score-create"
	
	// EventTypeSDKLog represents an SDK log event
	EventTypeSDKLog EventType = "sdk-log"
)

// MarshalJSON implements json.Marshaler for IngestionEvent
func (e *IngestionEvent) MarshalJSON() ([]byte, error) {
	type Alias IngestionEvent
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(e),
		Timestamp: e.Timestamp.UTC().Format(time.RFC3339Nano),
	})
}

// UnmarshalJSON implements json.Unmarshaler for IngestionEvent
func (e *IngestionEvent) UnmarshalJSON(data []byte) error {
	type Alias IngestionEvent
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias: (*Alias)(e),
	}
	
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	
	var err error
	e.Timestamp, err = time.Parse(time.RFC3339Nano, aux.Timestamp)
	return err
}

// Validate performs basic validation on the ingestion event
func (e *IngestionEvent) Validate() error {
	if e.ID == "" {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	
	if e.Type == "" {
		return &ValidationError{Field: "type", Message: "type is required"}
	}
	
	if e.Timestamp.IsZero() {
		return &ValidationError{Field: "timestamp", Message: "timestamp is required"}
	}
	
	if e.Body == nil {
		return &ValidationError{Field: "body", Message: "body is required"}
	}
	
	return nil
}

// ValidationError represents a validation error for ingestion events
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}