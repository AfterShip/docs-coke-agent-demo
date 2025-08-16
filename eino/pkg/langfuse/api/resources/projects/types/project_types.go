package types

import (
	"time"

	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

// Project represents a project in Langfuse
type Project struct {
	// Unique identifier for the project
	ID string `json:"id"`

	// Name of the project
	Name string `json:"name"`

	// Description of the project
	Description *string `json:"description,omitempty"`

	// Organization ID this project belongs to
	OrganizationID string `json:"organizationId"`

	// Public key for API access
	PublicKey string `json:"publicKey"`

	// Whether the project is active
	IsActive bool `json:"isActive"`

	// Settings for the project
	Settings *ProjectSettings `json:"settings,omitempty"`

	// Metadata associated with the project
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamp when the project was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the project was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// User who created the project
	CreatedBy *string `json:"createdBy,omitempty"`

	// Project statistics
	Stats *ProjectStats `json:"stats,omitempty"`
}

// ProjectSettings represents settings for a project
type ProjectSettings struct {
	// Data retention period in days
	DataRetentionDays *int `json:"dataRetentionDays,omitempty"`

	// Whether to enable sample rate limiting
	SampleRateLimit *float64 `json:"sampleRateLimit,omitempty"`

	// Maximum traces per month
	MaxTracesPerMonth *int `json:"maxTracesPerMonth,omitempty"`

	// Whether to enable evaluation features
	EvaluationEnabled bool `json:"evaluationEnabled"`

	// Whether to enable dataset features
	DatasetEnabled bool `json:"datasetEnabled"`

	// Whether to enable prompt management
	PromptManagementEnabled bool `json:"promptManagementEnabled"`

	// Webhook configuration
	WebhookConfig *WebhookConfig `json:"webhookConfig,omitempty"`

	// Notification settings
	NotificationSettings *NotificationSettings `json:"notificationSettings,omitempty"`

	// Custom configuration
	CustomConfig map[string]interface{} `json:"customConfig,omitempty"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	// Whether webhooks are enabled
	Enabled bool `json:"enabled"`

	// Webhook URL
	URL *string `json:"url,omitempty"`

	// Secret for webhook authentication
	Secret *string `json:"secret,omitempty"`

	// Events to send webhooks for
	Events []WebhookEvent `json:"events,omitempty"`

	// Headers to include in webhook requests
	Headers map[string]string `json:"headers,omitempty"`
}

// WebhookEvent represents webhook event types
type WebhookEvent string

const (
	WebhookEventTraceCreate WebhookEvent = "trace.create"
	WebhookEventTraceUpdate WebhookEvent = "trace.update"
	WebhookEventScoreCreate WebhookEvent = "score.create"
	WebhookEventScoreUpdate WebhookEvent = "score.update"
)

// NotificationSettings represents notification settings
type NotificationSettings struct {
	// Whether email notifications are enabled
	EmailEnabled bool `json:"emailEnabled"`

	// Whether Slack notifications are enabled
	SlackEnabled bool `json:"slackEnabled"`

	// Slack webhook URL
	SlackWebhookURL *string `json:"slackWebhookUrl,omitempty"`

	// Notification thresholds
	Thresholds *NotificationThresholds `json:"thresholds,omitempty"`
}

// NotificationThresholds represents thresholds for notifications
type NotificationThresholds struct {
	// Error rate threshold (percentage)
	ErrorRate *float64 `json:"errorRate,omitempty"`

	// Cost threshold per month
	CostThreshold *float64 `json:"costThreshold,omitempty"`

	// Usage threshold (traces per day)
	UsageThreshold *int `json:"usageThreshold,omitempty"`
}

// ProjectStats represents statistics for a project
type ProjectStats struct {
	// Total number of traces
	TotalTraces int `json:"totalTraces"`

	// Total number of observations
	TotalObservations int `json:"totalObservations"`

	// Total number of generations
	TotalGenerations int `json:"totalGenerations"`

	// Total number of scores
	TotalScores int `json:"totalScores"`

	// Total number of datasets
	TotalDatasets int `json:"totalDatasets"`

	// Total number of prompts
	TotalPrompts int `json:"totalPrompts"`

	// Total cost
	TotalCost *float64 `json:"totalCost,omitempty"`

	// Currency for cost
	Currency *string `json:"currency,omitempty"`

	// Last activity timestamp
	LastActivity *time.Time `json:"lastActivity,omitempty"`

	// Monthly statistics
	MonthlyStats *MonthlyStats `json:"monthlyStats,omitempty"`
}

// MonthlyStats represents monthly statistics
type MonthlyStats struct {
	// Current month traces
	TracesThisMonth int `json:"tracesThisMonth"`

	// Cost this month
	CostThisMonth *float64 `json:"costThisMonth,omitempty"`

	// Usage by day this month
	DailyUsage []DailyUsage `json:"dailyUsage,omitempty"`
}

// DailyUsage represents daily usage statistics
type DailyUsage struct {
	Date   time.Time `json:"date"`
	Traces int       `json:"traces"`
	Cost   *float64  `json:"cost,omitempty"`
}

// GetProjectsRequest represents a request to list projects
type GetProjectsRequest struct {
	OrganizationID string     `json:"organizationId,omitempty"`
	Page           *int       `json:"page,omitempty"`
	Limit          *int       `json:"limit,omitempty"`
	Name           *string    `json:"name,omitempty"`
	IsActive       *bool      `json:"isActive,omitempty"`
	FromTimestamp  *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp    *time.Time `json:"toTimestamp,omitempty"`
	IncludeStats   *bool      `json:"includeStats,omitempty"`
}

// GetProjectsResponse represents the response from listing projects
type GetProjectsResponse struct {
	Data []Project          `json:"data"`
	Meta types.MetaResponse `json:"meta"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Settings    *ProjectSettings       `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateProjectResponse represents the response from creating a project
type CreateProjectResponse struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	OrganizationID string                 `json:"organizationId"`
	PublicKey      string                 `json:"publicKey"`
	SecretKey      string                 `json:"secretKey"` // Only returned on creation
	IsActive       bool                   `json:"isActive"`
	Settings       *ProjectSettings       `json:"settings,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	CreatedBy      *string                `json:"createdBy,omitempty"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	IsActive    *bool                  `json:"isActive,omitempty"`
	Settings    *ProjectSettings       `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectApiKey represents an API key for a project
type ProjectApiKey struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"projectId"`
	PublicKey string     `json:"publicKey"`
	Name      string     `json:"name"`
	Note      *string    `json:"note,omitempty"`
	IsActive  bool       `json:"isActive"`
	LastUsed  *time.Time `json:"lastUsed,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	CreatedBy *string    `json:"createdBy,omitempty"`
}

// CreateApiKeyRequest represents a request to create an API key
type CreateApiKeyRequest struct {
	Name string  `json:"name"`
	Note *string `json:"note,omitempty"`
}

// CreateApiKeyResponse represents the response from creating an API key
type CreateApiKeyResponse struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	PublicKey string    `json:"publicKey"`
	SecretKey string    `json:"secretKey"` // Only returned on creation
	Name      string    `json:"name"`
	Note      *string   `json:"note,omitempty"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy *string   `json:"createdBy,omitempty"`
}

// GetProjectUsageRequest represents a request to get project usage
type GetProjectUsageRequest struct {
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	Granularity   *string    `json:"granularity,omitempty"` // "day", "hour"
}

// ProjectUsageResponse represents project usage data
type ProjectUsageResponse struct {
	ProjectID    string             `json:"projectId"`
	TotalTraces  int                `json:"totalTraces"`
	TotalCost    *float64           `json:"totalCost,omitempty"`
	Currency     *string            `json:"currency,omitempty"`
	UsageByDay   []DailyUsage       `json:"usageByDay,omitempty"`
	UsageByModel map[string]int     `json:"usageByModel,omitempty"`
	CostByModel  map[string]float64 `json:"costByModel,omitempty"`
	DateRange    *DateRange         `json:"dateRange,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validate validates the GetProjectsRequest
func (req *GetProjectsRequest) Validate() error {
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}

	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}

	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}

	return nil
}

// Validate validates the CreateProjectRequest
func (req *CreateProjectRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if len(req.Name) > 255 {
		return &ValidationError{Field: "name", Message: "name must be 255 characters or less"}
	}

	if req.Description != nil && len(*req.Description) > 2000 {
		return &ValidationError{Field: "description", Message: "description must be 2000 characters or less"}
	}

	if req.Settings != nil {
		if err := req.Settings.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the CreateApiKeyRequest
func (req *CreateApiKeyRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if len(req.Name) > 255 {
		return &ValidationError{Field: "name", Message: "name must be 255 characters or less"}
	}

	if req.Note != nil && len(*req.Note) > 1000 {
		return &ValidationError{Field: "note", Message: "note must be 1000 characters or less"}
	}

	return nil
}

// Validate validates the GetProjectUsageRequest
func (req *GetProjectUsageRequest) Validate() error {
	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}

	if req.Granularity != nil && (*req.Granularity != "day" && *req.Granularity != "hour") {
		return &ValidationError{Field: "granularity", Message: "granularity must be 'day' or 'hour'"}
	}

	return nil
}

// Validate validates the ProjectSettings
func (settings *ProjectSettings) Validate() error {
	if settings.DataRetentionDays != nil && *settings.DataRetentionDays < 1 {
		return &ValidationError{Field: "settings.dataRetentionDays", Message: "dataRetentionDays must be greater than 0"}
	}

	if settings.SampleRateLimit != nil && (*settings.SampleRateLimit < 0 || *settings.SampleRateLimit > 1) {
		return &ValidationError{Field: "settings.sampleRateLimit", Message: "sampleRateLimit must be between 0 and 1"}
	}

	if settings.MaxTracesPerMonth != nil && *settings.MaxTracesPerMonth < 1 {
		return &ValidationError{Field: "settings.maxTracesPerMonth", Message: "maxTracesPerMonth must be greater than 0"}
	}

	if settings.WebhookConfig != nil {
		if err := settings.WebhookConfig.Validate(); err != nil {
			return err
		}
	}

	if settings.NotificationSettings != nil {
		if err := settings.NotificationSettings.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the WebhookConfig
func (config *WebhookConfig) Validate() error {
	if config.Enabled && config.URL == nil {
		return &ValidationError{Field: "webhookConfig.url", Message: "url is required when webhooks are enabled"}
	}

	// Validate events
	validEvents := map[WebhookEvent]bool{
		WebhookEventTraceCreate: true,
		WebhookEventTraceUpdate: true,
		WebhookEventScoreCreate: true,
		WebhookEventScoreUpdate: true,
	}

	for _, event := range config.Events {
		if !validEvents[event] {
			return &ValidationError{Field: "webhookConfig.events", Message: "invalid webhook event: " + string(event)}
		}
	}

	return nil
}

// Validate validates the NotificationSettings
func (settings *NotificationSettings) Validate() error {
	if settings.SlackEnabled && settings.SlackWebhookURL == nil {
		return &ValidationError{Field: "notificationSettings.slackWebhookUrl", Message: "slackWebhookUrl is required when Slack is enabled"}
	}

	if settings.Thresholds != nil {
		if err := settings.Thresholds.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the NotificationThresholds
func (thresholds *NotificationThresholds) Validate() error {
	if thresholds.ErrorRate != nil && (*thresholds.ErrorRate < 0 || *thresholds.ErrorRate > 100) {
		return &ValidationError{Field: "thresholds.errorRate", Message: "errorRate must be between 0 and 100"}
	}

	if thresholds.CostThreshold != nil && *thresholds.CostThreshold < 0 {
		return &ValidationError{Field: "thresholds.costThreshold", Message: "costThreshold cannot be negative"}
	}

	if thresholds.UsageThreshold != nil && *thresholds.UsageThreshold < 1 {
		return &ValidationError{Field: "thresholds.usageThreshold", Message: "usageThreshold must be greater than 0"}
	}

	return nil
}

// Active returns true if the project is active
func (p *Project) Active() bool {
	return p.IsActive
}

// HasDatasets returns true if dataset features are enabled
func (p *Project) HasDatasets() bool {
	return p.Settings != nil && p.Settings.DatasetEnabled
}

// HasPromptManagement returns true if prompt management is enabled
func (p *Project) HasPromptManagement() bool {
	return p.Settings != nil && p.Settings.PromptManagementEnabled
}

// HasEvaluation returns true if evaluation features are enabled
func (p *Project) HasEvaluation() bool {
	return p.Settings != nil && p.Settings.EvaluationEnabled
}

// GetRetentionDays returns the data retention period in days
func (p *Project) GetRetentionDays() int {
	if p.Settings != nil && p.Settings.DataRetentionDays != nil {
		return *p.Settings.DataRetentionDays
	}
	return 90 // Default retention period
}

// GetSampleRate returns the sample rate limit
func (p *Project) GetSampleRate() float64 {
	if p.Settings != nil && p.Settings.SampleRateLimit != nil {
		return *p.Settings.SampleRateLimit
	}
	return 1.0 // Default no sampling
}

// NewCreateProjectRequest creates a new project creation request
func NewCreateProjectRequest(name string) *CreateProjectRequest {
	return &CreateProjectRequest{
		Name: name,
	}
}

// NewCreateApiKeyRequest creates a new API key creation request
func NewCreateApiKeyRequest(name string) *CreateApiKeyRequest {
	return &CreateApiKeyRequest{
		Name: name,
	}
}
