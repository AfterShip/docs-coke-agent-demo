package scores

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/scores/types"
	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

const (
	scoresBasePath      = "/api/public/scores"
	scoreByIDPath       = "/api/public/scores/%s"
	scoresAggregationPath = "/api/public/scores/aggregation"
	scoresStatsPath     = "/api/public/scores/stats"
)

// Client handles score-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new scores client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// Create creates a new score
func (c *Client) Create(ctx context.Context, req *types.CreateScoreRequest) (*types.CreateScoreResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.CreateScoreResponse{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(scoresBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create score: %w", err)
	}
	
	return response, nil
}

// List retrieves a list of scores based on the provided filters
func (c *Client) List(ctx context.Context, req *types.GetScoresRequest) (*types.GetScoresResponse, error) {
	if req == nil {
		req = &types.GetScoresRequest{}
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
	
	if req.TraceID != nil {
		queryParams["traceId"] = *req.TraceID
	}
	
	if req.ObservationID != nil {
		queryParams["observationId"] = *req.ObservationID
	}
	
	if req.Name != nil {
		queryParams["name"] = *req.Name
	}
	
	if req.DataType != nil {
		queryParams["dataType"] = string(*req.DataType)
	}
	
	if req.ConfigID != nil {
		queryParams["configId"] = *req.ConfigID
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.UserID != nil {
		queryParams["userId"] = *req.UserID
	}
	
	if req.Source != nil {
		queryParams["source"] = *req.Source
	}
	
	response := &types.GetScoresResponse{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(scoresBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list scores: %w", err)
	}
	
	return response, nil
}

// Get retrieves a specific score by ID
func (c *Client) Get(ctx context.Context, scoreID string) (*commonTypes.Score, error) {
	if scoreID == "" {
		return nil, fmt.Errorf("score ID cannot be empty")
	}
	
	response := &commonTypes.Score{}
	
	path := fmt.Sprintf(scoreByIDPath, url.PathEscape(scoreID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get score %s: %w", scoreID, err)
	}
	
	return response, nil
}

// Delete deletes a score by ID
func (c *Client) Delete(ctx context.Context, scoreID string) error {
	if scoreID == "" {
		return fmt.Errorf("score ID cannot be empty")
	}
	
	path := fmt.Sprintf(scoreByIDPath, url.PathEscape(scoreID))
	
	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)
	
	if err != nil {
		return fmt.Errorf("failed to delete score %s: %w", scoreID, err)
	}
	
	return nil
}

// GetAggregation retrieves aggregated score data
func (c *Client) GetAggregation(ctx context.Context, req *types.GetScoreAggregationRequest) (*types.GetScoreAggregationResponse, error) {
	if req == nil {
		req = &types.GetScoreAggregationRequest{}
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}
	
	if req.TraceID != nil {
		queryParams["traceId"] = *req.TraceID
	}
	
	if req.ObservationID != nil {
		queryParams["observationId"] = *req.ObservationID
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
	
	if req.UserID != nil {
		queryParams["userId"] = *req.UserID
	}
	
	if len(req.GroupBy) > 0 {
		queryParams["groupBy"] = strings.Join(req.GroupBy, ",")
	}
	
	response := &types.GetScoreAggregationResponse{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(scoresAggregationPath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get score aggregation: %w", err)
	}
	
	return response, nil
}

// GetStats retrieves statistics for scores
func (c *Client) GetStats(ctx context.Context, req *types.GetScoreStatsRequest) (*types.ScoreStats, error) {
	if req == nil {
		req = &types.GetScoreStatsRequest{}
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}
	
	if req.TraceID != nil {
		queryParams["traceId"] = *req.TraceID
	}
	
	if req.ObservationID != nil {
		queryParams["observationId"] = *req.ObservationID
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.UserID != nil {
		queryParams["userId"] = *req.UserID
	}
	
	response := &types.ScoreStats{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(scoresStatsPath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get score stats: %w", err)
	}
	
	return response, nil
}

// ListPaginated retrieves scores with enhanced pagination and filtering
func (c *Client) ListPaginated(ctx context.Context, req *types.PaginatedScoresRequest) (*types.GetScoresResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("paginated request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	// Convert to standard GetScoresRequest
	getReq := &types.GetScoresRequest{
		ProjectID: req.ProjectID,
		Page:      &req.Page,
		Limit:     &req.Limit,
	}
	
	if req.Filter != nil {
		if len(req.Filter.TraceIDs) == 1 {
			getReq.TraceID = &req.Filter.TraceIDs[0]
		}
		
		if len(req.Filter.ObservationIDs) == 1 {
			getReq.ObservationID = &req.Filter.ObservationIDs[0]
		}
		
		if len(req.Filter.Names) == 1 {
			getReq.Name = &req.Filter.Names[0]
		}
		
		if len(req.Filter.DataTypes) == 1 {
			getReq.DataType = &req.Filter.DataTypes[0]
		}
		
		if len(req.Filter.ConfigIDs) == 1 {
			getReq.ConfigID = &req.Filter.ConfigIDs[0]
		}
		
		if len(req.Filter.UserIDs) == 1 {
			getReq.UserID = &req.Filter.UserIDs[0]
		}
		
		if len(req.Filter.Sources) == 1 {
			getReq.Source = &req.Filter.Sources[0]
		}
		
		getReq.FromTimestamp = req.Filter.FromTimestamp
		getReq.ToTimestamp = req.Filter.ToTimestamp
	}
	
	return c.List(ctx, getReq)
}

// ListByTrace retrieves all scores for a specific trace
func (c *Client) ListByTrace(ctx context.Context, traceID string, limit int) (*types.GetScoresResponse, error) {
	if traceID == "" {
		return nil, fmt.Errorf("trace ID cannot be empty")
	}
	
	req := &types.GetScoresRequest{
		TraceID: &traceID,
		Limit:   &limit,
	}
	
	return c.List(ctx, req)
}

// ListByObservation retrieves all scores for a specific observation
func (c *Client) ListByObservation(ctx context.Context, observationID string, limit int) (*types.GetScoresResponse, error) {
	if observationID == "" {
		return nil, fmt.Errorf("observation ID cannot be empty")
	}
	
	req := &types.GetScoresRequest{
		ObservationID: &observationID,
		Limit:         &limit,
	}
	
	return c.List(ctx, req)
}

// CreateNumeric creates a numeric score
func (c *Client) CreateNumeric(ctx context.Context, traceID, name string, value float64) (*types.CreateScoreResponse, error) {
	req := types.NewNumericScoreRequest(traceID, name, value)
	return c.Create(ctx, req)
}

// CreateBoolean creates a boolean score
func (c *Client) CreateBoolean(ctx context.Context, traceID, name string, value bool) (*types.CreateScoreResponse, error) {
	req := types.NewBooleanScoreRequest(traceID, name, value)
	return c.Create(ctx, req)
}

// CreateCategorical creates a categorical score
func (c *Client) CreateCategorical(ctx context.Context, traceID, name, value string) (*types.CreateScoreResponse, error) {
	req := types.NewCategoricalScoreRequest(traceID, name, value)
	return c.Create(ctx, req)
}

// Exists checks if a score exists
func (c *Client) Exists(ctx context.Context, scoreID string) (bool, error) {
	if scoreID == "" {
		return false, fmt.Errorf("score ID cannot be empty")
	}
	
	_, err := c.Get(ctx, scoreID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}