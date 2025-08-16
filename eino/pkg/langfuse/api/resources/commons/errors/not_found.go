package errors

import "fmt"

// NotFoundError represents a resource not found error from the Langfuse API
type NotFoundError struct {
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	ResourceID string `json:"resourceId,omitempty"`
	Resource   string `json:"resource,omitempty"`
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	if e.ResourceID != "" && e.Resource != "" {
		return fmt.Sprintf("%s not found: %s (ID: %s)", e.Resource, e.Message, e.ResourceID)
	}
	if e.Resource != "" {
		return fmt.Sprintf("%s not found: %s", e.Resource, e.Message)
	}
	if e.Code != "" {
		return fmt.Sprintf("not found [%s]: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("not found: %s", e.Message)
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		Message: message,
	}
}

// NewNotFoundErrorWithCode creates a new NotFoundError with a specific error code
func NewNotFoundErrorWithCode(message, code string) *NotFoundError {
	return &NotFoundError{
		Message: message,
		Code:    code,
	}
}

// NewResourceNotFoundError creates a new NotFoundError for a specific resource
func NewResourceNotFoundError(resource, resourceID, message string) *NotFoundError {
	return &NotFoundError{
		Message:    message,
		ResourceID: resourceID,
		Resource:   resource,
	}
}

// NewTraceNotFoundError creates a new NotFoundError for a trace
func NewTraceNotFoundError(traceID string) *NotFoundError {
	return NewResourceNotFoundError("trace", traceID, "trace does not exist")
}

// NewObservationNotFoundError creates a new NotFoundError for an observation
func NewObservationNotFoundError(observationID string) *NotFoundError {
	return NewResourceNotFoundError("observation", observationID, "observation does not exist")
}

// NewSessionNotFoundError creates a new NotFoundError for a session
func NewSessionNotFoundError(sessionID string) *NotFoundError {
	return NewResourceNotFoundError("session", sessionID, "session does not exist")
}
