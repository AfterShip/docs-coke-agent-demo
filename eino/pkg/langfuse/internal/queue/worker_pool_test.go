package queue

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eino/pkg/langfuse/api/resources/ingestion/types"
)

// FailingMockClient simulates various failure scenarios
type FailingMockClient struct {
	*MockIngestionClient
	failureRate  float64 // 0.0 to 1.0
	panicRate    float64 // 0.0 to 1.0
	slowRate     float64 // 0.0 to 1.0
	slowDuration time.Duration
	callCounter  int64
}

// NewFailingMockClient creates a mock client that can simulate failures
func NewFailingMockClient() *FailingMockClient {
	return &FailingMockClient{
		MockIngestionClient: NewMockIngestionClient(),
		slowDuration:        100 * time.Millisecond,
	}
}

// SubmitBatch implements IngestionClient with configurable failures
func (f *FailingMockClient) SubmitBatch(ctx context.Context, events []types.IngestionEvent) (*types.IngestionResponse, error) {
	callNum := atomic.AddInt64(&f.callCounter, 1)

	// Simulate panic based on panic rate
	if f.panicRate > 0 && float64(callNum%100)/100.0 < f.panicRate {
		panic(fmt.Sprintf("simulated panic in worker %d", callNum))
	}

	// Simulate slow processing
	if f.slowRate > 0 && float64(callNum%100)/100.0 < f.slowRate {
		select {
		case <-time.After(f.slowDuration):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Simulate failure based on failure rate
	if f.failureRate > 0 && float64(callNum%100)/100.0 < f.failureRate {
		return nil, fmt.Errorf("simulated failure for call %d", callNum)
	}

	return f.MockIngestionClient.SubmitBatch(ctx, events)
}

func TestWorkerPool_BasicFunctionality(t *testing.T) {
	mockClient := NewMockIngestionClient()
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 3
	config.WorkBufferSize = 10

	pool := NewWorkerPool(mockClient, config)
	defer pool.Shutdown(context.Background())

	t.Run("worker pool initialization", func(t *testing.T) {
		assert.NotNil(t, pool)
		assert.False(t, pool.IsShuttingDown())

		stats := pool.Stats()
		assert.Equal(t, 0, stats.WorkersActive)
		assert.Equal(t, int64(0), stats.WorkItemsQueued)
		assert.Equal(t, int64(0), stats.WorkItemsProcessed)
	})

	t.Run("submit single work item", func(t *testing.T) {
		events := []types.IngestionEvent{
			TestIngestionEvent("worker-test-1", "trace-create"),
			TestIngestionEvent("worker-test-2", "observation-create"),
		}

		err := pool.SubmitWork(events)
		assert.NoError(t, err)

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		stats := pool.Stats()
		assert.Equal(t, int64(1), stats.WorkItemsQueued)
		assert.True(t, stats.WorkItemsProcessed >= 1)

		// Verify client received the work
		assert.Equal(t, 1, mockClient.GetCallCount())
		calls := mockClient.GetSubmitCalls()
		require.Len(t, calls, 1)
		assert.Len(t, calls[0], 2)
	})

	t.Run("submit multiple work items", func(t *testing.T) {
		mockClient.Reset()

		numItems := 10
		for i := 0; i < numItems; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("multi-test-%d", i), "trace-create"),
			}
			err := pool.SubmitWork(events)
			assert.NoError(t, err)
		}

		// Wait for processing
		time.Sleep(500 * time.Millisecond)

		stats := pool.Stats()
		assert.Equal(t, int64(numItems), stats.WorkItemsQueued)
		assert.True(t, stats.WorkItemsProcessed >= int64(numItems))
		assert.Equal(t, numItems, mockClient.GetCallCount())
	})
}

func TestWorkerPool_ConcurrentWorkers(t *testing.T) {
	mockClient := NewMockIngestionClient()
	mockClient.SetProcessingTime(50 * time.Millisecond) // Simulate processing time

	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 5
	config.WorkBufferSize = 50

	pool := NewWorkerPool(mockClient, config)
	defer pool.Shutdown(context.Background())

	t.Run("concurrent work processing", func(t *testing.T) {
		numItems := 25
		startTime := time.Now()

		// Submit work items rapidly
		for i := 0; i < numItems; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("concurrent-%d", i), "observation-create"),
			}
			err := pool.SubmitWork(events)
			assert.NoError(t, err)
		}

		// Wait for all processing to complete
		time.Sleep(2 * time.Second)

		duration := time.Since(startTime)
		stats := pool.Stats()
		maxConcurrent := mockClient.GetMaxConcurrentCalls()

		t.Logf("Concurrent processing results:")
		t.Logf("  Duration: %v", duration)
		t.Logf("  Work items processed: %d", stats.WorkItemsProcessed)
		t.Logf("  Max concurrent calls: %d", maxConcurrent)
		t.Logf("  Average processing time: %v", stats.AverageProcessingTime)

		// Verify all work was processed
		assert.Equal(t, int64(numItems), stats.WorkItemsQueued)
		assert.True(t, stats.WorkItemsProcessed >= int64(numItems))

		// Verify concurrent execution (should be > 1 for multiple workers)
		assert.True(t, maxConcurrent > 1, "Expected concurrent execution, got max concurrent = %d", maxConcurrent)
		assert.True(t, maxConcurrent <= 5, "Max concurrent should not exceed number of workers")

		// Performance check: with 5 workers, should be significantly faster than sequential
		sequentialTime := time.Duration(numItems) * 50 * time.Millisecond
		expectedMinSpeedup := 2.0 // At least 2x speedup expected
		actualSpeedup := float64(sequentialTime) / float64(duration)

		t.Logf("  Sequential time would be: %v", sequentialTime)
		t.Logf("  Actual speedup: %.2fx", actualSpeedup)

		assert.True(t, actualSpeedup >= expectedMinSpeedup,
			"Expected at least %.1fx speedup, got %.2fx", expectedMinSpeedup, actualSpeedup)
	})
}

func TestWorkerPool_ErrorHandlingAndRetries(t *testing.T) {
	failingClient := NewFailingMockClient()
	failingClient.failureRate = 0.3 // 30% failure rate

	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 2
	config.MaxRetries = 2
	config.RetryBackoff = 10 * time.Millisecond

	var workResults []bool
	var workErrors []error
	config.OnWorkEnd = func(result *WorkResult) {
		workResults = append(workResults, result.Success)
		workErrors = append(workErrors, result.Error)
	}

	pool := NewWorkerPool(failingClient, config)
	defer pool.Shutdown(context.Background())

	t.Run("retry failed work items", func(t *testing.T) {
		numItems := 20

		for i := 0; i < numItems; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("retry-test-%d", i), "trace-update"),
			}
			err := pool.SubmitWork(events)
			assert.NoError(t, err)
		}

		// Wait for processing and retries
		time.Sleep(3 * time.Second)

		stats := pool.Stats()

		t.Logf("Error handling results:")
		t.Logf("  Work items queued: %d", stats.WorkItemsQueued)
		t.Logf("  Work items processed: %d", stats.WorkItemsProcessed)
		t.Logf("  Work items failed: %d", stats.WorkItemsFailed)
		t.Logf("  Total client calls: %d", failingClient.GetCallCount())

		// Should have made more calls than work items due to retries
		assert.True(t, failingClient.GetCallCount() > numItems,
			"Expected retries to cause more calls than work items")

		// Should have both successes and failures due to failure rate
		assert.True(t, stats.WorkItemsProcessed > 0, "Expected some work to succeed")
		assert.True(t, stats.WorkItemsFailed > 0, "Expected some work to fail with 30% failure rate")

		// Verify callbacks were called
		assert.NotEmpty(t, workResults)
		assert.Contains(t, workResults, true)  // Some successes
		assert.Contains(t, workResults, false) // Some failures
	})
}

func TestWorkerPool_WorkerPanicRecovery(t *testing.T) {
	panicClient := NewFailingMockClient()
	panicClient.panicRate = 0.1 // 10% panic rate

	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 3

	var panicCount int64
	var panickedWorkers []int
	config.OnWorkerPanic = func(workerID int, err interface{}) {
		atomic.AddInt64(&panicCount, 1)
		panickedWorkers = append(panickedWorkers, workerID)
		t.Logf("Worker %d panicked: %v", workerID, err)
	}

	pool := NewWorkerPool(panicClient, config)
	defer pool.Shutdown(context.Background())

	t.Run("worker panic recovery", func(t *testing.T) {
		numItems := 50

		for i := 0; i < numItems; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("panic-test-%d", i), "observation-update"),
			}
			err := pool.SubmitWork(events)
			assert.NoError(t, err)
		}

		// Wait for processing
		time.Sleep(3 * time.Second)

		stats := pool.Stats()
		panics := atomic.LoadInt64(&panicCount)

		t.Logf("Panic recovery results:")
		t.Logf("  Worker panics: %d", panics)
		t.Logf("  Panicked workers: %v", panickedWorkers)
		t.Logf("  Work items processed: %d", stats.WorkItemsProcessed)
		t.Logf("  Total panics in stats: %d", stats.WorkerPanics)

		// Should have some panics due to panic rate
		assert.True(t, panics > 0, "Expected some worker panics")
		assert.Equal(t, panics, stats.WorkerPanics)

		// Should still process most work despite panics
		assert.True(t, stats.WorkItemsProcessed > int64(numItems*0.7),
			"Expected at least 70%% of work to be processed despite panics")

		// Pool should still be functional after panics
		assert.False(t, pool.IsShuttingDown())
	})
}

func TestWorkerPool_Shutdown(t *testing.T) {
	mockClient := NewMockIngestionClient()
	mockClient.SetProcessingTime(100 * time.Millisecond)

	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 3
	config.WorkBufferSize = 20

	pool := NewWorkerPool(mockClient, config)

	t.Run("graceful shutdown with pending work", func(t *testing.T) {
		// Submit work that will take time to process
		numItems := 10
		for i := 0; i < numItems; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("shutdown-test-%d", i), "trace-create"),
			}
			err := pool.SubmitWork(events)
			assert.NoError(t, err)
		}

		// Give workers time to start processing
		time.Sleep(50 * time.Millisecond)

		// Shutdown should wait for workers to finish current work
		shutdownStart := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := pool.Shutdown(ctx)
		shutdownDuration := time.Since(shutdownStart)

		assert.NoError(t, err)
		assert.True(t, pool.IsShuttingDown())

		t.Logf("Shutdown took %v", shutdownDuration)

		// Should have processed some work before shutdown
		stats := pool.Stats()
		assert.True(t, stats.WorkItemsProcessed > 0)

		// Should not accept new work after shutdown
		events := []types.IngestionEvent{
			TestIngestionEvent("after-shutdown", "trace-create"),
		}
		err = pool.SubmitWork(events)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shutting down")
	})

	t.Run("shutdown timeout", func(t *testing.T) {
		slowClient := NewMockIngestionClient()
		slowClient.SetProcessingTime(2 * time.Second) // Very slow

		pool2 := NewWorkerPool(slowClient, config)

		// Submit work that will take long to process
		events := []types.IngestionEvent{
			TestIngestionEvent("slow-shutdown-test", "trace-create"),
		}
		err := pool2.SubmitWork(events)
		assert.NoError(t, err)

		// Give time for processing to start
		time.Sleep(50 * time.Millisecond)

		// Try to shutdown with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err = pool2.Shutdown(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestWorkerPool_LoadBalancing(t *testing.T) {
	mockClient := NewMockIngestionClient()
	mockClient.SetProcessingTime(20 * time.Millisecond)

	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 4
	config.WorkBufferSize = 100

	// Track which workers process work
	var workerUsage = make(map[int]int)
	var mu sync.Mutex

	config.OnWorkEnd = func(result *WorkResult) {
		mu.Lock()
		workerUsage[result.WorkerID]++
		mu.Unlock()
	}

	pool := NewWorkerPool(mockClient, config)
	defer pool.Shutdown(context.Background())

	t.Run("work distribution across workers", func(t *testing.T) {
		numItems := 40 // 10 items per worker if evenly distributed

		for i := 0; i < numItems; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("load-balance-%d", i), "observation-create"),
			}
			err := pool.SubmitWork(events)
			assert.NoError(t, err)
		}

		// Wait for all processing
		time.Sleep(2 * time.Second)

		mu.Lock()
		usage := make(map[int]int)
		for k, v := range workerUsage {
			usage[k] = v
		}
		mu.Unlock()

		t.Logf("Worker usage distribution:")
		for workerID := 0; workerID < config.NumWorkers; workerID++ {
			count := usage[workerID]
			percentage := float64(count) / float64(numItems) * 100
			t.Logf("  Worker %d: %d items (%.1f%%)", workerID, count, percentage)
		}

		// All workers should have been used
		assert.Len(t, usage, config.NumWorkers, "All workers should have processed work")

		// Distribution should be reasonably balanced (within 50% of average)
		averagePerWorker := float64(numItems) / float64(config.NumWorkers)
		for workerID, count := range usage {
			deviation := math.Abs(float64(count)-averagePerWorker) / averagePerWorker
			assert.True(t, deviation < 0.5,
				"Worker %d processed %d items, expected ~%.1f (%.1f%% deviation)",
				workerID, count, averagePerWorker, deviation*100)
		}
	})
}

func TestWorkerPool_HighThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high throughput test in short mode")
	}

	mockClient := NewMockIngestionClient()
	mockClient.SetProcessingTime(1 * time.Millisecond) // Fast processing

	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 8
	config.WorkBufferSize = 500
	config.MaxRetries = 1 // Minimal retries for throughput test
	config.RetryBackoff = 1 * time.Millisecond

	pool := NewWorkerPool(mockClient, config)
	defer pool.Shutdown(context.Background())

	t.Run("high throughput stress test", func(t *testing.T) {
		const numItems = 1000
		const targetThroughput = 500 // items per second

		startTime := time.Now()

		// Submit work items as fast as possible
		var submitErrors int64
		for i := 0; i < numItems; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("throughput-%d", i), "score-create"),
				TestIngestionEvent(fmt.Sprintf("throughput-%d-b", i), "observation-update"),
			}

			if err := pool.SubmitWork(events); err != nil {
				atomic.AddInt64(&submitErrors, 1)
			}
		}

		submitDuration := time.Since(startTime)

		// Wait for processing to complete
		time.Sleep(5 * time.Second)

		totalDuration := time.Since(startTime)
		stats := pool.Stats()

		t.Logf("High throughput test results:")
		t.Logf("  Submit duration: %v", submitDuration)
		t.Logf("  Total duration: %v", totalDuration)
		t.Logf("  Submit errors: %d", submitErrors)
		t.Logf("  Work items queued: %d", stats.WorkItemsQueued)
		t.Logf("  Work items processed: %d", stats.WorkItemsProcessed)
		t.Logf("  Work items failed: %d", stats.WorkItemsFailed)
		t.Logf("  Average processing time: %v", stats.AverageProcessingTime)
		t.Logf("  Max concurrent calls: %d", mockClient.GetMaxConcurrentCalls())

		// Calculate throughput
		throughput := float64(numItems) / totalDuration.Seconds()
		t.Logf("  Actual throughput: %.1f items/second", throughput)

		// Assertions
		assert.Equal(t, int64(0), submitErrors, "Should not have submit errors")
		assert.Equal(t, int64(numItems), stats.WorkItemsQueued)
		assert.True(t, stats.WorkItemsProcessed >= int64(numItems*0.95),
			"Should process at least 95%% of items")

		// Performance requirement
		assert.True(t, throughput >= float64(targetThroughput),
			"Expected throughput >= %d items/sec, got %.1f", targetThroughput, throughput)

		// Should utilize multiple workers concurrently
		assert.True(t, mockClient.GetMaxConcurrentCalls() >= 4,
			"Expected high concurrency with 8 workers")
	})
}

func TestWorkerPool_QueueCapacity(t *testing.T) {
	mockClient := NewMockIngestionClient()
	mockClient.SetProcessingTime(100 * time.Millisecond) // Slow processing to fill queue

	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 2
	config.WorkBufferSize = 5 // Small buffer

	pool := NewWorkerPool(mockClient, config)
	defer pool.Shutdown(context.Background())

	t.Run("queue capacity limits", func(t *testing.T) {
		// Submit work rapidly to fill the queue
		var successfulSubmits int
		var failedSubmits int

		for i := 0; i < 20; i++ {
			events := []types.IngestionEvent{
				TestIngestionEvent(fmt.Sprintf("capacity-%d", i), "trace-create"),
			}

			err := pool.SubmitWork(events)
			if err != nil {
				failedSubmits++
			} else {
				successfulSubmits++
			}
		}

		t.Logf("Queue capacity test:")
		t.Logf("  Successful submits: %d", successfulSubmits)
		t.Logf("  Failed submits: %d", failedSubmits)
		t.Logf("  Current queue size: %d", pool.QueueSize())

		// Should have some failed submits due to queue capacity
		assert.True(t, failedSubmits > 0, "Expected some submits to fail due to queue capacity")
		assert.True(t, successfulSubmits > 0, "Expected some submits to succeed")
		assert.True(t, pool.QueueSize() <= config.WorkBufferSize,
			"Queue size should not exceed buffer size")

		// Wait for some processing and try again
		time.Sleep(500 * time.Millisecond)

		// Should be able to submit more work now
		events := []types.IngestionEvent{
			TestIngestionEvent("after-drain", "trace-create"),
		}
		err := pool.SubmitWork(events)
		assert.NoError(t, err, "Should be able to submit after queue drains")
	})
}
