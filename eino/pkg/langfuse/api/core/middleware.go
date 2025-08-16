package core

import (
	"fmt"

	"github.com/go-resty/resty/v2"

	"eino/pkg/langfuse/config"
)

// createRetryCondition creates a retry condition function based on config
func createRetryCondition(cfg *config.Config) resty.RetryConditionFunc {
	return func(r *resty.Response, err error) bool {
		if err != nil {
			return true // Retry on network errors
		}

		// Retry on specific status codes
		statusCode := r.StatusCode()
		return statusCode >= 500 || statusCode == 429 || statusCode == 408
	}
}

// parseHTTPError parses HTTP error responses
func parseHTTPError(resp *resty.Response) error {
	statusCode := resp.StatusCode()
	body := string(resp.Body())
	
	switch statusCode {
	case 400:
		return fmt.Errorf("bad request: %s", body)
	case 401:
		return fmt.Errorf("unauthorized: %s", body)
	case 403:
		return fmt.Errorf("forbidden: %s", body)
	case 404:
		return fmt.Errorf("not found: %s", body)
	case 429:
		return fmt.Errorf("too many requests: %s", body)
	case 500:
		return fmt.Errorf("internal server error: %s", body)
	case 502:
		return fmt.Errorf("bad gateway: %s", body)
	case 503:
		return fmt.Errorf("service unavailable: %s", body)
	case 504:
		return fmt.Errorf("gateway timeout: %s", body)
	default:
		return fmt.Errorf("HTTP %d error: %s", statusCode, body)
	}
}

// createErrorHandler creates an error handling middleware
func createErrorHandler() resty.ResponseMiddleware {
	return func(c *resty.Client, r *resty.Response) error {
		if r.StatusCode() >= 400 {
			return parseHTTPError(r)
		}
		return nil
	}
}

