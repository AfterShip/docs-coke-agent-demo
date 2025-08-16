package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"

	"eino/pkg/langfuse/internal/utils"
)

// RetryConfig configures retry behavior for HTTP requests
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts (0 = no retries, 1 = original + 1 retry)
	MaxAttempts int
	
	// BaseDelay is the base delay for exponential backoff
	BaseDelay time.Duration
	
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	
	// Multiplier for exponential backoff (typically 2.0)
	Multiplier float64
	
	// Jitter adds randomness to backoff delays to avoid thundering herd
	Jitter bool
	
	// MaxJitter is the maximum jitter factor (0.0-1.0)
	MaxJitter float64
	
	// RetryableHTTPCodes defines which HTTP status codes should trigger retries
	RetryableHTTPCodes []int
	
	// RetryableErrors defines error types that should trigger retries
	RetryableErrors []string
	
	// NonRetryableErrors defines error types that should NOT trigger retries
	NonRetryableErrors []string
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:        3,
		BaseDelay:          1 * time.Second,
		MaxDelay:           30 * time.Second,
		Multiplier:         2.0,
		Jitter:             true,
		MaxJitter:          0.3,
		RetryableHTTPCodes: []int{429, 500, 502, 503, 504},
		RetryableErrors: []string{
			"network_error",
			"timeout",
			"connection_refused",
			"connection_reset",
			"dns_error",
			"temporary_failure",
		},
		NonRetryableErrors: []string{
			"authentication_failed",
			"access_denied",
			"not_found",
			"bad_request",
			"validation_error",
		},
	}
}

// RetryDecision represents whether a request should be retried
type RetryDecision struct {
	ShouldRetry bool
	Delay       time.Duration
	Reason      string
}

// ErrorClassifier classifies errors for retry decisions
type ErrorClassifier struct {
	config *RetryConfig
}

// NewErrorClassifier creates a new error classifier
func NewErrorClassifier(config *RetryConfig) *ErrorClassifier {
	return &ErrorClassifier{config: config}
}

// ShouldRetry determines if an error should trigger a retry
func (ec *ErrorClassifier) ShouldRetry(resp *resty.Response, err error, attempt int) *RetryDecision {
	// Check if we've exceeded max attempts
	if attempt >= ec.config.MaxAttempts {
		return &RetryDecision{
			ShouldRetry: false,
			Reason:      fmt.Sprintf("exceeded max attempts (%d)", ec.config.MaxAttempts),
		}
	}

	// Check HTTP response codes
	if resp != nil {
		statusCode := resp.StatusCode()
		
		// Check if status code is explicitly retryable
		for _, code := range ec.config.RetryableHTTPCodes {
			if statusCode == code {
				delay := ec.calculateBackoff(attempt)
				return &RetryDecision{
					ShouldRetry: true,
					Delay:       delay,
					Reason:      fmt.Sprintf("retryable HTTP status %d", statusCode),
				}
			}
		}
		
		// 4xx errors are generally not retryable (except 429)
		if statusCode >= 400 && statusCode < 500 && statusCode != 429 {
			return &RetryDecision{
				ShouldRetry: false,
				Reason:      fmt.Sprintf("non-retryable HTTP status %d", statusCode),
			}
		}
	}

	// Check if error is nil (successful response)
	if err == nil {
		return &RetryDecision{
			ShouldRetry: false,
			Reason:      "no error occurred",
		}
	}

	// Classify the error
	errorType := ec.classifyError(err)
	
	// Check non-retryable errors first
	for _, nonRetryable := range ec.config.NonRetryableErrors {
		if strings.Contains(errorType, nonRetryable) {
			return &RetryDecision{
				ShouldRetry: false,
				Reason:      fmt.Sprintf("non-retryable error type: %s", errorType),
			}
		}
	}
	
	// Check retryable errors
	for _, retryable := range ec.config.RetryableErrors {
		if strings.Contains(errorType, retryable) {
			delay := ec.calculateBackoff(attempt)
			return &RetryDecision{
				ShouldRetry: true,
				Delay:       delay,
				Reason:      fmt.Sprintf("retryable error type: %s", errorType),
			}
		}
	}

	// Default to retry for network-level errors
	if ec.isNetworkError(err) {
		delay := ec.calculateBackoff(attempt)
		return &RetryDecision{
			ShouldRetry: true,
			Delay:       delay,
			Reason:      "network-level error",
		}
	}

	// Default to no retry for unknown errors
	return &RetryDecision{
		ShouldRetry: false,
		Reason:      fmt.Sprintf("unknown error type: %T", err),
	}
}

// classifyError determines the type of error for retry decisions
func (ec *ErrorClassifier) classifyError(err error) string {
	if err == nil {
		return "no_error"
	}

	// Check for context cancellation/timeout
	if errors.Is(err, context.Canceled) {
		return "request_canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return "timeout"
		}
		if netErr.Temporary() {
			return "temporary_failure"
		}
		return "network_error"
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns_error"
	}

	// Check for URL errors
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return ec.classifyError(urlErr.Err)
	}

	// Check for system call errors
	var syscallErr *net.OpError
	if errors.As(err, &syscallErr) {
		if errors.Is(syscallErr.Err, syscall.ECONNREFUSED) {
			return "connection_refused"
		}
		if errors.Is(syscallErr.Err, syscall.ECONNRESET) {
			return "connection_reset"
		}
		return "network_error"
	}

	// Check for SDK-specific errors
	if langfuseErr, ok := err.(*utils.LangfuseError); ok {
		switch langfuseErr.Type {
		case utils.ErrorTypeNetwork:
			return "network_error"
		case utils.ErrorTypeTimeout:
			return "timeout"
		case utils.ErrorTypeAuth:
			return "authentication_failed"
		case utils.ErrorTypeValidation:
			return "validation_error"
		case utils.ErrorTypeRateLimit:
			return "rate_limited"
		case utils.ErrorTypeServerError:
			return "server_error"
		default:
			return "sdk_error"
		}
	}

	// Check error message for common patterns
	errMsg := strings.ToLower(err.Error())
	
	if strings.Contains(errMsg, "timeout") {
		return "timeout"
	}
	if strings.Contains(errMsg, "connection refused") {
		return "connection_refused"
	}
	if strings.Contains(errMsg, "connection reset") {
		return "connection_reset"
	}
	if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "dns") {
		return "dns_error"
	}
	if strings.Contains(errMsg, "network") {
		return "network_error"
	}

	return "unknown_error"
}

// isNetworkError checks if the error is a network-level error that should generally be retried
func (ec *ErrorClassifier) isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	// URL errors (often wrap network errors)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return ec.isNetworkError(urlErr.Err)
	}

	// OpError (network operations)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	return false
}

// calculateBackoff calculates the delay for the given attempt using exponential backoff
func (ec *ErrorClassifier) calculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Calculate exponential backoff
	delay := float64(ec.config.BaseDelay) * math.Pow(ec.config.Multiplier, float64(attempt-1))
	
	// Apply jitter if enabled
	if ec.config.Jitter && ec.config.MaxJitter > 0 {
		jitter := rand.Float64() * ec.config.MaxJitter
		delay = delay * (1 + jitter)
	}
	
	// Ensure we don't exceed max delay
	if maxDelay := float64(ec.config.MaxDelay); delay > maxDelay {
		delay = maxDelay
	}
	
	return time.Duration(delay)
}

// WithRetryConfig updates the error classifier with new retry configuration
func (ec *ErrorClassifier) WithRetryConfig(config *RetryConfig) *ErrorClassifier {
	return &ErrorClassifier{config: config}
}

// RetryableErrorTypes returns the list of retryable error types
func (ec *ErrorClassifier) RetryableErrorTypes() []string {
	return ec.config.RetryableErrors
}

// NonRetryableErrorTypes returns the list of non-retryable error types
func (ec *ErrorClassifier) NonRetryableErrorTypes() []string {
	return ec.config.NonRetryableErrors
}

// RetryableHTTPCodes returns the list of retryable HTTP status codes
func (ec *ErrorClassifier) RetryableHTTPCodes() []int {
	return ec.config.RetryableHTTPCodes
}