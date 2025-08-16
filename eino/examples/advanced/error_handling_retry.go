package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/api/resources/commons/types"
)

// SimulatedLLMService demonstrates error handling and retry scenarios
type SimulatedLLMService struct {
	client      *client.Langfuse
	failureRate float64 // 0.0 to 1.0
}

func NewSimulatedLLMService(client *client.Langfuse, failureRate float64) *SimulatedLLMService {
	return &SimulatedLLMService{
		client:      client,
		failureRate: failureRate,
	}
}

func (s *SimulatedLLMService) ProcessRequest(ctx context.Context, request string, attempt int) error {
	trace := s.client.Trace(fmt.Sprintf("llm-request-attempt-%d", attempt)).
		WithInput(request).
		WithMetadata(map[string]interface{}{
			"attempt_number": attempt,
			"max_retries":    3,
			"failure_rate":   s.failureRate,
		})

	// Simulate different types of errors
	if rand.Float64() < s.failureRate {
		errorType := rand.Intn(4)
		var err error
		var level types.ObservationLevel = types.ObservationLevelError
		var statusMessage string

		switch errorType {
		case 0:
			err = errors.New("network timeout")
			statusMessage = "Network timeout occurred"
		case 1:
			err = errors.New("rate limit exceeded")
			statusMessage = "API rate limit exceeded"
		case 2:
			err = errors.New("service unavailable")
			statusMessage = "LLM service temporarily unavailable"
		case 3:
			err = errors.New("invalid api key")
			statusMessage = "Authentication failed"
			level = types.ObservationLevelError
		}

		// Create an error generation
		generation := trace.Generation("failed-llm-call").
			WithModel("gpt-4", map[string]interface{}{
				"temperature": 0.7,
				"attempt":     attempt,
			}).
			WithInput(request).
			WithStartTime(time.Now()).
			WithLevel(level).
			WithStatusMessage(statusMessage)

		// Simulate some processing time before failure
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		generation.WithEndTime(time.Now()).
			WithOutput(map[string]interface{}{
				"error":   err.Error(),
				"status":  "failed",
				"attempt": attempt,
			})

		if submitErr := generation.End(); submitErr != nil {
			log.Printf("Failed to submit error generation: %v", submitErr)
		}

		trace.WithOutput(map[string]interface{}{
			"status":        "failed",
			"error":         err.Error(),
			"attempt":       attempt,
			"will_retry":    attempt < 3,
		})

		if submitErr := trace.End(); submitErr != nil {
			log.Printf("Failed to submit error trace: %v", submitErr)
		}

		return fmt.Errorf("attempt %d failed: %w", attempt, err)
	}

	// Success case
	generation := trace.Generation("successful-llm-call").
		WithModel("gpt-4", map[string]interface{}{
			"temperature": 0.7,
			"attempt":     attempt,
		}).
		WithInput(request).
		WithStartTime(time.Now())

	// Simulate successful processing
	processingTime := time.Duration(100+rand.Intn(200)) * time.Millisecond
	time.Sleep(processingTime)

	response := fmt.Sprintf("Successful response to: %s (attempt %d)", request, attempt)

	generation.WithEndTime(time.Now()).
		WithOutput(response).
		WithUsage(&types.Usage{
			Input:     intPtr(len(request) / 4),
			Output:    intPtr(len(response) / 4),
			Total:     intPtr((len(request) + len(response)) / 4),
			Unit:      stringPtr("TOKENS"),
			InputCost: floatPtr(0.00003),
			OutputCost: floatPtr(0.00006),
			TotalCost: floatPtr(0.00009),
		}).
		WithLevel(types.ObservationLevelDefault)

	if err := generation.End(); err != nil {
		log.Printf("Failed to submit successful generation: %v", err)
	}

	trace.WithOutput(map[string]interface{}{
		"status":           "success",
		"response":         response,
		"attempt":          attempt,
		"processing_time":  processingTime.String(),
	})

	if err := trace.End(); err != nil {
		log.Printf("Failed to submit successful trace: %v", err)
	}

	return nil
}

// RetryableProcessor handles retry logic with exponential backoff
type RetryableProcessor struct {
	langfuse    *client.Langfuse
	llmService  *SimulatedLLMService
	maxRetries  int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

func NewRetryableProcessor(langfuse *client.Langfuse, failureRate float64) *RetryableProcessor {
	return &RetryableProcessor{
		langfuse:    langfuse,
		llmService:  NewSimulatedLLMService(langfuse, failureRate),
		maxRetries:  3,
		baseDelay:   100 * time.Millisecond,
		maxDelay:    5 * time.Second,
	}
}

func (rp *RetryableProcessor) ProcessWithRetry(ctx context.Context, request string) error {
	// Create main trace for the retry operation
	mainTrace := rp.langfuse.Trace("retryable-processing").
		WithInput(request).
		WithMetadata(map[string]interface{}{
			"max_retries":     rp.maxRetries,
			"base_delay":      rp.baseDelay.String(),
			"max_delay":       rp.maxDelay.String(),
			"strategy":        "exponential_backoff",
		})

	var lastErr error
	for attempt := 1; attempt <= rp.maxRetries+1; attempt++ {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			mainTrace.WithOutput(map[string]interface{}{
				"status":       "cancelled",
				"error":        err.Error(),
				"attempts":     attempt - 1,
				"final_result": "context_cancelled",
			})
			mainTrace.End()
			return err
		default:
		}

		// Try processing
		err := rp.llmService.ProcessRequest(ctx, request, attempt)
		if err == nil {
			// Success!
			mainTrace.WithOutput(map[string]interface{}{
				"status":       "success",
				"attempts":     attempt,
				"final_result": "success_after_retries",
			})
			
			if submitErr := mainTrace.End(); submitErr != nil {
				log.Printf("Failed to submit main trace: %v", submitErr)
			}

			// Add success score
			score := &types.Score{
				TraceID:  mainTrace.GetID(),
				Name:     "retry_success",
				Value:    true,
				DataType: types.ScoreDataTypeBoolean,
				Comment:  stringPtr(fmt.Sprintf("Succeeded after %d attempts", attempt)),
			}
			if err := rp.langfuse.Score(score); err != nil {
				log.Printf("Failed to submit score: %v", err)
			}

			return nil
		}

		lastErr = err
		fmt.Printf("Attempt %d failed: %v\n", attempt, err)

		// Don't wait after the last attempt
		if attempt <= rp.maxRetries {
			// Calculate exponential backoff delay
			delay := rp.calculateBackoff(attempt)
			fmt.Printf("Waiting %v before retry...\n", delay)
			
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				err := ctx.Err()
				mainTrace.WithOutput(map[string]interface{}{
					"status":       "cancelled",
					"error":        err.Error(),
					"attempts":     attempt,
					"final_result": "context_cancelled_during_backoff",
				})
				mainTrace.End()
				return err
			}
		}
	}

	// All attempts failed
	mainTrace.WithOutput(map[string]interface{}{
		"status":         "failed",
		"error":          lastErr.Error(),
		"attempts":       rp.maxRetries + 1,
		"final_result":   "exhausted_retries",
	})

	if err := mainTrace.End(); err != nil {
		log.Printf("Failed to submit main trace: %v", err)
	}

	// Add failure score
	score := &types.Score{
		TraceID:  mainTrace.GetID(),
		Name:     "retry_exhausted",
		Value:    true,
		DataType: types.ScoreDataTypeBoolean,
		Comment:  stringPtr(fmt.Sprintf("Failed after %d attempts", rp.maxRetries+1)),
	}
	if err := rp.langfuse.Score(score); err != nil {
		log.Printf("Failed to submit failure score: %v", err)
	}

	return fmt.Errorf("all retry attempts exhausted, last error: %w", lastErr)
}

func (rp *RetryableProcessor) calculateBackoff(attempt int) time.Duration {
	delay := rp.baseDelay * time.Duration(1<<uint(attempt-1)) // Exponential backoff
	if delay > rp.maxDelay {
		delay = rp.maxDelay
	}
	
	// Add jitter (±25%)
	jitter := time.Duration(rand.Int63n(int64(delay) / 2))
	delay = delay + jitter - jitter/2
	
	return delay
}

func main() {
	// Initialize Langfuse client
	langfuseClient, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
		client.WithRetrySettings(5, 2*time.Second), // SDK-level retries
	)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	fmt.Println("=== Error Handling and Retry Examples ===\n")

	// Test different failure rates
	failureRates := []float64{0.3, 0.7, 0.9} // 30%, 70%, 90% failure rates
	requests := []string{
		"Analyze the benefits of microservices",
		"Explain machine learning concepts",
		"Design a scalable web architecture",
	}

	for i, failureRate := range failureRates {
		fmt.Printf("Testing with %.0f%% failure rate:\n", failureRate*100)
		
		processor := NewRetryableProcessor(langfuseClient, failureRate)
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		request := requests[i]
		fmt.Printf("Processing request: %s\n", request)

		err := processor.ProcessWithRetry(ctx, request)
		if err != nil {
			fmt.Printf("❌ Final result: FAILED - %v\n", err)
		} else {
			fmt.Printf("✅ Final result: SUCCESS\n")
		}
		fmt.Println()
	}

	// Demonstrate circuit breaker pattern simulation
	fmt.Println("=== Circuit Breaker Pattern Simulation ===")
	circuitBreakerDemo(langfuseClient)

	// Wait for all events to be processed
	fmt.Println("Flushing events...")
	if err := langfuseClient.Flush(context.Background()); err != nil {
		log.Printf("Failed to flush: %v", err)
	}
	
	time.Sleep(2 * time.Second)

	// Show final statistics
	stats := langfuseClient.GetStats()
	fmt.Printf("\n=== Final Statistics ===\n")
	fmt.Printf("Traces Created: %d\n", stats.TracesCreated)
	fmt.Printf("Generations Created: %d\n", stats.GenerationsCreated)
	fmt.Printf("Events Submitted: %d\n", stats.EventsSubmitted)
	fmt.Printf("Events Failed: %d\n", stats.EventsFailed)
	fmt.Printf("Success Rate: %.2f%%\n", 
		float64(stats.EventsSubmitted)/float64(stats.EventsSubmitted+stats.EventsFailed)*100)
}

func circuitBreakerDemo(langfuseClient *client.Langfuse) {
	// Simulate circuit breaker states
	states := []string{"CLOSED", "OPEN", "HALF_OPEN"}
	
	for _, state := range states {
		trace := langfuseClient.Trace("circuit-breaker-demo").
			WithInput(map[string]interface{}{
				"circuit_state": state,
				"timestamp":     time.Now(),
			}).
			WithMetadata(map[string]interface{}{
				"pattern":       "circuit_breaker",
				"state":         state,
			})

		var result string
		var level types.ObservationLevel = types.ObservationLevelDefault

		switch state {
		case "CLOSED":
			result = "Request processed normally"
			level = types.ObservationLevelDefault
		case "OPEN":
			result = "Request rejected - circuit breaker open"
			level = types.ObservationLevelWarning
		case "HALF_OPEN":
			result = "Request allowed for testing - circuit breaker half-open"
			level = types.ObservationLevelDefault
		}

		span := trace.Span("circuit-breaker-check").
			WithInput(state).
			WithStartTime(time.Now()).
			WithLevel(level).
			WithMetadata(map[string]interface{}{
				"pattern": "circuit_breaker",
			})

		time.Sleep(10 * time.Millisecond) // Simulate processing

		span.WithOutput(result).WithEndTime(time.Now())
		
		if err := span.End(); err != nil {
			log.Printf("Failed to submit circuit breaker span: %v", err)
		}

		trace.WithOutput(map[string]interface{}{
			"circuit_state": state,
			"result":        result,
		})

		if err := trace.End(); err != nil {
			log.Printf("Failed to submit circuit breaker trace: %v", err)
		}

		fmt.Printf("Circuit Breaker %s: %s\n", state, result)
	}
}

// Helper functions
func intPtr(v int) *int { return &v }
func stringPtr(v string) *string { return &v }
func floatPtr(v float64) *float64 { return &v }