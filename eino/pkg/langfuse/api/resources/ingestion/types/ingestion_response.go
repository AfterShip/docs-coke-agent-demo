package types

import (
	"encoding/json"
	"time"
)

// IngestionResponse represents the response from the Langfuse ingestion API
type IngestionResponse struct {
	Success   bool                   `json:"success"`
	Errors    []IngestionError       `json:"errors,omitempty"`
	Usage     *IngestionUsage        `json:"usage,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// IngestionError represents an error that occurred during ingestion
type IngestionError struct {
	ID        string            `json:"id"`
	Status    int               `json:"status"`
	Message   string            `json:"message"`
	ErrorText string            `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	EventID   *string           `json:"event_id,omitempty"`
	EventType *EventType        `json:"event_type,omitempty"`
}

// IngestionUsage represents usage metrics from the ingestion response
type IngestionUsage struct {
	TokensIngested     *int     `json:"tokens_ingested,omitempty"`
	EventsProcessed    int      `json:"events_processed"`
	EventsSkipped      int      `json:"events_skipped,omitempty"`
	EventsFailed       int      `json:"events_failed,omitempty"`
	ProcessingTime     *float64 `json:"processing_time,omitempty"`
	EstimatedCost      *float64 `json:"estimated_cost,omitempty"`
	RateLimitRemaining *int     `json:"rate_limit_remaining,omitempty"`
}

// NewIngestionResponse creates a new ingestion response
func NewIngestionResponse(success bool) *IngestionResponse {
	return &IngestionResponse{
		Success:   success,
		Errors:    make([]IngestionError, 0),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now().UTC(),
	}
}

// AddError adds an error to the response
func (r *IngestionResponse) AddError(err IngestionError) {
	r.Errors = append(r.Errors, err)
	r.Success = false
}

// HasErrors returns true if the response contains errors
func (r *IngestionResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorCount returns the number of errors in the response
func (r *IngestionResponse) ErrorCount() int {
	return len(r.Errors)
}

// IsPartialSuccess returns true if some events were processed successfully but some failed
func (r *IngestionResponse) IsPartialSuccess() bool {
	if r.Usage == nil {
		return false
	}
	return r.Usage.EventsProcessed > 0 && r.Usage.EventsFailed > 0
}

// IsCompleteFailure returns true if no events were processed successfully
func (r *IngestionResponse) IsCompleteFailure() bool {
	if r.Usage == nil {
		return !r.Success
	}
	return r.Usage.EventsProcessed == 0 && r.Usage.EventsFailed > 0
}

// MarshalJSON implements json.Marshaler for IngestionResponse
func (r *IngestionResponse) MarshalJSON() ([]byte, error) {
	type Alias IngestionResponse
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(r),
		Timestamp: r.Timestamp.UTC().Format(time.RFC3339Nano),
	})
}

// UnmarshalJSON implements json.Unmarshaler for IngestionResponse
func (r *IngestionResponse) UnmarshalJSON(data []byte) error {
	type Alias IngestionResponse
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias: (*Alias)(r),
	}
	
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	
	var err error
	r.Timestamp, err = time.Parse(time.RFC3339Nano, aux.Timestamp)
	return err
}

// NewIngestionError creates a new ingestion error
func NewIngestionError(id string, status int, message string) *IngestionError {
	return &IngestionError{
		ID:      id,
		Status:  status,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithEventID adds event ID to the error
func (e *IngestionError) WithEventID(eventID string) *IngestionError {
	e.EventID = &eventID
	return e
}

// WithEventType adds event type to the error
func (e *IngestionError) WithEventType(eventType EventType) *IngestionError {
	e.EventType = &eventType
	return e
}

// WithError adds error string to the error
func (e *IngestionError) WithError(errorStr string) *IngestionError {
	e.ErrorText = errorStr
	return e
}

// WithDetail adds a detail to the error
func (e *IngestionError) WithDetail(key string, value interface{}) *IngestionError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Error implements the error interface for IngestionError
func (e *IngestionError) Error() string {
	if e.ErrorText != "" {
		return e.ErrorText
	}
	return e.Message
}

// NewIngestionUsage creates a new ingestion usage struct
func NewIngestionUsage() *IngestionUsage {
	return &IngestionUsage{
		EventsProcessed: 0,
		EventsSkipped:   0,
		EventsFailed:    0,
	}
}

// AddProcessedEvent increments the processed events counter
func (u *IngestionUsage) AddProcessedEvent() {
	u.EventsProcessed++
}

// AddSkippedEvent increments the skipped events counter
func (u *IngestionUsage) AddSkippedEvent() {
	u.EventsSkipped++
}

// AddFailedEvent increments the failed events counter
func (u *IngestionUsage) AddFailedEvent() {
	u.EventsFailed++
}

// TotalEvents returns the total number of events processed
func (u *IngestionUsage) TotalEvents() int {
	return u.EventsProcessed + u.EventsSkipped + u.EventsFailed
}

// SuccessRate returns the success rate as a percentage
func (u *IngestionUsage) SuccessRate() float64 {
	total := u.TotalEvents()
	if total == 0 {
		return 0.0
	}
	return (float64(u.EventsProcessed) / float64(total)) * 100.0
}

// Common HTTP status codes for ingestion errors
const (
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden          = 403
	StatusNotFound           = 404
	StatusUnprocessableEntity = 422
	StatusTooManyRequests    = 429
	StatusInternalServerError = 500
	StatusServiceUnavailable = 503
)