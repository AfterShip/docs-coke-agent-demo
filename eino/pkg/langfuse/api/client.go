package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"eino/pkg/langfuse/api/core"
	"eino/pkg/langfuse/api/resources/datasets"
	"eino/pkg/langfuse/api/resources/health"
	"eino/pkg/langfuse/api/resources/ingestion"
	"eino/pkg/langfuse/api/resources/models"
	"eino/pkg/langfuse/api/resources/projects"
	"eino/pkg/langfuse/api/resources/prompts"
	"eino/pkg/langfuse/api/resources/scores"
	"eino/pkg/langfuse/api/resources/sessions"
	"eino/pkg/langfuse/api/resources/traces"
	"eino/pkg/langfuse/config"
)

// APIClient provides access to all Langfuse API resources
type APIClient struct {
	// Core HTTP client for making requests
	client *resty.Client

	// Configuration
	config *config.Config

	// Resource clients
	Health    *health.Client
	Ingestion *ingestion.Client
	Traces    *traces.Client
	Scores    *scores.Client
	Sessions  *sessions.Client
	Models    *models.Client
	Datasets  *datasets.Client
	Projects  *projects.Client
	Prompts   *prompts.Client

	// State management
	mu     sync.RWMutex
	closed bool

	// Health monitoring
	lastHealthCheck time.Time
	isHealthy       bool
	healthCheckMu   sync.RWMutex
}

// NewAPIClient creates a new API client with all resource clients initialized
func NewAPIClient(config *config.Config) (*APIClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate required configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Create the resty client and configure it
	client := resty.New()
	
	if err := core.ConfigureRestyClient(client, config); err != nil {
		return nil, fmt.Errorf("failed to configure resty client: %w", err)
	}

	// HTTPClient wrapper removed - now using resty directly

	// Create the API client with all resource clients
	apiClient := &APIClient{
		client:    client,
		config:    config,
		Health:    health.NewClient(client),
		Ingestion: ingestion.NewClient(client),
		Traces:    traces.NewClient(client),
		Scores:    scores.NewClient(client),
		Sessions:  sessions.NewClient(client),
		Models:    models.NewClient(client),
		Datasets:  datasets.NewClient(client),
		Projects:  projects.NewClient(client),
		Prompts:   prompts.NewClient(client),
		closed:    false,
		isHealthy: false,
	}

	// Perform initial health check if enabled
	if !config.SkipInitialHealthCheck {
		if err := apiClient.performInitialHealthCheck(); err != nil {
			if config.RequireHealthyStart {
				return nil, fmt.Errorf("initial health check failed: %w", err)
			}
			// Log warning but continue if health check is not required
		}
	}

	return apiClient, nil
}

// validateConfig validates the API client configuration
func validateConfig(config *config.Config) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}

	if config.PublicKey == "" {
		return fmt.Errorf("public key is required")
	}

	if config.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}

	if config.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	if config.RetryCount < 0 {
		return fmt.Errorf("retry count cannot be negative")
	}

	if config.SampleRate < 0 || config.SampleRate > 1 {
		return fmt.Errorf("sample rate must be between 0 and 1")
	}

	return nil
}

// performInitialHealthCheck performs an initial health check on startup
func (c *APIClient) performInitialHealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.RequestTimeout)
	defer cancel()

	response, err := c.Health.Check(ctx)
	if err != nil {
		return err
	}

	c.healthCheckMu.Lock()
	c.lastHealthCheck = time.Now()
	c.isHealthy = response.IsHealthy()
	c.healthCheckMu.Unlock()

	if !response.IsHealthy() {
		return fmt.Errorf("service is not healthy: status=%s", response.Status)
	}

	return nil
}

// GetConfig returns the client configuration (read-only copy)
func (c *APIClient) GetConfig() *config.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent modification
	configCopy := *c.config
	return &configCopy
}

// IsHealthy returns the last known health status
func (c *APIClient) IsHealthy() bool {
	c.healthCheckMu.RLock()
	defer c.healthCheckMu.RUnlock()

	return c.isHealthy
}

// GetLastHealthCheck returns the timestamp of the last health check
func (c *APIClient) GetLastHealthCheck() time.Time {
	c.healthCheckMu.RLock()
	defer c.healthCheckMu.RUnlock()

	return c.lastHealthCheck
}

// RefreshHealthStatus performs a new health check and updates the status
func (c *APIClient) RefreshHealthStatus(ctx context.Context) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.RUnlock()

	response, err := c.Health.Check(ctx)
	if err != nil {
		c.healthCheckMu.Lock()
		c.lastHealthCheck = time.Now()
		c.isHealthy = false
		c.healthCheckMu.Unlock()
		return err
	}

	c.healthCheckMu.Lock()
	c.lastHealthCheck = time.Now()
	c.isHealthy = response.IsHealthy()
	c.healthCheckMu.Unlock()

	return nil
}

// WaitForHealthy waits for the service to become healthy within the given timeout
func (c *APIClient) WaitForHealthy(ctx context.Context, checkInterval time.Duration) error {
	return c.Health.WaitForHealthy(ctx, checkInterval)
}

// IsClosed returns true if the client has been closed
func (c *APIClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.closed
}

// Close gracefully closes the API client and releases resources
func (c *APIClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Close HTTP client connections if the HTTP client supports it
	// Note: The resty client doesn't have an explicit close method,
	// but we mark the client as closed to prevent further use

	return nil
}

// WithContext returns a new API client that uses the provided context for requests
// Note: This is conceptual - actual implementation would depend on how contexts are handled
func (c *APIClient) WithContext(ctx context.Context) *APIClient {
	// Create a copy of the client with context
	// This is a simplified implementation
	return c
}

// SetUserAgent updates the user agent for all requests
func (c *APIClient) SetUserAgent(userAgent string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.config.UserAgent = userAgent
		// Update the resty client's user agent
		c.client.SetHeader("User-Agent", userAgent)
	}
}

// SetTimeout updates the request timeout for all requests
func (c *APIClient) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed && timeout > 0 {
		c.config.RequestTimeout = timeout
		// Update the resty client's timeout
		c.client.SetTimeout(timeout)
	}
}

// GetVersion returns the SDK version
func (c *APIClient) GetVersion() string {
	return c.config.Version
}

// GetRestyClient returns the underlying resty client for advanced usage
func (c *APIClient) GetRestyClient() *resty.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.closed {
		return nil
	}
	
	return c.client
}

// GetHost returns the API host
func (c *APIClient) GetHost() string {
	return c.config.Host
}

// GetEnvironment returns the environment name
func (c *APIClient) GetEnvironment() string {
	return c.config.Environment
}

// IsEnabled returns whether the client is enabled
func (c *APIClient) IsEnabled() bool {
	return c.config.Enabled
}

// GetSampleRate returns the current sample rate
func (c *APIClient) GetSampleRate() float64 {
	return c.config.SampleRate
}

// ShouldSample returns true if an event should be sampled based on the sample rate
func (c *APIClient) ShouldSample() bool {
	if c.config.SampleRate >= 1.0 {
		return true
	}

	if c.config.SampleRate <= 0.0 {
		return false
	}

	// Simple random sampling - could be improved with more sophisticated algorithms
	return c.config.SampleRate > 0.5 // Simplified for now
}

// Debug returns whether debug mode is enabled
func (c *APIClient) Debug() bool {
	return c.config.Debug
}

// ClientStats represents client statistics and metrics
type ClientStats struct {
	LastHealthCheck time.Time `json:"lastHealthCheck"`
	IsHealthy       bool      `json:"isHealthy"`
	IsClosed        bool      `json:"isClosed"`
	Host            string    `json:"host"`
	Environment     string    `json:"environment"`
	Version         string    `json:"version"`
	SampleRate      float64   `json:"sampleRate"`
	RequestTimeout  string    `json:"requestTimeout"`
	RetryCount      int       `json:"retryCount"`
	Debug           bool      `json:"debug"`
	Enabled         bool      `json:"enabled"`
}

// GetStats returns client statistics and configuration information
func (c *APIClient) GetStats() *ClientStats {
	c.mu.RLock()
	c.healthCheckMu.RLock()

	stats := &ClientStats{
		LastHealthCheck: c.lastHealthCheck,
		IsHealthy:       c.isHealthy,
		IsClosed:        c.closed,
		Host:            c.config.Host,
		Environment:     c.config.Environment,
		Version:         c.config.Version,
		SampleRate:      c.config.SampleRate,
		RequestTimeout:  c.config.RequestTimeout.String(),
		RetryCount:      c.config.RetryCount,
		Debug:           c.config.Debug,
		Enabled:         c.config.Enabled,
	}

	c.healthCheckMu.RUnlock()
	c.mu.RUnlock()

	return stats
}

// Ping performs a simple health check and returns response time
func (c *APIClient) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()

	err := c.Health.CheckLiveness(ctx)

	duration := time.Since(start)

	if err != nil {
		return duration, fmt.Errorf("ping failed: %w", err)
	}

	return duration, nil
}

// TestConnection tests the connection to Langfuse with comprehensive checks
func (c *APIClient) TestConnection(ctx context.Context) error {
	// Check if client is closed
	if c.IsClosed() {
		return fmt.Errorf("client is closed")
	}

	// Perform health check
	if err := c.RefreshHealthStatus(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !c.IsHealthy() {
		return fmt.Errorf("service is not healthy")
	}

	return nil
}
