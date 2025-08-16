# Advanced Langfuse Go SDK Examples

This directory contains advanced usage examples demonstrating complex patterns, integrations, and production-ready implementations of the Langfuse Go SDK.

## Prerequisites

1. **Install Dependencies**: Ensure Go 1.18+ and run:
   ```bash
   go mod tidy
   ```

2. **Langfuse Setup**: Configure your Langfuse credentials:
   ```bash
   export LANGFUSE_PUBLIC_KEY="pk_your_public_key_here"
   export LANGFUSE_SECRET_KEY="sk_your_secret_key_here"
   export LANGFUSE_HOST="https://cloud.langfuse.com"
   ```

3. **Review Basic Examples**: Familiarize yourself with the [basic examples](../basic/) first.

## Advanced Examples Overview

### 1. LLM Chain Tracing (`llm_chain_tracing.go`)
Demonstrates complex multi-step LLM workflows with hierarchical tracing.

```bash
go run examples/advanced/llm_chain_tracing.go
```

**What it demonstrates:**
- Multi-step LLM processing chains
- Hierarchical trace and span relationships
- Complex workflow orchestration
- Parent-child span relationships
- Chain-level evaluation and scoring
- Real-world AI agent patterns

**Key Concepts:**
- Sequential processing steps
- Step-by-step result passing
- Comprehensive metadata tracking
- Performance monitoring across chains

### 2. Concurrent Processing (`concurrent_processing.go`)
Shows how to handle concurrent operations with proper tracing in multi-threaded scenarios.

```bash
go run examples/advanced/concurrent_processing.go
```

**What it demonstrates:**
- Worker pool patterns with tracing
- Concurrent job processing
- Batch processing with parallelism
- Thread-safe trace creation
- Performance monitoring in concurrent systems
- Aggregated statistics and reporting

**Key Concepts:**
- Worker pool implementation
- Concurrent trace management
- Batch result aggregation
- Performance metrics collection

### 3. Direct API Usage (`direct_api_usage.go`)
Comprehensive demonstration of direct API client usage for advanced scenarios.

```bash
go run examples/advanced/direct_api_usage.go
```

**What it demonstrates:**
- Direct API client access
- CRUD operations on Langfuse resources
- Dataset management
- Session handling
- Organization and project management
- Model information retrieval
- Health check implementations
- Statistical reporting

**Key Concepts:**
- Low-level API access
- Resource management
- System administration tasks
- Direct REST API operations

### 4. Error Handling & Retry Logic (`error_handling_retry.go`)
Advanced error handling, retry mechanisms, and resilience patterns.

```bash
go run examples/advanced/error_handling_retry.go
```

**What it demonstrates:**
- Sophisticated retry strategies
- Exponential backoff with jitter
- Circuit breaker pattern simulation
- Error classification and handling
- Resilience patterns
- Failure tracking and analysis
- Different error types and responses

**Key Concepts:**
- Retry policy implementation
- Backoff strategies
- Circuit breaker states
- Error metrics and monitoring

## Advanced Patterns & Concepts

### 1. Hierarchical Tracing
```go
// Main process trace
mainTrace := client.Trace("complex-workflow")

// Sub-processes as spans
analysisSpan := mainTrace.Span("analysis-phase")
processingSpan := mainTrace.Span("processing-phase")
validationSpan := mainTrace.Span("validation-phase")

// Nested operations
generation := analysisSpan.Generation("llm-analysis")
```

### 2. Concurrent Trace Management
```go
// Thread-safe trace creation
var wg sync.WaitGroup
for i := 0; i < workerCount; i++ {
    wg.Add(1)
    go func(workerID int) {
        defer wg.Done()
        trace := client.Trace(fmt.Sprintf("worker-%d", workerID))
        // Process work with proper tracing
    }(i)
}
wg.Wait()
```

### 3. Advanced Configuration
```go
client, err := client.NewWithOptions(
    client.WithCredentials("pk_...", "sk_..."),
    client.WithRetrySettings(5, 2*time.Second),
    client.WithFlushSettings(50, 30*time.Second),
    client.WithQueueSize(5000),
    client.WithTimeout(30*time.Second),
)
```

### 4. Error Handling Patterns
```go
for attempt := 1; attempt <= maxRetries; attempt++ {
    err := processRequest(ctx, request, attempt)
    if err == nil {
        // Success - record and return
        return nil
    }
    
    // Record failure and retry if appropriate
    if isRetryable(err) && attempt < maxRetries {
        delay := calculateBackoff(attempt)
        time.Sleep(delay)
        continue
    }
    
    return fmt.Errorf("all retries exhausted: %w", err)
}
```

## Production Considerations

### 1. Performance Optimization
- **Batch Processing**: Use appropriate flush settings for your workload
- **Queue Sizing**: Configure queue size based on throughput requirements
- **Connection Pooling**: Leverage HTTP client connection reuse
- **Asynchronous Operations**: Utilize non-blocking trace submission

### 2. Error Resilience
- **Retry Strategies**: Implement exponential backoff with jitter
- **Circuit Breakers**: Protect against cascading failures
- **Graceful Degradation**: Continue operation when tracing fails
- **Health Monitoring**: Regular health checks and status monitoring

### 3. Resource Management
- **Memory Usage**: Monitor and control queue memory consumption
- **Connection Limits**: Respect API rate limits and connection pools
- **Graceful Shutdown**: Ensure proper cleanup on application termination
- **Context Handling**: Use context for cancellation and timeouts

### 4. Monitoring & Observability
- **Statistics Tracking**: Monitor SDK performance metrics
- **Error Rates**: Track submission success/failure rates
- **Latency Monitoring**: Measure trace submission latency
- **Queue Health**: Monitor queue depth and processing rates

## Best Practices for Advanced Usage

### 1. Trace Design
- **Logical Grouping**: Group related operations in single traces
- **Meaningful Names**: Use descriptive trace and span names
- **Hierarchical Structure**: Organize spans in logical parent-child relationships
- **Metadata Strategy**: Include relevant context without over-indexing

### 2. Performance Management
- **Sampling Strategies**: Implement intelligent sampling for high-volume systems
- **Buffering**: Use appropriate buffer sizes for your workload
- **Batching**: Optimize batch sizes for network efficiency
- **Resource Cleanup**: Always call Shutdown() in production

### 3. Error Handling
- **Non-blocking**: Never let tracing failures block business logic
- **Retry Logic**: Implement intelligent retry with circuit breaking
- **Fallback Strategies**: Have fallback mechanisms for trace failures
- **Monitoring**: Monitor and alert on tracing system health

### 4. Security & Compliance
- **Credential Management**: Use environment variables for credentials
- **Data Privacy**: Avoid logging sensitive information in traces
- **Access Control**: Implement proper access controls for trace data
- **Audit Trails**: Maintain audit logs for compliance requirements

## Troubleshooting Advanced Scenarios

### High-Volume Environments
- Monitor queue depth and adjust flush settings
- Use sampling to reduce trace volume
- Implement circuit breakers for resilience
- Monitor memory usage and GC pressure

### Concurrent Operations
- Ensure thread-safe trace creation
- Avoid sharing builder instances across goroutines
- Use appropriate synchronization for shared resources
- Monitor for deadlocks and race conditions

### Network Issues
- Implement retry with exponential backoff
- Use health checks to detect connectivity issues
- Have fallback mechanisms for network failures
- Monitor network latency and connection health

### Performance Issues
- Profile application to identify bottlenecks
- Optimize batch sizes and flush intervals
- Use asynchronous operations where possible
- Monitor SDK overhead and adjust configuration

## Integration Examples

These advanced examples can be integrated into:

- **Microservices**: Distributed tracing across service boundaries
- **AI/ML Pipelines**: Complex data processing and model inference workflows  
- **API Gateways**: Request routing and transformation tracing
- **Background Workers**: Job processing and queue management
- **Real-time Systems**: Event processing and streaming applications

## Next Steps

1. **Middleware Integration**: Check out [middleware examples](../middleware/) for HTTP/gRPC integration
2. **Custom Patterns**: Adapt these patterns to your specific use cases
3. **Production Deployment**: Use these patterns as templates for production systems
4. **Performance Tuning**: Benchmark and optimize configurations for your workload

For more detailed API reference, see the package documentation and source code.