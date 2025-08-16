package types

import "time"

// DeleteTraceResponse represents the response from deleting a trace
type DeleteTraceResponse struct {
	Success   bool      `json:"success"`
	TraceID   string    `json:"traceId"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewDeleteTraceResponse creates a new delete trace response
func NewDeleteTraceResponse(traceID string, success bool) *DeleteTraceResponse {
	return &DeleteTraceResponse{
		Success:   success,
		TraceID:   traceID,
		Timestamp: time.Now().UTC(),
	}
}

// WithMessage adds a message to the delete response
func (r *DeleteTraceResponse) WithMessage(message string) *DeleteTraceResponse {
	r.Message = message
	return r
}

// IsSuccess returns true if the deletion was successful
func (r *DeleteTraceResponse) IsSuccess() bool {
	return r.Success
}

// GetTraceID returns the ID of the deleted trace
func (r *DeleteTraceResponse) GetTraceID() string {
	return r.TraceID
}