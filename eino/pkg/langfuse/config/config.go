package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"eino/pkg/langfuse/internal/utils"
)

// Config represents comprehensive configuration options for the Langfuse client.
//
// The configuration controls all aspects of the SDK behavior including API connectivity,
// batching and queuing, retry logic, debugging, and performance tuning.
//
// Configuration can be loaded from environment variables using LoadConfig() or
// created programmatically using NewConfig() with configuration options.
//
// Environment variables (with defaults):
//   - LANGFUSE_HOST: API endpoint URL (default: "https://cloud.langfuse.com")
//   - LANGFUSE_PUBLIC_KEY: API public key (required)
//   - LANGFUSE_SECRET_KEY: API secret key (required)
//   - LANGFUSE_DEBUG: Enable debug logging (default: false)
//   - LANGFUSE_ENABLED: Enable/disable SDK (default: true)
//   - LANGFUSE_FLUSH_AT: Batch size for auto-flush (default: 15)
//   - LANGFUSE_FLUSH_INTERVAL: Time interval for auto-flush (default: 10s)
//   - LANGFUSE_TIMEOUT: Request timeout (default: 10s)
//   - LANGFUSE_ENVIRONMENT: Environment name for traces (optional)
//   - LANGFUSE_RELEASE: Release version for traces (optional)
type Config struct {
	// API Configuration - Connection settings for the Langfuse service

	// Host is the base URL of the Langfuse API endpoint
	Host string

	// PublicKey is the API public key for authentication
	PublicKey string

	// SecretKey is the API secret key for authentication
	SecretKey string

	// APIVersion specifies the API version to use (currently unused)
	APIVersion string

	// HTTP Client Configuration - Settings for the underlying HTTP transport

	// Timeout is the timeout for individual HTTP requests (deprecated, use RequestTimeout)
	Timeout time.Duration

	// RetryCount is the maximum number of retry attempts for failed requests
	RetryCount int

	// RetryDelay is the initial delay between retry attempts (deprecated, use RetryWaitTime)
	RetryDelay time.Duration

	// MaxRetryDelay is the maximum delay between retry attempts (deprecated, use RetryMaxWaitTime)
	MaxRetryDelay time.Duration

	// HTTPUserAgent is the User-Agent header for HTTP requests (deprecated, use UserAgent)
	HTTPUserAgent string

	// Queue Configuration - Settings for async event processing and batching

	// FlushAt is the number of events that triggers an automatic flush to the API
	FlushAt int

	// FlushInterval is the maximum time to wait before flushing pending events
	FlushInterval time.Duration

	// QueueSize is the maximum number of events to buffer in memory
	QueueSize int

	// WorkerCount is the number of background workers for processing events (currently unused)
	WorkerCount int

	// Feature Flags - Enable/disable SDK features

	// Debug enables verbose logging for troubleshooting
	Debug bool

	// Enabled controls whether the SDK performs any actual work (allows conditional disabling)
	Enabled bool

	// BatchMode enables batch processing optimizations (currently unused)
	BatchMode bool

	// Advanced Configuration - Environment and versioning settings

	// Release identifies the application release version in traces
	Release string

	// Environment identifies the deployment environment in traces (e.g., "production", "staging")
	Environment string

	// RequestTimeout is the timeout for API requests
	RequestTimeout time.Duration

	// SDKName identifies the SDK in API requests (set automatically)
	SDKName string

	// SDKVersion identifies the SDK version in API requests (set automatically)
	SDKVersion string

	// Performance and Reliability Configuration

	// SampleRate controls what fraction of events to actually submit (0.0-1.0, default 1.0)
	SampleRate float64

	// UserAgent is the User-Agent header value for HTTP requests
	UserAgent string

	// Version is the SDK version identifier
	Version string

	// RetryWaitTime is the initial delay between retry attempts
	RetryWaitTime time.Duration

	// RetryMaxWaitTime is the maximum delay between retry attempts
	RetryMaxWaitTime       time.Duration
	SkipInitialHealthCheck bool
	RequireHealthyStart    bool
}

// ConfigOption represents a configuration option function
type ConfigOption func(*Config) error

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		// API defaults
		Host:       "https://cloud.langfuse.com",
		APIVersion: "v1",

		// HTTP defaults
		Timeout:       30 * time.Second,
		RetryCount:    3,
		RetryDelay:    1 * time.Second,
		MaxRetryDelay: 30 * time.Second,
		HTTPUserAgent: "langfuse-go-sdk",

		// Queue defaults
		FlushAt:       100,
		FlushInterval: 10 * time.Second,
		QueueSize:     1000,
		WorkerCount:   1,

		// Feature flags
		Debug:     false,
		Enabled:   true,
		BatchMode: true,

		// Advanced defaults
		RequestTimeout: 10 * time.Second,
		SDKName:        "langfuse-go",
		SDKVersion:     "1.0.0",

		// Additional API client defaults
		SampleRate:             1.0,
		UserAgent:              "langfuse-go/1.0.0",
		Version:                "1.0.0",
		RetryWaitTime:          1 * time.Second,
		RetryMaxWaitTime:       10 * time.Second,
		SkipInitialHealthCheck: false,
		RequireHealthyStart:    false,
	}
}

// NewConfig creates a new Config with environment variables loaded and options applied
func NewConfig(options ...ConfigOption) (*Config, error) {
	config := DefaultConfig()

	// Load from environment variables first
	if err := config.LoadFromEnvironment(); err != nil {
		return nil, err
	}

	// Apply options
	for _, option := range options {
		if err := option(config); err != nil {
			return nil, err
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadFromEnvironment loads configuration from environment variables
func (c *Config) LoadFromEnvironment() error {
	// API Configuration
	if host := os.Getenv("LANGFUSE_HOST"); host != "" {
		c.Host = strings.TrimSuffix(host, "/")
	}
	if publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY"); publicKey != "" {
		c.PublicKey = publicKey
	}
	if secretKey := os.Getenv("LANGFUSE_SECRET_KEY"); secretKey != "" {
		c.SecretKey = secretKey
	}

	// HTTP Configuration
	if timeout := os.Getenv("LANGFUSE_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Timeout = d
		}
	}
	if retryCount := os.Getenv("LANGFUSE_RETRY_COUNT"); retryCount != "" {
		if count, err := strconv.Atoi(retryCount); err == nil && count >= 0 {
			c.RetryCount = count
		}
	}

	// Queue Configuration
	if flushAt := os.Getenv("LANGFUSE_FLUSH_AT"); flushAt != "" {
		if count, err := strconv.Atoi(flushAt); err == nil && count > 0 {
			c.FlushAt = count
		}
	}
	if flushInterval := os.Getenv("LANGFUSE_FLUSH_INTERVAL"); flushInterval != "" {
		if d, err := time.ParseDuration(flushInterval); err == nil {
			c.FlushInterval = d
		}
	}
	if queueSize := os.Getenv("LANGFUSE_QUEUE_SIZE"); queueSize != "" {
		if size, err := strconv.Atoi(queueSize); err == nil && size > 0 {
			c.QueueSize = size
		}
	}
	if workerCount := os.Getenv("LANGFUSE_WORKER_COUNT"); workerCount != "" {
		if count, err := strconv.Atoi(workerCount); err == nil && count > 0 {
			c.WorkerCount = count
		}
	}

	// Feature Flags
	if debug := os.Getenv("LANGFUSE_DEBUG"); debug != "" {
		c.Debug = strings.ToLower(debug) == "true" || debug == "1"
	}
	if enabled := os.Getenv("LANGFUSE_ENABLED"); enabled != "" {
		c.Enabled = strings.ToLower(enabled) == "true" || enabled == "1"
	}
	if batchMode := os.Getenv("LANGFUSE_BATCH_MODE"); batchMode != "" {
		c.BatchMode = strings.ToLower(batchMode) == "true" || batchMode == "1"
	}

	// Advanced Configuration
	if release := os.Getenv("LANGFUSE_RELEASE"); release != "" {
		c.Release = release
	}
	if environment := os.Getenv("LANGFUSE_ENVIRONMENT"); environment != "" {
		c.Environment = environment
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.PublicKey == "" {
		return utils.NewConfigurationError("publicKey", "public key is required")
	}
	if c.SecretKey == "" {
		return utils.NewConfigurationError("secretKey", "secret key is required")
	}
	if c.Host == "" {
		return utils.NewConfigurationError("host", "host is required")
	}
	if !strings.HasPrefix(c.Host, "http://") && !strings.HasPrefix(c.Host, "https://") {
		return utils.NewConfigurationError("host", "host must include protocol (http:// or https://)")
	}
	if c.Timeout <= 0 {
		return utils.NewConfigurationErrorWithExpected("timeout", "timeout must be positive", "> 0", c.Timeout.String())
	}
	if c.FlushAt <= 0 {
		return utils.NewConfigurationErrorWithExpected("flushAt", "flush at must be positive", "> 0", strconv.Itoa(c.FlushAt))
	}
	if c.FlushInterval <= 0 {
		return utils.NewConfigurationErrorWithExpected("flushInterval", "flush interval must be positive", "> 0", c.FlushInterval.String())
	}
	if c.QueueSize <= 0 {
		return utils.NewConfigurationErrorWithExpected("queueSize", "queue size must be positive", "> 0", strconv.Itoa(c.QueueSize))
	}
	if c.WorkerCount <= 0 {
		return utils.NewConfigurationErrorWithExpected("workerCount", "worker count must be positive", "> 0", strconv.Itoa(c.WorkerCount))
	}

	return nil
}

// WithHost sets the Langfuse API host
func WithHost(host string) ConfigOption {
	return func(c *Config) error {
		if host == "" {
			return utils.NewConfigurationError("host", "host cannot be empty")
		}
		c.Host = strings.TrimSuffix(host, "/")
		return nil
	}
}

// WithCredentials sets the API credentials
func WithCredentials(publicKey, secretKey string) ConfigOption {
	return func(c *Config) error {
		if publicKey == "" {
			return utils.NewConfigurationError("publicKey", "public key cannot be empty")
		}
		if secretKey == "" {
			return utils.NewConfigurationError("secretKey", "secret key cannot be empty")
		}
		c.PublicKey = publicKey
		c.SecretKey = secretKey
		return nil
	}
}

// WithPublicKey sets the public key
func WithPublicKey(publicKey string) ConfigOption {
	return func(c *Config) error {
		if publicKey == "" {
			return utils.NewConfigurationError("publicKey", "public key cannot be empty")
		}
		c.PublicKey = publicKey
		return nil
	}
}

// WithSecretKey sets the secret key
func WithSecretKey(secretKey string) ConfigOption {
	return func(c *Config) error {
		if secretKey == "" {
			return utils.NewConfigurationError("secretKey", "secret key cannot be empty")
		}
		c.SecretKey = secretKey
		return nil
	}
}

// WithTimeout sets the HTTP timeout
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) error {
		if timeout <= 0 {
			return utils.NewConfigurationError("timeout", "timeout must be positive")
		}
		c.Timeout = timeout
		return nil
	}
}

// WithRetryConfig sets retry configuration
func WithRetryConfig(count int, delay, maxDelay time.Duration) ConfigOption {
	return func(c *Config) error {
		if count < 0 {
			return utils.NewConfigurationError("retryCount", "retry count cannot be negative")
		}
		if delay < 0 {
			return utils.NewConfigurationError("retryDelay", "retry delay cannot be negative")
		}
		if maxDelay < delay {
			return utils.NewConfigurationError("maxRetryDelay", "max retry delay must be >= retry delay")
		}
		c.RetryCount = count
		c.RetryDelay = delay
		c.MaxRetryDelay = maxDelay
		return nil
	}
}

// WithQueueConfig sets queue configuration
func WithQueueConfig(flushAt int, flushInterval time.Duration, queueSize, workerCount int) ConfigOption {
	return func(c *Config) error {
		if flushAt <= 0 {
			return utils.NewConfigurationError("flushAt", "flush at must be positive")
		}
		if flushInterval <= 0 {
			return utils.NewConfigurationError("flushInterval", "flush interval must be positive")
		}
		if queueSize <= 0 {
			return utils.NewConfigurationError("queueSize", "queue size must be positive")
		}
		if workerCount <= 0 {
			return utils.NewConfigurationError("workerCount", "worker count must be positive")
		}
		c.FlushAt = flushAt
		c.FlushInterval = flushInterval
		c.QueueSize = queueSize
		c.WorkerCount = workerCount
		return nil
	}
}

// WithDebug enables or disables debug mode
func WithDebug(enabled bool) ConfigOption {
	return func(c *Config) error {
		c.Debug = enabled
		return nil
	}
}

// WithEnabled enables or disables the SDK
func WithEnabled(enabled bool) ConfigOption {
	return func(c *Config) error {
		c.Enabled = enabled
		return nil
	}
}

// WithBatchMode enables or disables batch mode
func WithBatchMode(enabled bool) ConfigOption {
	return func(c *Config) error {
		c.BatchMode = enabled
		return nil
	}
}

// WithRelease sets the release version
func WithRelease(release string) ConfigOption {
	return func(c *Config) error {
		c.Release = release
		return nil
	}
}

// WithEnvironment sets the environment
func WithEnvironment(environment string) ConfigOption {
	return func(c *Config) error {
		c.Environment = environment
		return nil
	}
}

// WithUserAgent sets the HTTP user agent
func WithUserAgent(userAgent string) ConfigOption {
	return func(c *Config) error {
		if userAgent == "" {
			return utils.NewConfigurationError("userAgent", "user agent cannot be empty")
		}
		c.HTTPUserAgent = userAgent
		return nil
	}
}
