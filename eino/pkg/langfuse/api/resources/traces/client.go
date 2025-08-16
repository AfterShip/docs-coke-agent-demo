package traces

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/traces/types"
	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

const (
	tracesBasePath  = "/api/public/traces"
	traceByIDPath   = "/api/public/traces/%s"
	tracesStatsPath = "/api/public/traces/stats"
)

// Client handles trace-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new traces client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// List retrieves a list of traces based on the provided filters
func (c *Client) List(ctx context.Context, req *types.GetTracesRequest) (*types.GetTracesResponse, error) {
	if req == nil {
		req = &types.GetTracesRequest{}
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
	
	if req.UserID != nil {
		queryParams["userId"] = *req.UserID
	}
	
	if req.Name != nil {
		queryParams["name"] = *req.Name
	}
	
	if req.SessionID != nil {
		queryParams["sessionId"] = *req.SessionID
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.OrderBy != nil {
		queryParams["orderBy"] = *req.OrderBy
	}
	
	if len(req.Tags) > 0 {
		queryParams["tags"] = strings.Join(req.Tags, ",")
	}
	
	response := &types.GetTracesResponse{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(tracesBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list traces: %w", err)
	}
	
	return response, nil
}

// Get retrieves a specific trace by ID
func (c *Client) Get(ctx context.Context, traceID string) (*commonTypes.Trace, error) {
	if traceID == "" {
		return nil, fmt.Errorf("trace ID cannot be empty")
	}
	
	response := &commonTypes.Trace{}
	
	path := fmt.Sprintf(traceByIDPath, url.PathEscape(traceID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get trace %s: %w", traceID, err)
	}
	
	return response, nil
}

// GetWithObservations retrieves a trace with all its observations
func (c *Client) GetWithObservations(ctx context.Context, traceID string) (*types.TraceWithObservations, error) {
	if traceID == "" {
		return nil, fmt.Errorf("trace ID cannot be empty")
	}
	
	response := &types.TraceWithObservations{}
	
	path := fmt.Sprintf(traceByIDPath, url.PathEscape(traceID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("includeObservations", "true").
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get trace with observations %s: %w", traceID, err)
	}
	
	return response, nil
}

// Create creates a new trace
func (c *Client) Create(ctx context.Context, req *types.CreateTraceRequest) (*commonTypes.Trace, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &commonTypes.Trace{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(tracesBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create trace: %w", err)
	}
	
	return response, nil
}

// Update updates an existing trace
func (c *Client) Update(ctx context.Context, req *types.UpdateTraceRequest) (*commonTypes.Trace, error) {
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &commonTypes.Trace{}
	
	path := fmt.Sprintf(traceByIDPath, url.PathEscape(req.TraceID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update trace %s: %w", req.TraceID, err)
	}
	
	return response, nil
}

// Delete deletes a trace by ID
func (c *Client) Delete(ctx context.Context, traceID string) (*types.DeleteTraceResponse, error) {
	if traceID == "" {
		return nil, fmt.Errorf("trace ID cannot be empty")
	}
	
	response := &types.DeleteTraceResponse{}
	
	path := fmt.Sprintf(traceByIDPath, url.PathEscape(traceID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Delete(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to delete trace %s: %w", traceID, err)
	}
	
	return response, nil
}

// GetStats retrieves statistics for traces
func (c *Client) GetStats(ctx context.Context, req *types.GetTraceStatsRequest) (*types.TraceStats, error) {
	if req == nil {
		req = &types.GetTraceStatsRequest{}
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}
	
	if req.UserID != nil {
		queryParams["userId"] = *req.UserID
	}
	
	if req.SessionID != nil {
		queryParams["sessionId"] = *req.SessionID
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if len(req.Tags) > 0 {
		queryParams["tags"] = strings.Join(req.Tags, ",")
	}
	
	response := &types.TraceStats{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(tracesStatsPath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get trace stats: %w", err)
	}
	
	return response, nil
}

// ListPaginated retrieves traces with enhanced pagination and filtering
func (c *Client) ListPaginated(ctx context.Context, req *types.PaginatedTracesRequest) (*types.GetTracesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("paginated request cannot be nil")
	}
	
	// Convert to standard GetTracesRequest
	getReq := &types.GetTracesRequest{
		ProjectID: req.ProjectID,
		Page:      &req.Page,
		Limit:     &req.Limit,
	}
	
	if req.SortOrder != "" {
		orderBy := string(req.SortOrder)
		getReq.OrderBy = &orderBy
	}
	
	if req.Filter != nil {
		if len(req.Filter.UserIDs) == 1 {
			getReq.UserID = &req.Filter.UserIDs[0]
		}
		
		if len(req.Filter.SessionIDs) == 1 {
			getReq.SessionID = &req.Filter.SessionIDs[0]
		}
		
		if len(req.Filter.Names) == 1 {
			getReq.Name = &req.Filter.Names[0]
		}
		
		getReq.Tags = req.Filter.Tags
		getReq.FromTimestamp = req.Filter.FromTimestamp
		getReq.ToTimestamp = req.Filter.ToTimestamp
	}
	
	return c.List(ctx, getReq)
}

// Exists checks if a trace exists
func (c *Client) Exists(ctx context.Context, traceID string) (bool, error) {
	if traceID == "" {
		return false, fmt.Errorf("trace ID cannot be empty")
	}
	
	_, err := c.Get(ctx, traceID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// ListBySession retrieves all traces for a specific session
func (c *Client) ListBySession(ctx context.Context, sessionID string, limit int) (*types.GetTracesResponse, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}
	
	req := &types.GetTracesRequest{
		SessionID: &sessionID,
		Limit:     &limit,
	}
	
	return c.List(ctx, req)
}

// ListByUser retrieves all traces for a specific user
func (c *Client) ListByUser(ctx context.Context, userID string, limit int) (*types.GetTracesResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	
	req := &types.GetTracesRequest{
		UserID: &userID,
		Limit:  &limit,
	}
	
	return c.List(ctx, req)
}