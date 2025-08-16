package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/api/resources/commons/types"
)

// WorkerPool demonstrates concurrent processing with Langfuse tracing
type WorkerPool struct {
	langfuse     *client.Langfuse
	workerCount  int
	jobQueue     chan Job
	results      chan Result
	wg           sync.WaitGroup
}

type Job struct {
	ID      int
	Query   string
	TraceID string
}

type Result struct {
	JobID    int
	Response string
	Error    error
}

func NewWorkerPool(langfuse *client.Langfuse, workerCount int) *WorkerPool {
	return &WorkerPool{
		langfuse:    langfuse,
		workerCount: workerCount,
		jobQueue:    make(chan Job, workerCount*2),
		results:     make(chan Result, workerCount*2),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	// Start workers
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.results)
}

func (wp *WorkerPool) AddJob(job Job) {
	wp.jobQueue <- job
}

func (wp *WorkerPool) GetResults() <-chan Result {
	return wp.results
}

func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	
	for job := range wp.jobQueue {
		result := wp.processJob(ctx, job, workerID)
		select {
		case wp.results <- result:
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) processJob(ctx context.Context, job Job, workerID int) Result {
	// Create a span for this worker's processing
	trace := wp.langfuse.Trace(fmt.Sprintf("concurrent-job-%s", job.TraceID))
	
	span := trace.Span(fmt.Sprintf("worker-process")).
		WithInput(map[string]interface{}{
			"job_id":    job.ID,
			"worker_id": workerID,
			"query":     job.Query,
		}).
		WithStartTime(time.Now()).
		WithMetadata(map[string]interface{}{
			"worker_id":       workerID,
			"processing_type": "concurrent",
		})

	// Simulate processing work
	processingTime := time.Duration(50+job.ID*10) * time.Millisecond
	time.Sleep(processingTime)

	response := fmt.Sprintf("Worker %d processed job %d: %s", workerID, job.ID, job.Query)

	span.WithOutput(map[string]interface{}{
		"response":        response,
		"processing_time": processingTime.String(),
		"status":         "completed",
	}).
		WithEndTime(time.Now()).
		WithLevel("DEFAULT")

	if err := span.End(); err != nil {
		log.Printf("Worker %d failed to submit span for job %d: %v", workerID, job.ID, err)
	}

	// Create a generation to simulate LLM call
	generation := trace.Generation(fmt.Sprintf("llm-call-job-%d", job.ID)).
		WithModel("gpt-4o-mini", map[string]interface{}{
			"temperature": 0.5,
			"worker_id":   workerID,
		}).
		WithInput(job.Query).
		WithStartTime(time.Now())

	// Simulate LLM processing
	llmTime := time.Duration(100+job.ID*5) * time.Millisecond
	time.Sleep(llmTime)

	llmResponse := fmt.Sprintf("LLM response for: %s", job.Query)
	
	generation.WithOutput(llmResponse).
		WithEndTime(time.Now()).
		WithUsage(&types.Usage{
			Input:     intPtr(len(job.Query) / 4),
			Output:    intPtr(len(llmResponse) / 4),
			Total:     intPtr((len(job.Query) + len(llmResponse)) / 4),
			Unit:      stringPtr("TOKENS"),
			InputCost: floatPtr(0.00001 * float64(job.ID)),
			OutputCost: floatPtr(0.00002 * float64(job.ID)),
			TotalCost: floatPtr(0.00003 * float64(job.ID)),
		})

	if err := generation.End(); err != nil {
		log.Printf("Worker %d failed to submit generation for job %d: %v", workerID, job.ID, err)
	}

	if err := trace.End(); err != nil {
		log.Printf("Worker %d failed to submit trace for job %d: %v", workerID, job.ID, err)
	}

	return Result{
		JobID:    job.ID,
		Response: response + " | " + llmResponse,
		Error:    nil,
	}
}

// Batch processor for handling multiple concurrent operations
type BatchProcessor struct {
	langfuse *client.Langfuse
}

func NewBatchProcessor(langfuse *client.Langfuse) *BatchProcessor {
	return &BatchProcessor{langfuse: langfuse}
}

func (bp *BatchProcessor) ProcessBatch(ctx context.Context, queries []string, batchSize int) error {
	// Create main trace for the entire batch
	mainTrace := bp.langfuse.Trace("concurrent-batch-processing").
		WithInput(map[string]interface{}{
			"total_queries":  len(queries),
			"batch_size":     batchSize,
			"worker_count":   4,
		}).
		WithMetadata(map[string]interface{}{
			"processing_type": "concurrent_batch",
			"start_time":      time.Now(),
		})

	traceID := mainTrace.GetID()

	// Create worker pool
	workerPool := NewWorkerPool(bp.langfuse, 4)
	workerPool.Start(ctx)

	// Submit jobs
	go func() {
		for i, query := range queries {
			job := Job{
				ID:      i,
				Query:   query,
				TraceID: traceID,
			}
			workerPool.AddJob(job)
		}
		workerPool.Stop()
	}()

	// Collect results
	var results []Result
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range workerPool.GetResults() {
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			
			if result.Error != nil {
				log.Printf("Job %d failed: %v", result.JobID, result.Error)
			} else {
				fmt.Printf("✓ Job %d completed: %s\n", result.JobID, result.Response)
			}
		}
	}()

	wg.Wait()

	// Complete main trace
	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	mainTrace.WithOutput(map[string]interface{}{
		"total_processed": len(results),
		"successful":      successCount,
		"failed":         errorCount,
		"completion_rate": float64(successCount) / float64(len(results)),
	})

	if err := mainTrace.End(); err != nil {
		return fmt.Errorf("failed to submit main trace: %w", err)
	}

	// Add batch-level scores
	completionRate := float64(successCount) / float64(len(results))
	scores := []*types.Score{
		{
			TraceID:  traceID,
			Name:     "batch_completion_rate",
			Value:    completionRate,
			DataType: types.ScoreDataTypeNumeric,
			Comment:  stringPtr(fmt.Sprintf("Completed %d/%d jobs successfully", successCount, len(results))),
		},
		{
			TraceID:  traceID,
			Name:     "processing_efficiency",
			Value:    "high",
			DataType: types.ScoreDataTypeCategorical,
			Comment:  stringPtr("Concurrent processing completed efficiently"),
		},
	}

	for _, score := range scores {
		if err := bp.langfuse.Score(score); err != nil {
			log.Printf("Failed to submit score %s: %v", score.Name, err)
		}
	}

	fmt.Printf("Batch processing completed: %d successful, %d failed\n", successCount, errorCount)
	return nil
}

func main() {
	// Initialize Langfuse client
	langfuseClient, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
		client.WithFlushSettings(10, 5*time.Second), // More frequent flushing for demo
		client.WithQueueSize(1000),
	)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	// Test queries for concurrent processing
	queries := []string{
		"What are the benefits of microservices architecture?",
		"How do I optimize database queries for better performance?",
		"What are the best practices for API design?",
		"How can I implement effective logging in a distributed system?",
		"What are the key principles of DevOps?",
		"How do I design a scalable authentication system?",
		"What are the trade-offs between SQL and NoSQL databases?",
		"How can I implement circuit breaker patterns?",
		"What are the benefits of containerization with Docker?",
		"How do I monitor microservices effectively?",
	}

	fmt.Printf("Starting concurrent processing of %d queries...\n", len(queries))

	// Create batch processor
	processor := NewBatchProcessor(langfuseClient)

	// Process batch concurrently
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := processor.ProcessBatch(ctx, queries, 4); err != nil {
		log.Fatal("Batch processing failed:", err)
	}

	// Force flush and wait
	if err := langfuseClient.Flush(context.Background()); err != nil {
		log.Printf("Failed to flush: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Display final statistics
	stats := langfuseClient.GetStats()
	fmt.Printf("\n=== Final Statistics ===\n")
	fmt.Printf("Traces Created: %d\n", stats.TracesCreated)
	fmt.Printf("Spans Created: %d\n", stats.SpansCreated)
	fmt.Printf("Generations Created: %d\n", stats.GenerationsCreated)
	fmt.Printf("Events Submitted: %d\n", stats.EventsSubmitted)
	fmt.Printf("Events Failed: %d\n", stats.EventsFailed)
	fmt.Printf("Last Activity: %s\n", stats.LastActivity.Format(time.RFC3339))
}

// Helper functions
func intPtr(v int) *int { return &v }
func stringPtr(v string) *string { return &v }
func floatPtr(v float64) *float64 { return &v }