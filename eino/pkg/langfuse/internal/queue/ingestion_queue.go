package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"eino/pkg/langfuse/api/resources/ingestion/types"
)

// Common queue errors
var (
	ErrQueueClosed = errors.New("queue is closed")
)

// IngestionClient interface defines the methods needed to submit ingestion requests
type IngestionClient interface {
	SubmitBatch(ctx context.Context, events []types.IngestionEvent) (*types.IngestionResponse, error)
}

// IngestionQueue manages batching and async submission of ingestion events
type IngestionQueue struct {
	client        IngestionClient
	buffer        []types.IngestionEvent
	mu            sync.RWMutex
	flushAt       int
	flushInterval time.Duration

	// Background processing
	ticker     *time.Ticker
	stopCh     chan struct{}
	flushCh    chan struct{}
	shutdownCh chan struct{}
	wg         sync.WaitGroup

	// State management
	closed bool

	// Statistics
	stats *QueueStats

	// Configuration
	maxRetries   int
	retryBackoff time.Duration

	// Event hooks
	onFlushStart func(batchSize int)
	onFlushEnd   func(batchSize int, success bool, err error)
	onEventDrop  func(event types.IngestionEvent, reason string)
}

// QueueStats tracks queue performance metrics
type QueueStats struct {
	mu               sync.RWMutex
	EventsQueued     int64
	EventsProcessed  int64
	EventsFailed     int64
	EventsDropped    int64
	BatchesSubmitted int64
	BatchesFailed    int64
	TotalFlushTime   time.Duration
	AverageFlushTime time.Duration
	LastFlushTime    time.Time
	QueueSize        int
	MaxQueueSize     int
}

// QueueConfig holds configuration for the ingestion queue
type QueueConfig struct {
	FlushAt       int
	FlushInterval time.Duration
	MaxRetries    int
	RetryBackoff  time.Duration
	MaxQueueSize  int
	OnFlushStart  func(batchSize int)
	OnFlushEnd    func(batchSize int, success bool, err error)
	OnEventDrop   func(event types.IngestionEvent, reason string)
}

// DefaultQueueConfig returns a default queue configuration
func DefaultQueueConfig() *QueueConfig {
	return &QueueConfig{
		FlushAt:       15,
		FlushInterval: 10 * time.Second,
		MaxRetries:    3,
		RetryBackoff:  1 * time.Second,
		MaxQueueSize:  1000,
	}
}

// NewIngestionQueue creates a new ingestion queue with the given configuration
func NewIngestionQueue(client IngestionClient, config *QueueConfig) *IngestionQueue {
	if config == nil {
		config = DefaultQueueConfig()
	}

	queue := &IngestionQueue{
		client:        client,
		buffer:        make([]types.IngestionEvent, 0, config.FlushAt),
		flushAt:       config.FlushAt,
		flushInterval: config.FlushInterval,
		maxRetries:    config.MaxRetries,
		retryBackoff:  config.RetryBackoff,
		stopCh:        make(chan struct{}),
		flushCh:       make(chan struct{}, 1),
		shutdownCh:    make(chan struct{}),
		closed:        false,
		stats:         &QueueStats{MaxQueueSize: config.MaxQueueSize},
		onFlushStart:  config.OnFlushStart,
		onFlushEnd:    config.OnFlushEnd,
		onEventDrop:   config.OnEventDrop,
	}

	// Start background worker
	queue.startWorker()

	return queue
}

// Enqueue adds an event to the queue for processing
func (q *IngestionQueue) Enqueue(event types.IngestionEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	// Validate the event before queueing
	if err := event.Validate(); err != nil {
		q.stats.mu.Lock()
		q.stats.EventsFailed++
		q.stats.mu.Unlock()
		return fmt.Errorf("event validation failed: %w", err)
	}

	// Check queue size limits
	if len(q.buffer) >= q.stats.MaxQueueSize {
		// Drop the oldest event to make room
		droppedEvent := q.buffer[0]
		q.buffer = q.buffer[1:]
		q.stats.mu.Lock()
		q.stats.EventsDropped++
		q.stats.mu.Unlock()

		if q.onEventDrop != nil {
			q.onEventDrop(droppedEvent, "queue_full")
		}
	}

	// Add event to buffer
	q.buffer = append(q.buffer, event)
	q.stats.mu.Lock()
	q.stats.EventsQueued++
	q.stats.QueueSize = len(q.buffer)
	if q.stats.QueueSize > q.stats.MaxQueueSize {
		q.stats.MaxQueueSize = q.stats.QueueSize
	}
	q.stats.mu.Unlock()

	// Trigger flush if buffer is full
	if len(q.buffer) >= q.flushAt {
		select {
		case q.flushCh <- struct{}{}:
		default:
			// Channel full, flush already triggered
		}
	}

	return nil
}

// Flush forces an immediate flush of all pending events
func (q *IngestionQueue) Flush() error {
	// Trigger flush and wait for completion
	select {
	case q.flushCh <- struct{}{}:
	default:
		// Channel full, flush already triggered
	}

	// Give some time for the flush to complete
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Size returns the current number of events in the queue
func (q *IngestionQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.buffer)
}

// Stats returns a copy of the current queue statistics
func (q *IngestionQueue) Stats() QueueStats {
	q.stats.mu.RLock()
	defer q.stats.mu.RUnlock()

	// Create a copy to avoid data races
	stats := *q.stats
	return stats
}

// Shutdown gracefully shuts down the queue, flushing any pending events
func (q *IngestionQueue) Shutdown(ctx context.Context) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	q.mu.Unlock()

	// Stop the ticker
	if q.ticker != nil {
		q.ticker.Stop()
	}

	// Signal shutdown
	close(q.shutdownCh)

	// Wait for worker to finish or timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// startWorker starts the background worker goroutine
func (q *IngestionQueue) startWorker() {
	q.ticker = time.NewTicker(q.flushInterval)
	q.wg.Add(1)

	go q.worker()
}

// worker is the main background processing loop
func (q *IngestionQueue) worker() {
	defer q.wg.Done()
	defer q.ticker.Stop()

	for {
		select {
		case <-q.ticker.C:
			q.periodicFlush()
		case <-q.flushCh:
			q.forceFlush()
		case <-q.shutdownCh:
			q.finalFlush()
			return
		}
	}
}

// periodicFlush performs a periodic flush if there are events in the buffer
func (q *IngestionQueue) periodicFlush() {
	q.mu.RLock()
	hasEvents := len(q.buffer) > 0
	q.mu.RUnlock()

	if hasEvents {
		q.flushBuffer()
	}
}

// forceFlush performs an immediate flush
func (q *IngestionQueue) forceFlush() {
	q.flushBuffer()
}

// finalFlush performs a final flush during shutdown
func (q *IngestionQueue) finalFlush() {
	q.flushBuffer()
}

// flushBuffer sends the current buffer to the ingestion client
func (q *IngestionQueue) flushBuffer() {
	q.mu.Lock()
	if len(q.buffer) == 0 {
		q.mu.Unlock()
		return
	}

	// Take a copy of the buffer and clear it
	events := make([]types.IngestionEvent, len(q.buffer))
	copy(events, q.buffer)
	q.buffer = q.buffer[:0] // Clear buffer but keep capacity
	batchSize := len(events)
	q.mu.Unlock()

	// Update stats
	q.stats.mu.Lock()
	q.stats.QueueSize = 0
	q.stats.BatchesSubmitted++
	q.stats.mu.Unlock()

	// Call flush start hook
	if q.onFlushStart != nil {
		q.onFlushStart(batchSize)
	}

	startTime := time.Now()
	success := false
	var flushErr error

	// Submit with retries
	ctx := context.Background() // TODO: Make this configurable
	for attempt := 0; attempt <= q.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(q.retryBackoff * time.Duration(attempt))
		}

		response, err := q.client.SubmitBatch(ctx, events)
		if err == nil && response != nil && response.Success {
			// Success
			q.stats.mu.Lock()
			q.stats.EventsProcessed += int64(batchSize)
			q.stats.LastFlushTime = time.Now()
			flushTime := time.Since(startTime)
			q.stats.TotalFlushTime += flushTime
			q.stats.AverageFlushTime = q.stats.TotalFlushTime / time.Duration(q.stats.BatchesSubmitted)
			q.stats.mu.Unlock()

			success = true
			break
		}

		flushErr = err
		if response != nil && response.HasErrors() {
			// Handle partial failures
			q.handlePartialFailure(response, events)
		}
	}

	if !success {
		// All retries failed
		q.stats.mu.Lock()
		q.stats.EventsFailed += int64(batchSize)
		q.stats.BatchesFailed++
		q.stats.mu.Unlock()

		// Drop events that couldn't be processed
		for _, event := range events {
			if q.onEventDrop != nil {
				q.onEventDrop(event, "max_retries_exceeded")
			}
		}
	}

	// Call flush end hook
	if q.onFlushEnd != nil {
		q.onFlushEnd(batchSize, success, flushErr)
	}
}

// handlePartialFailure handles cases where some events succeeded and some failed
func (q *IngestionQueue) handlePartialFailure(response *types.IngestionResponse, events []types.IngestionEvent) {
	if response.Usage != nil {
		q.stats.mu.Lock()
		q.stats.EventsProcessed += int64(response.Usage.EventsProcessed)
		q.stats.EventsFailed += int64(response.Usage.EventsFailed)
		q.stats.mu.Unlock()
	}

	// Log failed events for debugging
	for _, ingestionErr := range response.Errors {
		if ingestionErr.EventID != nil {
			// Find the corresponding event and potentially retry or drop it
			for _, event := range events {
				if event.ID == *ingestionErr.EventID {
					if q.onEventDrop != nil {
						q.onEventDrop(event, fmt.Sprintf("ingestion_error: %s", ingestionErr.Message))
					}
					break
				}
			}
		}
	}
}

// IsEmpty returns true if the queue is empty
func (q *IngestionQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.buffer) == 0
}

// IsClosed returns true if the queue is closed
func (q *IngestionQueue) IsClosed() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.closed
}
