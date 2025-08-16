package errors

import "fmt"

// UnauthorizedError represents an authentication/authorization error from the Langfuse API
type UnauthorizedError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// Error implements the error interface
func (e *UnauthorizedError) Error() string {
	if e.Reason != "" && e.Code != "" {
		return fmt.Sprintf("unauthorized [%s] %s: %s", e.Code, e.Reason, e.Message)
	}
	if e.Reason != "" {
		return fmt.Sprintf("unauthorized (%s): %s", e.Reason, e.Message)
	}
	if e.Code != "" {
		return fmt.Sprintf("unauthorized [%s]: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("unauthorized: %s", e.Message)
}

// NewUnauthorizedError creates a new UnauthorizedError
func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		Message: message,
	}
}

// NewUnauthorizedErrorWithCode creates a new UnauthorizedError with a specific error code
func NewUnauthorizedErrorWithCode(message, code string) *UnauthorizedError {
	return &UnauthorizedError{
		Message: message,
		Code:    code,
	}
}

// NewUnauthorizedErrorWithReason creates a new UnauthorizedError with a specific reason
func NewUnauthorizedErrorWithReason(message, reason string) *UnauthorizedError {
	return &UnauthorizedError{
		Message: message,
		Reason:  reason,
	}
}

// NewInvalidCredentialsError creates a new UnauthorizedError for invalid credentials
func NewInvalidCredentialsError() *UnauthorizedError {
	return NewUnauthorizedErrorWithReason(
		"authentication failed with provided credentials",
		"invalid_credentials",
	)
}

// NewExpiredTokenError creates a new UnauthorizedError for expired tokens
func NewExpiredTokenError() *UnauthorizedError {
	return NewUnauthorizedErrorWithReason(
		"authentication token has expired",
		"token_expired",
	)
}

// NewInsufficientPermissionsError creates a new UnauthorizedError for insufficient permissions
func NewInsufficientPermissionsError(operation string) *UnauthorizedError {
	return NewUnauthorizedErrorWithReason(
		fmt.Sprintf("insufficient permissions to perform operation: %s", operation),
		"insufficient_permissions",
	)
}
