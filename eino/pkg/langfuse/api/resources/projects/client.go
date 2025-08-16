package projects

import (
	"context"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
	"eino/pkg/langfuse/api/resources/projects/types"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/url"
	"strconv"
)

const (
	projectsBasePath      = "/api/public/projects"
	projectByIDPath       = "/api/public/projects/%s"
	projectApiKeysPath    = "/api/public/projects/%s/api-keys"
	projectApiKeyByIDPath = "/api/public/projects/%s/api-keys/%s"
	projectUsagePath      = "/api/public/projects/%s/usage"
)

// Client handles project-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new projects client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// List retrieves a list of projects
func (c *Client) List(ctx context.Context, req *types.GetProjectsRequest) (*types.GetProjectsResponse, error) {
	if req == nil {
		req = &types.GetProjectsRequest{}
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Build query parameters
	queryParams := make(map[string]string)

	if req.OrganizationID != "" {
		queryParams["organizationId"] = req.OrganizationID
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

	if req.IsActive != nil {
		queryParams["isActive"] = strconv.FormatBool(*req.IsActive)
	}

	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	if req.IncludeStats != nil {
		queryParams["includeStats"] = strconv.FormatBool(*req.IncludeStats)
	}

	response := &types.GetProjectsResponse{}

	request := c.client.R().
		SetContext(ctx).
		SetResult(response)

	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}

	_, err := request.Get(projectsBasePath)

	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return response, nil
}

// Get retrieves a specific project by ID
func (c *Client) Get(ctx context.Context, projectID string) (*types.Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	response := &types.Project{}

	path := fmt.Sprintf(projectByIDPath, url.PathEscape(projectID))

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectID, err)
	}

	return response, nil
}

// Create creates a new project
func (c *Client) Create(ctx context.Context, req *types.CreateProjectRequest) (*types.CreateProjectResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	response := &types.CreateProjectResponse{}

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(projectsBasePath)

	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return response, nil
}

// Update updates an existing project
func (c *Client) Update(ctx context.Context, projectID string, req *types.UpdateProjectRequest) (*types.Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	response := &types.Project{}

	path := fmt.Sprintf(projectByIDPath, url.PathEscape(projectID))

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)

	if err != nil {
		return nil, fmt.Errorf("failed to update project %s: %w", projectID, err)
	}

	return response, nil
}

// Delete deletes a project by ID
func (c *Client) Delete(ctx context.Context, projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}

	path := fmt.Sprintf(projectByIDPath, url.PathEscape(projectID))

	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)

	if err != nil {
		return fmt.Errorf("failed to delete project %s: %w", projectID, err)
	}

	return nil
}

// ListApiKeys retrieves API keys for a project
func (c *Client) ListApiKeys(ctx context.Context, projectID string) ([]types.ProjectApiKey, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	var response []types.ProjectApiKey

	path := fmt.Sprintf(projectApiKeysPath, url.PathEscape(projectID))

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("failed to list API keys for project %s: %w", projectID, err)
	}

	return response, nil
}

// CreateApiKey creates a new API key for a project
func (c *Client) CreateApiKey(ctx context.Context, projectID string, req *types.CreateApiKeyRequest) (*types.CreateApiKeyResponse, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	response := &types.CreateApiKeyResponse{}

	path := fmt.Sprintf(projectApiKeysPath, url.PathEscape(projectID))

	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(path)

	if err != nil {
		return nil, fmt.Errorf("failed to create API key for project %s: %w", projectID, err)
	}

	return response, nil
}

// GetApiKey retrieves a specific API key
func (c *Client) GetApiKey(ctx context.Context, projectID, keyID string) (*types.ProjectApiKey, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	if keyID == "" {
		return nil, fmt.Errorf("key ID cannot be empty")
	}

	response := &types.ProjectApiKey{}

	path := fmt.Sprintf(projectApiKeyByIDPath, url.PathEscape(projectID), url.PathEscape(keyID))

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("failed to get API key %s for project %s: %w", keyID, projectID, err)
	}

	return response, nil
}

// DeleteApiKey deletes an API key
func (c *Client) DeleteApiKey(ctx context.Context, projectID, keyID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}

	if keyID == "" {
		return fmt.Errorf("key ID cannot be empty")
	}

	path := fmt.Sprintf(projectApiKeyByIDPath, url.PathEscape(projectID), url.PathEscape(keyID))

	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)

	if err != nil {
		return fmt.Errorf("failed to delete API key %s for project %s: %w", keyID, projectID, err)
	}

	return nil
}

// GetUsage retrieves usage statistics for a project
func (c *Client) GetUsage(ctx context.Context, projectID string, req *types.GetProjectUsageRequest) (*types.ProjectUsageResponse, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	if req == nil {
		req = &types.GetProjectUsageRequest{}
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Build query parameters
	queryParams := make(map[string]string)

	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}

	if req.Granularity != nil {
		queryParams["granularity"] = *req.Granularity
	}

	response := &types.ProjectUsageResponse{}

	path := fmt.Sprintf(projectUsagePath, url.PathEscape(projectID))

	request := c.client.R().
		SetContext(ctx).
		SetResult(response)

	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}

	_, err := request.Get(path)

	if err != nil {
		return nil, fmt.Errorf("failed to get usage for project %s: %w", projectID, err)
	}

	return response, nil
}

// Exists checks if a project exists
func (c *Client) Exists(ctx context.Context, projectID string) (bool, error) {
	if projectID == "" {
		return false, fmt.Errorf("project ID cannot be empty")
	}

	_, err := c.Get(ctx, projectID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ListActive lists all active projects
func (c *Client) ListActive(ctx context.Context) (*types.GetProjectsResponse, error) {
	req := &types.GetProjectsRequest{
		IsActive: func() *bool { b := true; return &b }(),
	}

	return c.List(ctx, req)
}

// ListByOrganization lists projects for a specific organization
func (c *Client) ListByOrganization(ctx context.Context, organizationID string) (*types.GetProjectsResponse, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID cannot be empty")
	}

	req := &types.GetProjectsRequest{
		OrganizationID: organizationID,
	}

	return c.List(ctx, req)
}

// FindByName finds projects by name (partial match)
func (c *Client) FindByName(ctx context.Context, name string) (*types.GetProjectsResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	req := &types.GetProjectsRequest{
		Name: &name,
	}

	return c.List(ctx, req)
}

// CreateSimple creates a simple project with just a name
func (c *Client) CreateSimple(ctx context.Context, name string) (*types.CreateProjectResponse, error) {
	req := types.NewCreateProjectRequest(name)
	return c.Create(ctx, req)
}

// CreateWithSettings creates a project with specific settings
func (c *Client) CreateWithSettings(ctx context.Context, name string, settings *types.ProjectSettings) (*types.CreateProjectResponse, error) {
	req := &types.CreateProjectRequest{
		Name:     name,
		Settings: settings,
	}

	return c.Create(ctx, req)
}

// ActivateProject activates a project
func (c *Client) ActivateProject(ctx context.Context, projectID string) (*types.Project, error) {
	req := &types.UpdateProjectRequest{
		IsActive: func() *bool { b := true; return &b }(),
	}

	return c.Update(ctx, projectID, req)
}

// DeactivateProject deactivates a project
func (c *Client) DeactivateProject(ctx context.Context, projectID string) (*types.Project, error) {
	req := &types.UpdateProjectRequest{
		IsActive: func() *bool { b := false; return &b }(),
	}

	return c.Update(ctx, projectID, req)
}

// UpdateSettings updates project settings
func (c *Client) UpdateSettings(ctx context.Context, projectID string, settings *types.ProjectSettings) (*types.Project, error) {
	req := &types.UpdateProjectRequest{
		Settings: settings,
	}

	return c.Update(ctx, projectID, req)
}

// EnableFeature enables a specific feature for a project
func (c *Client) EnableFeature(ctx context.Context, projectID, feature string) (*types.Project, error) {
	// Get current project to preserve existing settings
	project, err := c.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current project: %w", err)
	}

	settings := project.Settings
	if settings == nil {
		settings = &types.ProjectSettings{}
	}

	// Enable the specified feature
	switch feature {
	case "datasets":
		settings.DatasetEnabled = true
	case "prompts":
		settings.PromptManagementEnabled = true
	case "evaluation":
		settings.EvaluationEnabled = true
	default:
		return nil, fmt.Errorf("unknown feature: %s", feature)
	}

	return c.UpdateSettings(ctx, projectID, settings)
}

// DisableFeature disables a specific feature for a project
func (c *Client) DisableFeature(ctx context.Context, projectID, feature string) (*types.Project, error) {
	// Get current project to preserve existing settings
	project, err := c.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current project: %w", err)
	}

	settings := project.Settings
	if settings == nil {
		settings = &types.ProjectSettings{}
	}

	// Disable the specified feature
	switch feature {
	case "datasets":
		settings.DatasetEnabled = false
	case "prompts":
		settings.PromptManagementEnabled = false
	case "evaluation":
		settings.EvaluationEnabled = false
	default:
		return nil, fmt.Errorf("unknown feature: %s", feature)
	}

	return c.UpdateSettings(ctx, projectID, settings)
}

// CreateApiKeySimple creates a simple API key with just a name
func (c *Client) CreateApiKeySimple(ctx context.Context, projectID, name string) (*types.CreateApiKeyResponse, error) {
	req := types.NewCreateApiKeyRequest(name)
	return c.CreateApiKey(ctx, projectID, req)
}

// GetProjectStats gets statistics for a project (with stats included)
func (c *Client) GetProjectStats(ctx context.Context, projectID string) (*types.Project, error) {
	// For a real implementation, this might be a separate endpoint
	// For now, we'll use the regular Get method
	return c.Get(ctx, projectID)
}

// ListWithStats lists projects including statistics
func (c *Client) ListWithStats(ctx context.Context) (*types.GetProjectsResponse, error) {
	req := &types.GetProjectsRequest{
		IncludeStats: func() *bool { b := true; return &b }(),
	}

	return c.List(ctx, req)
}
