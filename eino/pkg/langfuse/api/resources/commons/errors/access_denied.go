package errors

import "fmt"

// AccessDeniedError represents an access denied error from the Langfuse API
type AccessDeniedError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Error implements the error interface
func (e *AccessDeniedError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("access denied [%s]: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("access denied: %s", e.Message)
}

// NewAccessDeniedError creates a new AccessDeniedError
func NewAccessDeniedError(message string) *AccessDeniedError {
	return &AccessDeniedError{
		Message: message,
	}
}

// NewAccessDeniedErrorWithCode creates a new AccessDeniedError with a specific error code
func NewAccessDeniedErrorWithCode(message, code string) *AccessDeniedError {
	return &AccessDeniedError{
		Message: message,
		Code:    code,
	}
}
