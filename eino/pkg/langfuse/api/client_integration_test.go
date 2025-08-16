package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"eino/pkg/langfuse/config"
)

func TestAPIClientIntegration(t *testing.T) {
	// Integration test for complete APIClient initialization flow
	t.Run("complete initialization flow", func(t *testing.T) {
		config := &config.Config{
			Host:                   "https://api.example.com",
			PublicKey:              "test-public-key",
			SecretKey:              "test-secret-key",
			RequestTimeout:         30 * time.Second,
			RetryCount:             3,
			RetryDelay:             1 * time.Second,
			MaxRetryDelay:          10 * time.Second,
			SampleRate:             1.0,
			HTTPUserAgent:          "langfuse-go-test",
			Debug:                  false,
			SkipInitialHealthCheck: true,
		}

		// Test APIClient creation
		client, err := NewAPIClient(config)
		if err != nil {
			t.Fatalf("NewAPIClient() failed: %v", err)
		}
		defer client.Close()

		// Verify resty client configuration
		restyClient := client.GetRestyClient()
		if restyClient == nil {
			t.Fatal("GetRestyClient() returned nil")
		}

		// Test configuration application
		if restyClient.BaseURL != config.Host {
			t.Errorf("BaseURL not configured correctly: got %v, want %v", 
				restyClient.BaseURL, config.Host)
		}

		if restyClient.RetryCount != config.RetryCount {
			t.Errorf("RetryCount not configured correctly: got %v, want %v", 
				restyClient.RetryCount, config.RetryCount)
		}

		// Test resource client integration
		if client.Health == nil {
			t.Error("Health client not initialized")
		}

		if client.Ingestion == nil {
			t.Error("Ingestion client not initialized")
		}
	})
}

func TestAPIClientConfigurationIntegration(t *testing.T) {
	// Test that configuration changes are properly applied to resty client
	t.Run("configuration changes integration", func(t *testing.T) {
		config := &config.Config{
			Host:                   "https://initial.example.com",
			PublicKey:              "initial-public-key",
			SecretKey:              "initial-secret-key",
			RequestTimeout:         15 * time.Second,
			HTTPUserAgent:          "initial-agent",
			SkipInitialHealthCheck: true,
		}

		client, err := NewAPIClient(config)
		if err != nil {
			t.Fatalf("NewAPIClient() failed: %v", err)
		}
		defer client.Close()

		restyClient := client.GetRestyClient()

		// Test SetUserAgent integration
		newAgent := "updated-test-agent"
		client.SetUserAgent(newAgent)

		updatedAgent := restyClient.Header.Get("User-Agent")
		if updatedAgent != newAgent {
			t.Errorf("SetUserAgent() integration failed: got %v, want %v", 
				updatedAgent, newAgent)
		}

		// Test SetTimeout integration
		newTimeout := 45 * time.Second
		client.SetTimeout(newTimeout)

		// Verify timeout was updated in config
		if client.GetConfig().RequestTimeout != newTimeout {
			t.Errorf("SetTimeout() config integration failed: got %v, want %v", 
				client.GetConfig().RequestTimeout, newTimeout)
		}
	})
}

func TestAPIClientResourceClientIntegration(t *testing.T) {
	// Test that resource clients work with the new configuration
	t.Run("resource clients integration", func(t *testing.T) {
		config := &config.Config{
			Host:                   "https://resource.example.com",
			PublicKey:              "resource-public-key",
			SecretKey:              "resource-secret-key",
			RequestTimeout:         20 * time.Second,
			HTTPUserAgent:          "resource-test-agent",
			SkipInitialHealthCheck: true,
		}

		client, err := NewAPIClient(config)
		if err != nil {
			t.Fatalf("NewAPIClient() failed: %v", err)
		}
		defer client.Close()

		// Test that all resource clients are properly initialized
		resourceClients := map[string]interface{}{
			"Health":    client.Health,
			"Ingestion": client.Ingestion,
			"Traces":    client.Traces,
			"Scores":    client.Scores,
			"Sessions":  client.Sessions,
		}

		for name, resourceClient := range resourceClients {
			if resourceClient == nil {
				t.Errorf("%s resource client is nil", name)
			}
		}

		// Test that client state management works correctly
		if client.IsClosed() {
			t.Error("Client should not be closed initially")
		}

		// Test Close() functionality
		err = client.Close()
		if err != nil {
			t.Errorf("Close() failed: %v", err)
		}

		if !client.IsClosed() {
			t.Error("Client should be closed after Close()")
		}

		// Test that GetRestyClient returns nil after close
		if client.GetRestyClient() != nil {
			t.Error("GetRestyClient() should return nil after close")
		}
	})
}

func TestAPIClientErrorHandling(t *testing.T) {
	// Test error handling in various scenarios
	t.Run("configuration error handling", func(t *testing.T) {
		invalidConfigs := []*config.Config{
			nil,
			{}, // Missing required fields
			{
				Host:      "invalid-url", // Invalid URL format
				PublicKey: "test",
				SecretKey: "test",
			},
			{
				Host:           "https://example.com",
				PublicKey:      "", // Missing public key
				SecretKey:      "test",
				RequestTimeout: 30 * time.Second,
			},
			{
				Host:           "https://example.com",
				PublicKey:      "test",
				SecretKey:      "", // Missing secret key
				RequestTimeout: 30 * time.Second,
			},
			{
				Host:           "https://example.com",
				PublicKey:      "test",
				SecretKey:      "test",
				RequestTimeout: 0, // Invalid timeout
			},
		}

		for i, invalidConfig := range invalidConfigs {
			t.Run(fmt.Sprintf("invalid_config_%d", i), func(t *testing.T) {
				client, err := NewAPIClient(invalidConfig)
				if err == nil {
					t.Error("NewAPIClient() should have failed with invalid config")
					if client != nil {
						client.Close()
					}
				}
			})
		}
	})
}

func TestAPIClientStatsIntegration(t *testing.T) {
	// Test that client statistics work correctly
	t.Run("client stats integration", func(t *testing.T) {
		config := &config.Config{
			Host:                   "https://stats.example.com",
			PublicKey:              "stats-public-key",
			SecretKey:              "stats-secret-key",
			RequestTimeout:         25 * time.Second,
			RetryCount:             2,
			HTTPUserAgent:          "stats-test-agent",
			Environment:            "test",
			Version:                "1.0.0",
			SampleRate:             0.8,
			Debug:                  true,
			Enabled:                true,
			SkipInitialHealthCheck: true,
		}

		client, err := NewAPIClient(config)
		if err != nil {
			t.Fatalf("NewAPIClient() failed: %v", err)
		}
		defer client.Close()

		stats := client.GetStats()
		if stats == nil {
			t.Fatal("GetStats() returned nil")
		}

		// Verify stats reflect configuration
		if stats.Host != config.Host {
			t.Errorf("Stats Host = %v, want %v", stats.Host, config.Host)
		}

		if stats.Environment != config.Environment {
			t.Errorf("Stats Environment = %v, want %v", stats.Environment, config.Environment)
		}

		if stats.Version != config.Version {
			t.Errorf("Stats Version = %v, want %v", stats.Version, config.Version)
		}

		if stats.SampleRate != config.SampleRate {
			t.Errorf("Stats SampleRate = %v, want %v", stats.SampleRate, config.SampleRate)
		}

		if stats.RetryCount != config.RetryCount {
			t.Errorf("Stats RetryCount = %v, want %v", stats.RetryCount, config.RetryCount)
		}

		if stats.Debug != config.Debug {
			t.Errorf("Stats Debug = %v, want %v", stats.Debug, config.Debug)
		}

		if stats.Enabled != config.Enabled {
			t.Errorf("Stats Enabled = %v, want %v", stats.Enabled, config.Enabled)
		}

		if stats.IsClosed != client.IsClosed() {
			t.Errorf("Stats IsClosed = %v, want %v", stats.IsClosed, client.IsClosed())
		}
	})
}