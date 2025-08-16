package types

import (
	"time"

	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

// GetScoresRequest represents a request to get scores
type GetScoresRequest struct {
	ProjectID     string                     `json:"projectId,omitempty"`
	Page          *int                       `json:"page,omitempty"`
	Limit         *int                       `json:"limit,omitempty"`
	TraceID       *string                    `json:"traceId,omitempty"`
	ObservationID *string                    `json:"observationId,omitempty"`
	Name          *string                    `json:"name,omitempty"`
	DataType      *commonTypes.ScoreDataType `json:"dataType,omitempty"`
	ConfigID      *string                    `json:"configId,omitempty"`
	FromTimestamp *time.Time                 `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time                 `json:"toTimestamp,omitempty"`
	UserID        *string                    `json:"userId,omitempty"`
	Source        *string                    `json:"source,omitempty"`
}

// GetScoresResponse represents the response from getting scores
type GetScoresResponse struct {
	Data []commonTypes.Score     `json:"data"`
	Meta types.MetaResponse      `json:"meta"`
}

// ScoreAggregation represents aggregated score data
type ScoreAggregation struct {
	Name         string                    `json:"name"`
	DataType     commonTypes.ScoreDataType `json:"dataType"`
	Count        int                       `json:"count"`
	AverageValue *float64                  `json:"averageValue,omitempty"`
	MinValue     interface{}               `json:"minValue,omitempty"`
	MaxValue     interface{}               `json:"maxValue,omitempty"`
	Distribution map[string]int            `json:"distribution,omitempty"`
}

// GetScoreAggregationRequest represents a request to get score aggregations
type GetScoreAggregationRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	TraceID       *string    `json:"traceId,omitempty"`
	ObservationID *string    `json:"observationId,omitempty"`
	Name          *string    `json:"name,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	GroupBy       []string   `json:"groupBy,omitempty"`
}

// GetScoreAggregationResponse represents the response from getting score aggregations
type GetScoreAggregationResponse struct {
	Data []ScoreAggregation `json:"data"`
}

// ScoreStats represents statistics about scores
type ScoreStats struct {
	TotalCount        int                        `json:"totalCount"`
	UniqueNames       int                        `json:"uniqueNames"`
	ScoresByDataType  map[string]int             `json:"scoresByDataType"`
	ScoresByName      map[string]int             `json:"scoresByName"`
	AveragesByName    map[string]float64         `json:"averagesByName"`
	LatestScores      []commonTypes.Score        `json:"latestScores"`
	DateRange         *DateRange                 `json:"dateRange,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// GetScoreStatsRequest represents a request to get score statistics
type GetScoreStatsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	TraceID       *string    `json:"traceId,omitempty"`
	ObservationID *string    `json:"observationId,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
}

// ScoreFilter represents filters for score queries
type ScoreFilter struct {
	TraceIDs       []string                   `json:"traceIds,omitempty"`
	ObservationIDs []string                   `json:"observationIds,omitempty"`
	Names          []string                   `json:"names,omitempty"`
	DataTypes      []commonTypes.ScoreDataType `json:"dataTypes,omitempty"`
	ConfigIDs      []string                   `json:"configIds,omitempty"`
	FromTimestamp  *time.Time                 `json:"fromTimestamp,omitempty"`
	ToTimestamp    *time.Time                 `json:"toTimestamp,omitempty"`
	UserIDs        []string                   `json:"userIds,omitempty"`
	Sources        []string                   `json:"sources,omitempty"`
	MinValue       interface{}                `json:"minValue,omitempty"`
	MaxValue       interface{}                `json:"maxValue,omitempty"`
}

// ScoreSortOrder represents sort order options for scores
type ScoreSortOrder string

const (
	ScoreSortOrderTimestampAsc  ScoreSortOrder = "timestamp_asc"
	ScoreSortOrderTimestampDesc ScoreSortOrder = "timestamp_desc"
	ScoreSortOrderValueAsc      ScoreSortOrder = "value_asc"
	ScoreSortOrderValueDesc     ScoreSortOrder = "value_desc"
	ScoreSortOrderNameAsc       ScoreSortOrder = "name_asc"
	ScoreSortOrderNameDesc      ScoreSortOrder = "name_desc"
)

// PaginatedScoresRequest represents a paginated request for scores
type PaginatedScoresRequest struct {
	ProjectID string         `json:"projectId,omitempty"`
	Filter    *ScoreFilter   `json:"filter,omitempty"`
	SortOrder ScoreSortOrder `json:"sortOrder,omitempty"`
	Page      int            `json:"page,omitempty"`
	Limit     int            `json:"limit,omitempty"`
}

// Validate validates the GetScoresRequest
func (req *GetScoresRequest) Validate() error {
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

// Validate validates the GetScoreAggregationRequest
func (req *GetScoreAggregationRequest) Validate() error {
	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}
	
	// Validate groupBy fields
	validGroupBy := map[string]bool{
		"name":     true,
		"dataType": true,
		"traceId":  true,
		"userId":   true,
		"source":   true,
		"date":     true,
	}
	
	for _, field := range req.GroupBy {
		if !validGroupBy[field] {
			return &ValidationError{Field: "groupBy", Message: "invalid groupBy field: " + field}
		}
	}
	
	return nil
}

// Validate validates the PaginatedScoresRequest
func (req *PaginatedScoresRequest) Validate() error {
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
	}
	
	return nil
}

// HasNumericValues returns true if the aggregation has numeric values
func (sa *ScoreAggregation) HasNumericValues() bool {
	return sa.DataType == commonTypes.ScoreDataTypeNumeric && sa.AverageValue != nil
}

// GetDistributionEntries returns the distribution as a sorted slice of entries
func (sa *ScoreAggregation) GetDistributionEntries() []DistributionEntry {
	if sa.Distribution == nil {
		return nil
	}
	
	entries := make([]DistributionEntry, 0, len(sa.Distribution))
	for value, count := range sa.Distribution {
		entries = append(entries, DistributionEntry{
			Value: value,
			Count: count,
		})
	}
	
	return entries
}

// DistributionEntry represents a single entry in a distribution
type DistributionEntry struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}