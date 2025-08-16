package types

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreDataType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		dataType ScoreDataType
		expected string
	}{
		{"numeric type", ScoreDataTypeNumeric, "NUMERIC"},
		{"categorical type", ScoreDataTypeCategorical, "CATEGORICAL"},
		{"boolean type", ScoreDataTypeBoolean, "BOOLEAN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.dataType))
		})
	}
}

func TestScoreSource_Constants(t *testing.T) {
	tests := []struct {
		name     string
		source   ScoreSource
		expected string
	}{
		{"api source", ScoreSourceAPI, "API"},
		{"annotation source", ScoreSourceAnnotation, "ANNOTATION"},
		{"review source", ScoreSourceReview, "REVIEW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.source))
		})
	}
}

func TestScore_JSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		score Score
	}{
		{
			name: "complete numeric score",
			score: Score{
				ID:            "score-123",
				Timestamp:     time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Name:          "accuracy",
				Value:         json.RawMessage(`0.95`),
				DataType:      ScoreDataTypeNumeric,
				Source:        scoreSourcePtr(ScoreSourceAPI),
				TraceID:       "trace-456",
				ObservationID: stringPtr("obs-789"),
				Comment:       stringPtr("High accuracy score"),
				AuthorUserID:  stringPtr("user-123"),
				ConfigID:      stringPtr("config-456"),
				QueueID:       stringPtr("queue-789"),
			},
		},
		{
			name: "categorical score",
			score: Score{
				ID:        "score-456",
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Name:      "sentiment",
				Value:     json.RawMessage(`"positive"`),
				DataType:  ScoreDataTypeCategorical,
				Source:    scoreSourcePtr(ScoreSourceAnnotation),
				TraceID:   "trace-789",
			},
		},
		{
			name: "boolean score",
			score: Score{
				ID:        "score-789",
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Name:      "is_helpful",
				Value:     json.RawMessage(`true`),
				DataType:  ScoreDataTypeBoolean,
				Source:    scoreSourcePtr(ScoreSourceReview),
				TraceID:   "trace-123",
				Comment:   stringPtr("Marked as helpful"),
			},
		},
		{
			name: "minimal score",
			score: Score{
				ID:        "minimal-score",
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Name:      "quality",
				Value:     json.RawMessage(`5`),
				DataType:  ScoreDataTypeNumeric,
				TraceID:   "trace-minimal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.score)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test unmarshaling
			var unmarshaled Score
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.score.ID, unmarshaled.ID)
			assert.Equal(t, tt.score.Name, unmarshaled.Name)
			assert.Equal(t, tt.score.DataType, unmarshaled.DataType)
			assert.Equal(t, tt.score.TraceID, unmarshaled.TraceID)
			assert.Equal(t, tt.score.Timestamp.UTC(), unmarshaled.Timestamp.UTC())
			assert.Equal(t, tt.score.Value, unmarshaled.Value)

			if tt.score.Source != nil {
				require.NotNil(t, unmarshaled.Source)
				assert.Equal(t, *tt.score.Source, *unmarshaled.Source)
			} else {
				assert.Nil(t, unmarshaled.Source)
			}

			if tt.score.ObservationID != nil {
				require.NotNil(t, unmarshaled.ObservationID)
				assert.Equal(t, *tt.score.ObservationID, *unmarshaled.ObservationID)
			} else {
				assert.Nil(t, unmarshaled.ObservationID)
			}

			if tt.score.Comment != nil {
				require.NotNil(t, unmarshaled.Comment)
				assert.Equal(t, *tt.score.Comment, *unmarshaled.Comment)
			} else {
				assert.Nil(t, unmarshaled.Comment)
			}
		})
	}
}

func TestScoreCreateRequest_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		request ScoreCreateRequest
	}{
		{
			name: "complete request",
			request: ScoreCreateRequest{
				ID:            stringPtr("score-123"),
				Timestamp:     timePtr(time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)),
				Name:          "accuracy",
				Value:         0.95,
				DataType:      scoreDataTypePtr(ScoreDataTypeNumeric),
				Source:        scoreSourcePtr(ScoreSourceAPI),
				TraceID:       "trace-456",
				ObservationID: stringPtr("obs-789"),
				Comment:       stringPtr("High accuracy score"),
				AuthorUserID:  stringPtr("user-123"),
				ConfigID:      stringPtr("config-456"),
				QueueID:       stringPtr("queue-789"),
			},
		},
		{
			name: "minimal request",
			request: ScoreCreateRequest{
				Name:    "quality",
				Value:   5,
				TraceID: "trace-123",
			},
		},
		{
			name: "categorical request",
			request: ScoreCreateRequest{
				Name:     "sentiment",
				Value:    "positive",
				DataType: scoreDataTypePtr(ScoreDataTypeCategorical),
				TraceID:  "trace-789",
			},
		},
		{
			name: "boolean request",
			request: ScoreCreateRequest{
				Name:     "is_helpful",
				Value:    true,
				DataType: scoreDataTypePtr(ScoreDataTypeBoolean),
				TraceID:  "trace-456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.request)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test unmarshaling
			var unmarshaled ScoreCreateRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.request.Name, unmarshaled.Name)
			assert.Equal(t, tt.request.TraceID, unmarshaled.TraceID)
			assert.Equal(t, tt.request.Value, unmarshaled.Value)

			if tt.request.ID != nil {
				require.NotNil(t, unmarshaled.ID)
				assert.Equal(t, *tt.request.ID, *unmarshaled.ID)
			} else {
				assert.Nil(t, unmarshaled.ID)
			}

			if tt.request.DataType != nil {
				require.NotNil(t, unmarshaled.DataType)
				assert.Equal(t, *tt.request.DataType, *unmarshaled.DataType)
			} else {
				assert.Nil(t, unmarshaled.DataType)
			}
		})
	}
}

func TestScoreUpdateRequest_JSONSerialization(t *testing.T) {
	request := ScoreUpdateRequest{
		Name:    stringPtr("updated_accuracy"),
		Value:   0.98,
		Comment: stringPtr("Updated score with better evaluation"),
	}

	// Test marshaling
	data, err := json.Marshal(request)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var unmarshaled ScoreUpdateRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, *request.Name, *unmarshaled.Name)
	assert.Equal(t, request.Value, unmarshaled.Value)
	assert.Equal(t, *request.Comment, *unmarshaled.Comment)
}

func TestScoreConfig_JSONSerialization(t *testing.T) {
	tests := []struct {
		name   string
		config ScoreConfig
	}{
		{
			name: "numeric config with range",
			config: ScoreConfig{
				ID:          "config-123",
				Name:        "accuracy_score",
				DataType:    ScoreDataTypeNumeric,
				IsArchived:  boolPtr(false),
				MinValue:    float64Ptr(0.0),
				MaxValue:    float64Ptr(1.0),
				Description: stringPtr("Accuracy score from 0 to 1"),
			},
		},
		{
			name: "categorical config",
			config: ScoreConfig{
				ID:         "config-456",
				Name:       "sentiment_score",
				DataType:   ScoreDataTypeCategorical,
				IsArchived: boolPtr(false),
				Categories: []string{"positive", "negative", "neutral"},
				Description: stringPtr("Sentiment analysis score"),
			},
		},
		{
			name: "boolean config",
			config: ScoreConfig{
				ID:          "config-789",
				Name:        "helpful_score",
				DataType:    ScoreDataTypeBoolean,
				IsArchived:  boolPtr(false),
				Description: stringPtr("Whether the response is helpful"),
			},
		},
		{
			name: "archived config",
			config: ScoreConfig{
				ID:         "config-archived",
				Name:       "old_score",
				DataType:   ScoreDataTypeNumeric,
				IsArchived: boolPtr(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test unmarshaling
			var unmarshaled ScoreConfig
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.config.ID, unmarshaled.ID)
			assert.Equal(t, tt.config.Name, unmarshaled.Name)
			assert.Equal(t, tt.config.DataType, unmarshaled.DataType)

			if tt.config.IsArchived != nil {
				require.NotNil(t, unmarshaled.IsArchived)
				assert.Equal(t, *tt.config.IsArchived, *unmarshaled.IsArchived)
			} else {
				assert.Nil(t, unmarshaled.IsArchived)
			}

			if tt.config.MinValue != nil {
				require.NotNil(t, unmarshaled.MinValue)
				assert.Equal(t, *tt.config.MinValue, *unmarshaled.MinValue)
			} else {
				assert.Nil(t, unmarshaled.MinValue)
			}

			if tt.config.MaxValue != nil {
				require.NotNil(t, unmarshaled.MaxValue)
				assert.Equal(t, *tt.config.MaxValue, *unmarshaled.MaxValue)
			} else {
				assert.Nil(t, unmarshaled.MaxValue)
			}

			if tt.config.Categories != nil {
				assert.Equal(t, tt.config.Categories, unmarshaled.Categories)
			} else {
				assert.Nil(t, unmarshaled.Categories)
			}

			if tt.config.Description != nil {
				require.NotNil(t, unmarshaled.Description)
				assert.Equal(t, *tt.config.Description, *unmarshaled.Description)
			} else {
				assert.Nil(t, unmarshaled.Description)
			}
		})
	}
}

func TestNumericScore(t *testing.T) {
	tests := []struct {
		name    string
		scoreName string
		value   float64
		traceID string
	}{
		{
			name:      "typical numeric score",
			scoreName: "accuracy",
			value:     0.95,
			traceID:   "trace-123",
		},
		{
			name:      "zero value",
			scoreName: "error_rate",
			value:     0.0,
			traceID:   "trace-456",
		},
		{
			name:      "negative value",
			scoreName: "loss",
			value:     -0.5,
			traceID:   "trace-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := NumericScore(tt.scoreName, tt.value, tt.traceID)

			require.NotNil(t, request)
			assert.Equal(t, tt.scoreName, request.Name)
			assert.Equal(t, tt.value, request.Value)
			assert.Equal(t, tt.traceID, request.TraceID)

			require.NotNil(t, request.DataType)
			assert.Equal(t, ScoreDataTypeNumeric, *request.DataType)

			require.NotNil(t, request.Source)
			assert.Equal(t, ScoreSourceAPI, *request.Source)
		})
	}
}

func TestCategoricalScore(t *testing.T) {
	tests := []struct {
		name      string
		scoreName string
		value     string
		traceID   string
	}{
		{
			name:      "sentiment score",
			scoreName: "sentiment",
			value:     "positive",
			traceID:   "trace-123",
		},
		{
			name:      "quality rating",
			scoreName: "quality",
			value:     "excellent",
			traceID:   "trace-456",
		},
		{
			name:      "empty string value",
			scoreName: "category",
			value:     "",
			traceID:   "trace-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CategoricalScore(tt.scoreName, tt.value, tt.traceID)

			require.NotNil(t, request)
			assert.Equal(t, tt.scoreName, request.Name)
			assert.Equal(t, tt.value, request.Value)
			assert.Equal(t, tt.traceID, request.TraceID)

			require.NotNil(t, request.DataType)
			assert.Equal(t, ScoreDataTypeCategorical, *request.DataType)

			require.NotNil(t, request.Source)
			assert.Equal(t, ScoreSourceAPI, *request.Source)
		})
	}
}

func TestBooleanScore(t *testing.T) {
	tests := []struct {
		name      string
		scoreName string
		value     bool
		traceID   string
	}{
		{
			name:      "helpful score true",
			scoreName: "is_helpful",
			value:     true,
			traceID:   "trace-123",
		},
		{
			name:      "helpful score false",
			scoreName: "is_helpful",
			value:     false,
			traceID:   "trace-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := BooleanScore(tt.scoreName, tt.value, tt.traceID)

			require.NotNil(t, request)
			assert.Equal(t, tt.scoreName, request.Name)
			assert.Equal(t, tt.value, request.Value)
			assert.Equal(t, tt.traceID, request.TraceID)

			require.NotNil(t, request.DataType)
			assert.Equal(t, ScoreDataTypeBoolean, *request.DataType)

			require.NotNil(t, request.Source)
			assert.Equal(t, ScoreSourceAPI, *request.Source)
		})
	}
}

func TestScore_EdgeCases(t *testing.T) {
	t.Run("score with null json value", func(t *testing.T) {
		score := Score{
			ID:        "test-score",
			Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Name:      "test",
			Value:     json.RawMessage(`null`),
			DataType:  ScoreDataTypeNumeric,
			TraceID:   "trace-123",
		}

		data, err := json.Marshal(score)
		require.NoError(t, err)

		var unmarshaled Score
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, json.RawMessage(`null`), unmarshaled.Value)
	})

	t.Run("score with complex json value", func(t *testing.T) {
		complexValue := json.RawMessage(`{"nested": {"score": 0.95, "confidence": 0.8}}`)
		
		score := Score{
			ID:        "test-score",
			Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Name:      "complex_score",
			Value:     complexValue,
			DataType:  ScoreDataTypeNumeric,
			TraceID:   "trace-123",
		}

		data, err := json.Marshal(score)
		require.NoError(t, err)

		var unmarshaled Score
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, complexValue, unmarshaled.Value)
	})

	t.Run("config with empty categories", func(t *testing.T) {
		config := ScoreConfig{
			ID:         "config-empty-categories",
			Name:       "empty_categories",
			DataType:   ScoreDataTypeCategorical,
			Categories: []string{},
		}

		data, err := json.Marshal(config)
		require.NoError(t, err)

		var unmarshaled ScoreConfig
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.NotNil(t, unmarshaled.Categories)
		assert.Len(t, unmarshaled.Categories, 0)
	})

	t.Run("config with many categories", func(t *testing.T) {
		manyCategories := make([]string, 100)
		for i := 0; i < 100; i++ {
			manyCategories[i] = fmt.Sprintf("category_%d", i)
		}

		config := ScoreConfig{
			ID:         "config-many-categories",
			Name:       "many_categories",
			DataType:   ScoreDataTypeCategorical,
			Categories: manyCategories,
		}

		data, err := json.Marshal(config)
		require.NoError(t, err)

		var unmarshaled ScoreConfig
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Len(t, unmarshaled.Categories, 100)
		assert.Equal(t, "category_0", unmarshaled.Categories[0])
		assert.Equal(t, "category_99", unmarshaled.Categories[99])
	})
}

// Helper functions specific to score tests
func scoreSourcePtr(source ScoreSource) *ScoreSource {
	return &source
}

func scoreDataTypePtr(dataType ScoreDataType) *ScoreDataType {
	return &dataType
}