package types

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrace_JSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		trace Trace
	}{
		{
			name: "complete trace",
			trace: Trace{
				ID:         "test-trace-id",
				ExternalID: stringPtr("external-123"),
				Timestamp:  time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Name:       stringPtr("test-trace"),
				Input:      json.RawMessage(`{"input": "test"}`),
				Output:     json.RawMessage(`{"output": "result"}`),
				SessionID:  stringPtr("session-123"),
				UserID:     stringPtr("user-456"),
				Metadata: map[string]interface{}{
					"key1": "value1",
					"key2": 123,
				},
				Tags:    []string{"tag1", "tag2"},
				Version: stringPtr("1.0.0"),
				Release: stringPtr("v1.0.0"),
				Public:  boolPtr(true),
			},
		},
		{
			name: "minimal trace",
			trace: Trace{
				ID:        "minimal-trace-id",
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "trace with empty metadata and tags",
			trace: Trace{
				ID:        "trace-with-empty",
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Metadata:  map[string]interface{}{},
				Tags:      []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.trace)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test unmarshaling
			var unmarshaled Trace
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.trace.ID, unmarshaled.ID)
			assert.Equal(t, tt.trace.Timestamp.UTC(), unmarshaled.Timestamp.UTC())
			
			if tt.trace.Name != nil {
				assert.Equal(t, *tt.trace.Name, *unmarshaled.Name)
			}
			
			if tt.trace.SessionID != nil {
				assert.Equal(t, *tt.trace.SessionID, *unmarshaled.SessionID)
			}
			
			if tt.trace.UserID != nil {
				assert.Equal(t, *tt.trace.UserID, *unmarshaled.UserID)
			}
		})
	}
}

func TestTraceCreateRequest_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		request TraceCreateRequest
	}{
		{
			name: "complete request",
			request: TraceCreateRequest{
				ID:         stringPtr("test-trace-id"),
				ExternalID: stringPtr("external-123"),
				Timestamp:  timePtr(time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)),
				Name:       stringPtr("test-trace"),
				Input:      map[string]interface{}{"input": "test"},
				Output:     map[string]interface{}{"output": "result"},
				SessionID:  stringPtr("session-123"),
				UserID:     stringPtr("user-456"),
				Metadata: map[string]interface{}{
					"key1": "value1",
					"key2": 123,
				},
				Tags:    []string{"tag1", "tag2"},
				Version: stringPtr("1.0.0"),
				Release: stringPtr("v1.0.0"),
				Public:  boolPtr(false),
			},
		},
		{
			name: "minimal request",
			request: TraceCreateRequest{
				Name: stringPtr("minimal-trace"),
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
			var unmarshaled TraceCreateRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify key fields
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

func TestTraceUpdateRequest_JSONSerialization(t *testing.T) {
	request := TraceUpdateRequest{
		Name:      stringPtr("updated-trace"),
		Input:     map[string]interface{}{"updated": "input"},
		Output:    map[string]interface{}{"updated": "output"},
		SessionID: stringPtr("new-session"),
		UserID:    stringPtr("new-user"),
		Metadata: map[string]interface{}{
			"updated": true,
		},
		Tags:    []string{"updated"},
		Version: stringPtr("2.0.0"),
		Release: stringPtr("v2.0.0"),
		Public:  boolPtr(true),
	}

	// Test marshaling
	data, err := json.Marshal(request)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var unmarshaled TraceUpdateRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, *request.Name, *unmarshaled.Name)
	assert.Equal(t, *request.SessionID, *unmarshaled.SessionID)
	assert.Equal(t, *request.UserID, *unmarshaled.UserID)
	assert.Equal(t, request.Tags, unmarshaled.Tags)
}

func TestTraceListResponse_JSONSerialization(t *testing.T) {
	response := TraceListResponse{
		Data: []Trace{
			{
				ID:        "trace-1",
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Name:      stringPtr("trace-1"),
			},
			{
				ID:        "trace-2",
				Timestamp: time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC),
				Name:      stringPtr("trace-2"),
			},
		},
		Meta: struct {
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			TotalItems int `json:"totalItems"`
			TotalPages int `json:"totalPages"`
		}{
			Page:       1,
			Limit:      10,
			TotalItems: 25,
			TotalPages: 3,
		},
	}

	// Test marshaling
	data, err := json.Marshal(response)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var unmarshaled TraceListResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify response
	assert.Len(t, unmarshaled.Data, 2)
	assert.Equal(t, "trace-1", unmarshaled.Data[0].ID)
	assert.Equal(t, "trace-2", unmarshaled.Data[1].ID)
	assert.Equal(t, 1, unmarshaled.Meta.Page)
	assert.Equal(t, 10, unmarshaled.Meta.Limit)
	assert.Equal(t, 25, unmarshaled.Meta.TotalItems)
	assert.Equal(t, 3, unmarshaled.Meta.TotalPages)
}

func TestTrace_EdgeCases(t *testing.T) {
	t.Run("nil pointer fields", func(t *testing.T) {
		trace := Trace{
			ID:        "test-trace",
			Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			// All pointer fields are nil
		}

		data, err := json.Marshal(trace)
		require.NoError(t, err)

		var unmarshaled Trace
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, trace.ID, unmarshaled.ID)
		assert.Nil(t, unmarshaled.Name)
		assert.Nil(t, unmarshaled.ExternalID)
		assert.Nil(t, unmarshaled.SessionID)
		assert.Nil(t, unmarshaled.UserID)
	})

	t.Run("empty json raw message", func(t *testing.T) {
		trace := Trace{
			ID:        "test-trace",
			Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Input:     json.RawMessage(`{}`),
			Output:    json.RawMessage(`null`),
		}

		data, err := json.Marshal(trace)
		require.NoError(t, err)

		var unmarshaled Trace
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, json.RawMessage(`{}`), unmarshaled.Input)
		assert.Equal(t, json.RawMessage(`null`), unmarshaled.Output)
	})

	t.Run("large metadata", func(t *testing.T) {
		largeMetadata := make(map[string]interface{})
		for i := 0; i < 100; i++ {
			largeMetadata[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
		}

		trace := Trace{
			ID:        "test-trace",
			Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Metadata:  largeMetadata,
		}

		data, err := json.Marshal(trace)
		require.NoError(t, err)

		var unmarshaled Trace
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Len(t, unmarshaled.Metadata, 100)
		assert.Equal(t, "value0", unmarshaled.Metadata["key0"])
		assert.Equal(t, "value99", unmarshaled.Metadata["key99"])
	})
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func timePtr(t time.Time) *time.Time {
	return &t
}