package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObservationType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		obsType  ObservationType
		expected string
	}{
		{"span type", ObservationTypeSpan, "SPAN"},
		{"generation type", ObservationTypeGeneration, "GENERATION"},
		{"event type", ObservationTypeEvent, "EVENT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.obsType))
		})
	}
}

func TestObservationLevel_Constants(t *testing.T) {
	tests := []struct {
		name     string
		level    ObservationLevel
		expected string
	}{
		{"debug level", ObservationLevelDebug, "DEBUG"},
		{"default level", ObservationLevelDefault, "DEFAULT"},
		{"warning level", ObservationLevelWarning, "WARNING"},
		{"error level", ObservationLevelError, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.level))
		})
	}
}

func TestObservation_JSONSerialization(t *testing.T) {
	tests := []struct {
		name        string
		observation Observation
	}{
		{
			name: "complete observation",
			observation: Observation{
				ID:                  "obs-123",
				TraceID:             "trace-456",
				Type:                ObservationTypeGeneration,
				ExternalID:          stringPtr("ext-123"),
				Name:                stringPtr("llm-call"),
				StartTime:           time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				EndTime:             timePtr(time.Date(2024, 1, 15, 12, 0, 5, 0, time.UTC)),
				CompletionStartTime: timePtr(time.Date(2024, 1, 15, 12, 0, 2, 0, time.UTC)),
				Model:               stringPtr("gpt-4"),
				ModelParameters: map[string]interface{}{
					"temperature": 0.7,
					"max_tokens":  1000,
				},
				Input:  json.RawMessage(`{"prompt": "Hello, world!"}`),
				Output: json.RawMessage(`{"response": "Hi there!"}`),
				Usage: &Usage{
					Input:  intPtr(10),
					Output: intPtr(20),
					Total:  intPtr(30),
				},
				Metadata: map[string]interface{}{
					"model_version": "gpt-4-0613",
					"temperature":   0.7,
				},
				ParentObservationID: stringPtr("parent-obs-789"),
				Level:               observationLevelPtr(ObservationLevelDefault),
				StatusMessage:       stringPtr("completed successfully"),
				Version:             stringPtr("1.0.0"),
			},
		},
		{
			name: "minimal observation",
			observation: Observation{
				ID:        "minimal-obs",
				TraceID:   "trace-123",
				Type:      ObservationTypeSpan,
				StartTime: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "event observation",
			observation: Observation{
				ID:            "event-obs",
				TraceID:       "trace-123",
				Type:          ObservationTypeEvent,
				Name:          stringPtr("user_action"),
				StartTime:     time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Input:         json.RawMessage(`{"action": "click", "button": "submit"}`),
				Level:         observationLevelPtr(ObservationLevelWarning),
				StatusMessage: stringPtr("button clicked"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.observation)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test unmarshaling
			var unmarshaled Observation
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.observation.ID, unmarshaled.ID)
			assert.Equal(t, tt.observation.TraceID, unmarshaled.TraceID)
			assert.Equal(t, tt.observation.Type, unmarshaled.Type)
			assert.Equal(t, tt.observation.StartTime.UTC(), unmarshaled.StartTime.UTC())

			if tt.observation.Name != nil {
				require.NotNil(t, unmarshaled.Name)
				assert.Equal(t, *tt.observation.Name, *unmarshaled.Name)
			}

			if tt.observation.Model != nil {
				require.NotNil(t, unmarshaled.Model)
				assert.Equal(t, *tt.observation.Model, *unmarshaled.Model)
			}

			if tt.observation.EndTime != nil {
				require.NotNil(t, unmarshaled.EndTime)
				assert.Equal(t, tt.observation.EndTime.UTC(), unmarshaled.EndTime.UTC())
			}
		})
	}
}

func TestObservationCreateRequest_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		request ObservationCreateRequest
	}{
		{
			name: "complete generation request",
			request: ObservationCreateRequest{
				ID:                  stringPtr("obs-123"),
				TraceID:             "trace-456",
				Type:                ObservationTypeGeneration,
				ExternalID:          stringPtr("ext-123"),
				Name:                stringPtr("llm-call"),
				StartTime:           timePtr(time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)),
				EndTime:             timePtr(time.Date(2024, 1, 15, 12, 0, 5, 0, time.UTC)),
				CompletionStartTime: timePtr(time.Date(2024, 1, 15, 12, 0, 2, 0, time.UTC)),
				Model:               stringPtr("gpt-4"),
				ModelParameters: map[string]interface{}{
					"temperature": 0.7,
					"max_tokens":  1000,
				},
				Input:  map[string]interface{}{"prompt": "Hello, world!"},
				Output: map[string]interface{}{"response": "Hi there!"},
				Usage: &Usage{
					Input:  intPtr(10),
					Output: intPtr(20),
					Total:  intPtr(30),
				},
				Metadata: map[string]interface{}{
					"model_version": "gpt-4-0613",
				},
				ParentObservationID: stringPtr("parent-obs-789"),
				Level:               observationLevelPtr(ObservationLevelDefault),
				StatusMessage:       stringPtr("completed successfully"),
				Version:             stringPtr("1.0.0"),
			},
		},
		{
			name: "minimal span request",
			request: ObservationCreateRequest{
				TraceID: "trace-123",
				Type:    ObservationTypeSpan,
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
			var unmarshaled ObservationCreateRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.request.TraceID, unmarshaled.TraceID)
			assert.Equal(t, tt.request.Type, unmarshaled.Type)

			if tt.request.ID != nil {
				require.NotNil(t, unmarshaled.ID)
				assert.Equal(t, *tt.request.ID, *unmarshaled.ID)
			}

			if tt.request.Name != nil {
				require.NotNil(t, unmarshaled.Name)
				assert.Equal(t, *tt.request.Name, *unmarshaled.Name)
			}
		})
	}
}

func TestObservationUpdateRequest_JSONSerialization(t *testing.T) {
	request := ObservationUpdateRequest{
		Name:                stringPtr("updated-observation"),
		EndTime:             timePtr(time.Date(2024, 1, 15, 12, 0, 10, 0, time.UTC)),
		CompletionStartTime: timePtr(time.Date(2024, 1, 15, 12, 0, 8, 0, time.UTC)),
		Model:               stringPtr("gpt-4-turbo"),
		ModelParameters: map[string]interface{}{
			"temperature": 0.5,
		},
		Input:  map[string]interface{}{"updated": "input"},
		Output: map[string]interface{}{"updated": "output"},
		Usage: &Usage{
			Input:  intPtr(15),
			Output: intPtr(25),
			Total:  intPtr(40),
		},
		Metadata: map[string]interface{}{
			"updated": true,
		},
		Level:         observationLevelPtr(ObservationLevelWarning),
		StatusMessage: stringPtr("updated status"),
		Version:       stringPtr("2.0.0"),
	}

	// Test marshaling
	data, err := json.Marshal(request)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var unmarshaled ObservationUpdateRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, *request.Name, *unmarshaled.Name)
	assert.Equal(t, *request.Model, *unmarshaled.Model)
	assert.Equal(t, request.EndTime.UTC(), unmarshaled.EndTime.UTC())
	assert.Equal(t, *request.Level, *unmarshaled.Level)
	assert.Equal(t, *request.StatusMessage, *unmarshaled.StatusMessage)
}

func TestObservation_EdgeCases(t *testing.T) {
	t.Run("nil usage field", func(t *testing.T) {
		obs := Observation{
			ID:        "test-obs",
			TraceID:   "test-trace",
			Type:      ObservationTypeSpan,
			StartTime: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Usage:     nil,
		}

		data, err := json.Marshal(obs)
		require.NoError(t, err)

		var unmarshaled Observation
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Nil(t, unmarshaled.Usage)
	})

	t.Run("empty model parameters", func(t *testing.T) {
		obs := Observation{
			ID:              "test-obs",
			TraceID:         "test-trace",
			Type:            ObservationTypeGeneration,
			StartTime:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			ModelParameters: map[string]interface{}{},
		}

		data, err := json.Marshal(obs)
		require.NoError(t, err)

		var unmarshaled Observation
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.NotNil(t, unmarshaled.ModelParameters)
		assert.Len(t, unmarshaled.ModelParameters, 0)
	})

	t.Run("complex model parameters", func(t *testing.T) {
		obs := Observation{
			ID:        "test-obs",
			TraceID:   "test-trace",
			Type:      ObservationTypeGeneration,
			StartTime: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			ModelParameters: map[string]interface{}{
				"temperature":       0.7,
				"max_tokens":        1000,
				"top_p":             0.9,
				"frequency_penalty": 0.0,
				"presence_penalty":  0.0,
				"stop":              []string{"\n", ".", "!"},
				"logit_bias": map[string]interface{}{
					"50256": -100,
				},
			},
		}

		data, err := json.Marshal(obs)
		require.NoError(t, err)

		var unmarshaled Observation
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, 0.7, unmarshaled.ModelParameters["temperature"])
		assert.Equal(t, float64(1000), unmarshaled.ModelParameters["max_tokens"])
		
		stop, ok := unmarshaled.ModelParameters["stop"].([]interface{})
		require.True(t, ok)
		assert.Len(t, stop, 3)
	})
}

func TestObservationType_Validation(t *testing.T) {
	validTypes := []ObservationType{
		ObservationTypeSpan,
		ObservationTypeGeneration,
		ObservationTypeEvent,
	}

	for _, obsType := range validTypes {
		obs := Observation{
			ID:        "test-obs",
			TraceID:   "test-trace",
			Type:      obsType,
			StartTime: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(obs)
		require.NoError(t, err)

		var unmarshaled Observation
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, obsType, unmarshaled.Type)
	}
}

func TestObservationLevel_Validation(t *testing.T) {
	validLevels := []ObservationLevel{
		ObservationLevelDebug,
		ObservationLevelDefault,
		ObservationLevelWarning,
		ObservationLevelError,
	}

	for _, level := range validLevels {
		obs := Observation{
			ID:        "test-obs",
			TraceID:   "test-trace",
			Type:      ObservationTypeSpan,
			StartTime: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Level:     &level,
		}

		data, err := json.Marshal(obs)
		require.NoError(t, err)

		var unmarshaled Observation
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		require.NotNil(t, unmarshaled.Level)
		assert.Equal(t, level, *unmarshaled.Level)
	}
}

// Helper functions specific to observation tests
func intPtr(i int) *int {
	return &i
}

func observationLevelPtr(level ObservationLevel) *ObservationLevel {
	return &level
}