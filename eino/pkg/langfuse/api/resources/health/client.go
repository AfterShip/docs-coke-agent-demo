package health

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/health/types"
)

// API path constants
const (
	healthBasePath = "/api/public/health"
)

// Client handles health check operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new health client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// Check performs a health check against the Langfuse API
func (c *Client) Check(ctx context.Context) (*types.HealthResponse, error) {
	response := &types.HealthResponse{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(healthBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("health check request failed: %w", err)
	}
	
	return response, nil
}

// CheckWithTimeout performs a health check with a specific timeout
func (c *Client) CheckWithTimeout(timeout time.Duration) (*types.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return c.Check(ctx)
}

// IsHealthy performs a simple health check and returns true if the service is healthy
func (c *Client) IsHealthy(ctx context.Context) (bool, error) {
	response, err := c.Check(ctx)
	if err != nil {
		return false, err
	}
	
	return response.IsHealthy(), nil
}

// WaitForHealthy waits for the service to become healthy within the given timeout
func (c *Client) WaitForHealthy(ctx context.Context, checkInterval time.Duration) error {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			healthy, err := c.IsHealthy(ctx)
			if err != nil {
				// Continue checking even if there's an error
				continue
			}
			
			if healthy {
				return nil
			}
		}
	}
}

// CheckLiveness performs a basic liveness check (simple ping)
func (c *Client) CheckLiveness(ctx context.Context) error {
	_, err := c.client.R().
		SetContext(ctx).
		Get(healthBasePath)
	
	if err != nil {
		return fmt.Errorf("liveness check failed: %w", err)
	}
	
	return nil
}

// CheckReadiness performs a readiness check to ensure the service can handle requests
func (c *Client) CheckReadiness(ctx context.Context) error {
	response, err := c.Check(ctx)
	if err != nil {
		return fmt.Errorf("readiness check failed: %w", err)
	}
	
	if !response.IsHealthy() {
		return fmt.Errorf("service is not ready: status=%s", response.Status)
	}
	
	// Check if any critical services are unhealthy
	if response.HasUnhealthyServices() {
		unhealthyServices := response.GetUnhealthyServices()
		return fmt.Errorf("critical services are unhealthy: %v", unhealthyServices)
	}
	
	return nil
}

// GetServiceHealth returns the health status of a specific service component
func (c *Client) GetServiceHealth(ctx context.Context, serviceName string) (*types.ServiceHealth, error) {
	response, err := c.Check(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get service health: %w", err)
	}
	
	serviceHealth, exists := response.GetServiceHealth(serviceName)
	if !exists {
		return nil, fmt.Errorf("service '%s' not found in health response", serviceName)
	}
	
	return &serviceHealth, nil
}

// Monitor continuously monitors the health status and calls the provided callback
func (c *Client) Monitor(ctx context.Context, interval time.Duration, callback func(*types.HealthResponse, error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	// Initial check
	response, err := c.Check(ctx)
	callback(response, err)
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			response, err := c.Check(ctx)
			callback(response, err)
		}
	}
}