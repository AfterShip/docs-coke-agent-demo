package core

import (
	"fmt"

	"github.com/go-resty/resty/v2"

	"eino/pkg/langfuse/config"
)

// ConfigureRestyClient configures a resty client with Langfuse-specific settings
func ConfigureRestyClient(client *resty.Client, cfg *config.Config) error {
	if client == nil {
		return fmt.Errorf("client cannot be nil")
	}
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Basic configuration
	client.
		SetBaseURL(cfg.Host).
		SetTimeout(cfg.Timeout).
		SetHeader("User-Agent", cfg.HTTPUserAgent).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	// Authentication
	if cfg.PublicKey != "" && cfg.SecretKey != "" {
		client.SetBasicAuth(cfg.PublicKey, cfg.SecretKey)
	}

	// Retry configuration using resty's built-in retry
	if cfg.RetryCount > 0 {
		client.
			SetRetryCount(cfg.RetryCount).
			SetRetryWaitTime(cfg.RetryDelay).
			SetRetryMaxWaitTime(cfg.MaxRetryDelay).
			AddRetryCondition(createRetryCondition(cfg))
	}

	// Debug mode
	if cfg.Debug {
		client.SetDebug(true)
	}

	// Error handling middleware
	client.OnAfterResponse(createErrorHandler())

	return nil
}