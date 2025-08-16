package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"eino/pkg/langfuse/api/resources/ingestion/types"
)

// WorkerPool manages a pool of workers for processing ingestion batches
type WorkerPool struct {
	client       IngestionClient
	workers      []*Worker
	workCh       chan *WorkItem
	resultCh     chan *WorkResult
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	
	// Configuration
	numWorkers   int
	maxRetries   int
	retryBackoff time.Duration
	
	// Statistics
	stats        *WorkerPoolStats
	
	// Event hooks
	onWorkStart  func(item *WorkItem)
	onWorkEnd    func(result *WorkResult)
	onWorkerPanic func(workerID int, err interface{})
}

// WorkItem represents a batch of events to be processed
type WorkItem struct {
	ID       string
	Events   []types.IngestionEvent
	Retries  int
	Created  time.Time
	Started  time.Time
}

// WorkResult represents the result of processing a work item
type WorkResult struct {
	Item       *WorkItem
	Response   *types.IngestionResponse
	Success    bool
	Error      error
	ProcessingTime time.Duration
	WorkerID   int
}

// Worker represents a single worker in the pool
type Worker struct {
	ID       int
	pool     *WorkerPool
	workCh   <-chan *WorkItem
	resultCh chan<- *WorkResult
}

// WorkerPoolStats tracks worker pool performance metrics
type WorkerPoolStats struct {
	mu                  sync.RWMutex
	WorkersActive       int
	WorkItemsQueued     int64
	WorkItemsProcessed  int64
	WorkItemsFailed     int64
	TotalProcessingTime time.Duration
	AverageProcessingTime time.Duration
	WorkerPanics        int64
	LastProcessedTime   time.Time
}

// WorkerPoolConfig holds configuration for the worker pool
type WorkerPoolConfig struct {
	NumWorkers     int
	MaxRetries     int
	RetryBackoff   time.Duration
	WorkBufferSize int
	OnWorkStart    func(item *WorkItem)
	OnWorkEnd      func(result *WorkResult)
	OnWorkerPanic  func(workerID int, err interface{})
}

// DefaultWorkerPoolConfig returns a default worker pool configuration
func DefaultWorkerPoolConfig() *WorkerPoolConfig {
	return &WorkerPoolConfig{
		NumWorkers:     3,
		MaxRetries:     3,
		RetryBackoff:   1 * time.Second,
		WorkBufferSize: 100,
	}
}

// NewWorkerPool creates a new worker pool with the given configuration
func NewWorkerPool(client IngestionClient, config *WorkerPoolConfig) *WorkerPool {
	if config == nil {
		config = DefaultWorkerPoolConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &WorkerPool{
		client:       client,
		numWorkers:   config.NumWorkers,
		maxRetries:   config.MaxRetries,
		retryBackoff: config.RetryBackoff,
		workCh:       make(chan *WorkItem, config.WorkBufferSize),
		resultCh:     make(chan *WorkResult, config.WorkBufferSize),
		ctx:          ctx,
		cancel:       cancel,
		stats:        &WorkerPoolStats{},
		onWorkStart:  config.OnWorkStart,
		onWorkEnd:    config.OnWorkEnd,
		onWorkerPanic: config.OnWorkerPanic,
	}
	
	// Create and start workers
	pool.workers = make([]*Worker, config.NumWorkers)
	for i := 0; i < config.NumWorkers; i++ {
		worker := &Worker{
			ID:       i,
			pool:     pool,
			workCh:   pool.workCh,
			resultCh: pool.resultCh,
		}
		pool.workers[i] = worker
		pool.wg.Add(1)
		go worker.run()
	}
	
	// Start result processor
	pool.wg.Add(1)
	go pool.processResults()
	
	return pool
}

// SubmitWork submits a batch of events for processing
func (wp *WorkerPool) SubmitWork(events []types.IngestionEvent) error {
	if len(events) == 0 {
		return fmt.Errorf("cannot submit empty work batch")
	}
	
	item := &WorkItem{
		ID:      generateWorkItemID(),
		Events:  events,
		Retries: 0,
		Created: time.Now(),
	}
	
	select {
	case wp.workCh <- item:
		wp.stats.mu.Lock()
		wp.stats.WorkItemsQueued++
		wp.stats.mu.Unlock()
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("work queue is full")
	}
}

// Stats returns a copy of the current worker pool statistics
func (wp *WorkerPool) Stats() WorkerPoolStats {
	wp.stats.mu.RLock()
	defer wp.stats.mu.RUnlock()
	
	// Create a copy to avoid data races
	stats := *wp.stats
	return stats
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	// Close work channel to stop accepting new work
	close(wp.workCh)
	
	// Cancel context to signal workers to stop
	wp.cancel()
	
	// Wait for all workers to finish or timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// run is the main worker loop
func (w *Worker) run() {
	defer w.pool.wg.Done()
	defer w.recoverPanic()
	
	for {
		select {
		case item, ok := <-w.workCh:
			if !ok {
				// Channel closed, shutdown
				return
			}
			w.processWork(item)
		case <-w.pool.ctx.Done():
			return
		}
	}
}

// processWork processes a single work item
func (w *Worker) processWork(item *WorkItem) {
	item.Started = time.Now()
	
	w.pool.stats.mu.Lock()
	w.pool.stats.WorkersActive++
	w.pool.stats.mu.Unlock()
	
	defer func() {
		w.pool.stats.mu.Lock()
		w.pool.stats.WorkersActive--
		w.pool.stats.mu.Unlock()
	}()
	
	// Call work start hook
	if w.pool.onWorkStart != nil {
		w.pool.onWorkStart(item)
	}
	
	startTime := time.Now()
	response, err := w.pool.client.SubmitBatch(w.pool.ctx, item.Events)
	processingTime := time.Since(startTime)
	
	result := &WorkResult{
		Item:           item,
		Response:       response,
		Success:        err == nil && response != nil && response.Success,
		Error:          err,
		ProcessingTime: processingTime,
		WorkerID:       w.ID,
	}
	
	// Send result for processing
	select {
	case w.resultCh <- result:
	case <-w.pool.ctx.Done():
		return
	}
}

// processResults processes work results and handles retries
func (wp *WorkerPool) processResults() {
	defer wp.wg.Done()
	
	for {
		select {
		case result, ok := <-wp.resultCh:
			if !ok {
				return
			}
			wp.handleResult(result)
		case <-wp.ctx.Done():
			return
		}
	}
}

// handleResult handles a work result, potentially retrying failed items
func (wp *WorkerPool) handleResult(result *WorkResult) {
	// Update statistics
	wp.stats.mu.Lock()
	wp.stats.LastProcessedTime = time.Now()
	wp.stats.TotalProcessingTime += result.ProcessingTime
	
	if result.Success {
		wp.stats.WorkItemsProcessed++
		if wp.stats.WorkItemsProcessed > 0 {
			wp.stats.AverageProcessingTime = wp.stats.TotalProcessingTime / time.Duration(wp.stats.WorkItemsProcessed)
		}
	} else {
		wp.stats.WorkItemsFailed++
	}
	wp.stats.mu.Unlock()
	
	// Handle retries for failed items
	if !result.Success && result.Item.Retries < wp.maxRetries {
		// Retry after backoff
		result.Item.Retries++
		
		go func() {
			time.Sleep(wp.retryBackoff * time.Duration(result.Item.Retries))
			
			select {
			case wp.workCh <- result.Item:
			case <-wp.ctx.Done():
			}
		}()
	}
	
	// Call work end hook
	if wp.onWorkEnd != nil {
		wp.onWorkEnd(result)
	}
}

// recoverPanic recovers from panics in worker goroutines
func (w *Worker) recoverPanic() {
	if r := recover(); r != nil {
		w.pool.stats.mu.Lock()
		w.pool.stats.WorkerPanics++
		w.pool.stats.mu.Unlock()
		
		if w.pool.onWorkerPanic != nil {
			w.pool.onWorkerPanic(w.ID, r)
		}
		
		// Restart the worker
		w.pool.wg.Add(1)
		go w.run()
	}
}

// IsShuttingDown returns true if the worker pool is shutting down
func (wp *WorkerPool) IsShuttingDown() bool {
	select {
	case <-wp.ctx.Done():
		return true
	default:
		return false
	}
}

// QueueSize returns the current size of the work queue
func (wp *WorkerPool) QueueSize() int {
	return len(wp.workCh)
}

// generateWorkItemID generates a unique ID for work items
func generateWorkItemID() string {
	return fmt.Sprintf("work_%d", time.Now().UnixNano())
}

// AdvancedIngestionQueue combines the basic queue with worker pool for high-throughput scenarios
type AdvancedIngestionQueue struct {
	basicQueue *IngestionQueue
	workerPool *WorkerPool
	config     *AdvancedQueueConfig
	mu         sync.RWMutex
	closed     bool
}

// AdvancedQueueConfig holds configuration for the advanced ingestion queue
type AdvancedQueueConfig struct {
	BasicQueueConfig  *QueueConfig
	WorkerPoolConfig  *WorkerPoolConfig
	UseWorkerPool     bool
	WorkerPoolThreshold int // Switch to worker pool when queue size exceeds this
}

// NewAdvancedIngestionQueue creates a new advanced ingestion queue
func NewAdvancedIngestionQueue(client IngestionClient, config *AdvancedQueueConfig) *AdvancedIngestionQueue {
	if config == nil {
		config = &AdvancedQueueConfig{
			BasicQueueConfig:    DefaultQueueConfig(),
			WorkerPoolConfig:    DefaultWorkerPoolConfig(),
			UseWorkerPool:       true,
			WorkerPoolThreshold: 50,
		}
	}
	
	basicQueue := NewIngestionQueue(client, config.BasicQueueConfig)
	var workerPool *WorkerPool
	if config.UseWorkerPool {
		workerPool = NewWorkerPool(client, config.WorkerPoolConfig)
	}
	
	return &AdvancedIngestionQueue{
		basicQueue: basicQueue,
		workerPool: workerPool,
		config:     config,
		closed:     false,
	}
}

// Enqueue adds an event to the appropriate queue based on current load
func (aq *AdvancedIngestionQueue) Enqueue(event types.IngestionEvent) error {
	aq.mu.RLock()
	defer aq.mu.RUnlock()
	
	if aq.closed {
		return fmt.Errorf("advanced queue is closed")
	}
	
	// Use worker pool for high load scenarios
	if aq.config.UseWorkerPool && aq.basicQueue.Size() > aq.config.WorkerPoolThreshold {
		return aq.workerPool.SubmitWork([]types.IngestionEvent{event})
	}
	
	return aq.basicQueue.Enqueue(event)
}

// Shutdown gracefully shuts down both queues
func (aq *AdvancedIngestionQueue) Shutdown(ctx context.Context) error {
	aq.mu.Lock()
	if aq.closed {
		aq.mu.Unlock()
		return nil
	}
	aq.closed = true
	aq.mu.Unlock()
	
	var errs []error
	
	if aq.workerPool != nil {
		if err := aq.workerPool.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("worker pool shutdown error: %w", err))
		}
	}
	
	if err := aq.basicQueue.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("basic queue shutdown error: %w", err))
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	
	return nil
}