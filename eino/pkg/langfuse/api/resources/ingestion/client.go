package ingestion

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/ingestion/types"
	"eino/pkg/langfuse/internal/utils"
)

// API path constants
const (
	ingestionBasePath = "/api/public/ingestion"
	healthBasePath    = "/api/public/health"
)

// Client handles ingestion API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new ingestion client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// Submit submits an ingestion request to the Langfuse API
func (c *Client) Submit(ctx context.Context, req *types.IngestionRequest) (*types.IngestionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("ingestion request cannot be nil")
	}
	
	// Validate the request before submission
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.IngestionResponse{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(ingestionBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to submit ingestion request: %w", err)
	}
	
	return response, nil
}

// SubmitBatch submits a batch of ingestion events
func (c *Client) SubmitBatch(ctx context.Context, events []types.IngestionEvent) (*types.IngestionResponse, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("cannot submit empty batch")
	}
	
	if len(events) > types.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum allowed size %d", 
			len(events), types.MaxBatchSize)
	}
	
	// Create request with metadata
	req := types.NewIngestionRequest(events)
	
	return c.Submit(ctx, req)
}

// SubmitBatchWithMetadata submits a batch with custom metadata
func (c *Client) SubmitBatchWithMetadata(ctx context.Context, events []types.IngestionEvent, metadata *types.IngestionBatchMetadata) (*types.IngestionResponse, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("cannot submit empty batch")
	}
	
	if len(events) > types.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum allowed size %d", 
			len(events), types.MaxBatchSize)
	}
	
	// Create request with custom metadata
	req := types.NewIngestionRequestWithMetadata(events, metadata)
	
	return c.Submit(ctx, req)
}

// SubmitTrace submits a trace creation event
func (c *Client) SubmitTrace(ctx context.Context, event *types.TraceCreateEvent) (*types.IngestionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("trace event cannot be nil")
	}
	
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("trace event validation failed: %w", err)
	}
	
	ingestionEvent := event.ToIngestionEvent()
	return c.SubmitBatch(ctx, []types.IngestionEvent{ingestionEvent})
}

// SubmitTraceUpdate submits a trace update event
func (c *Client) SubmitTraceUpdate(ctx context.Context, event *types.TraceUpdateEvent) (*types.IngestionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("trace update event cannot be nil")
	}
	
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("trace update event validation failed: %w", err)
	}
	
	ingestionEvent := event.ToIngestionEvent()
	return c.SubmitBatch(ctx, []types.IngestionEvent{ingestionEvent})
}

// SubmitObservation submits an observation event (span, generation, or event)
func (c *Client) SubmitObservation(ctx context.Context, event interface{}) (*types.IngestionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("observation event cannot be nil")
	}
	
	var ingestionEvent types.IngestionEvent
	
	switch e := event.(type) {
	case *types.ObservationCreateEvent:
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("observation create event validation failed: %w", err)
		}
		ingestionEvent = e.ToIngestionEvent()
		
	case *types.ObservationUpdateEvent:
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("observation update event validation failed: %w", err)
		}
		ingestionEvent = e.ToIngestionEvent()
		
	case *types.SpanCreateEvent:
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("span create event validation failed: %w", err)
		}
		ingestionEvent = e.ToIngestionEvent()
		
	case *types.SpanUpdateEvent:
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("span update event validation failed: %w", err)
		}
		ingestionEvent = e.ToIngestionEvent()
		
	case *types.GenerationCreateEvent:
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("generation create event validation failed: %w", err)
		}
		ingestionEvent = e.ToIngestionEvent()
		
	case *types.GenerationUpdateEvent:
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("generation update event validation failed: %w", err)
		}
		ingestionEvent = e.ToIngestionEvent()
		
	case *types.EventCreateEvent:
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("event create validation failed: %w", err)
		}
		ingestionEvent = e.ToIngestionEvent()
		
	default:
		return nil, fmt.Errorf("unsupported observation event type: %T", event)
	}
	
	return c.SubmitBatch(ctx, []types.IngestionEvent{ingestionEvent})
}

// SubmitScore submits a score creation event
func (c *Client) SubmitScore(ctx context.Context, event *types.ScoreCreateEvent) (*types.IngestionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("score event cannot be nil")
	}
	
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("score event validation failed: %w", err)
	}
	
	ingestionEvent := event.ToIngestionEvent()
	return c.SubmitBatch(ctx, []types.IngestionEvent{ingestionEvent})
}

// SubmitMultipleEvents submits multiple events of different types in a single batch
func (c *Client) SubmitMultipleEvents(ctx context.Context, events []interface{}) (*types.IngestionResponse, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("cannot submit empty events list")
	}
	
	ingestionEvents := make([]types.IngestionEvent, 0, len(events))
	
	for i, event := range events {
		if event == nil {
			return nil, fmt.Errorf("event at index %d cannot be nil", i)
		}
		
		var ingestionEvent types.IngestionEvent
		
		switch e := event.(type) {
		case *types.TraceCreateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("trace create event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.TraceUpdateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("trace update event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.ObservationCreateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("observation create event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.ObservationUpdateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("observation update event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.SpanCreateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("span create event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.SpanUpdateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("span update event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.GenerationCreateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("generation create event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.GenerationUpdateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("generation update event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.EventCreateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("event create at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case *types.ScoreCreateEvent:
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("score create event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e.ToIngestionEvent()
			
		case types.IngestionEvent:
			// Direct ingestion event
			if err := e.Validate(); err != nil {
				return nil, fmt.Errorf("ingestion event at index %d validation failed: %w", i, err)
			}
			ingestionEvent = e
			
		default:
			return nil, fmt.Errorf("unsupported event type at index %d: %T", i, event)
		}
		
		ingestionEvents = append(ingestionEvents, ingestionEvent)
	}
	
	return c.SubmitBatch(ctx, ingestionEvents)
}

// SubmitWithRetry submits an ingestion request with automatic retries
func (c *Client) SubmitWithRetry(ctx context.Context, req *types.IngestionRequest, maxRetries int, backoff time.Duration) (*types.IngestionResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff * time.Duration(attempt)):
				// Continue with retry
			}
		}
		
		response, err := c.Submit(ctx, req)
		if err == nil && response != nil && response.Success {
			return response, nil
		}
		
		lastErr = err
		
		// Don't retry on validation errors or client errors
		if err != nil {
			if apiErr, ok := err.(*utils.SDKError); ok {
				if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
					// Client error, don't retry (except rate limits)
					break
				}
			}
		}
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Health checks if the ingestion endpoint is available
func (c *Client) Health(ctx context.Context) error {
	_, err := c.client.R().
		SetContext(ctx).
		Get(healthBasePath)
	
	if err != nil {
		return fmt.Errorf("ingestion health check failed: %w", err)
	}
	
	return nil
}