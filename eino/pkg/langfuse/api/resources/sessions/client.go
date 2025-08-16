package sessions

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/sessions/types"
	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

const (
	sessionsBasePath  = "/api/public/sessions"
	sessionByIDPath   = "/api/public/sessions/%s"
	sessionsStatsPath = "/api/public/sessions/stats"
)

// Client handles session-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new sessions client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// List retrieves a list of sessions based on the provided filters
func (c *Client) List(ctx context.Context, req *types.GetSessionsRequest) (*types.GetSessionsResponse, error) {
	if req == nil {
		req = &types.GetSessionsRequest{}
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
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.OrderBy != nil {
		queryParams["orderBy"] = *req.OrderBy
	}
	
	response := &types.GetSessionsResponse{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(sessionsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	
	return response, nil
}

// Get retrieves a specific session by ID
func (c *Client) Get(ctx context.Context, sessionID string) (*commonTypes.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}
	
	response := &commonTypes.Session{}
	
	path := fmt.Sprintf(sessionByIDPath, url.PathEscape(sessionID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get session %s: %w", sessionID, err)
	}
	
	return response, nil
}

// GetWithTraces retrieves a session with all its traces
func (c *Client) GetWithTraces(ctx context.Context, sessionID string) (*types.SessionWithTraces, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}
	
	response := &types.SessionWithTraces{}
	
	path := fmt.Sprintf(sessionByIDPath, url.PathEscape(sessionID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("includeTraces", "true").
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get session with traces %s: %w", sessionID, err)
	}
	
	return response, nil
}

// Create creates a new session
func (c *Client) Create(ctx context.Context, req *types.CreateSessionRequest) (*commonTypes.Session, error) {
	if req == nil {
		req = &types.CreateSessionRequest{}
	}
	
	response := &commonTypes.Session{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(sessionsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	return response, nil
}

// Update updates an existing session
func (c *Client) Update(ctx context.Context, req *types.UpdateSessionRequest) (*commonTypes.Session, error) {
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &commonTypes.Session{}
	
	path := fmt.Sprintf(sessionByIDPath, url.PathEscape(req.SessionID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update session %s: %w", req.SessionID, err)
	}
	
	return response, nil
}

// Delete deletes a session by ID
func (c *Client) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	
	path := fmt.Sprintf(sessionByIDPath, url.PathEscape(sessionID))
	
	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)
	
	if err != nil {
		return fmt.Errorf("failed to delete session %s: %w", sessionID, err)
	}
	
	return nil
}

// GetStats retrieves statistics for sessions
func (c *Client) GetStats(ctx context.Context, req *types.GetSessionStatsRequest) (*types.SessionStats, error) {
	if req == nil {
		req = &types.GetSessionStatsRequest{}
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.ProjectID != "" {
		queryParams["projectId"] = req.ProjectID
	}
	
	if req.UserID != nil {
		queryParams["userId"] = *req.UserID
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	response := &types.SessionStats{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(sessionsStatsPath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}
	
	return response, nil
}

// ListPaginated retrieves sessions with enhanced pagination and filtering
func (c *Client) ListPaginated(ctx context.Context, req *types.PaginatedSessionsRequest) (*types.GetSessionsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("paginated request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	// Convert to standard GetSessionsRequest
	getReq := &types.GetSessionsRequest{
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
		
		getReq.FromTimestamp = req.Filter.FromTimestamp
		getReq.ToTimestamp = req.Filter.ToTimestamp
	}
	
	return c.List(ctx, getReq)
}

// ListByUser retrieves all sessions for a specific user
func (c *Client) ListByUser(ctx context.Context, userID string, limit int) (*types.GetSessionsResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	
	req := &types.GetSessionsRequest{
		UserID: &userID,
		Limit:  &limit,
	}
	
	return c.List(ctx, req)
}

// Exists checks if a session exists
func (c *Client) Exists(ctx context.Context, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, fmt.Errorf("session ID cannot be empty")
	}
	
	_, err := c.Get(ctx, sessionID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// CreateForUser creates a session for a specific user
func (c *Client) CreateForUser(ctx context.Context, userID string) (*commonTypes.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	
	req := &types.CreateSessionRequest{
		UserID: &userID,
	}
	
	return c.Create(ctx, req)
}

// UpdateUserID updates the user ID for a session
func (c *Client) UpdateUserID(ctx context.Context, sessionID, userID string) (*commonTypes.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}
	
	req := &types.UpdateSessionRequest{
		SessionID: sessionID,
		UserID:    &userID,
	}
	
	return c.Update(ctx, req)
}

// GetTraceCount gets the number of traces in a session
func (c *Client) GetTraceCount(ctx context.Context, sessionID string) (int, error) {
	sessionWithTraces, err := c.GetWithTraces(ctx, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get session traces: %w", err)
	}
	
	return sessionWithTraces.GetTraceCount(), nil
}