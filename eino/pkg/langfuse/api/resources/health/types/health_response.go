package types

import "time"

// HealthResponse represents the response from the health check endpoint
type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Version     string                 `json:"version,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      *time.Duration         `json:"uptime,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Services    map[string]ServiceHealth `json:"services,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus represents the health status
type HealthStatus string

const (
	// HealthStatusHealthy indicates the service is healthy
	HealthStatusHealthy HealthStatus = "healthy"
	
	// HealthStatusUnhealthy indicates the service is unhealthy
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	
	// HealthStatusDegraded indicates the service is degraded
	HealthStatusDegraded HealthStatus = "degraded"
	
	// HealthStatusMaintenance indicates the service is in maintenance mode
	HealthStatusMaintenance HealthStatus = "maintenance"
)

// ServiceHealth represents the health status of a specific service component
type ServiceHealth struct {
	Status      HealthStatus `json:"status"`
	LastChecked time.Time    `json:"lastChecked,omitempty"`
	Message     string       `json:"message,omitempty"`
	ResponseTime *time.Duration `json:"responseTime,omitempty"`
}

// IsHealthy returns true if the overall health status is healthy
func (hr *HealthResponse) IsHealthy() bool {
	return hr.Status == HealthStatusHealthy
}

// IsUnhealthy returns true if the overall health status is unhealthy
func (hr *HealthResponse) IsUnhealthy() bool {
	return hr.Status == HealthStatusUnhealthy
}

// IsDegraded returns true if the overall health status is degraded
func (hr *HealthResponse) IsDegraded() bool {
	return hr.Status == HealthStatusDegraded
}

// IsInMaintenance returns true if the service is in maintenance mode
func (hr *HealthResponse) IsInMaintenance() bool {
	return hr.Status == HealthStatusMaintenance
}

// GetServiceHealth returns the health status of a specific service
func (hr *HealthResponse) GetServiceHealth(serviceName string) (ServiceHealth, bool) {
	if hr.Services == nil {
		return ServiceHealth{}, false
	}
	
	health, exists := hr.Services[serviceName]
	return health, exists
}

// HasUnhealthyServices returns true if any service is unhealthy
func (hr *HealthResponse) HasUnhealthyServices() bool {
	if hr.Services == nil {
		return false
	}
	
	for _, service := range hr.Services {
		if service.Status == HealthStatusUnhealthy {
			return true
		}
	}
	
	return false
}

// GetUnhealthyServices returns a list of unhealthy service names
func (hr *HealthResponse) GetUnhealthyServices() []string {
	if hr.Services == nil {
		return nil
	}
	
	var unhealthy []string
	for name, service := range hr.Services {
		if service.Status == HealthStatusUnhealthy {
			unhealthy = append(unhealthy, name)
		}
	}
	
	return unhealthy
}