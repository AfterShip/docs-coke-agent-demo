package models

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/models/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

// API path constants
const (
	modelsBasePath        = "/api/public/models"
	modelsItemPath        = "/api/public/models/%s"
	modelsMatchPath       = "/api/public/models/match"
	modelsUsageStatsPath  = "/api/public/models/usage-stats"
)

// Client handles model-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new models client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// List retrieves a list of models
func (c *Client) List(ctx context.Context, req *types.GetModelsRequest) (*types.GetModelsResponse, error) {
	if req == nil {
		req = &types.GetModelsRequest{}
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
	
	if req.ModelName != nil {
		queryParams["modelName"] = *req.ModelName
	}
	
	if req.Provider != nil {
		queryParams["provider"] = *req.Provider
	}
	
	if req.ModelFamily != nil {
		queryParams["modelFamily"] = *req.ModelFamily
	}
	
	if req.Unit != nil {
		queryParams["unit"] = string(*req.Unit)
	}
	
	if req.FromDate != nil {
		queryParams["fromDate"] = req.FromDate.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToDate != nil {
		queryParams["toDate"] = req.ToDate.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.IncludeDeprecated != nil {
		queryParams["includeDeprecated"] = strconv.FormatBool(*req.IncludeDeprecated)
	}
	
	response := &types.GetModelsResponse{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(modelsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	
	return response, nil
}

// Get retrieves a specific model by ID
func (c *Client) Get(ctx context.Context, modelID string) (*types.Model, error) {
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}
	
	response := &types.Model{}
	
	path := fmt.Sprintf(modelsItemPath, url.PathEscape(modelID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get model %s: %w", modelID, err)
	}
	
	return response, nil
}

// Create creates a new model
func (c *Client) Create(ctx context.Context, req *types.CreateModelRequest) (*types.CreateModelResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.CreateModelResponse{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(modelsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}
	
	return response, nil
}

// Update updates an existing model
func (c *Client) Update(ctx context.Context, modelID string, req *types.UpdateModelRequest) (*types.Model, error) {
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}
	
	response := &types.Model{}
	
	path := fmt.Sprintf(modelsItemPath, url.PathEscape(modelID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update model %s: %w", modelID, err)
	}
	
	return response, nil
}

// Delete deletes a model by ID
func (c *Client) Delete(ctx context.Context, modelID string) error {
	if modelID == "" {
		return fmt.Errorf("model ID cannot be empty")
	}
	
	path := fmt.Sprintf(modelsItemPath, url.PathEscape(modelID))
	
	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)
	
	if err != nil {
		return fmt.Errorf("failed to delete model %s: %w", modelID, err)
	}
	
	return nil
}

// Match finds a matching model configuration for a given model name
func (c *Client) Match(ctx context.Context, req *types.ModelMatchRequest) (*types.ModelMatchResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("match request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.ModelMatchResponse{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(modelsMatchPath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to match model: %w", err)
	}
	
	return response, nil
}

// GetUsageStats retrieves usage statistics for models
func (c *Client) GetUsageStats(ctx context.Context, req *types.GetModelUsageStatsRequest) (*types.ModelUsageStats, error) {
	if req == nil {
		req = &types.GetModelUsageStatsRequest{}
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}
	
	if req.ModelID != nil {
		queryParams["modelId"] = *req.ModelID
	}
	
	if req.ModelName != nil {
		queryParams["modelName"] = *req.ModelName
	}
	
	if req.Provider != nil {
		queryParams["provider"] = *req.Provider
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	response := &types.ModelUsageStats{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(modelsUsageStatsPath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get model usage stats: %w", err)
	}
	
	return response, nil
}

// Exists checks if a model exists
func (c *Client) Exists(ctx context.Context, modelID string) (bool, error) {
	if modelID == "" {
		return false, fmt.Errorf("model ID cannot be empty")
	}
	
	_, err := c.Get(ctx, modelID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// FindByName finds models by name (exact match)
func (c *Client) FindByName(ctx context.Context, modelName string) (*types.GetModelsResponse, error) {
	if modelName == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}
	
	req := &types.GetModelsRequest{
		ModelName: &modelName,
	}
	
	return c.List(ctx, req)
}

// FindByProvider finds models by provider
func (c *Client) FindByProvider(ctx context.Context, provider string) (*types.GetModelsResponse, error) {
	if provider == "" {
		return nil, fmt.Errorf("provider cannot be empty")
	}
	
	req := &types.GetModelsRequest{
		Provider: &provider,
	}
	
	return c.List(ctx, req)
}

// ListActive lists all active (non-deprecated) models
func (c *Client) ListActive(ctx context.Context) (*types.GetModelsResponse, error) {
	req := &types.GetModelsRequest{
		IncludeDeprecated: func() *bool { b := false; return &b }(),
	}
	
	return c.List(ctx, req)
}

// ListByUnit lists models by pricing unit
func (c *Client) ListByUnit(ctx context.Context, unit types.ModelPricingUnit) (*types.GetModelsResponse, error) {
	req := &types.GetModelsRequest{
		Unit: &unit,
	}
	
	return c.List(ctx, req)
}

// ListTokenBasedModels lists all token-based pricing models
func (c *Client) ListTokenBasedModels(ctx context.Context) (*types.GetModelsResponse, error) {
	return c.ListByUnit(ctx, types.ModelPricingUnitTokens)
}

// ListRequestBasedModels lists all request-based pricing models
func (c *Client) ListRequestBasedModels(ctx context.Context) (*types.GetModelsResponse, error) {
	return c.ListByUnit(ctx, types.ModelPricingUnitRequests)
}

// CreateTokenBasedModel creates a token-based pricing model
func (c *Client) CreateTokenBasedModel(ctx context.Context, name, matchPattern string, inputPrice, outputPrice float64, currency string) (*types.CreateModelResponse, error) {
	req := types.NewTokenBasedModel(name, matchPattern, inputPrice, outputPrice, currency)
	return c.Create(ctx, req)
}

// CreateRequestBasedModel creates a request-based pricing model
func (c *Client) CreateRequestBasedModel(ctx context.Context, name, matchPattern string, totalPrice float64, currency string) (*types.CreateModelResponse, error) {
	req := types.NewRequestBasedModel(name, matchPattern, totalPrice, currency)
	return c.Create(ctx, req)
}

// MatchModel finds the best matching model configuration for a model name
func (c *Client) MatchModel(ctx context.Context, modelName string) (*types.ModelMatchResponse, error) {
	req := &types.ModelMatchRequest{
		ModelName: modelName,
	}
	
	return c.Match(ctx, req)
}

// DeprecateModel marks a model as deprecated by setting its end date
func (c *Client) DeprecateModel(ctx context.Context, modelID string, endDate *time.Time) (*types.Model, error) {
	if endDate == nil {
		now := time.Now()
		endDate = &now
	}
	
	req := &types.UpdateModelRequest{
		EndDate: endDate,
	}
	
	return c.Update(ctx, modelID, req)
}

// UpdatePricing updates the pricing for a model
func (c *Client) UpdatePricing(ctx context.Context, modelID string, inputPrice, outputPrice *float64, totalPrice *float64) (*types.Model, error) {
	req := &types.UpdateModelRequest{
		InputPrice:  inputPrice,
		OutputPrice: outputPrice,
		TotalPrice:  totalPrice,
	}
	
	return c.Update(ctx, modelID, req)
}

// GetModelsByFamily finds models by model family
func (c *Client) GetModelsByFamily(ctx context.Context, modelFamily string) (*types.GetModelsResponse, error) {
	if modelFamily == "" {
		return nil, fmt.Errorf("model family cannot be empty")
	}
	
	req := &types.GetModelsRequest{
		ModelFamily: &modelFamily,
	}
	
	return c.List(ctx, req)
}