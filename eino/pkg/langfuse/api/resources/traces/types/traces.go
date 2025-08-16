package types

import (
	"encoding/json"
	"time"

	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

// GetTracesRequest represents a request to get traces
type GetTracesRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	Page          *int       `json:"page,omitempty"`
	Limit         *int       `json:"limit,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	Name          *string    `json:"name,omitempty"`
	SessionID     *string    `json:"sessionId,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	OrderBy       *string    `json:"orderBy,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
}

// GetTracesResponse represents the response from getting traces
type GetTracesResponse struct {
	Data []commonTypes.Trace `json:"data"`
	Meta types.MetaResponse  `json:"meta"`
}

// GetTraceRequest represents a request to get a specific trace
type GetTraceRequest struct {
	TraceID string `json:"traceId"`
}

// DeleteTraceRequest represents a request to delete a trace
type DeleteTraceRequest struct {
	TraceID string `json:"traceId"`
}

// CreateTraceRequest represents a request to create a trace directly via API
type CreateTraceRequest struct {
	ID          *string                `json:"id,omitempty"`
	Name        string                 `json:"name"`
	UserID      *string                `json:"userId,omitempty"`
	SessionID   *string                `json:"sessionId,omitempty"`
	Input       interface{}            `json:"input,omitempty"`
	Output      interface{}            `json:"output,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Environment *string                `json:"environment,omitempty"`
	Release     *string                `json:"release,omitempty"`
	Version     *string                `json:"version,omitempty"`
	Public      *bool                  `json:"public,omitempty"`
	Timestamp   *time.Time             `json:"timestamp,omitempty"`
}

// UpdateTraceRequest represents a request to update a trace
type UpdateTraceRequest struct {
	TraceID     string                 `json:"traceId"`
	Name        *string                `json:"name,omitempty"`
	UserID      *string                `json:"userId,omitempty"`
	SessionID   *string                `json:"sessionId,omitempty"`
	Input       interface{}            `json:"input,omitempty"`
	Output      interface{}            `json:"output,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Environment *string                `json:"environment,omitempty"`
	Release     *string                `json:"release,omitempty"`
	Version     *string                `json:"version,omitempty"`
	Public      *bool                  `json:"public,omitempty"`
}

// TraceWithObservations represents a trace with its observations
type TraceWithObservations struct {
	commonTypes.Trace
	Observations []commonTypes.Observation `json:"observations"`
}

// TraceStats represents statistics about traces
type TraceStats struct {
	TotalCount       int            `json:"totalCount"`
	UniqueUsers      int            `json:"uniqueUsers"`
	UniqueSessions   int            `json:"uniqueSessions"`
	AverageLatency   *time.Duration `json:"averageLatency,omitempty"`
	TotalTokens      *int           `json:"totalTokens,omitempty"`
	TotalCost        *float64       `json:"totalCost,omitempty"`
	DateRange        *DateRange     `json:"dateRange,omitempty"`
	TagDistribution  map[string]int `json:"tagDistribution,omitempty"`
	UserDistribution map[string]int `json:"userDistribution,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// GetTraceStatsRequest represents a request to get trace statistics
type GetTraceStatsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	SessionID     *string    `json:"sessionId,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
}

// TraceFilter represents filters for trace queries
type TraceFilter struct {
	UserIDs       []string       `json:"userIds,omitempty"`
	SessionIDs    []string       `json:"sessionIds,omitempty"`
	Names         []string       `json:"names,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	FromTimestamp *time.Time     `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time     `json:"toTimestamp,omitempty"`
	MinDuration   *time.Duration `json:"minDuration,omitempty"`
	MaxDuration   *time.Duration `json:"maxDuration,omitempty"`
	HasErrors     *bool          `json:"hasErrors,omitempty"`
	Public        *bool          `json:"public,omitempty"`
	Environment   *string        `json:"environment,omitempty"`
	Release       *string        `json:"release,omitempty"`
	Version       *string        `json:"version,omitempty"`
}

// TraceSortOrder represents sort order options for traces
type TraceSortOrder string

const (
	TraceSortOrderTimestampAsc  TraceSortOrder = "timestamp_asc"
	TraceSortOrderTimestampDesc TraceSortOrder = "timestamp_desc"
	TraceSortOrderDurationAsc   TraceSortOrder = "duration_asc"
	TraceSortOrderDurationDesc  TraceSortOrder = "duration_desc"
	TraceSortOrderNameAsc       TraceSortOrder = "name_asc"
	TraceSortOrderNameDesc      TraceSortOrder = "name_desc"
	TraceSortOrderUserIDAsc     TraceSortOrder = "user_id_asc"
	TraceSortOrderUserIDDesc    TraceSortOrder = "user_id_desc"
)

// PaginatedTracesRequest represents a paginated request for traces
type PaginatedTracesRequest struct {
	ProjectID string         `json:"projectId,omitempty"`
	Filter    *TraceFilter   `json:"filter,omitempty"`
	SortOrder TraceSortOrder `json:"sortOrder,omitempty"`
	Page      int            `json:"page,omitempty"`
	Limit     int            `json:"limit,omitempty"`
}

// Validate validates the GetTracesRequest
func (req *GetTracesRequest) Validate() error {
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

// Validate validates the CreateTraceRequest
func (req *CreateTraceRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	return nil
}

// Validate validates the UpdateTraceRequest
func (req *UpdateTraceRequest) Validate() error {
	if req.TraceID == "" {
		return &ValidationError{Field: "traceId", Message: "traceId is required"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ToCommonTrace converts CreateTraceRequest to common Trace type
func (req *CreateTraceRequest) ToCommonTrace() *commonTypes.Trace {
	trace := &commonTypes.Trace{
		Name:      &req.Name,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Metadata:  req.Metadata,
		Tags:      req.Tags,
		Release:   req.Release,
		Version:   req.Version,
		Public:    req.Public,
	}

	// Convert interface{} to json.RawMessage for Input
	if req.Input != nil {
		if inputBytes, err := json.Marshal(req.Input); err == nil {
			trace.Input = json.RawMessage(inputBytes)
		}
	}

	// Convert interface{} to json.RawMessage for Output
	if req.Output != nil {
		if outputBytes, err := json.Marshal(req.Output); err == nil {
			trace.Output = json.RawMessage(outputBytes)
		}
	}

	if req.ID != nil {
		trace.ID = *req.ID
	}

	if req.Timestamp != nil {
		trace.Timestamp = *req.Timestamp
	} else {
		trace.Timestamp = time.Now()
	}

	return trace
}
