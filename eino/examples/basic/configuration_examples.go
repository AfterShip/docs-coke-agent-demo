package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"eino/pkg/langfuse/client"
)

func main() {
	fmt.Println("=== Langfuse Configuration Examples ===\n")

	// Example 1: Environment-based configuration (recommended)
	environmentBasedConfig()

	// Example 2: Programmatic configuration
	programmaticConfig()

	// Example 3: Configuration with options
	configWithOptions()

	// Example 4: Production-ready configuration
	productionConfig()
}

func environmentBasedConfig() {
	fmt.Println("1. Environment-based Configuration:")
	fmt.Println("   Set these environment variables:")
	fmt.Println("   LANGFUSE_PUBLIC_KEY=pk_your_public_key")
	fmt.Println("   LANGFUSE_SECRET_KEY=sk_your_secret_key")
	fmt.Println("   LANGFUSE_HOST=https://cloud.langfuse.com")
	fmt.Println("   LANGFUSE_DEBUG=true")
	fmt.Println("   LANGFUSE_ENVIRONMENT=production")

	// Load configuration from environment
	config, err := client.LoadConfig()
	if err != nil {
		fmt.Printf("   ❌ Failed to load config: %v\n", err)
		return
	}

	// You can still override specific settings
	config.Debug = true
	config.Environment = "development"

	langfuseClient, err := client.New(config)
	if err != nil {
		fmt.Printf("   ❌ Failed to create client: %v\n", err)
		return
	}
	defer langfuseClient.Shutdown(context.Background())

	fmt.Printf("   ✅ Client created with environment: %s\n\n", langfuseClient.GetEnvironment())
}

func programmaticConfig() {
	fmt.Println("2. Programmatic Configuration:")

	config := &client.Config{
		Host:           "https://cloud.langfuse.com",
		PublicKey:      "your-public-key",
		SecretKey:      "your-secret-key",
		Debug:          true,
		Enabled:        true,
		Environment:    "development",
		FlushAt:        15,
		FlushInterval:  10 * time.Second,
		RequestTimeout: 10 * time.Second,
		RetryCount:     3,
		RetryDelay:     1 * time.Second,
		QueueSize:      1000,
	}

	langfuseClient, err := client.New(config)
	if err != nil {
		fmt.Printf("   ❌ Failed to create client: %v\n", err)
		return
	}
	defer langfuseClient.Shutdown(context.Background())

	fmt.Printf("   ✅ Client created with flush interval: %v\n\n", config.FlushInterval)
}

func configWithOptions() {
	fmt.Println("3. Configuration with Options (recommended):")

	langfuseClient, err := client.NewWithOptions(
		client.WithHost("https://cloud.langfuse.com"),
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithDebug(true),
		client.WithEnvironment("staging"),
		client.WithTimeout(15*time.Second),
		client.WithFlushSettings(20, 5*time.Second),
		client.WithRetrySettings(3, time.Second),
	)
	if err != nil {
		fmt.Printf("   ❌ Failed to create client: %v\n", err)
		return
	}
	defer langfuseClient.Shutdown(context.Background())

	config := langfuseClient.GetConfig()
	fmt.Printf("   ✅ Client created with timeout: %v, flush at: %d\n\n", 
		config.RequestTimeout, config.FlushAt)
}

func productionConfig() {
	fmt.Println("4. Production-ready Configuration:")

	langfuseClient, err := client.NewWithOptions(
		// Connection settings
		client.WithHost("https://cloud.langfuse.com"),
		client.WithCredentials("your-public-key", "your-secret-key"),
		
		// Environment settings
		client.WithEnvironment("production"),
		client.WithRelease("v1.2.3"),
		
		// Performance settings
		client.WithTimeout(30*time.Second),
		client.WithFlushSettings(50, 30*time.Second), // Larger batches, less frequent
		client.WithRetrySettings(5, 2*time.Second),   // More retries for reliability
		
		// Production optimizations
		client.WithDebug(false),    // Disable debug in production
		client.WithEnabled(true),   // Ensure SDK is enabled
		client.WithQueueSize(5000), // Larger queue for high throughput
	)
	if err != nil {
		fmt.Printf("   ❌ Failed to create client: %v\n", err)
		return
	}
	defer langfuseClient.Shutdown(context.Background())

	// Test the health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := langfuseClient.HealthCheck(ctx); err != nil {
		fmt.Printf("   ⚠️  Health check failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Health check passed")
	}

	config := langfuseClient.GetConfig()
	fmt.Printf("   ✅ Production client ready - Environment: %s, Release: %s\n\n", 
		config.Environment, config.Release)
}