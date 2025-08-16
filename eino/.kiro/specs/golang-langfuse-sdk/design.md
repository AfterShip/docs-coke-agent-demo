# Golang Langfuse SDK Design Document

## Overview

This document outlines the design for a comprehensive Golang SDK for Langfuse, an observability platform for AI applications. The SDK enables developers to trace, debug, and monitor their Go-based AI applications, providing detailed insights into LLM calls, agent behaviors, and application performance.

The design follows modern SDK design patterns with focus on developer experience, performance, and maintainability.

## Core Concepts and Terminology

**Core Concepts:**
- **Trace**: A complete execution flow representing a single user request or application run
- **Observation**: Any monitored unit of work within a trace (spans, generations, events)
- **Span**: A general-purpose observation representing any operation or process
- **Generation**: A specialized observation for LLM inference calls
- **Event**: A point-in-time observation representing discrete occurrences
- **Score**: A quantitative or qualitative evaluation metric applied to observations
- **Session**: A collection of traces grouped by user session or conversation
- **Prompt**: A managed template for LLM interactions
- **Dataset**: A collection of test cases for evaluation and benchmarking
- **Usage Metrics**: Token counts, costs, and performance measurements
- **Media Content**: Binary attachments (images, audio) associated with observations

## SDK Architecture

The SDK follows a layered architecture optimized for developer experience and performance, with an `api` directory structure that mirrors the Python Langfuse implementation:

```
pkg/langfuse/
├── client/              # Public SDK interface
│   ├── langfuse.go      # Main client
│   ├── trace.go         # Trace builder
│   ├── span.go          # Span builder  
│   ├── generation.go    # Generation builder
│   └── config.go        # Configuration
├── api/                 # API layer - mirrors Python structure
│   ├── client.go        # HTTP client wrapper
│   ├── resources/       # Resource-specific API clients
│   │   ├── annotations/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── annotation_queue.go
│   │   │       ├── create_annotation_request.go
│   │   │       └── paginated_annotations.go
│   │   ├── comments/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── create_comment_request.go
│   │   │       └── comment_response.go
│   │   ├── commons/     # Shared types and errors
│   │   │   ├── errors/
│   │   │   │   ├── access_denied.go
│   │   │   │   ├── not_found.go
│   │   │   │   └── unauthorized.go
│   │   │   └── types/
│   │   │       ├── trace.go
│   │   │       ├── observation.go
│   │   │       ├── score.go
│   │   │       ├── usage.go
│   │   │       ├── session.go
│   │   │       └── dataset.go
│   │   ├── datasets/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── create_dataset_request.go
│   │   │       ├── dataset_item.go
│   │   │       └── paginated_datasets.go
│   │   ├── health/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       └── health_response.go
│   │   ├── ingestion/   # Core ingestion API
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── ingestion_event.go
│   │   │       ├── ingestion_request.go
│   │   │       ├── ingestion_response.go
│   │   │       ├── trace_event.go
│   │   │       ├── observation_event.go
│   │   │       └── score_event.go
│   │   ├── media/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── media_upload_request.go
│   │   │       └── media_response.go
│   │   ├── metrics/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       └── metrics_response.go
│   │   ├── models/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── create_model_request.go
│   │   │       └── paginated_models.go
│   │   ├── observations/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── observations.go
│   │   │       └── observations_view.go
│   │   ├── organizations/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── membership.go
│   │   │       └── organization_project.go
│   │   ├── projects/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── api_key.go
│   │   │       └── project.go
│   │   ├── prompts/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── prompt.go
│   │   │       ├── chat_prompt.go
│   │   │       ├── text_prompt.go
│   │   │       └── create_prompt_request.go
│   │   ├── scores/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── create_score_request.go
│   │   │       ├── score_config.go
│   │   │       └── get_scores_response.go
│   │   ├── sessions/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       └── paginated_sessions.go
│   │   ├── traces/
│   │   │   ├── client.go
│   │   │   └── types/
│   │   │       ├── traces.go
│   │   │       └── delete_trace_response.go
│   │   └── utils/
│   │       └── pagination/
│   │           └── types/
│   │               └── meta_response.go
│   └── core/            # Core HTTP client functionality
│       ├── client_wrapper.go
│       ├── http_client.go
│       ├── auth.go
│       ├── request_options.go
│       ├── query_encoder.go
│       └── datetime_utils.go
├── internal/
│   ├── queue/           # Async processing
│   │   ├── ingestion_queue.go
│   │   └── worker_pool.go
│   ├── telemetry/       # OpenTelemetry integration
│   │   ├── processor.go
│   │   └── exporter.go
│   └── utils/           # Internal utilities
│       ├── ids.go
│       ├── validation.go
│       └── errors.go
└── middleware/          # Framework integrations
    ├── http.go          # HTTP middleware
    └── grpc.go          # gRPC interceptors
```

## Core Data Models

### Trace Model
```go
type Trace struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    UserID       *string               `json:"userId,omitempty"`
    SessionID    *string               `json:"sessionId,omitempty"`
    Input        interface{}           `json:"input,omitempty"`
    Output       interface{}           `json:"output,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
    Tags         []string              `json:"tags,omitempty"`
    Environment  string                `json:"environment,omitempty"`
    Release      *string               `json:"release,omitempty"`
    Version      *string               `json:"version,omitempty"`
    Public       *bool                 `json:"public,omitempty"`
    Timestamp    time.Time             `json:"timestamp"`
    Observations []Observation         `json:"observations"`
}
```

### Observation Model
```go
type Observation struct {
    ID                   string                 `json:"id"`
    TraceID              string                 `json:"traceId"`
    ParentObservationID  *string               `json:"parentObservationId,omitempty"`
    Type                 ObservationType       `json:"type"`
    Name                 string                `json:"name"`
    StartTime            time.Time             `json:"startTime"`
    EndTime              *time.Time            `json:"endTime,omitempty"`
    CompletionStartTime  *time.Time            `json:"completionStartTime,omitempty"`
    Model                *string               `json:"model,omitempty"`
    ModelParameters      map[string]interface{} `json:"modelParameters,omitempty"`
    Input                interface{}           `json:"input,omitempty"`
    Output               interface{}           `json:"output,omitempty"`
    Usage                *Usage                `json:"usage,omitempty"`
    Level                ObservationLevel      `json:"level,omitempty"`
    StatusMessage        *string               `json:"statusMessage,omitempty"`
    Version              *string               `json:"version,omitempty"`
    Metadata             map[string]interface{} `json:"metadata,omitempty"`
    Environment          string                `json:"environment,omitempty"`
}

type ObservationType string

const (
    ObservationTypeSpan       ObservationType = "SPAN"
    ObservationTypeGeneration ObservationType = "GENERATION"
    ObservationTypeEvent      ObservationType = "EVENT"
)

type ObservationLevel string

const (
    ObservationLevelDebug   ObservationLevel = "DEBUG"
    ObservationLevelDefault ObservationLevel = "DEFAULT"
    ObservationLevelWarning ObservationLevel = "WARNING"
    ObservationLevelError   ObservationLevel = "ERROR"
)
```

### Usage Model
```go
type Usage struct {
    Input      *int     `json:"input,omitempty"`
    Output     *int     `json:"output,omitempty"`
    Total      *int     `json:"total,omitempty"`
    Unit       *string  `json:"unit,omitempty"`
    InputCost  *float64 `json:"inputCost,omitempty"`
    OutputCost *float64 `json:"outputCost,omitempty"`
    TotalCost  *float64 `json:"totalCost,omitempty"`
}
```

### Score Model
```go
type Score struct {
    ID             string           `json:"id"`
    TraceID        string           `json:"traceId"`
    ObservationID  *string         `json:"observationId,omitempty"`
    Name           string          `json:"name"`
    Value          interface{}     `json:"value"`
    DataType       ScoreDataType   `json:"dataType"`
    Comment        *string         `json:"comment,omitempty"`
    ConfigID       *string         `json:"configId,omitempty"`
    Timestamp      time.Time       `json:"timestamp"`
}

type ScoreDataType string

const (
    ScoreDataTypeNumeric     ScoreDataType = "NUMERIC"
    ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL" 
    ScoreDataTypeBoolean     ScoreDataType = "BOOLEAN"
)
```

## Public API Design

### Main Client Integration

The main client now integrates with the API layer while maintaining the builder pattern interface:

```go
// client/langfuse.go
type Langfuse struct {
    config    *Config
    apiClient *api.APIClient
    queue     *queue.IngestionQueue
    mu        sync.RWMutex
    closed    bool
}

func New(config *Config) (*Langfuse, error) {
    apiClient, err := api.NewAPIClient(config)
    if err != nil {
        return nil, err
    }
    
    queue := queue.NewIngestionQueue(apiClient.Ingestion, config.FlushAt, config.FlushInterval)
    
    return &Langfuse{
        config:    config,
        apiClient: apiClient,
        queue:     queue,
        closed:    false,
    }, nil
}

func (lf *Langfuse) Trace(name string) *TraceBuilder {
    return &TraceBuilder{
        trace: &commons.Trace{
            ID:          generateID(),
            Name:        name,
            Timestamp:   time.Now(),
            Environment: lf.config.Environment,
        },
        lf: lf,
    }
}

func (lf *Langfuse) Generation(name string) *GenerationBuilder {
    trace := lf.Trace("generation-trace-" + generateID())
    return trace.Generation(name)
}

// Direct API access for advanced usage
func (lf *Langfuse) API() *api.APIClient {
    return lf.apiClient
}

func (lf *Langfuse) Score(score *commons.Score) error {
    return lf.queue.Enqueue(createScoreEvent(score))
}

func (lf *Langfuse) Shutdown(ctx context.Context) error {
    lf.mu.Lock()
    defer lf.mu.Unlock()
    
    if lf.closed {
        return nil
    }
    
    // Flush pending data
    if err := lf.queue.Shutdown(ctx); err != nil {
        return err
    }
    
    lf.closed = true
    return nil
}
```

### Builder Pattern API
```go
type TraceBuilder struct {
    trace *Trace
    lf    *Langfuse
}

func (tb *TraceBuilder) WithUser(userID string) *TraceBuilder {
    tb.trace.UserID = &userID
    return tb
}

func (tb *TraceBuilder) WithSession(sessionID string) *TraceBuilder {
    tb.trace.SessionID = &sessionID
    return tb
}

func (tb *TraceBuilder) WithMetadata(metadata map[string]interface{}) *TraceBuilder {
    tb.trace.Metadata = metadata
    return tb
}

func (tb *TraceBuilder) WithTags(tags ...string) *TraceBuilder {
    tb.trace.Tags = tags
    return tb
}

func (tb *TraceBuilder) WithInput(input interface{}) *TraceBuilder {
    tb.trace.Input = input
    return tb
}

func (tb *TraceBuilder) WithOutput(output interface{}) *TraceBuilder {
    tb.trace.Output = output
    return tb
}

func (tb *TraceBuilder) Span(name string) *SpanBuilder {
    // Create span builder within this trace
}

func (tb *TraceBuilder) Generation(name string) *GenerationBuilder {
    // Create generation builder within this trace
}

func (tb *TraceBuilder) Event(name string) *EventBuilder {
    // Create event builder within this trace
}

func (tb *TraceBuilder) End() error {
    // Complete and submit the trace
}
```

### Generation Builder
```go
type GenerationBuilder struct {
    observation *Observation
    tb          *TraceBuilder
}

func (gb *GenerationBuilder) WithModel(name string, parameters map[string]interface{}) *GenerationBuilder {
    gb.observation.Model = &name
    gb.observation.ModelParameters = parameters
    return gb
}

func (gb *GenerationBuilder) WithInput(input interface{}) *GenerationBuilder {
    gb.observation.Input = input
    return gb
}

func (gb *GenerationBuilder) WithOutput(output interface{}) *GenerationBuilder {
    gb.observation.Output = input
    return gb
}

func (gb *GenerationBuilder) WithUsage(usage *Usage) *GenerationBuilder {
    gb.observation.Usage = usage
    return gb
}

func (gb *GenerationBuilder) WithCompletionStartTime(t time.Time) *GenerationBuilder {
    gb.observation.CompletionStartTime = &t
    return gb
}

func (gb *GenerationBuilder) End() (*TraceBuilder, error) {
    // Complete generation and return to trace builder
}
```

### Configuration
```go
type Config struct {
    // Connection settings
    Host      string `env:"LANGFUSE_HOST" default:"https://cloud.langfuse.com"`
    PublicKey string `env:"LANGFUSE_PUBLIC_KEY"`
    SecretKey string `env:"LANGFUSE_SECRET_KEY"`
    
    // Behavior settings
    Debug             bool          `env:"LANGFUSE_DEBUG" default:"false"`
    Enabled           bool          `env:"LANGFUSE_ENABLED" default:"true"`
    SampleRate        float64       `env:"LANGFUSE_SAMPLE_RATE" default:"1.0"`
    FlushAt           int           `env:"LANGFUSE_FLUSH_AT" default:"15"`
    FlushInterval     time.Duration `env:"LANGFUSE_FLUSH_INTERVAL" default:"10s"`
    RequestTimeout    time.Duration `env:"LANGFUSE_TIMEOUT" default:"10s"`
    
    // HTTP Client settings (Resty-specific)
    RetryCount        int           `env:"LANGFUSE_RETRY_COUNT" default:"3"`
    RetryWaitTime     time.Duration `env:"LANGFUSE_RETRY_WAIT" default:"1s"`
    RetryMaxWaitTime  time.Duration `env:"LANGFUSE_RETRY_MAX_WAIT" default:"10s"`
    UserAgent         string        `env:"LANGFUSE_USER_AGENT" default:"langfuse-go/1.0.0"`
    
    // Environment settings
    Environment string `env:"LANGFUSE_ENVIRONMENT"`
    Release     string `env:"LANGFUSE_RELEASE"`
    Version     string `env:"LANGFUSE_VERSION" default:"1.0.0"`
    
    // Advanced settings
    EnableOpenTelemetry bool `env:"LANGFUSE_ENABLE_OTEL" default:"false"`
}

func LoadConfig() (*Config, error) {
    config := &Config{}
    
    // Load from environment variables with defaults
    config.Host = getEnvOrDefault("LANGFUSE_HOST", "https://cloud.langfuse.com")
    config.PublicKey = os.Getenv("LANGFUSE_PUBLIC_KEY")
    config.SecretKey = os.Getenv("LANGFUSE_SECRET_KEY")
    config.Debug = getEnvBoolOrDefault("LANGFUSE_DEBUG", false)
    config.Enabled = getEnvBoolOrDefault("LANGFUSE_ENABLED", true)
    config.SampleRate = getEnvFloatOrDefault("LANGFUSE_SAMPLE_RATE", 1.0)
    config.FlushAt = getEnvIntOrDefault("LANGFUSE_FLUSH_AT", 15)
    config.FlushInterval = getEnvDurationOrDefault("LANGFUSE_FLUSH_INTERVAL", 10*time.Second)
    config.RequestTimeout = getEnvDurationOrDefault("LANGFUSE_TIMEOUT", 10*time.Second)
    config.RetryCount = getEnvIntOrDefault("LANGFUSE_RETRY_COUNT", 3)
    config.RetryWaitTime = getEnvDurationOrDefault("LANGFUSE_RETRY_WAIT", 1*time.Second)
    config.RetryMaxWaitTime = getEnvDurationOrDefault("LANGFUSE_RETRY_MAX_WAIT", 10*time.Second)
    config.UserAgent = getEnvOrDefault("LANGFUSE_USER_AGENT", "langfuse-go/1.0.0")
    config.Environment = os.Getenv("LANGFUSE_ENVIRONMENT")
    config.Release = os.Getenv("LANGFUSE_RELEASE")
    config.Version = getEnvOrDefault("LANGFUSE_VERSION", "1.0.0")
    config.EnableOpenTelemetry = getEnvBoolOrDefault("LANGFUSE_ENABLE_OTEL", false)
    
    return config, nil
}

func NewConfig(opts ...ConfigOption) (*Config, error) {
    config, err := LoadConfig()
    if err != nil {
        return nil, err
    }
    
    // Apply configuration options
    for _, opt := range opts {
        opt(config)
    }
    
    return config, nil
}

type ConfigOption func(*Config)

func WithHost(host string) ConfigOption {
    return func(c *Config) {
        c.Host = host
    }
}

func WithCredentials(publicKey, secretKey string) ConfigOption {
    return func(c *Config) {
        c.PublicKey = publicKey
        c.SecretKey = secretKey
    }
}

func WithDebug(debug bool) ConfigOption {
    return func(c *Config) {
        c.Debug = debug
    }
}
```

## API Layer Design

The API layer provides a clean separation between the high-level SDK interface and the HTTP transport layer, with each resource having its own dedicated client and types.

### Core API Client

```go
// api/client.go
type APIClient struct {
    core        *core.HTTPClient
    Annotations *annotations.Client
    Comments    *comments.Client  
    Datasets    *datasets.Client
    Health      *health.Client
    Ingestion   *ingestion.Client
    Media       *media.Client
    Metrics     *metrics.Client
    Models      *models.Client
    Observations *observations.Client
    Organizations *organizations.Client
    Projects    *projects.Client
    Prompts     *prompts.Client
    Scores      *scores.Client
    Sessions    *sessions.Client
    Traces      *traces.Client
}

func NewAPIClient(config *Config) (*APIClient, error) {
    httpClient := core.NewHTTPClient(config)
    
    return &APIClient{
        core:          httpClient,
        Annotations:   annotations.NewClient(httpClient),
        Comments:      comments.NewClient(httpClient),
        Datasets:      datasets.NewClient(httpClient),
        Health:        health.NewClient(httpClient),
        Ingestion:     ingestion.NewClient(httpClient),
        Media:         media.NewClient(httpClient),
        Metrics:       metrics.NewClient(httpClient),
        Models:        models.NewClient(httpClient),
        Observations:  observations.NewClient(httpClient),
        Organizations: organizations.NewClient(httpClient),
        Projects:      projects.NewClient(httpClient),
        Prompts:       prompts.NewClient(httpClient),
        Scores:        scores.NewClient(httpClient),
        Sessions:      sessions.NewClient(httpClient),
        Traces:        traces.NewClient(httpClient),
    }, nil
}
```

### Resource Client Example

```go
// api/resources/ingestion/client.go
type Client struct {
    httpClient *core.HTTPClient
}

func NewClient(httpClient *core.HTTPClient) *Client {
    return &Client{httpClient: httpClient}
}

func (c *Client) Submit(ctx context.Context, req *types.IngestionRequest) (*types.IngestionResponse, error) {
    response := &types.IngestionResponse{}
    err := c.httpClient.Post(ctx, "/api/public/ingestion", req, response)
    if err != nil {
        return nil, fmt.Errorf("failed to submit ingestion batch: %w", err)
    }
    return response, nil
}

func (c *Client) SubmitBatch(ctx context.Context, events []types.IngestionEvent) (*types.IngestionResponse, error) {
    req := &types.IngestionRequest{
        Batch: events,
        Metadata: map[string]interface{}{
            "sdk_version": "go-1.0.0",
            "timestamp": time.Now().Unix(),
        },
    }
    return c.Submit(ctx, req)
}

// api/resources/ingestion/types/ingestion_request.go
type IngestionRequest struct {
    Batch    []IngestionEvent       `json:"batch"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type IngestionEvent struct {
    ID        string      `json:"id"`
    Type      string      `json:"type"`  // "trace-create", "observation-create", "score-create"
    Timestamp time.Time   `json:"timestamp"`
    Body      interface{} `json:"body"`
}

type IngestionResponse struct {
    Success bool                   `json:"success"`
    Errors  []IngestionError      `json:"errors,omitempty"`
    Usage   *IngestionUsage       `json:"usage,omitempty"`
}
```

## Core HTTP Client Layer

```go
// api/core/http_client.go
import (
    "context"
    "encoding/base64"
    "fmt"
    "time"
    
    "github.com/go-resty/resty/v2"
)

type HTTPClient struct {
    client *resty.Client
    config *Config
}

type RequestOptions struct {
    Method      string
    Path        string
    QueryParams map[string]string
    Body        interface{}
    Result      interface{}
    Headers     map[string]string
}

func NewHTTPClient(config *Config) *HTTPClient {
    client := resty.New().
        SetBaseURL(config.Host).
        SetTimeout(config.RequestTimeout).
        SetHeader("Content-Type", "application/json").
        SetHeader("User-Agent", fmt.Sprintf("langfuse-go/%s", config.Version))
    
    // Set up authentication
    if config.PublicKey != "" && config.SecretKey != "" {
        auth := base64.StdEncoding.EncodeToString([]byte(config.PublicKey + ":" + config.SecretKey))
        client.SetHeader("Authorization", "Basic "+auth)
    }
    
    // Configure retry settings
    client.SetRetryCount(3).
        SetRetryWaitTime(1 * time.Second).
        SetRetryMaxWaitTime(10 * time.Second).
        AddRetryCondition(func(r *resty.Response, err error) bool {
            return r.StatusCode() >= 500 || err != nil
        })
    
    // Add debug logging if enabled
    if config.Debug {
        client.SetDebug(true)
    }
    
    return &HTTPClient{
        client: client,
        config: config,
    }
}

func (c *HTTPClient) DoRequest(ctx context.Context, opts *RequestOptions) error {
    req := c.client.R().SetContext(ctx)
    
    // Set query parameters
    if opts.QueryParams != nil {
        req.SetQueryParams(opts.QueryParams)
    }
    
    // Set additional headers
    if opts.Headers != nil {
        req.SetHeaders(opts.Headers)
    }
    
    // Set request body
    if opts.Body != nil {
        req.SetBody(opts.Body)
    }
    
    // Set result destination
    if opts.Result != nil {
        req.SetResult(opts.Result)
    }
    
    // Set error response handler
    req.SetError(&APIError{})
    
    // Execute request
    var resp *resty.Response
    var err error
    
    switch opts.Method {
    case "GET":
        resp, err = req.Get(opts.Path)
    case "POST":
        resp, err = req.Post(opts.Path)
    case "PUT":
        resp, err = req.Put(opts.Path)
    case "DELETE":
        resp, err = req.Delete(opts.Path)
    case "PATCH":
        resp, err = req.Patch(opts.Path)
    default:
        return fmt.Errorf("unsupported HTTP method: %s", opts.Method)
    }
    
    if err != nil {
        return fmt.Errorf("HTTP request failed: %w", err)
    }
    
    // Handle HTTP error responses
    if resp.IsError() {
        if apiErr, ok := resp.Error().(*APIError); ok {
            return apiErr
        }
        return fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status())
    }
    
    return nil
}

// api/core/auth.go
type APIError struct {
    Code    int                    `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}

// Helper methods for common HTTP operations
func (c *HTTPClient) Get(ctx context.Context, path string, result interface{}) error {
    return c.DoRequest(ctx, &RequestOptions{
        Method: "GET",
        Path:   path,
        Result: result,
    })
}

func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
    return c.DoRequest(ctx, &RequestOptions{
        Method: "POST",
        Path:   path,
        Body:   body,
        Result: result,
    })
}

func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
    return c.DoRequest(ctx, &RequestOptions{
        Method: "PUT",
        Path:   path,
        Body:   body,
        Result: result,
    })
}

func (c *HTTPClient) Delete(ctx context.Context, path string, result interface{}) error {
    return c.DoRequest(ctx, &RequestOptions{
        Method: "DELETE",
        Path:   path,
        Result: result,
    })
}
```

## Async Processing

```go
type IngestionQueue struct {
    buffer   []IngestionEvent
    mu       sync.RWMutex
    flushAt  int
    interval time.Duration
    client   *HTTPClient
    
    ticker   *time.Ticker
    stopCh   chan struct{}
    flushCh  chan struct{}
}

func NewIngestionQueue(client *HTTPClient, flushAt int, interval time.Duration) *IngestionQueue {
    // Initialize queue with background worker
}

func (q *IngestionQueue) Enqueue(event IngestionEvent) error {
    // Add event to buffer, trigger flush if needed
}

func (q *IngestionQueue) Flush() error {
    // Force flush all pending events
}

func (q *IngestionQueue) worker() {
    // Background worker to handle periodic flushing
    for {
        select {
        case <-q.ticker.C:
            q.periodicFlush()
        case <-q.flushCh:
            q.forceFlush()
        case <-q.stopCh:
            q.finalFlush()
            return
        }
    }
}
```

## Error Handling

### SDK-Specific Error Types
```go
type LangfuseError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType string

const (
    ErrorTypeValidation    ErrorType = "VALIDATION"
    ErrorTypeNetwork      ErrorType = "NETWORK"
    ErrorTypeAuth         ErrorType = "AUTH"
    ErrorTypeRateLimit    ErrorType = "RATE_LIMIT"
    ErrorTypeServerError  ErrorType = "SERVER_ERROR"
    ErrorTypeTimeout      ErrorType = "TIMEOUT"
)

func (e *LangfuseError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *LangfuseError) Is(target error) bool {
    t, ok := target.(*LangfuseError)
    return ok && e.Type == t.Type
}
```

### Retry Strategy
```go
type RetryConfig struct {
    MaxAttempts     int
    BaseDelay       time.Duration
    MaxDelay        time.Duration
    BackoffFactor   float64
    RetryableErrors []ErrorType
}

func (c *HTTPClient) submitWithRetry(ctx context.Context, req *IngestionRequest) error {
    var lastErr error
    
    for attempt := 0; attempt < c.retryConfig.MaxAttempts; attempt++ {
        if attempt > 0 {
            delay := c.calculateBackoff(attempt)
            timer := time.NewTimer(delay)
            select {
            case <-ctx.Done():
                timer.Stop()
                return ctx.Err()
            case <-timer.C:
            }
        }
        
        err := c.Submit(ctx, req)
        if err == nil {
            return nil
        }
        
        lastErr = err
        if !c.shouldRetry(err) {
            break
        }
    }
    
    return lastErr
}
```

## Testing Strategy

### Unit Testing

**Builder Pattern Tests:**
```go
func TestTraceBuilder_FluentAPI(t *testing.T) {
    client := createTestClient()
    
    trace := client.Trace("test-trace").
        WithUser("user123").
        WithSession("session456").
        WithMetadata(map[string]interface{}{
            "key": "value",
        }).
        WithTags("tag1", "tag2")
    
    assert.Equal(t, "test-trace", trace.trace.Name)
    assert.Equal(t, "user123", *trace.trace.UserID)
    assert.Equal(t, "session456", *trace.trace.SessionID)
    assert.Equal(t, []string{"tag1", "tag2"}, trace.trace.Tags)
}

func TestGenerationBuilder_ModelParameters(t *testing.T) {
    client := createTestClient()
    
    generation := client.Generation("test-generation").
        WithModel("gpt-4", map[string]interface{}{
            "temperature": 0.7,
            "max_tokens": 1000,
        }).
        WithInput("Hello, world!").
        WithUsage(&Usage{
            Input:  intPtr(10),
            Output: intPtr(20),
            Total:  intPtr(30),
        })
    
    assert.Equal(t, "gpt-4", *generation.observation.Model)
    assert.Equal(t, 0.7, generation.observation.ModelParameters["temperature"])
    assert.Equal(t, 30, *generation.observation.Usage.Total)
}
```

**Configuration Tests:**
```go
func TestConfig_LoadFromEnvironment(t *testing.T) {
    os.Setenv("LANGFUSE_HOST", "https://test.langfuse.com")
    os.Setenv("LANGFUSE_PUBLIC_KEY", "pk_test_123")
    os.Setenv("LANGFUSE_SECRET_KEY", "sk_test_456")
    defer func() {
        os.Unsetenv("LANGFUSE_HOST")
        os.Unsetenv("LANGFUSE_PUBLIC_KEY")
        os.Unsetenv("LANGFUSE_SECRET_KEY")
    }()
    
    config, err := LoadConfig()
    assert.NoError(t, err)
    assert.Equal(t, "https://test.langfuse.com", config.Host)
    assert.Equal(t, "pk_test_123", config.PublicKey)
    assert.Equal(t, "sk_test_456", config.SecretKey)
}
```

### Integration Testing

**HTTP Client Tests:**
```go
func TestHTTPClient_SubmitBatch(t *testing.T) {
    // Start test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/api/public/ingestion", r.URL.Path)
        assert.Contains(t, r.Header.Get("Authorization"), "Basic")
        
        var req IngestionRequest
        json.NewDecoder(r.Body).Decode(&req)
        assert.Len(t, req.Batch, 1)
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": true,
        })
    }))
    defer server.Close()
    
    client := NewHTTPClient(server.URL, &BasicAuthProvider{
        publicKey: "test-key",
        secretKey: "test-secret",
    })
    
    req := &IngestionRequest{
        Batch: []IngestionEvent{{
            ID:        "test-event-id",
            Type:      "trace-create",
            Timestamp: time.Now(),
            Body:      map[string]interface{}{"name": "test"},
        }},
    }
    
    err := client.Submit(context.Background(), req)
    assert.NoError(t, err)
}
```

### Performance Testing

```go
func BenchmarkTraceCreation(b *testing.B) {
    client := createTestClient()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        trace := client.Trace(fmt.Sprintf("trace-%d", i))
        trace.WithUser("user123").
              WithMetadata(map[string]interface{}{"iteration": i}).
              End()
    }
}

func TestConcurrentTraceCreation(t *testing.T) {
    client := createTestClient()
    
    const numGoroutines = 100
    const tracesPerGoroutine = 10
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(routineID int) {
            defer wg.Done()
            
            for j := 0; j < tracesPerGoroutine; j++ {
                trace := client.Trace(fmt.Sprintf("trace-%d-%d", routineID, j))
                if err := trace.End(); err != nil {
                    errors <- err
                    return
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("Concurrent trace creation failed: %v", err)
    }
}
```

### Mock Testing Utilities

```go
// Mock HTTP client for testing
type MockHTTPClient struct {
    mock.Mock
}

func (m *MockHTTPClient) Submit(ctx context.Context, req *IngestionRequest) error {
    args := m.Called(ctx, req)
    return args.Error(0)
}

// Test data builders
func createTestTrace() *Trace {
    return &Trace{
        ID:          "test-trace-id",
        Name:        "test-trace",
        Environment: "test",
        Timestamp:   time.Now(),
        Observations: []Observation{},
    }
}

func createTestClient() *Langfuse {
    config := &Config{
        Host:      "https://test.langfuse.com",
        PublicKey: "test-public-key",
        SecretKey: "test-secret-key",
        Enabled:   true,
    }
    
    client, _ := New(config)
    return client
}
```

## Implementation Phases

### Phase 1: Core SDK Foundation
- Basic data models (Trace, Observation, Score, Usage)
- HTTP transport layer with authentication
- Configuration management
- Simple trace creation and submission

### Phase 2: Advanced Features
- Builder pattern API with fluent interface
- Async processing with ingestion queue
- Error handling and retry mechanisms
- Basic validation and ID generation

### Phase 3: Performance & Reliability
- Batch processing optimization
- Circuit breaker pattern
- Comprehensive error handling
- Memory-efficient buffering

### Phase 4: Integrations & Ecosystem
- OpenTelemetry integration
- HTTP/gRPC middleware
- Framework-specific helpers
- Comprehensive testing suite

## Design Decisions and Rationales

### Why Builder Pattern for Public API?
- **Fluent Interface**: Natural, readable API that matches developer expectations
- **Flexibility**: Easy to add optional parameters without breaking changes
- **Type Safety**: Compile-time validation of required vs optional parameters
- **Immutability**: Builders can create immutable data structures

### Why Async by Default?
- **Performance**: Non-blocking operations don't impact application performance
- **Reliability**: Buffering and batching improve reliability and reduce API load
- **User Experience**: Applications remain responsive during trace submission
- **Scalability**: Handle high-throughput scenarios without blocking

### Why Structured Error Types?
- **Error Handling**: Clear categorization of different error types
- **Retry Logic**: Enables smart retry strategies based on error type
- **Debugging**: Better error messages and debugging information
- **Integration**: Proper error propagation to calling code

### Why API Directory Structure?
- **Resource Separation**: Each Langfuse resource (traces, scores, datasets, etc.) has its own dedicated client and types
- **Code Organization**: Clear separation of concerns with dedicated `client.go` and `types/` for each resource
- **Type Safety**: Strong typing for all API request/response structures
- **Maintainability**: Easy to maintain and extend individual resource APIs independently
- **Python Compatibility**: Mirrors the proven Python Langfuse SDK structure for consistency

### Why Internal Package Structure?
- **API Stability**: Public API remains stable while internal implementation can evolve
- **Encapsulation**: Hide implementation details from SDK users
- **Testing**: Easier to test internal components separately
- **Modularity**: Clear separation between public interface and implementation

### API Usage Patterns

The new structure supports both high-level builder patterns and direct API access:

```go
// High-level builder pattern (recommended for most users)
trace := client.Trace("my-trace").
    WithUser("user123").
    WithSession("session456")

generation := trace.Generation("my-generation").
    WithModel("gpt-4", map[string]interface{}{
        "temperature": 0.7,
    }).
    WithInput("Hello world")

// Direct API access (for advanced users)
traces, err := client.API().Traces.List(ctx, &traces.ListTracesRequest{
    ProjectId: "project123",
    Limit:     50,
})

// Resource-specific operations
datasets, err := client.API().Datasets.List(ctx, &datasets.ListDatasetsRequest{
    ProjectId: "project123",
})

score, err := client.API().Scores.Create(ctx, &scores.CreateScoreRequest{
    TraceId: "trace123",
    Name:    "quality",
    Value:   0.95,
})
```

This design provides a solid foundation for a production-ready Golang Langfuse SDK focused on developer experience, performance, and reliability while maintaining clean architecture and comprehensive testing coverage.