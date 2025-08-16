package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

// LangfuseError represents a Langfuse-specific error with categorization
type LangfuseError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Cause   error     `json:"-"`
}

// ErrorType categorizes different types of errors
type ErrorType string

const (
	ErrorTypeValidation  ErrorType = "VALIDATION"
	ErrorTypeNetwork     ErrorType = "NETWORK"
	ErrorTypeAuth        ErrorType = "AUTH"
	ErrorTypeRateLimit   ErrorType = "RATE_LIMIT"
	ErrorTypeServerError ErrorType = "SERVER_ERROR"
	ErrorTypeTimeout     ErrorType = "TIMEOUT"
)

// Error implements the error interface
func (e *LangfuseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Is implements error checking
func (e *LangfuseError) Is(target error) bool {
	t, ok := target.(*LangfuseError)
	return ok && e.Type == t.Type
}

// Unwrap implements error unwrapping
func (e *LangfuseError) Unwrap() error {
	return e.Cause
}

// SDKError represents a base error type for the Langfuse SDK
type SDKError struct {
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	RequestID  string `json:"requestId,omitempty"`
	Cause      error  `json:"-"`
}

// Error implements the error interface
func (e *SDKError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("SDK error [%d]: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("SDK error: %s", e.Message)
}

// Unwrap implements the error unwrapping interface for Go 1.13+
func (e *SDKError) Unwrap() error {
	return e.Cause
}

// NewSDKError creates a new SDKError
func NewSDKError(message string) *SDKError {
	return &SDKError{
		Message: message,
	}
}

// NewSDKErrorWithCode creates a new SDKError with a specific error code
func NewSDKErrorWithCode(message, code string) *SDKError {
	return &SDKError{
		Message: message,
		Code:    code,
	}
}

// NewSDKErrorFromHTTP creates a new SDKError from an HTTP response
func NewSDKErrorFromHTTP(statusCode int, body []byte, requestID string) error {
	baseErr := &SDKError{
		StatusCode: statusCode,
		RequestID:  requestID,
	}

	// Try to parse the error response body
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		if message, ok := errorResponse["message"].(string); ok {
			baseErr.Message = message
		}
		if code, ok := errorResponse["code"].(string); ok {
			baseErr.Code = code
		}
	} else {
		baseErr.Message = string(body)
	}

	// Return specific error types based on status code
	switch statusCode {
	case http.StatusUnauthorized:
		return &commonErrors.UnauthorizedError{
			Message: baseErr.Message,
			Code:    baseErr.Code,
		}
	case http.StatusForbidden:
		return &commonErrors.AccessDeniedError{
			Message: baseErr.Message,
			Code:    baseErr.Code,
		}
	case http.StatusNotFound:
		return &commonErrors.NotFoundError{
			Message: baseErr.Message,
			Code:    baseErr.Code,
		}
	default:
		return baseErr
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation error for field '%s' with value '%s': %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewValidationErrorWithValue creates a new ValidationError with a specific value
func NewValidationErrorWithValue(field, message, value string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ConfigurationError represents configuration-related errors
type ConfigurationError struct {
	Parameter string `json:"parameter"`
	Message   string `json:"message"`
	Expected  string `json:"expected,omitempty"`
	Actual    string `json:"actual,omitempty"`
}

// Error implements the error interface
func (e *ConfigurationError) Error() string {
	if e.Expected != "" && e.Actual != "" {
		return fmt.Sprintf("configuration error for parameter '%s': expected %s, got %s. %s",
			e.Parameter, e.Expected, e.Actual, e.Message)
	}
	return fmt.Sprintf("configuration error for parameter '%s': %s", e.Parameter, e.Message)
}

// NewConfigurationError creates a new ConfigurationError
func NewConfigurationError(parameter, message string) *ConfigurationError {
	return &ConfigurationError{
		Parameter: parameter,
		Message:   message,
	}
}

// NewConfigurationErrorWithExpected creates a new ConfigurationError with expected/actual values
func NewConfigurationErrorWithExpected(parameter, message, expected, actual string) *ConfigurationError {
	return &ConfigurationError{
		Parameter: parameter,
		Message:   message,
		Expected:  expected,
		Actual:    actual,
	}
}

// NetworkError represents network-related errors
type NetworkError struct {
	Operation string `json:"operation"`
	URL       string `json:"url,omitempty"`
	Message   string `json:"message"`
	Cause     error  `json:"-"`
}

// Error implements the error interface
func (e *NetworkError) Error() string {
	if e.URL != "" {
		return fmt.Sprintf("network error during %s to %s: %s", e.Operation, e.URL, e.Message)
	}
	return fmt.Sprintf("network error during %s: %s", e.Operation, e.Message)
}

// Unwrap implements the error unwrapping interface
func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// NewNetworkError creates a new NetworkError
func NewNetworkError(operation, message string, cause error) *NetworkError {
	return &NetworkError{
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// NewNetworkErrorWithURL creates a new NetworkError with a specific URL
func NewNetworkErrorWithURL(operation, url, message string, cause error) *NetworkError {
	return &NetworkError{
		Operation: operation,
		URL:       url,
		Message:   message,
		Cause:     cause,
	}
}
