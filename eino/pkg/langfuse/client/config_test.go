package client

import (
	"os"
	"testing"
	"time"

	"eino/pkg/langfuse/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test API defaults
	assert.Equal(t, "https://cloud.langfuse.com", config.Host)
	assert.Equal(t, "v1", config.APIVersion)

	// Test HTTP defaults
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.RetryCount)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.MaxRetryDelay)
	assert.Equal(t, "langfuse-go-sdk", config.HTTPUserAgent)

	// Test Queue defaults
	assert.Equal(t, 100, config.FlushAt)
	assert.Equal(t, 10*time.Second, config.FlushInterval)
	assert.Equal(t, 1000, config.QueueSize)
	assert.Equal(t, 1, config.WorkerCount)

	// Test Feature flags
	assert.False(t, config.Debug)
	assert.True(t, config.Enabled)
	assert.True(t, config.BatchMode)

	// Test Advanced defaults
	assert.Equal(t, 10*time.Second, config.RequestTimeout)
	assert.Equal(t, "langfuse-go", config.SDKName)
	assert.Equal(t, "1.0.0", config.SDKVersion)

	// Test additional API client defaults
	assert.Equal(t, 1.0, config.SampleRate)
	assert.Equal(t, "langfuse-go/1.0.0", config.UserAgent)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, 1*time.Second, config.RetryWaitTime)
	assert.Equal(t, 10*time.Second, config.RetryMaxWaitTime)
	assert.False(t, config.SkipInitialHealthCheck)
	assert.False(t, config.RequireHealthyStart)
}

func TestConfig_LoadFromEnvironment(t *testing.T) {
	// Save original env vars
	originalVars := saveEnvironmentVars()
	defer restoreEnvironmentVars(originalVars)

	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, config *Config)
	}{
		{
			name: "all environment variables",
			envVars: map[string]string{
				"LANGFUSE_HOST":           "https://custom.langfuse.com",
				"LANGFUSE_PUBLIC_KEY":     "pk_test_12345",
				"LANGFUSE_SECRET_KEY":     "sk_test_67890",
				"LANGFUSE_TIMEOUT":        "60s",
				"LANGFUSE_RETRY_COUNT":    "5",
				"LANGFUSE_FLUSH_AT":       "50",
				"LANGFUSE_FLUSH_INTERVAL": "5s",
				"LANGFUSE_QUEUE_SIZE":     "500",
				"LANGFUSE_WORKER_COUNT":   "2",
				"LANGFUSE_DEBUG":          "true",
				"LANGFUSE_ENABLED":        "false",
				"LANGFUSE_BATCH_MODE":     "false",
				"LANGFUSE_RELEASE":        "v2.0.0",
				"LANGFUSE_ENVIRONMENT":    "staging",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "https://custom.langfuse.com", config.Host)
				assert.Equal(t, "pk_test_12345", config.PublicKey)
				assert.Equal(t, "sk_test_67890", config.SecretKey)
				assert.Equal(t, 60*time.Second, config.Timeout)
				assert.Equal(t, 5, config.RetryCount)
				assert.Equal(t, 50, config.FlushAt)
				assert.Equal(t, 5*time.Second, config.FlushInterval)
				assert.Equal(t, 500, config.QueueSize)
				assert.Equal(t, 2, config.WorkerCount)
				assert.True(t, config.Debug)
				assert.False(t, config.Enabled)
				assert.False(t, config.BatchMode)
				assert.Equal(t, "v2.0.0", config.Release)
				assert.Equal(t, "staging", config.Environment)
			},
		},
		{
			name: "host with trailing slash",
			envVars: map[string]string{
				"LANGFUSE_HOST": "https://custom.langfuse.com/",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "https://custom.langfuse.com", config.Host)
			},
		},
		{
			name: "boolean values with '1'",
			envVars: map[string]string{
				"LANGFUSE_DEBUG":      "1",
				"LANGFUSE_ENABLED":    "1",
				"LANGFUSE_BATCH_MODE": "1",
			},
			validate: func(t *testing.T, config *Config) {
				assert.True(t, config.Debug)
				assert.True(t, config.Enabled)
				assert.True(t, config.BatchMode)
			},
		},
		{
			name: "boolean values case insensitive",
			envVars: map[string]string{
				"LANGFUSE_DEBUG":      "True",
				"LANGFUSE_ENABLED":    "TRUE",
				"LANGFUSE_BATCH_MODE": "tRuE",
			},
			validate: func(t *testing.T, config *Config) {
				assert.True(t, config.Debug)
				assert.True(t, config.Enabled)
				assert.True(t, config.BatchMode)
			},
		},
		{
			name: "invalid values ignored",
			envVars: map[string]string{
				"LANGFUSE_TIMEOUT":        "invalid",
				"LANGFUSE_RETRY_COUNT":    "invalid",
				"LANGFUSE_FLUSH_AT":       "-5", // negative value ignored
				"LANGFUSE_FLUSH_INTERVAL": "invalid",
				"LANGFUSE_QUEUE_SIZE":     "0", // zero value ignored
				"LANGFUSE_WORKER_COUNT":   "-1", // negative value ignored
			},
			validate: func(t *testing.T, config *Config) {
				// Should keep default values when invalid values provided
				assert.Equal(t, 30*time.Second, config.Timeout)
				assert.Equal(t, 3, config.RetryCount)
				assert.Equal(t, 100, config.FlushAt)
				assert.Equal(t, 10*time.Second, config.FlushInterval)
				assert.Equal(t, 1000, config.QueueSize)
				assert.Equal(t, 1, config.WorkerCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all Langfuse environment variables
			clearLangfuseEnvVars()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := DefaultConfig()
			err := config.LoadFromEnvironment()
			require.NoError(t, err)

			tt.validate(t, config)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *Config
		expectError bool
		errorField  string
	}{
		{
			name: "valid config",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				return config
			},
			expectError: false,
		},
		{
			name: "missing public key",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.SecretKey = "sk_test_67890"
				return config
			},
			expectError: true,
			errorField:  "publicKey",
		},
		{
			name: "missing secret key",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				return config
			},
			expectError: true,
			errorField:  "secretKey",
		},
		{
			name: "empty host",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.Host = ""
				return config
			},
			expectError: true,
			errorField:  "host",
		},
		{
			name: "host without protocol",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.Host = "cloud.langfuse.com"
				return config
			},
			expectError: true,
			errorField:  "host",
		},
		{
			name: "zero timeout",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.Timeout = 0
				return config
			},
			expectError: true,
			errorField:  "timeout",
		},
		{
			name: "negative timeout",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.Timeout = -1 * time.Second
				return config
			},
			expectError: true,
			errorField:  "timeout",
		},
		{
			name: "zero flush at",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.FlushAt = 0
				return config
			},
			expectError: true,
			errorField:  "flushAt",
		},
		{
			name: "negative flush interval",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.FlushInterval = -1 * time.Second
				return config
			},
			expectError: true,
			errorField:  "flushInterval",
		},
		{
			name: "zero queue size",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.QueueSize = 0
				return config
			},
			expectError: true,
			errorField:  "queueSize",
		},
		{
			name: "zero worker count",
			setupConfig: func() *Config {
				config := DefaultConfig()
				config.PublicKey = "pk_test_12345"
				config.SecretKey = "sk_test_67890"
				config.WorkerCount = 0
				return config
			},
			expectError: true,
			errorField:  "workerCount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()
			err := config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					configErr, ok := err.(*utils.ConfigurationError)
					assert.True(t, ok, "Expected ConfigurationError")
					assert.Equal(t, tt.errorField, configErr.Parameter)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	// Save original env vars
	originalVars := saveEnvironmentVars()
	defer restoreEnvironmentVars(originalVars)

	t.Run("with environment variables and options", func(t *testing.T) {
		// Clear all Langfuse environment variables
		clearLangfuseEnvVars()

		// Set environment variables
		os.Setenv("LANGFUSE_HOST", "https://env.langfuse.com")
		os.Setenv("LANGFUSE_PUBLIC_KEY", "pk_env_12345")
		os.Setenv("LANGFUSE_DEBUG", "true")

		config, err := NewConfig(
			WithSecretKey("sk_option_67890"),
			WithTimeout(45*time.Second),
			WithDebug(false), // Should override env variable
		)

		require.NoError(t, err)
		assert.Equal(t, "https://env.langfuse.com", config.Host)
		assert.Equal(t, "pk_env_12345", config.PublicKey)
		assert.Equal(t, "sk_option_67890", config.SecretKey)
		assert.Equal(t, 45*time.Second, config.Timeout)
		assert.False(t, config.Debug) // Option overrides env
	})

	t.Run("validation error", func(t *testing.T) {
		// Clear all Langfuse environment variables
		clearLangfuseEnvVars()

		_, err := NewConfig(
			WithPublicKey("pk_test_12345"),
			// Missing secret key
		)

		assert.Error(t, err)
		configErr, ok := err.(*utils.ConfigurationError)
		assert.True(t, ok)
		assert.Equal(t, "secretKey", configErr.Parameter)
	})

	t.Run("option error", func(t *testing.T) {
		_, err := NewConfig(
			WithHost(""), // Empty host should cause error
		)

		assert.Error(t, err)
		configErr, ok := err.(*utils.ConfigurationError)
		assert.True(t, ok)
		assert.Equal(t, "host", configErr.Parameter)
	})
}

func TestConfigOptions(t *testing.T) {
	tests := []struct {
		name        string
		option      ConfigOption
		expectError bool
		validate    func(t *testing.T, config *Config)
	}{
		{
			name:        "WithHost valid",
			option:      WithHost("https://test.langfuse.com"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "https://test.langfuse.com", config.Host)
			},
		},
		{
			name:        "WithHost with trailing slash",
			option:      WithHost("https://test.langfuse.com/"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "https://test.langfuse.com", config.Host)
			},
		},
		{
			name:        "WithHost empty",
			option:      WithHost(""),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithCredentials valid",
			option:      WithCredentials("pk_test", "sk_test"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "pk_test", config.PublicKey)
				assert.Equal(t, "sk_test", config.SecretKey)
			},
		},
		{
			name:        "WithCredentials empty public key",
			option:      WithCredentials("", "sk_test"),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithCredentials empty secret key",
			option:      WithCredentials("pk_test", ""),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithPublicKey valid",
			option:      WithPublicKey("pk_test_12345"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "pk_test_12345", config.PublicKey)
			},
		},
		{
			name:        "WithPublicKey empty",
			option:      WithPublicKey(""),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithSecretKey valid",
			option:      WithSecretKey("sk_test_67890"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "sk_test_67890", config.SecretKey)
			},
		},
		{
			name:        "WithSecretKey empty",
			option:      WithSecretKey(""),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithTimeout valid",
			option:      WithTimeout(60 * time.Second),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 60*time.Second, config.Timeout)
			},
		},
		{
			name:        "WithTimeout zero",
			option:      WithTimeout(0),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithTimeout negative",
			option:      WithTimeout(-1 * time.Second),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithRetryConfig valid",
			option:      WithRetryConfig(5, 2*time.Second, 60*time.Second),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 5, config.RetryCount)
				assert.Equal(t, 2*time.Second, config.RetryDelay)
				assert.Equal(t, 60*time.Second, config.MaxRetryDelay)
			},
		},
		{
			name:        "WithRetryConfig negative count",
			option:      WithRetryConfig(-1, 1*time.Second, 10*time.Second),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithRetryConfig negative delay",
			option:      WithRetryConfig(3, -1*time.Second, 10*time.Second),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithRetryConfig max delay less than delay",
			option:      WithRetryConfig(3, 10*time.Second, 5*time.Second),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithQueueConfig valid",
			option:      WithQueueConfig(50, 5*time.Second, 500, 2),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 50, config.FlushAt)
				assert.Equal(t, 5*time.Second, config.FlushInterval)
				assert.Equal(t, 500, config.QueueSize)
				assert.Equal(t, 2, config.WorkerCount)
			},
		},
		{
			name:        "WithQueueConfig zero flush at",
			option:      WithQueueConfig(0, 5*time.Second, 500, 2),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithQueueConfig zero flush interval",
			option:      WithQueueConfig(50, 0, 500, 2),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithQueueConfig zero queue size",
			option:      WithQueueConfig(50, 5*time.Second, 0, 2),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithQueueConfig zero worker count",
			option:      WithQueueConfig(50, 5*time.Second, 500, 0),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "WithDebug true",
			option:      WithDebug(true),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.True(t, config.Debug)
			},
		},
		{
			name:        "WithDebug false",
			option:      WithDebug(false),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.False(t, config.Debug)
			},
		},
		{
			name:        "WithEnabled true",
			option:      WithEnabled(true),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.True(t, config.Enabled)
			},
		},
		{
			name:        "WithEnabled false",
			option:      WithEnabled(false),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.False(t, config.Enabled)
			},
		},
		{
			name:        "WithBatchMode true",
			option:      WithBatchMode(true),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.True(t, config.BatchMode)
			},
		},
		{
			name:        "WithBatchMode false",
			option:      WithBatchMode(false),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.False(t, config.BatchMode)
			},
		},
		{
			name:        "WithRelease",
			option:      WithRelease("v2.1.0"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "v2.1.0", config.Release)
			},
		},
		{
			name:        "WithEnvironment",
			option:      WithEnvironment("production"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "production", config.Environment)
			},
		},
		{
			name:        "WithUserAgent valid",
			option:      WithUserAgent("custom-agent/1.0.0"),
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "custom-agent/1.0.0", config.HTTPUserAgent)
			},
		},
		{
			name:        "WithUserAgent empty",
			option:      WithUserAgent(""),
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			err := tt.option(config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

func TestConfig_MultipleOptions(t *testing.T) {
	config, err := NewConfig(
		WithHost("https://test.langfuse.com"),
		WithCredentials("pk_test_12345", "sk_test_67890"),
		WithTimeout(45*time.Second),
		WithRetryConfig(5, 2*time.Second, 30*time.Second),
		WithQueueConfig(25, 5*time.Second, 250, 2),
		WithDebug(true),
		WithEnabled(true),
		WithBatchMode(false),
		WithRelease("v1.5.0"),
		WithEnvironment("test"),
		WithUserAgent("test-agent/1.0.0"),
	)

	require.NoError(t, err)

	// Validate all options were applied
	assert.Equal(t, "https://test.langfuse.com", config.Host)
	assert.Equal(t, "pk_test_12345", config.PublicKey)
	assert.Equal(t, "sk_test_67890", config.SecretKey)
	assert.Equal(t, 45*time.Second, config.Timeout)
	assert.Equal(t, 5, config.RetryCount)
	assert.Equal(t, 2*time.Second, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.MaxRetryDelay)
	assert.Equal(t, 25, config.FlushAt)
	assert.Equal(t, 5*time.Second, config.FlushInterval)
	assert.Equal(t, 250, config.QueueSize)
	assert.Equal(t, 2, config.WorkerCount)
	assert.True(t, config.Debug)
	assert.True(t, config.Enabled)
	assert.False(t, config.BatchMode)
	assert.Equal(t, "v1.5.0", config.Release)
	assert.Equal(t, "test", config.Environment)
	assert.Equal(t, "test-agent/1.0.0", config.HTTPUserAgent)
}

func TestConfig_EdgeCases(t *testing.T) {
	t.Run("empty string environment variables", func(t *testing.T) {
		// Save original env vars
		originalVars := saveEnvironmentVars()
		defer restoreEnvironmentVars(originalVars)

		// Clear all Langfuse environment variables
		clearLangfuseEnvVars()

		// Set empty string values
		os.Setenv("LANGFUSE_HOST", "")
		os.Setenv("LANGFUSE_PUBLIC_KEY", "")
		os.Setenv("LANGFUSE_SECRET_KEY", "")
		os.Setenv("LANGFUSE_RELEASE", "")
		os.Setenv("LANGFUSE_ENVIRONMENT", "")

		config := DefaultConfig()
		err := config.LoadFromEnvironment()
		require.NoError(t, err)

		// Empty string env vars should not override defaults
		assert.Equal(t, "https://cloud.langfuse.com", config.Host)
		assert.Equal(t, "", config.PublicKey) // These should remain empty
		assert.Equal(t, "", config.SecretKey)
		assert.Equal(t, "", config.Release)
		assert.Equal(t, "", config.Environment)
	})

	t.Run("extremely large values", func(t *testing.T) {
		// Save original env vars
		originalVars := saveEnvironmentVars()
		defer restoreEnvironmentVars(originalVars)

		// Clear all Langfuse environment variables
		clearLangfuseEnvVars()

		// Set large values
		os.Setenv("LANGFUSE_TIMEOUT", "24h")
		os.Setenv("LANGFUSE_RETRY_COUNT", "1000")
		os.Setenv("LANGFUSE_FLUSH_AT", "10000")
		os.Setenv("LANGFUSE_FLUSH_INTERVAL", "1h")
		os.Setenv("LANGFUSE_QUEUE_SIZE", "100000")
		os.Setenv("LANGFUSE_WORKER_COUNT", "100")

		config := DefaultConfig()
		err := config.LoadFromEnvironment()
		require.NoError(t, err)

		// Should parse large values correctly
		assert.Equal(t, 24*time.Hour, config.Timeout)
		assert.Equal(t, 1000, config.RetryCount)
		assert.Equal(t, 10000, config.FlushAt)
		assert.Equal(t, 1*time.Hour, config.FlushInterval)
		assert.Equal(t, 100000, config.QueueSize)
		assert.Equal(t, 100, config.WorkerCount)
	})
}

// Helper functions for testing

func saveEnvironmentVars() map[string]string {
	vars := map[string]string{
		"LANGFUSE_HOST":           os.Getenv("LANGFUSE_HOST"),
		"LANGFUSE_PUBLIC_KEY":     os.Getenv("LANGFUSE_PUBLIC_KEY"),
		"LANGFUSE_SECRET_KEY":     os.Getenv("LANGFUSE_SECRET_KEY"),
		"LANGFUSE_TIMEOUT":        os.Getenv("LANGFUSE_TIMEOUT"),
		"LANGFUSE_RETRY_COUNT":    os.Getenv("LANGFUSE_RETRY_COUNT"),
		"LANGFUSE_FLUSH_AT":       os.Getenv("LANGFUSE_FLUSH_AT"),
		"LANGFUSE_FLUSH_INTERVAL": os.Getenv("LANGFUSE_FLUSH_INTERVAL"),
		"LANGFUSE_QUEUE_SIZE":     os.Getenv("LANGFUSE_QUEUE_SIZE"),
		"LANGFUSE_WORKER_COUNT":   os.Getenv("LANGFUSE_WORKER_COUNT"),
		"LANGFUSE_DEBUG":          os.Getenv("LANGFUSE_DEBUG"),
		"LANGFUSE_ENABLED":        os.Getenv("LANGFUSE_ENABLED"),
		"LANGFUSE_BATCH_MODE":     os.Getenv("LANGFUSE_BATCH_MODE"),
		"LANGFUSE_RELEASE":        os.Getenv("LANGFUSE_RELEASE"),
		"LANGFUSE_ENVIRONMENT":    os.Getenv("LANGFUSE_ENVIRONMENT"),
	}
	return vars
}

func restoreEnvironmentVars(vars map[string]string) {
	for key, value := range vars {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func clearLangfuseEnvVars() {
	envVars := []string{
		"LANGFUSE_HOST",
		"LANGFUSE_PUBLIC_KEY",
		"LANGFUSE_SECRET_KEY",
		"LANGFUSE_TIMEOUT",
		"LANGFUSE_RETRY_COUNT",
		"LANGFUSE_FLUSH_AT",
		"LANGFUSE_FLUSH_INTERVAL",
		"LANGFUSE_QUEUE_SIZE",
		"LANGFUSE_WORKER_COUNT",
		"LANGFUSE_DEBUG",
		"LANGFUSE_ENABLED",
		"LANGFUSE_BATCH_MODE",
		"LANGFUSE_RELEASE",
		"LANGFUSE_ENVIRONMENT",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}