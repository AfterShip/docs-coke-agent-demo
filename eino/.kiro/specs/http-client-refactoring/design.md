# HTTP Client Refactoring Design Document

## Overview

This design document outlines the refactoring of the current HTTP client implementation in `pkg/langfuse` to directly use the `go-resty/resty/v2` library instead of maintaining a separate `HTTPClient` wrapper layer. The goal is to simplify the codebase, reduce maintenance overhead, and leverage resty's built-in features more directly.

## Current Architecture Analysis

### Current Implementation
- **HTTPClient (`pkg/langfuse/api/core/http_client.go`)**: A wrapper around resty that provides:
  - Custom retry logic with circuit breaker
  - Request options abstraction
  - Error handling and classification
  - Timeout management
  - Authentication handling
  
- **ClientWrapper (`pkg/langfuse/api/core/client_wrapper.go`)**: A higher-level wrapper providing:
  - Convenience methods (GetJSON, PostJSON, etc.)
  - Consistent error handling
  - Health check functionality

### Current Usage Pattern
```go
// Current pattern
httpClient := core.NewHTTPClient(config)
apiClient := api.NewAPIClient(config) // Uses httpClient internally
```

### Issues with Current Design
1. **Double wrapping**: Resty → HTTPClient → ClientWrapper creates unnecessary abstraction layers
2. **Feature duplication**: Custom retry logic when resty has built-in retry capabilities
3. **Maintenance overhead**: Need to maintain compatibility with resty updates
4. **Complexity**: Multiple abstraction layers make debugging difficult

## Architecture

### Target Architecture
Direct usage of resty client with minimal configuration helpers:

```go
// Target pattern
restyClient := resty.New()
// Configure directly with helper functions
configureAuthentication(restyClient, config)
configureRetry(restyClient, config)
configureCircuitBreaker(restyClient, config)
```

### Design Principles
1. **Direct resty usage**: Eliminate wrapper layers where possible
2. **Configuration helpers**: Provide utility functions for common configurations
3. **Backward compatibility**: Maintain existing API surface where reasonable
4. **Leverage resty features**: Use built-in retry, circuit breaker, and middleware features

## Components and Interfaces

### 1. Resty Client Configuration
Replace `HTTPClient` with direct resty configuration:

```go
// pkg/langfuse/api/core/client_config.go
package core

import (
    "github.com/go-resty/resty/v2"
    "eino/pkg/langfuse/config"
)

// ConfigureRestyClient configures a resty client with Langfuse-specific settings
func ConfigureRestyClient(client *resty.Client, cfg *config.Config) error {
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
```

### 2. API Client Refactoring
Simplify APIClient to use resty directly:

```go
// pkg/langfuse/api/client.go
package api

import (
    "github.com/go-resty/resty/v2"
    "eino/pkg/langfuse/api/core"
    "eino/pkg/langfuse/config"
)

type APIClient struct {
    client *resty.Client
    config *config.Config
    
    // Resource clients
    Health    *health.Client
    Ingestion *ingestion.Client
    // ... other resources
}

func NewAPIClient(config *config.Config) (*APIClient, error) {
    client := resty.New()
    defer client.Close() // Ensure cleanup
    
    if err := core.ConfigureRestyClient(client, config); err != nil {
        return nil, err
    }

    return &APIClient{
        client:    client,
        config:    config,
        Health:    health.NewClient(client),
        Ingestion: ingestion.NewClient(client),
        // ... initialize other resources
    }, nil
}
```

### 3. Resource Client Refactoring
Update resource clients to use resty directly:

```go
// pkg/langfuse/api/resources/ingestion/client.go
package ingestion

import (
    "context"
    "github.com/go-resty/resty/v2"
    "eino/pkg/langfuse/api/resources/ingestion/types"
)

type Client struct {
    client *resty.Client
}

func NewClient(client *resty.Client) *Client {
    return &Client{client: client}
}

func (c *Client) Submit(ctx context.Context, req *types.IngestionRequest) (*types.IngestionResponse, error) {
    var response types.IngestionResponse
    
    _, err := c.client.R().
        SetContext(ctx).
        SetBody(req).
        SetResult(&response).
        Post("/api/public/ingestion")
    
    return &response, err
}
```

### 4. Configuration Helpers
Provide utility functions for common configurations:

```go
// pkg/langfuse/api/core/middleware.go
package core

import (
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

// createErrorHandler creates an error handling middleware
func createErrorHandler() resty.ResponseMiddleware {
    return func(c *resty.Client, r *resty.Response) error {
        if r.StatusCode() >= 400 {
            return parseHTTPError(r)
        }
        return nil
    }
}
```

## Data Models

### Configuration Model
Update config to align with resty's capabilities:

```go
// pkg/langfuse/config/config.go
type Config struct {
    // Connection settings
    Host            string
    PublicKey       string
    SecretKey       string
    
    // HTTP settings
    Timeout         time.Duration
    HTTPUserAgent   string
    
    // Retry settings (leverage resty's retry mechanism)
    RetryCount      int
    RetryDelay      time.Duration
    MaxRetryDelay   time.Duration
    
    // Debug settings
    Debug           bool
    
    // Circuit breaker settings (if using resty's circuit breaker)
    CircuitBreakerEnabled bool
    // ... other settings
}
```

## Error Handling

### Error Classification
Simplify error handling using resty's built-in features:

```go
// pkg/langfuse/internal/utils/errors.go
func parseHTTPError(r *resty.Response) error {
    switch r.StatusCode() {
    case 400:
        return NewValidationErrorWithValue("bad_request", "invalid request parameters", r.String())
    case 401:
        return NewNetworkError("authentication_failed", "authentication failed", 
            fmt.Errorf("HTTP %d: %s", r.StatusCode(), r.String()))
    // ... other cases
    default:
        return NewNetworkError("http_error", fmt.Sprintf("HTTP %d error", r.StatusCode()),
            fmt.Errorf("HTTP %d: %s", r.StatusCode(), r.String()))
    }
}
```

### Middleware-based Error Handling
Use resty's middleware for consistent error processing:

```go
func setupErrorMiddleware(client *resty.Client) {
    client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
        if r.StatusCode() >= 400 {
            return parseHTTPError(r)
        }
        return nil
    })
}
```

## Testing Strategy

### Unit Testing
1. **Configuration Testing**: Test resty client configuration with various config combinations
2. **Middleware Testing**: Test custom middleware functions in isolation
3. **Resource Client Testing**: Test resource clients with mock resty responses

### Integration Testing
1. **End-to-end Testing**: Test complete API flows using the refactored client
2. **Retry Testing**: Verify retry behavior using test servers
3. **Error Handling Testing**: Test error scenarios and proper error propagation

### Migration Testing
1. **Backward Compatibility**: Ensure existing tests continue to pass
2. **Performance Testing**: Compare performance before and after refactoring
3. **Feature Parity**: Verify all current features work as expected

### Test Structure
```go
// pkg/langfuse/api/core/client_config_test.go
func TestConfigureRestyClient(t *testing.T) {
    tests := []struct {
        name   string
        config *config.Config
        want   func(*resty.Client) bool
    }{
        {
            name:   "basic configuration",
            config: &config.Config{Host: "https://api.example.com"},
            want:   func(c *resty.Client) bool { return c.BaseURL() == "https://api.example.com" },
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := resty.New()
            defer client.Close()
            
            err := ConfigureRestyClient(client, tt.config)
            assert.NoError(t, err)
            assert.True(t, tt.want(client))
        })
    }
}
```

## Migration Plan

### Phase 1: Preparation
1. Create new configuration helpers
2. Update dependencies in go.mod
3. Create compatibility layer for smooth transition

### Phase 2: Core Refactoring
1. Replace HTTPClient with direct resty usage in APIClient
2. Update resource clients to accept *resty.Client
3. Implement new middleware and error handling

### Phase 3: Cleanup
1. Remove HTTPClient and ClientWrapper
2. Remove unused abstractions
3. Update documentation and examples

### Phase 4: Testing and Validation
1. Run comprehensive test suite
2. Performance benchmarking
3. Integration testing with existing applications

### Migration Checklist
- [ ] Update go.mod dependencies
- [ ] Create configuration helper functions
- [ ] Refactor APIClient to use resty directly
- [ ] Update all resource clients
- [ ] Implement error handling middleware
- [ ] Remove HTTPClient wrapper
- [ ] Remove ClientWrapper
- [ ] Update tests
- [ ] Update documentation
- [ ] Performance validation

## Benefits and Trade-offs

### Benefits
1. **Reduced Complexity**: Eliminate unnecessary abstraction layers
2. **Better Performance**: Reduced overhead from wrapper functions
3. **Easier Maintenance**: Less code to maintain and debug
4. **Feature Access**: Direct access to resty's latest features
5. **Better Documentation**: Leverage resty's extensive documentation

### Trade-offs
1. **API Changes**: Some method signatures may change
2. **Learning Curve**: Developers need to understand resty's API
3. **Migration Effort**: Requires careful migration of existing code
4. **Less Control**: Some custom logic may need to be reimplemented using resty's patterns

## Conclusion

This refactoring will significantly simplify the HTTP client architecture while maintaining functionality. By directly using resty's battle-tested features, we reduce maintenance overhead and improve performance. The migration should be done incrementally to ensure backward compatibility and thorough testing.