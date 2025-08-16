package prompts

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
	"eino/pkg/langfuse/api/resources/prompts/types"
	"github.com/go-resty/resty/v2"
)

const (
	promptsBasePath       = "/api/public/v2/prompts"
	promptsUsageStatsPath = "/api/public/v2/prompts/usage-stats"
)

// Client handles prompt-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new prompts client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// List retrieves a list of prompts
func (c *Client) List(ctx context.Context, req *types.GetPromptsRequest) (*types.GetPromptsResponse, error) {
	if req == nil {
		req = &types.GetPromptsRequest{}
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Build query parameters
	queryParams := make(map[string]string)

	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}

	if req.Page != nil {
		queryParams["page"] = strconv.Itoa(*req.Page)
	}

	if req.Limit != nil {
		queryParams["limit"] = strconv.Itoa(*req.Limit)
	}

	if req.Name != nil {
		queryParams["name"] = *req.Name
	}

	if req.Version != nil {
		queryParams["version"] = strconv.Itoa(*req.Version)
	}

	if req.Type != nil {
		queryParams["type"] = *req.Type
	}

	if req.IsActive != nil {
		queryParams["isActive"] = strconv.FormatBool(*req.IsActive)
	}

	if len(req.Labels) > 0 {
		queryParams["labels"] = strings.Join(req.Labels, ",")
	}

	if len(req.Tags) > 0 {
		queryParams["tags"] = strings.Join(req.Tags, ",")
	}

	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	response := &types.GetPromptsResponse{}

	request := c.client.R().
		SetContext(ctx).
		SetResult(response)

	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}

	_, err := request.Get(promptsBasePath)

	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	return response, nil
}

// Get retrieves a specific prompt by name and version
func (c *Client) Get(ctx context.Context, name string, version *int) (*types.Prompt, error) {
	if name == "" {
		return nil, fmt.Errorf("prompt name cannot be empty")
	}

	response := &types.Prompt{}

	path := fmt.Sprintf("%s/%s", promptsBasePath, url.PathEscape(name))

	queryParams := make(map[string]string)
	if version != nil {
		queryParams["version"] = strconv.Itoa(*version)
	}

	request := c.client.R().
		SetContext(ctx).
		SetResult(response)

	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}

	_, err := request.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt %s: %w", name, err)
	}

	return response, nil
}

// GetByID retrieves a specific prompt by ID
func (c *Client) GetByID(ctx context.Context, promptID string) (*types.Prompt, error) {
	if promptID == "" {
		return nil, fmt.Errorf("prompt ID cannot be empty")
	}

	response := &types.Prompt{}

	path := fmt.Sprintf("%s/%s", promptsBasePath, url.PathEscape(promptID))

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("failed to get prompt %s: %w", promptID, err)
	}

	return response, nil
}

// Create creates a new text prompt
func (c *Client) Create(ctx context.Context, req *types.CreatePromptRequest) (*types.CreatePromptResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	response := &types.CreatePromptResponse{}

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(promptsBasePath)

	if err != nil {
		return nil, fmt.Errorf("failed to create prompt: %w", err)
	}

	return response, nil
}

// CreateChat creates a new chat prompt
func (c *Client) CreateChat(ctx context.Context, req *types.CreateChatPromptRequest) (*types.CreatePromptResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Convert chat prompt request to standard create request
	createReq := &types.CreatePromptRequest{
		Name:        req.Name,
		Description: req.Description,
		Type:        "chat",
		Config:      req.Config,
		Labels:      req.Labels,
		Tags:        req.Tags,
		IsActive:    req.IsActive,
		Metadata:    req.Metadata,
	}

	// Set the messages as prompt content (JSON string)
	// In a real implementation, this might be handled differently
	createReq.Prompt = fmt.Sprintf("chat_messages:%+v", req.Messages)

	response := &types.CreatePromptResponse{}

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(createReq).
		SetResult(response).
		Post(promptsBasePath)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat prompt: %w", err)
	}

	return response, nil
}

// Update updates an existing prompt
func (c *Client) Update(ctx context.Context, promptID string, req *types.UpdatePromptRequest) (*types.Prompt, error) {
	if promptID == "" {
		return nil, fmt.Errorf("prompt ID cannot be empty")
	}

	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	response := &types.Prompt{}

	path := fmt.Sprintf("%s/%s", promptsBasePath, url.PathEscape(promptID))

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)

	if err != nil {
		return nil, fmt.Errorf("failed to update prompt %s: %w", promptID, err)
	}

	return response, nil
}

// UpdateChat updates an existing chat prompt
func (c *Client) UpdateChat(ctx context.Context, promptID string, req *types.UpdateChatPromptRequest) (*types.Prompt, error) {
	if promptID == "" {
		return nil, fmt.Errorf("prompt ID cannot be empty")
	}

	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Convert chat update request to standard update request
	updateReq := &types.UpdatePromptRequest{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		Labels:      req.Labels,
		Tags:        req.Tags,
		IsActive:    req.IsActive,
		Metadata:    req.Metadata,
	}

	// Set the messages as prompt content if provided
	if req.Messages != nil {
		promptContent := fmt.Sprintf("chat_messages:%+v", req.Messages)
		updateReq.Prompt = &promptContent
	}

	response := &types.Prompt{}

	path := fmt.Sprintf("%s/%s", promptsBasePath, url.PathEscape(promptID))

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(updateReq).
		SetResult(response).
		Patch(path)

	if err != nil {
		return nil, fmt.Errorf("failed to update chat prompt %s: %w", promptID, err)
	}

	return response, nil
}

// Delete deletes a prompt by ID
func (c *Client) Delete(ctx context.Context, promptID string) error {
	if promptID == "" {
		return fmt.Errorf("prompt ID cannot be empty")
	}

	path := fmt.Sprintf("%s/%s", promptsBasePath, url.PathEscape(promptID))

	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)

	if err != nil {
		return fmt.Errorf("failed to delete prompt %s: %w", promptID, err)
	}

	return nil
}

// Deploy sets a specific version of a prompt as active
func (c *Client) Deploy(ctx context.Context, promptID string, req *types.PromptDeploymentRequest) (*types.PromptDeploymentResponse, error) {
	if promptID == "" {
		return nil, fmt.Errorf("prompt ID cannot be empty")
	}

	if req == nil {
		req = &types.PromptDeploymentRequest{}
	}

	response := &types.PromptDeploymentResponse{}

	path := fmt.Sprintf("%s/%s/deploy", promptsBasePath, url.PathEscape(promptID))

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(path)

	if err != nil {
		return nil, fmt.Errorf("failed to deploy prompt %s: %w", promptID, err)
	}

	return response, nil
}

// GetVersions retrieves all versions of a prompt
func (c *Client) GetVersions(ctx context.Context, promptName string) ([]types.Prompt, error) {
	if promptName == "" {
		return nil, fmt.Errorf("prompt name cannot be empty")
	}

	req := &types.GetPromptsRequest{
		Name:  &promptName,
		Limit: func() *int { l := 1000; return &l }(), // Get all versions
	}

	response, err := c.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions for prompt %s: %w", promptName, err)
	}

	return response.Data, nil
}

// GetLatest retrieves the latest version of a prompt by name
func (c *Client) GetLatest(ctx context.Context, promptName string) (*types.Prompt, error) {
	return c.Get(ctx, promptName, nil) // nil version gets latest
}

// GetActive retrieves the currently active version of a prompt
func (c *Client) GetActive(ctx context.Context, promptName string) (*types.Prompt, error) {
	req := &types.GetPromptsRequest{
		Name:     &promptName,
		IsActive: func() *bool { b := true; return &b }(),
		Limit:    func() *int { l := 1; return &l }(),
	}

	response, err := c.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get active version for prompt %s: %w", promptName, err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no active version found for prompt %s", promptName)
	}

	return &response.Data[0], nil
}

// GetUsageStats retrieves usage statistics for prompts
func (c *Client) GetUsageStats(ctx context.Context, req *types.GetPromptUsageStatsRequest) (*types.PromptUsageStats, error) {
	if req == nil {
		req = &types.GetPromptUsageStatsRequest{}
	}

	// Build query parameters
	queryParams := make(map[string]string)

	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}

	if req.PromptName != nil {
		queryParams["promptName"] = *req.PromptName
	}

	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	response := &types.PromptUsageStats{}

	request := c.client.R().
		SetContext(ctx).
		SetResult(response)

	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}

	_, err := request.Get(promptsUsageStatsPath)

	if err != nil {
		return nil, fmt.Errorf("failed to get prompt usage stats: %w", err)
	}

	return response, nil
}

// Exists checks if a prompt exists
func (c *Client) Exists(ctx context.Context, promptName string) (bool, error) {
	if promptName == "" {
		return false, fmt.Errorf("prompt name cannot be empty")
	}

	_, err := c.GetLatest(ctx, promptName)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ExistsByID checks if a prompt exists by ID
func (c *Client) ExistsByID(ctx context.Context, promptID string) (bool, error) {
	if promptID == "" {
		return false, fmt.Errorf("prompt ID cannot be empty")
	}

	_, err := c.GetByID(ctx, promptID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CreateTextPrompt creates a simple text prompt
func (c *Client) CreateTextPrompt(ctx context.Context, name, prompt string) (*types.CreatePromptResponse, error) {
	req := types.NewTextPromptRequest(name, prompt)
	return c.Create(ctx, req)
}

// CreateChatPromptWithMessages creates a chat prompt with messages
func (c *Client) CreateChatPromptWithMessages(ctx context.Context, name string, messages []types.ChatMessage) (*types.CreatePromptResponse, error) {
	req := types.NewChatPromptRequest(name, messages)
	return c.CreateChat(ctx, req)
}

// SearchByTag searches for prompts by tag
func (c *Client) SearchByTag(ctx context.Context, tag string) (*types.GetPromptsResponse, error) {
	req := &types.GetPromptsRequest{
		Tags: []string{tag},
	}

	return c.List(ctx, req)
}

// SearchByLabel searches for prompts by label
func (c *Client) SearchByLabel(ctx context.Context, label string) (*types.GetPromptsResponse, error) {
	req := &types.GetPromptsRequest{
		Labels: []string{label},
	}

	return c.List(ctx, req)
}

// ListActive lists all active prompts
func (c *Client) ListActive(ctx context.Context) (*types.GetPromptsResponse, error) {
	req := &types.GetPromptsRequest{
		IsActive: func() *bool { b := true; return &b }(),
	}

	return c.List(ctx, req)
}

// ListByType lists all prompts of a specific type
func (c *Client) ListByType(ctx context.Context, promptType string) (*types.GetPromptsResponse, error) {
	req := &types.GetPromptsRequest{
		Type: &promptType,
	}

	return c.List(ctx, req)
}
