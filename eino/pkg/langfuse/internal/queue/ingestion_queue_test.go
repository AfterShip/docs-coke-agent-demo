package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eino/pkg/langfuse/api/resources/ingestion/types"
)

// ErrQueueClosed is used when trying to operate on a closed queue
var ErrQueueClosed = fmt.Errorf("queue is closed")

// MockIngestionClient implements IngestionClient interface for testing
type MockIngestionClient struct {
	mu             sync.RWMutex
	submitCalls    [][]types.IngestionEvent
	responses      []*types.IngestionResponse
	errors         []error
	callIndex      int64
	shouldFail     bool
	failAfterCalls int
	processingTime time.Duration

	// Concurrent testing support
	concurrentCalls int64
	maxConcurrent   int64
	callDurations   []time.Duration
}

// NewMockIngestionClient creates a new mock ingestion client
func NewMockIngestionClient() *MockIngestionClient {
	return &MockIngestionClient{
		submitCalls:    make([][]types.IngestionEvent, 0),
		responses:      make([]*types.IngestionResponse, 0),
		errors:         make([]error, 0),
		callDurations:  make([]time.Duration, 0),
		processingTime: 10 * time.Millisecond, // Default processing time
	}
}

// SubmitBatch implements the IngestionClient interface
func (m *MockIngestionClient) SubmitBatch(ctx context.Context, events []types.IngestionEvent) (*types.IngestionResponse, error) {
	startTime := time.Now()
	callIndex := atomic.AddInt64(&m.callIndex, 1) - 1

	// Track concurrent calls
	concurrent := atomic.AddInt64(&m.concurrentCalls, 1)
	defer atomic.AddInt64(&m.concurrentCalls, -1)

	// Update max concurrent calls
	for {
		current := atomic.LoadInt64(&m.maxConcurrent)
		if concurrent <= current || atomic.CompareAndSwapInt64(&m.maxConcurrent, current, concurrent) {
			break
		}
	}

	// Simulate processing time
	select {
	case <-time.After(m.processingTime):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	eventsCopy := make([]types.IngestionEvent, len(events))
	copy(eventsCopy, events)
	m.submitCalls = append(m.submitCalls, eventsCopy)
	m.callDurations = append(m.callDurations, time.Since(startTime))

	// Check if we should fail
	if m.shouldFail || (m.failAfterCalls > 0 && int(callIndex) >= m.failAfterCalls) {
		err := fmt.Errorf("mock ingestion error for call %d", callIndex)
		m.errors = append(m.errors, err)
		return nil, err
	}

	// Return successful response
	response := types.NewIngestionResponse(true)
	response.Usage = &types.IngestionUsage{
		EventsProcessed: len(events),
		EventsSkipped:   0,
		EventsFailed:    0,
	}

	m.responses = append(m.responses, response)
	m.errors = append(m.errors, nil)

	return response, nil
}

// GetSubmitCalls returns all calls made to SubmitBatch
func (m *MockIngestionClient) GetSubmitCalls() [][]types.IngestionEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([][]types.IngestionEvent, len(m.submitCalls))
	copy(calls, m.submitCalls)
	return calls
}

// GetCallCount returns the total number of calls made
func (m *MockIngestionClient) GetCallCount() int {
	return int(atomic.LoadInt64(&m.callIndex))
}

// SetShouldFail configures whether subsequent calls should fail
func (m *MockIngestionClient) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

// SetFailAfterCalls configures to fail after a certain number of successful calls
func (m *MockIngestionClient) SetFailAfterCalls(calls int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failAfterCalls = calls
}

// SetProcessingTime sets the artificial processing delay
func (m *MockIngestionClient) SetProcessingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processingTime = duration
}

// GetMaxConcurrentCalls returns the maximum concurrent calls observed
func (m *MockIngestionClient) GetMaxConcurrentCalls() int64 {
	return atomic.LoadInt64(&m.maxConcurrent)
}

// GetAverageCallDuration returns the average duration of all calls
func (m *MockIngestionClient) GetAverageCallDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.callDurations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range m.callDurations {
		total += d
	}
	return total / time.Duration(len(m.callDurations))
}

// Reset clears all recorded calls and responses
func (m *MockIngestionClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.submitCalls = make([][]types.IngestionEvent, 0)
	m.responses = make([]*types.IngestionResponse, 0)
	m.errors = make([]error, 0)
	m.callDurations = make([]time.Duration, 0)
	atomic.StoreInt64(&m.callIndex, 0)
	atomic.StoreInt64(&m.concurrentCalls, 0)
	atomic.StoreInt64(&m.maxConcurrent, 0)
	m.shouldFail = false
	m.failAfterCalls = 0
}

// CreateCreateTestIngestionEvent creates a test ingestion event
func CreateCreateTestIngestionEvent(id, eventType string) types.IngestionEvent {
	return types.IngestionEvent{
		ID:        id,
		Type:      types.EventType(eventType),
		Timestamp: time.Now(),
		Body: map[string]interface{}{
			"name": fmt.Sprintf("test-event-%s", id),
			"data": fmt.Sprintf("test-data-%s", id),
		},
	}
}

func TestIngestionQueue_BasicFunctionality(t *testing.T) {
	mockClient := NewMockIngestionClient()
	config := DefaultQueueConfig()
	config.FlushAt = 3
	config.FlushInterval = 100 * time.Millisecond

	queue := NewIngestionQueue(mockClient, config)
	defer queue.Shutdown(context.Background())

	t.Run("enqueue single event", func(t *testing.T) {
		event := CreateTestIngestionEvent("test-1", "trace-create")

		err := queue.Enqueue(event)
		assert.NoError(t, err)

		assert.Equal(t, 1, queue.Size())
		assert.False(t, queue.IsEmpty())
	})

	t.Run("flush triggers at threshold", func(t *testing.T) {
		// Add 2 more events to trigger flush (total 3)
		event2 := CreateTestIngestionEvent("test-2", "trace-create")
		event3 := CreateTestIngestionEvent("test-3", "trace-create")

		err := queue.Enqueue(event2)
		assert.NoError(t, err)

		err = queue.Enqueue(event3)
		assert.NoError(t, err)

		// Wait for flush to complete
		time.Sleep(150 * time.Millisecond)

		// Verify queue is empty and client was called
		assert.Equal(t, 0, queue.Size())
		assert.True(t, queue.IsEmpty())
		assert.Equal(t, 1, mockClient.GetCallCount())

		calls := mockClient.GetSubmitCalls()
		require.Len(t, calls, 1)
		assert.Len(t, calls[0], 3)
	})

	t.Run("periodic flush", func(t *testing.T) {
		// Add a single event that shouldn't trigger size-based flush
		event := CreateTestIngestionEvent("periodic-test", "trace-create")
		err := queue.Enqueue(event)
		assert.NoError(t, err)

		// Wait for periodic flush
		time.Sleep(150 * time.Millisecond)

		// Verify periodic flush occurred
		assert.Equal(t, 0, queue.Size())
		assert.Equal(t, 2, mockClient.GetCallCount()) // Previous flush + this one
	})
}

func TestIngestionQueue_ConcurrentEnqueue(t *testing.T) {
	mockClient := NewMockIngestionClient()
	config := DefaultQueueConfig()
	config.FlushAt = 50
	config.FlushInterval = 500 * time.Millisecond

	queue := NewIngestionQueue(mockClient, config)
	defer queue.Shutdown(context.Background())

	const numGoroutines = 10
	const eventsPerGoroutine = 100
	const totalEvents = numGoroutines * eventsPerGoroutine

	var wg sync.WaitGroup
	errChan := make(chan error, totalEvents)

	t.Run("concurrent enqueue operations", func(t *testing.T) {
		startTime := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < eventsPerGoroutine; j++ {
					eventID := fmt.Sprintf("routine-%d-event-%d", routineID, j)
					event := CreateTestIngestionEvent(eventID, "observation-create")

					if err := queue.Enqueue(event); err != nil {
						errChan <- err
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		duration := time.Since(startTime)
		t.Logf("Enqueued %d events in %v", totalEvents, duration)

		// Check for errors
		var errors []error
		for err := range errChan {
			errors = append(errors, err)
		}
		assert.Empty(t, errors, "Expected no enqueue errors")

		// Wait for all flushes to complete
		time.Sleep(1 * time.Second)

		// Verify statistics
		stats := queue.Stats()
		assert.Equal(t, int64(totalEvents), stats.EventsQueued)
		assert.True(t, stats.EventsProcessed > 0, "Expected some events to be processed")

		t.Logf("Queue stats: Queued=%d, Processed=%d, Batches=%d",
			stats.EventsQueued, stats.EventsProcessed, stats.BatchesSubmitted)
	})
}

func TestIngestionQueue_ErrorHandling(t *testing.T) {
	mockClient := NewMockIngestionClient()
	config := DefaultQueueConfig()
	config.FlushAt = 2
	config.FlushInterval = 100 * time.Millisecond
	config.MaxRetries = 3

	var flushResults []bool
	var flushErrors []error
	config.OnFlushEnd = func(batchSize int, success bool, err error) {
		flushResults = append(flushResults, success)
		flushErrors = append(flushErrors, err)
	}

	queue := NewIngestionQueue(mockClient, config)
	defer queue.Shutdown(context.Background())

	t.Run("retry on client errors", func(t *testing.T) {
		// Configure client to fail initially then succeed
		mockClient.SetFailAfterCalls(0)

		event1 := CreateTestIngestionEvent("retry-test-1", "trace-create")
		event2 := CreateTestIngestionEvent("retry-test-2", "trace-create")

		err := queue.Enqueue(event1)
		assert.NoError(t, err)

		err = queue.Enqueue(event2)
		assert.NoError(t, err)

		// Wait for retries to complete
		time.Sleep(2 * time.Second)

		// Check that multiple attempts were made
		callCount := mockClient.GetCallCount()
		assert.True(t, callCount > 1, "Expected multiple retry attempts, got %d", callCount)

		stats := queue.Stats()
		assert.True(t, stats.BatchesFailed > 0, "Expected some batches to fail")

		// Check flush callbacks
		require.NotEmpty(t, flushResults)
		assert.Contains(t, flushResults, false) // At least one failure
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		mockClient.Reset()
		mockClient.SetShouldFail(true) // Always fail

		var droppedEvents []types.IngestionEvent
		config.OnEventDrop = func(event types.IngestionEvent, reason string) {
			droppedEvents = append(droppedEvents, event)
		}

		// Create new queue with event drop callback
		queue2 := NewIngestionQueue(mockClient, config)
		defer queue2.Shutdown(context.Background())

		event := CreateTestIngestionEvent("max-retry-test", "trace-create")
		err := queue2.Enqueue(event)
		assert.NoError(t, err)

		// Force flush and wait
		queue2.Flush()
		time.Sleep(1 * time.Second)

		// Verify event was dropped after max retries
		assert.NotEmpty(t, droppedEvents, "Expected events to be dropped")
		assert.Contains(t, droppedEvents[0].ID, "max-retry-test")
	})
}

func TestIngestionQueue_ShutdownBehavior(t *testing.T) {
	mockClient := NewMockIngestionClient()
	config := DefaultQueueConfig()
	config.FlushAt = 100                    // High threshold to prevent auto-flush
	config.FlushInterval = 10 * time.Second // Long interval

	queue := NewIngestionQueue(mockClient, config)

	t.Run("graceful shutdown flushes pending events", func(t *testing.T) {
		// Add events that won't trigger auto-flush
		for i := 0; i < 5; i++ {
			event := CreateTestIngestionEvent(fmt.Sprintf("shutdown-test-%d", i), "trace-create")
			err := queue.Enqueue(event)
			assert.NoError(t, err)
		}

		assert.Equal(t, 5, queue.Size())

		// Shutdown should flush pending events
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := queue.Shutdown(ctx)
		assert.NoError(t, err)

		// Verify events were flushed
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 1, mockClient.GetCallCount())

		calls := mockClient.GetSubmitCalls()
		require.Len(t, calls, 1)
		assert.Len(t, calls[0], 5)

		// Verify queue is closed
		assert.True(t, queue.IsClosed())

		// Verify enqueue fails on closed queue
		event := CreateTestIngestionEvent("after-shutdown", "trace-create")
		err = queue.Enqueue(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "queue is closed")
	})

	t.Run("shutdown timeout", func(t *testing.T) {
		slowClient := NewMockIngestionClient()
		slowClient.SetProcessingTime(2 * time.Second) // Very slow processing

		queue2 := NewIngestionQueue(slowClient, config)

		// Add event to queue
		event := CreateTestIngestionEvent("timeout-test", "trace-create")
		err := queue2.Enqueue(event)
		assert.NoError(t, err)

		// Force flush to start slow processing
		queue2.Flush()
		time.Sleep(50 * time.Millisecond) // Give time for processing to start

		// Try to shutdown with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err = queue2.Shutdown(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestIngestionQueue_QueueSizeLimits(t *testing.T) {
	mockClient := NewMockIngestionClient()
	config := DefaultQueueConfig()
	config.FlushAt = 1000                   // High threshold
	config.FlushInterval = 10 * time.Second // Long interval
	config.MaxQueueSize = 5                 // Small queue

	var droppedEvents []types.IngestionEvent
	config.OnEventDrop = func(event types.IngestionEvent, reason string) {
		droppedEvents = append(droppedEvents, event)
	}

	queue := NewIngestionQueue(mockClient, config)
	defer queue.Shutdown(context.Background())

	t.Run("queue drops old events when full", func(t *testing.T) {
		// Fill up the queue beyond capacity
		for i := 0; i < 10; i++ {
			event := CreateTestIngestionEvent(fmt.Sprintf("capacity-test-%d", i), "trace-create")
			err := queue.Enqueue(event)
			assert.NoError(t, err)
		}

		// Queue should maintain max size
		assert.LessOrEqual(t, queue.Size(), 5)

		// Some events should have been dropped
		assert.NotEmpty(t, droppedEvents)

		stats := queue.Stats()
		assert.True(t, stats.EventsDropped > 0)
		assert.Equal(t, int64(10), stats.EventsQueued)
	})
}

func TestIngestionQueue_PerformanceMetrics(t *testing.T) {
	mockClient := NewMockIngestionClient()
	mockClient.SetProcessingTime(50 * time.Millisecond) // Simulate realistic processing time

	config := DefaultQueueConfig()
	config.FlushAt = 10
	config.FlushInterval = 100 * time.Millisecond

	queue := NewIngestionQueue(mockClient, config)
	defer queue.Shutdown(context.Background())

	t.Run("performance statistics tracking", func(t *testing.T) {
		numEvents := 50
		startTime := time.Now()

		// Enqueue events in batches
		for i := 0; i < numEvents; i++ {
			event := CreateTestIngestionEvent(fmt.Sprintf("perf-test-%d", i), "observation-create")
			err := queue.Enqueue(event)
			assert.NoError(t, err)
		}

		// Wait for processing to complete
		time.Sleep(2 * time.Second)

		totalTime := time.Since(startTime)
		stats := queue.Stats()

		// Verify statistics
		assert.Equal(t, int64(numEvents), stats.EventsQueued)
		assert.True(t, stats.EventsProcessed > 0)
		assert.True(t, stats.BatchesSubmitted > 0)
		assert.True(t, stats.TotalFlushTime > 0)
		assert.True(t, stats.AverageFlushTime > 0)
		assert.False(t, stats.LastFlushTime.IsZero())

		t.Logf("Performance metrics:")
		t.Logf("  Total time: %v", totalTime)
		t.Logf("  Events queued: %d", stats.EventsQueued)
		t.Logf("  Events processed: %d", stats.EventsProcessed)
		t.Logf("  Batches submitted: %d", stats.BatchesSubmitted)
		t.Logf("  Total flush time: %v", stats.TotalFlushTime)
		t.Logf("  Average flush time: %v", stats.AverageFlushTime)
		t.Logf("  Client max concurrent: %d", mockClient.GetMaxConcurrentCalls())
		t.Logf("  Client avg call duration: %v", mockClient.GetAverageCallDuration())

		// Performance assertions
		eventsPerSecond := float64(numEvents) / totalTime.Seconds()
		assert.True(t, eventsPerSecond > 10, "Expected at least 10 events/second, got %.2f", eventsPerSecond)
	})
}

func TestIngestionQueue_ConcurrentFlushes(t *testing.T) {
	mockClient := NewMockIngestionClient()
	config := DefaultQueueConfig()
	config.FlushAt = 5
	config.FlushInterval = 50 * time.Millisecond
	config.MaxRetries = 1

	queue := NewIngestionQueue(mockClient, config)
	defer queue.Shutdown(context.Background())

	t.Run("multiple concurrent flush triggers", func(t *testing.T) {
		const numOperations = 100
		var wg sync.WaitGroup

		// Rapidly enqueue events and trigger manual flushes
		wg.Add(2)

		// Enqueue events rapidly
		go func() {
			defer wg.Done()
			for i := 0; i < numOperations; i++ {
				event := CreateTestIngestionEvent(fmt.Sprintf("concurrent-flush-%d", i), "trace-create")
				queue.Enqueue(event)
				time.Sleep(1 * time.Millisecond)
			}
		}()

		// Trigger manual flushes
		go func() {
			defer wg.Done()
			for i := 0; i < numOperations/10; i++ {
				queue.Flush()
				time.Sleep(5 * time.Millisecond)
			}
		}()

		wg.Wait()

		// Wait for all operations to complete
		time.Sleep(2 * time.Second)

		stats := queue.Stats()
		assert.Equal(t, int64(numOperations), stats.EventsQueued)

		// Should have handled concurrent operations without errors
		assert.True(t, stats.EventsProcessed > 0)
		assert.True(t, stats.BatchesSubmitted > 0)

		t.Logf("Concurrent flush test - Batches: %d, Max concurrent calls: %d",
			stats.BatchesSubmitted, mockClient.GetMaxConcurrentCalls())
	})
}

func TestIngestionQueue_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	mockClient := NewMockIngestionClient()
	mockClient.SetProcessingTime(1 * time.Millisecond) // Fast processing for stress test

	config := DefaultQueueConfig()
	config.FlushAt = 20
	config.FlushInterval = 10 * time.Millisecond
	config.MaxQueueSize = 1000

	queue := NewIngestionQueue(mockClient, config)
	defer queue.Shutdown(context.Background())

	t.Run("high-throughput stress test", func(t *testing.T) {
		const numGoroutines = 20
		const eventsPerGoroutine = 500
		const totalEvents = numGoroutines * eventsPerGoroutine

		var wg sync.WaitGroup
		var enqueueErrors int64

		startTime := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < eventsPerGoroutine; j++ {
					event := CreateTestIngestionEvent(
						fmt.Sprintf("stress-%d-%d", routineID, j),
						"observation-update",
					)

					if err := queue.Enqueue(event); err != nil {
						atomic.AddInt64(&enqueueErrors, 1)
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// Wait for processing to complete
		time.Sleep(5 * time.Second)

		stats := queue.Stats()

		t.Logf("Stress test results:")
		t.Logf("  Duration: %v", duration)
		t.Logf("  Enqueue errors: %d", enqueueErrors)
		t.Logf("  Events queued: %d", stats.EventsQueued)
		t.Logf("  Events processed: %d", stats.EventsProcessed)
		t.Logf("  Events dropped: %d", stats.EventsDropped)
		t.Logf("  Batches submitted: %d", stats.BatchesSubmitted)
		t.Logf("  Throughput: %.2f events/second", float64(totalEvents)/duration.Seconds())
		t.Logf("  Max concurrent client calls: %d", mockClient.GetMaxConcurrentCalls())

		// Assertions
		assert.Equal(t, int64(0), enqueueErrors, "Expected no enqueue errors")
		assert.True(t, stats.EventsProcessed > 0, "Expected events to be processed")

		// Performance requirement: Should handle at least 1000 events/second
		eventsPerSecond := float64(totalEvents) / duration.Seconds()
		assert.True(t, eventsPerSecond >= 1000,
			"Expected at least 1000 events/second, got %.2f", eventsPerSecond)
	})
}
