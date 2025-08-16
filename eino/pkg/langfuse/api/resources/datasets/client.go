package datasets

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/datasets/types"
	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

const (
	datasetsBasePath         = "/api/public/datasets"
	datasetsWithIDPath       = "/api/public/datasets/%s"
	datasetsItemsPath        = "/api/public/datasets/%s/items"
	datasetsItemWithIDPath   = "/api/public/datasets/%s/items/%s"
	datasetsRunsPath         = "/api/public/datasets/%s/runs"
	datasetsRunWithIDPath    = "/api/public/datasets/%s/runs/%s"
	datasetsRunItemsPath     = "/api/public/datasets/%s/runs/%s/items"
	datasetsStatsPath        = "/api/public/datasets/stats"
)

// Client handles dataset-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new datasets client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// List retrieves a list of datasets
func (c *Client) List(ctx context.Context, req *types.GetDatasetsRequest) (*types.GetDatasetsResponse, error) {
	if req == nil {
		req = &types.GetDatasetsRequest{}
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
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	response := &types.GetDatasetsResponse{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(datasetsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}
	
	return response, nil
}

// Get retrieves a specific dataset by ID
func (c *Client) Get(ctx context.Context, datasetID string) (*commonTypes.Dataset, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	response := &commonTypes.Dataset{}
	
	path := fmt.Sprintf(datasetsWithIDPath, url.PathEscape(datasetID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset %s: %w", datasetID, err)
	}
	
	return response, nil
}

// Create creates a new dataset
func (c *Client) Create(ctx context.Context, req *types.CreateDatasetRequest) (*types.CreateDatasetResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.CreateDatasetResponse{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(datasetsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}
	
	return response, nil
}

// Update updates an existing dataset
func (c *Client) Update(ctx context.Context, datasetID string, req *types.UpdateDatasetRequest) (*commonTypes.Dataset, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}
	
	response := &commonTypes.Dataset{}
	
	path := fmt.Sprintf(datasetsWithIDPath, url.PathEscape(datasetID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update dataset %s: %w", datasetID, err)
	}
	
	return response, nil
}

// Delete deletes a dataset by ID
func (c *Client) Delete(ctx context.Context, datasetID string) error {
	if datasetID == "" {
		return fmt.Errorf("dataset ID cannot be empty")
	}
	
	path := fmt.Sprintf(datasetsWithIDPath, url.PathEscape(datasetID))
	
	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)
	
	if err != nil {
		return fmt.Errorf("failed to delete dataset %s: %w", datasetID, err)
	}
	
	return nil
}

// ListItems retrieves items for a specific dataset
func (c *Client) ListItems(ctx context.Context, datasetID string, req *types.GetDatasetItemsRequest) (*types.GetDatasetItemsResponse, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if req == nil {
		req = &types.GetDatasetItemsRequest{DatasetID: datasetID}
	} else {
		req.DatasetID = datasetID // Ensure consistency
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.Page != nil {
		queryParams["page"] = strconv.Itoa(*req.Page)
	}
	
	if req.Limit != nil {
		queryParams["limit"] = strconv.Itoa(*req.Limit)
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	response := &types.GetDatasetItemsResponse{}
	
	path := fmt.Sprintf(datasetsItemsPath, url.PathEscape(datasetID))
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list items for dataset %s: %w", datasetID, err)
	}
	
	return response, nil
}

// GetItem retrieves a specific dataset item by ID
func (c *Client) GetItem(ctx context.Context, datasetID, itemID string) (*commonTypes.DatasetItem, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if itemID == "" {
		return nil, fmt.Errorf("item ID cannot be empty")
	}
	
	response := &commonTypes.DatasetItem{}
	
	path := fmt.Sprintf(datasetsItemWithIDPath, url.PathEscape(datasetID), url.PathEscape(itemID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get item %s from dataset %s: %w", itemID, datasetID, err)
	}
	
	return response, nil
}

// CreateItem creates a new dataset item
func (c *Client) CreateItem(ctx context.Context, datasetID string, req *types.CreateDatasetItemRequest) (*types.CreateDatasetItemResponse, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.CreateDatasetItemResponse{}
	
	path := fmt.Sprintf(datasetsItemsPath, url.PathEscape(datasetID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create item in dataset %s: %w", datasetID, err)
	}
	
	return response, nil
}

// UpdateItem updates an existing dataset item
func (c *Client) UpdateItem(ctx context.Context, datasetID, itemID string, req *types.UpdateDatasetItemRequest) (*commonTypes.DatasetItem, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if itemID == "" {
		return nil, fmt.Errorf("item ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}
	
	response := &commonTypes.DatasetItem{}
	
	path := fmt.Sprintf(datasetsItemWithIDPath, url.PathEscape(datasetID), url.PathEscape(itemID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update item %s in dataset %s: %w", itemID, datasetID, err)
	}
	
	return response, nil
}

// DeleteItem deletes a dataset item by ID
func (c *Client) DeleteItem(ctx context.Context, datasetID, itemID string) error {
	if datasetID == "" {
		return fmt.Errorf("dataset ID cannot be empty")
	}
	
	if itemID == "" {
		return fmt.Errorf("item ID cannot be empty")
	}
	
	path := fmt.Sprintf(datasetsItemWithIDPath, url.PathEscape(datasetID), url.PathEscape(itemID))
	
	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)
	
	if err != nil {
		return fmt.Errorf("failed to delete item %s from dataset %s: %w", itemID, datasetID, err)
	}
	
	return nil
}

// ListRuns retrieves runs for a specific dataset
func (c *Client) ListRuns(ctx context.Context, datasetID string, req *types.GetDatasetRunsRequest) (*types.GetDatasetRunsResponse, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if req == nil {
		req = &types.GetDatasetRunsRequest{DatasetID: datasetID}
	} else {
		req.DatasetID = datasetID // Ensure consistency
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.Page != nil {
		queryParams["page"] = strconv.Itoa(*req.Page)
	}
	
	if req.Limit != nil {
		queryParams["limit"] = strconv.Itoa(*req.Limit)
	}
	
	if req.Name != nil {
		queryParams["name"] = *req.Name
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	response := &types.GetDatasetRunsResponse{}
	
	path := fmt.Sprintf(datasetsRunsPath, url.PathEscape(datasetID))
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list runs for dataset %s: %w", datasetID, err)
	}
	
	return response, nil
}

// GetRun retrieves a specific dataset run by ID
func (c *Client) GetRun(ctx context.Context, datasetID, runID string) (*commonTypes.DatasetRun, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if runID == "" {
		return nil, fmt.Errorf("run ID cannot be empty")
	}
	
	response := &commonTypes.DatasetRun{}
	
	path := fmt.Sprintf(datasetsRunWithIDPath, url.PathEscape(datasetID), url.PathEscape(runID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get run %s from dataset %s: %w", runID, datasetID, err)
	}
	
	return response, nil
}

// CreateRun creates a new dataset run
func (c *Client) CreateRun(ctx context.Context, datasetID string, req *types.CreateDatasetRunRequest) (*types.CreateDatasetRunResponse, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.CreateDatasetRunResponse{}
	
	path := fmt.Sprintf(datasetsRunsPath, url.PathEscape(datasetID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create run in dataset %s: %w", datasetID, err)
	}
	
	return response, nil
}

// ListRunItems retrieves items for a specific dataset run
func (c *Client) ListRunItems(ctx context.Context, datasetID, runID string, req *types.GetDatasetRunItemsRequest) (*types.GetDatasetRunItemsResponse, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if runID == "" {
		return nil, fmt.Errorf("run ID cannot be empty")
	}
	
	if req == nil {
		req = &types.GetDatasetRunItemsRequest{DatasetRunID: runID}
	} else {
		req.DatasetRunID = runID // Ensure consistency
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.Page != nil {
		queryParams["page"] = strconv.Itoa(*req.Page)
	}
	
	if req.Limit != nil {
		queryParams["limit"] = strconv.Itoa(*req.Limit)
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	response := &types.GetDatasetRunItemsResponse{}
	
	path := fmt.Sprintf(datasetsRunItemsPath, url.PathEscape(datasetID), url.PathEscape(runID))
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list items for run %s in dataset %s: %w", runID, datasetID, err)
	}
	
	return response, nil
}

// CreateRunItem creates a new dataset run item
func (c *Client) CreateRunItem(ctx context.Context, datasetID, runID string, req *types.CreateDatasetRunItemRequest) (*types.CreateDatasetRunItemResponse, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset ID cannot be empty")
	}
	
	if runID == "" {
		return nil, fmt.Errorf("run ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.CreateDatasetRunItemResponse{}
	
	path := fmt.Sprintf(datasetsRunItemsPath, url.PathEscape(datasetID), url.PathEscape(runID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create run item in dataset %s run %s: %w", datasetID, runID, err)
	}
	
	return response, nil
}

// GetStats retrieves statistics for datasets
func (c *Client) GetStats(ctx context.Context, req *types.GetDatasetStatsRequest) (*types.DatasetStats, error) {
	if req == nil {
		req = &types.GetDatasetStatsRequest{}
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	response := &types.DatasetStats{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(datasetsStatsPath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset stats: %w", err)
	}
	
	return response, nil
}

// Exists checks if a dataset exists
func (c *Client) Exists(ctx context.Context, datasetID string) (bool, error) {
	if datasetID == "" {
		return false, fmt.Errorf("dataset ID cannot be empty")
	}
	
	_, err := c.Get(ctx, datasetID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// CreateFromTrace creates a dataset item from a trace
func (c *Client) CreateFromTrace(ctx context.Context, datasetID, traceID string, metadata map[string]interface{}) (*types.CreateDatasetItemResponse, error) {
	req := &types.CreateDatasetItemRequest{
		SourceTraceID: &traceID,
		Metadata:      metadata,
	}
	
	return c.CreateItem(ctx, datasetID, req)
}

// CreateFromObservation creates a dataset item from an observation
func (c *Client) CreateFromObservation(ctx context.Context, datasetID, observationID string, metadata map[string]interface{}) (*types.CreateDatasetItemResponse, error) {
	req := &types.CreateDatasetItemRequest{
		SourceObservationID: &observationID,
		Metadata:            metadata,
	}
	
	return c.CreateItem(ctx, datasetID, req)
}