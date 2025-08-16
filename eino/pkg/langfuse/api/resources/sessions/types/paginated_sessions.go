package types

import (
	"time"

	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

// GetSessionsRequest represents a request to get sessions
type GetSessionsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	Page          *int       `json:"page,omitempty"`
	Limit         *int       `json:"limit,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	OrderBy       *string    `json:"orderBy,omitempty"`
}

// GetSessionsResponse represents the response from getting sessions
type GetSessionsResponse struct {
	Data []commonTypes.Session `json:"data"`
	Meta types.MetaResponse    `json:"meta"`
}

// GetSessionRequest represents a request to get a specific session
type GetSessionRequest struct {
	SessionID string `json:"sessionId"`
}

// SessionWithTraces represents a session with its traces
type SessionWithTraces struct {
	commonTypes.Session
	Traces []commonTypes.Trace `json:"traces"`
}

// SessionStats represents statistics about sessions
type SessionStats struct {
	TotalCount       int                    `json:"totalCount"`
	UniqueUsers      int                    `json:"uniqueUsers"`
	AverageTraces    float64                `json:"averageTraces"`
	AverageLatency   *time.Duration         `json:"averageLatency,omitempty"`
	TotalTokens      *int                   `json:"totalTokens,omitempty"`
	TotalCost        *float64               `json:"totalCost,omitempty"`
	DateRange        *DateRange             `json:"dateRange,omitempty"`
	UserDistribution map[string]int         `json:"userDistribution,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// GetSessionStatsRequest represents a request to get session statistics
type GetSessionStatsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// SessionFilter represents filters for session queries
type SessionFilter struct {
	UserIDs       []string   `json:"userIds,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	MinTraces     *int       `json:"minTraces,omitempty"`
	MaxTraces     *int       `json:"maxTraces,omitempty"`
	MinDuration   *time.Duration `json:"minDuration,omitempty"`
	MaxDuration   *time.Duration `json:"maxDuration,omitempty"`
}

// SessionSortOrder represents sort order options for sessions
type SessionSortOrder string

const (
	SessionSortOrderCreatedAtAsc   SessionSortOrder = "created_at_asc"
	SessionSortOrderCreatedAtDesc  SessionSortOrder = "created_at_desc"
	SessionSortOrderUpdatedAtAsc   SessionSortOrder = "updated_at_asc"
	SessionSortOrderUpdatedAtDesc  SessionSortOrder = "updated_at_desc"
	SessionSortOrderTraceCountAsc  SessionSortOrder = "trace_count_asc"
	SessionSortOrderTraceCountDesc SessionSortOrder = "trace_count_desc"
	SessionSortOrderUserIDAsc      SessionSortOrder = "user_id_asc"
	SessionSortOrderUserIDDesc     SessionSortOrder = "user_id_desc"
)

// PaginatedSessionsRequest represents a paginated request for sessions
type PaginatedSessionsRequest struct {
	ProjectID string           `json:"projectId,omitempty"`
	Filter    *SessionFilter   `json:"filter,omitempty"`
	SortOrder SessionSortOrder `json:"sortOrder,omitempty"`
	Page      int             `json:"page,omitempty"`
	Limit     int             `json:"limit,omitempty"`
}

// CreateSessionRequest represents a request to create a session
type CreateSessionRequest struct {
	ID     *string `json:"id,omitempty"`
	UserID *string `json:"userId,omitempty"`
}

// UpdateSessionRequest represents a request to update a session
type UpdateSessionRequest struct {
	SessionID string  `json:"sessionId"`
	UserID    *string `json:"userId,omitempty"`
}

// Validate validates the GetSessionsRequest
func (req *GetSessionsRequest) Validate() error {
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

// Validate validates the UpdateSessionRequest
func (req *UpdateSessionRequest) Validate() error {
	if req.SessionID == "" {
		return &ValidationError{Field: "sessionId", Message: "sessionId is required"}
	}
	
	return nil
}

// Validate validates the PaginatedSessionsRequest
func (req *PaginatedSessionsRequest) Validate() error {
	if req.Limit < 1 || req.Limit > 1000 {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}
	
	if req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}
	
	if req.Filter != nil {
		if req.Filter.FromTimestamp != nil && req.Filter.ToTimestamp != nil && req.Filter.FromTimestamp.After(*req.Filter.ToTimestamp) {
			return &ValidationError{Field: "filter.timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
		}
		
		if req.Filter.MinTraces != nil && req.Filter.MaxTraces != nil && *req.Filter.MinTraces > *req.Filter.MaxTraces {
			return &ValidationError{Field: "filter.traces", Message: "minTraces cannot be greater than maxTraces"}
		}
		
		if req.Filter.MinDuration != nil && req.Filter.MaxDuration != nil && *req.Filter.MinDuration > *req.Filter.MaxDuration {
			return &ValidationError{Field: "filter.duration", Message: "minDuration cannot be greater than maxDuration"}
		}
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

// ToCommonSession converts CreateSessionRequest to common Session type
func (req *CreateSessionRequest) ToCommonSession() *commonTypes.Session {
	session := &commonTypes.Session{
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	
	if req.ID != nil {
		session.ID = *req.ID
	}
	
	if req.UserID != nil {
		session.UserID = req.UserID
	}
	
	return session
}

// GetTraceCount returns the number of traces in the session
func (s *SessionWithTraces) GetTraceCount() int {
	return len(s.Traces)
}

// GetFirstTrace returns the first trace in the session (by timestamp)
func (s *SessionWithTraces) GetFirstTrace() *commonTypes.Trace {
	if len(s.Traces) == 0 {
		return nil
	}
	
	earliest := &s.Traces[0]
	for i := 1; i < len(s.Traces); i++ {
		if s.Traces[i].Timestamp.Before(earliest.Timestamp) {
			earliest = &s.Traces[i]
		}
	}
	
	return earliest
}

// GetLastTrace returns the last trace in the session (by timestamp)
func (s *SessionWithTraces) GetLastTrace() *commonTypes.Trace {
	if len(s.Traces) == 0 {
		return nil
	}
	
	latest := &s.Traces[0]
	for i := 1; i < len(s.Traces); i++ {
		if s.Traces[i].Timestamp.After(latest.Timestamp) {
			latest = &s.Traces[i]
		}
	}
	
	return latest
}

// GetDuration returns the duration of the session (from first to last trace)
func (s *SessionWithTraces) GetDuration() *time.Duration {
	first := s.GetFirstTrace()
	last := s.GetLastTrace()
	
	if first == nil || last == nil || first.ID == last.ID {
		return nil
	}
	
	duration := last.Timestamp.Sub(first.Timestamp)
	return &duration
}

// GetUniqueUsers returns the count of unique users from the session stats
func (ss *SessionStats) GetUniqueUsers() int {
	return ss.UniqueUsers
}

// GetTopUsers returns the top N users by session count
func (ss *SessionStats) GetTopUsers(n int) []UserSessionCount {
	if ss.UserDistribution == nil {
		return nil
	}
	
	users := make([]UserSessionCount, 0, len(ss.UserDistribution))
	for userID, count := range ss.UserDistribution {
		users = append(users, UserSessionCount{
			UserID: userID,
			Count:  count,
		})
	}
	
	// Simple sort by count (descending)
	for i := 0; i < len(users)-1; i++ {
		for j := i + 1; j < len(users); j++ {
			if users[j].Count > users[i].Count {
				users[i], users[j] = users[j], users[i]
			}
		}
	}
	
	if n > len(users) {
		n = len(users)
	}
	
	return users[:n]
}

// UserSessionCount represents a user with their session count
type UserSessionCount struct {
	UserID string `json:"userId"`
	Count  int    `json:"count"`
}