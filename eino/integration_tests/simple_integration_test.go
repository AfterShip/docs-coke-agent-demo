// +build integration

package integration_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eino/pkg/langfuse/api/resources/ingestion/types"
)

// SimpleTestServer provides a minimal mock server for testing
type SimpleTestServer struct {
	server         *httptest.Server
	receivedEvents []types.IngestionEvent
	mu             sync.RWMutex
}

func NewSimpleTestServer() *SimpleTestServer {
	ts := &SimpleTestServer{
		receivedEvents: make([]types.IngestionEvent, 0),
	}

	mux := http.NewServeMux()

	// Ingestion endpoint
	mux.HandleFunc("/api/public/ingestion", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req types.IngestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}

		// Store received events for verification
		ts.receivedEvents = append(ts.receivedEvents, req.Batch...)

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.IngestionResponse{
			Success: true,
			Errors:  nil,
		})
	})

	ts.server = httptest.NewServer(mux)
	return ts
}

func (ts *SimpleTestServer) Close() {
	ts.server.Close()
}

func (ts *SimpleTestServer) URL() string {
	return ts.server.URL
}

func (ts *SimpleTestServer) GetReceivedEvents() []types.IngestionEvent {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	events := make([]types.IngestionEvent, len(ts.receivedEvents))
	copy(events, ts.receivedEvents)
	return events
}

// Test basic ingestion API functionality
func TestIntegration_IngestionAPI(t *testing.T) {
	server := NewSimpleTestServer()
	defer server.Close()

	// Create a simple ingestion request directly
	events := []types.IngestionEvent{
		{
			ID:        "test-event-1",
			Type:      types.EventType("trace-create"),
			Timestamp: time.Now(),
			Body: map[string]interface{}{
				"id":   "test-trace-id",
				"name": "test-trace",
				"userId": "test-user",
				"input": "test input",
			},
		},
	}

	req := types.IngestionRequest{
		Batch: events,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// Send the request to the server
	client := &http.Client{Timeout: 5 * time.Second}
	
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	resp, err := client.Post(
		server.URL()+"/api/public/ingestion",
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var responseBody types.IngestionResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	require.NoError(t, err)
	assert.True(t, responseBody.Success)

	// Verify the server received the events
	receivedEvents := server.GetReceivedEvents()
	require.Len(t, receivedEvents, 1)
	
	event := receivedEvents[0]
	assert.Equal(t, "test-event-1", event.ID)
	assert.Equal(t, types.EventType("trace-create"), event.Type)
	
	eventBody, ok := event.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-trace", eventBody["name"])
	assert.Equal(t, "test-user", eventBody["userId"])
}

// Test multiple concurrent requests
func TestIntegration_ConcurrentRequests(t *testing.T) {
	server := NewSimpleTestServer()
	defer server.Close()

	const numRequests = 10
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := types.IngestionEvent{
				ID:        fmt.Sprintf("concurrent-event-%d", id),
				Type:      types.EventType("trace-create"),
				Timestamp: time.Now(),
				Body: map[string]interface{}{
					"id":   fmt.Sprintf("trace-%d", id),
					"name": fmt.Sprintf("concurrent-trace-%d", id),
					"userId": fmt.Sprintf("user-%d", id),
				},
			}

			req := types.IngestionRequest{
				Batch: []types.IngestionEvent{event},
			}

			client := &http.Client{Timeout: 5 * time.Second}
			
			jsonData, err := json.Marshal(req)
			if err != nil {
				errors <- err
				return
			}

			resp, err := client.Post(
				server.URL()+"/api/public/ingestion",
				"application/json",
				strings.NewReader(string(jsonData)),
			)
			if err != nil {
				errors <- err
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("request %d failed with status %d", id, resp.StatusCode)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
	}

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Verify all events were received
	receivedEvents := server.GetReceivedEvents()
	assert.Len(t, receivedEvents, numRequests, "Expected all %d events to be received", numRequests)

	// Verify all events are unique
	eventIDs := make(map[string]bool)
	for _, event := range receivedEvents {
		assert.False(t, eventIDs[event.ID], "Duplicate event ID found: %s", event.ID)
		eventIDs[event.ID] = true
	}
}

// Test error handling
func TestIntegration_ErrorHandling(t *testing.T) {
	server := NewSimpleTestServer()
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Test invalid JSON
	resp, err := client.Post(
		server.URL()+"/api/public/ingestion",
		"application/json",
		strings.NewReader("{invalid json}"),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test wrong method
	resp, err = client.Get(server.URL() + "/api/public/ingestion")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

