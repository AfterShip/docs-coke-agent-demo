package types

import (
	"fmt"
	"time"

	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

const (
	PromptTypeChatMessage = "chatmessage"
)

// Prompt represents a prompt template for LLM interactions
type Prompt struct {
	// Unique identifier for the prompt
	ID string `json:"id"`

	// Name of the prompt
	Name string `json:"name"`

	// Version of the prompt
	Version int `json:"version"`

	// Description of the prompt
	Description *string `json:"description,omitempty"`

	// Project ID this prompt belongs to
	ProjectID string `json:"projectId"`

	// Prompt template content
	Prompt []ChatMessage `json:"prompt"`

	// Prompt type (e.g., "text", "chat")
	Type string `json:"type"`

	// Configuration for the prompt
	Config *PromptConfig `json:"config,omitempty"`

	// Labels associated with the prompt
	Labels []string `json:"labels,omitempty"`

	// Tags for categorization
	Tags []string `json:"tags,omitempty"`

	// Whether this is the production version
	IsActive bool `json:"isActive"`

	// Metadata associated with the prompt
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamp when the prompt was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the prompt was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// User who created the prompt
	CreatedBy *string `json:"createdBy,omitempty"`
}

// PromptConfig represents configuration for a prompt
type PromptConfig struct {
	// Model configuration
	Model *string `json:"model,omitempty"`

	// Temperature for generation
	Temperature *float64 `json:"temperature,omitempty"`

	// Maximum tokens to generate
	MaxTokens *int `json:"maxTokens,omitempty"`

	// Top-p sampling parameter
	TopP *float64 `json:"topP,omitempty"`

	// Frequency penalty
	FrequencyPenalty *float64 `json:"frequencyPenalty,omitempty"`

	// Presence penalty
	PresencePenalty *float64 `json:"presencePenalty,omitempty"`

	// Stop sequences
	Stop []string `json:"stop,omitempty"`

	// Additional model parameters
	ModelParameters map[string]interface{} `json:"modelParameters,omitempty"`
}

// ChatMessage represents a message in a chat prompt
type ChatMessage struct {
	// Role of the message sender
	Role string `json:"role"` // "system", "user", "assistant"

	// Content of the message
	Content string `json:"content"`

	// type of the message
	Type string `json:"type,omitempty"` // "chatmessage"
}

// GetPromptsRequest represents a request to list prompts
type GetPromptsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	Page          *int       `json:"page,omitempty"`
	Limit         *int       `json:"limit,omitempty"`
	Name          *string    `json:"name,omitempty"`
	Version       *int       `json:"version,omitempty"`
	Type          *string    `json:"type,omitempty"`
	IsActive      *bool      `json:"isActive,omitempty"`
	Labels        []string   `json:"labels,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// GetPromptsResponse represents the response from listing prompts
type GetPromptsResponse struct {
	Data []Prompt           `json:"data"`
	Meta types.MetaResponse `json:"meta"`
}

// CreatePromptRequest represents a request to create a prompt
type CreatePromptRequest struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Prompt      string                 `json:"prompt"`
	Type        string                 `json:"type"` // "text", "chat"
	Config      *PromptConfig          `json:"config,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	IsActive    *bool                  `json:"isActive,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateChatPromptRequest represents a request to create a chat prompt
type CreateChatPromptRequest struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Messages    []ChatMessage          `json:"messages"`
	Config      *PromptConfig          `json:"config,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	IsActive    *bool                  `json:"isActive,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreatePromptResponse represents the response from creating a prompt
type CreatePromptResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     int                    `json:"version"`
	Description *string                `json:"description,omitempty"`
	ProjectID   string                 `json:"projectId"`
	Prompt      string                 `json:"prompt"`
	Type        string                 `json:"type"`
	Config      *PromptConfig          `json:"config,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	CreatedBy   *string                `json:"createdBy,omitempty"`
}

// UpdatePromptRequest represents a request to update a prompt
type UpdatePromptRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Prompt      *string                `json:"prompt,omitempty"`
	Config      *PromptConfig          `json:"config,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	IsActive    *bool                  `json:"isActive,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateChatPromptRequest represents a request to update a chat prompt
type UpdateChatPromptRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Messages    []ChatMessage          `json:"messages,omitempty"`
	Config      *PromptConfig          `json:"config,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	IsActive    *bool                  `json:"isActive,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptVersionRequest represents a request to get a specific version of a prompt
type PromptVersionRequest struct {
	Name    string `json:"name"`
	Version *int   `json:"version,omitempty"` // If nil, gets latest version
}

// PromptUsageStats represents usage statistics for prompts
type PromptUsageStats struct {
	TotalPrompts  int                `json:"totalPrompts"`
	ActivePrompts int                `json:"activePrompts"`
	TotalVersions int                `json:"totalVersions"`
	PromptsByType map[string]int     `json:"promptsByType"`
	UsageByPrompt map[string]int     `json:"usageByPrompt"`
	RecentlyUsed  []PromptUsageEntry `json:"recentlyUsed"`
	DateRange     *DateRange         `json:"dateRange,omitempty"`
}

// PromptUsageEntry represents a single prompt usage entry
type PromptUsageEntry struct {
	PromptID   string    `json:"promptId"`
	PromptName string    `json:"promptName"`
	Version    int       `json:"version"`
	UsageCount int       `json:"usageCount"`
	LastUsed   time.Time `json:"lastUsed"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// GetPromptUsageStatsRequest represents a request to get prompt usage statistics
type GetPromptUsageStatsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	PromptName    *string    `json:"promptName,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// PromptDeploymentRequest represents a request to deploy a prompt version
type PromptDeploymentRequest struct {
	Version *int `json:"version,omitempty"` // If nil, deploys latest version
}

// PromptDeploymentResponse represents the response from deploying a prompt
type PromptDeploymentResponse struct {
	PromptID        string    `json:"promptId"`
	Name            string    `json:"promptName"`
	DeployedVersion int       `json:"deployedVersion"`
	PreviousVersion *int      `json:"previousVersion,omitempty"`
	DeployedAt      time.Time `json:"deployedAt"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validate validates the GetPromptsRequest
func (req *GetPromptsRequest) Validate() error {
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}

	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}

	if req.Version != nil && *req.Version < 1 {
		return &ValidationError{Field: "version", Message: "version must be greater than 0"}
	}

	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}

	if req.Type != nil && (*req.Type != "text" && *req.Type != "chat") {
		return &ValidationError{Field: "type", Message: "type must be 'text' or 'chat'"}
	}

	return nil
}

// Validate validates the CreatePromptRequest
func (req *CreatePromptRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if len(req.Name) > 255 {
		return &ValidationError{Field: "name", Message: "name must be 255 characters or less"}
	}

	if req.Prompt == "" {
		return &ValidationError{Field: "prompt", Message: "prompt content is required"}
	}

	if req.Type != "text" && req.Type != "chat" {
		return &ValidationError{Field: "type", Message: "type must be 'text' or 'chat'"}
	}

	if req.Description != nil && len(*req.Description) > 2000 {
		return &ValidationError{Field: "description", Message: "description must be 2000 characters or less"}
	}

	if req.Config != nil {
		if err := req.Config.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the CreateChatPromptRequest
func (req *CreateChatPromptRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if len(req.Name) > 255 {
		return &ValidationError{Field: "name", Message: "name must be 255 characters or less"}
	}

	if len(req.Messages) == 0 {
		return &ValidationError{Field: "messages", Message: "at least one message is required"}
	}

	for i, msg := range req.Messages {
		if msg.Role == "" {
			return &ValidationError{Field: fmt.Sprintf("messages[%d].role", i), Message: "role is required"}
		}

		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return &ValidationError{Field: fmt.Sprintf("messages[%d].role", i), Message: "role must be 'system', 'user', or 'assistant'"}
		}

		if msg.Content == "" {
			return &ValidationError{Field: fmt.Sprintf("messages[%d].content", i), Message: "content is required"}
		}
	}

	if req.Description != nil && len(*req.Description) > 2000 {
		return &ValidationError{Field: "description", Message: "description must be 2000 characters or less"}
	}

	if req.Config != nil {
		if err := req.Config.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the PromptVersionRequest
func (req *PromptVersionRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if req.Version != nil && *req.Version < 1 {
		return &ValidationError{Field: "version", Message: "version must be greater than 0"}
	}

	return nil
}

// Validate validates the PromptConfig
func (config *PromptConfig) Validate() error {
	if config.Temperature != nil && (*config.Temperature < 0 || *config.Temperature > 2) {
		return &ValidationError{Field: "config.temperature", Message: "temperature must be between 0 and 2"}
	}

	if config.MaxTokens != nil && *config.MaxTokens < 1 {
		return &ValidationError{Field: "config.maxTokens", Message: "maxTokens must be greater than 0"}
	}

	if config.TopP != nil && (*config.TopP < 0 || *config.TopP > 1) {
		return &ValidationError{Field: "config.topP", Message: "topP must be between 0 and 1"}
	}

	if config.FrequencyPenalty != nil && (*config.FrequencyPenalty < -2 || *config.FrequencyPenalty > 2) {
		return &ValidationError{Field: "config.frequencyPenalty", Message: "frequencyPenalty must be between -2 and 2"}
	}

	if config.PresencePenalty != nil && (*config.PresencePenalty < -2 || *config.PresencePenalty > 2) {
		return &ValidationError{Field: "config.presencePenalty", Message: "presencePenalty must be between -2 and 2"}
	}

	return nil
}

// NewTextPromptRequest creates a new text prompt request
func NewTextPromptRequest(name, prompt string) *CreatePromptRequest {
	return &CreatePromptRequest{
		Name:   name,
		Prompt: prompt,
		Type:   "text",
	}
}

// NewChatPromptRequest creates a new chat prompt request
func NewChatPromptRequest(name string, messages []ChatMessage) *CreateChatPromptRequest {
	return &CreateChatPromptRequest{
		Name:     name,
		Messages: messages,
	}
}

// NewSystemMessage creates a system message
func NewSystemMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    "system",
		Content: content,
	}
}

// NewUserMessage creates a user message
func NewUserMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    "user",
		Content: content,
	}
}

// NewAssistantMessage creates an assistant message
func NewAssistantMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    "assistant",
		Content: content,
	}
}

// IsText returns true if the prompt is a text prompt
func (p *Prompt) IsText() bool {
	return p.Type == "text"
}

// IsChat returns true if the prompt is a chat prompt
func (p *Prompt) IsChat() bool {
	return p.Type == "chat"
}

// HasLabel returns true if the prompt has the specified label
func (p *Prompt) HasLabel(label string) bool {
	for _, l := range p.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// HasTag returns true if the prompt has the specified tag
func (p *Prompt) HasTag(tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
