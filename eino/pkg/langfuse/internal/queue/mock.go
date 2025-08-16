package queue

import (
	"context"
	"sync"

	"eino/pkg/langfuse/api/resources/ingestion/types"
)

// MockQueue implements a mock ingestion queue for testing
type MockQueue struct {
	events   []types.IngestionEvent
	mu       sync.RWMutex
	enqueued int
	flushed  int
	closed   bool
}

// NewMockQueue creates a new mock queue for testing
func NewMockQueue() *MockQueue {
	return &MockQueue{
		events: make([]types.IngestionEvent, 0),
	}
}

// Enqueue adds an event to the mock queue
func (mq *MockQueue) Enqueue(event types.IngestionEvent) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	if mq.closed {
		return ErrQueueClosed
	}
	
	mq.events = append(mq.events, event)
	mq.enqueued++
	return nil
}

// Flush simulates flushing all events
func (mq *MockQueue) Flush(ctx context.Context) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	if mq.closed {
		return ErrQueueClosed
	}
	
	mq.flushed += len(mq.events)
	mq.events = make([]types.IngestionEvent, 0)
	return nil
}

// Shutdown simulates shutting down the queue
func (mq *MockQueue) Shutdown(ctx context.Context) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	if mq.closed {
		return nil
	}
	
	// Flush remaining events
	mq.flushed += len(mq.events)
	mq.events = make([]types.IngestionEvent, 0)
	mq.closed = true
	return nil
}

// GetEvents returns all events currently in the queue
func (mq *MockQueue) GetEvents() []types.IngestionEvent {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	
	events := make([]types.IngestionEvent, len(mq.events))
	copy(events, mq.events)
	return events
}

// GetEnqueuedCount returns the total number of events enqueued
func (mq *MockQueue) GetEnqueuedCount() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.enqueued
}

// GetFlushedCount returns the total number of events flushed
func (mq *MockQueue) GetFlushedCount() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.flushed
}

// IsClosed returns whether the queue is closed
func (mq *MockQueue) IsClosed() bool {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.closed
}

// Reset clears all events and counters
func (mq *MockQueue) Reset() {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	mq.events = make([]types.IngestionEvent, 0)
	mq.enqueued = 0
	mq.flushed = 0
	mq.closed = false
}