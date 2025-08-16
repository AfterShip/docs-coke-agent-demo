# Basic Langfuse Go SDK Examples

This directory contains basic usage examples for the Langfuse Go SDK, demonstrating core functionality and common use cases.

## Prerequisites

1. **Install Dependencies**: Make sure you have Go 1.18+ installed and run:
   ```bash
   go mod tidy
   ```

2. **Set up Langfuse Credentials**: You need a Langfuse account and API keys. Get them from:
   - [Langfuse Cloud](https://cloud.langfuse.com) (hosted)
   - Or your self-hosted Langfuse instance

3. **Environment Variables** (recommended):
   ```bash
   export LANGFUSE_PUBLIC_KEY="pk_your_public_key_here"
   export LANGFUSE_SECRET_KEY="sk_your_secret_key_here"
   export LANGFUSE_HOST="https://cloud.langfuse.com"
   export LANGFUSE_ENVIRONMENT="development"
   ```

## Examples Overview

### 1. Simple Trace (`simple_trace.go`)
Basic trace creation with metadata and user information.

```bash
go run examples/basic/simple_trace.go
```

**What it demonstrates:**
- Creating a simple trace
- Adding metadata, user ID, and session ID
- Basic error handling
- Client shutdown

### 2. Basic Generation (`basic_generation.go`)
LLM generation tracking with model parameters and usage metrics.

```bash
go run examples/basic/basic_generation.go
```

**What it demonstrates:**
- Generation builder for LLM calls
- Model parameters tracking
- Input/output capture
- Token usage and cost tracking
- Environment-based configuration

### 3. Trace with Spans (`trace_with_span.go`)
Complex operation tracking with multiple spans in a trace.

```bash
go run examples/basic/trace_with_span.go
```

**What it demonstrates:**
- Creating traces with multiple spans
- Hierarchical operation tracking
- Different span types and levels
- Time-based tracking

### 4. Scoring Example (`scoring_example.go`)
Adding evaluation scores to traces and generations.

```bash
go run examples/basic/scoring_example.go
```

**What it demonstrates:**
- Creating numeric, categorical, and boolean scores
- Attaching scores to traces
- Evaluation workflows
- Different score data types

### 5. Configuration Examples (`configuration_examples.go`)
Different ways to configure the Langfuse client.

```bash
go run examples/basic/configuration_examples.go
```

**What it demonstrates:**
- Environment-based configuration
- Programmatic configuration
- Configuration options pattern
- Production-ready settings
- Health checks

## Running the Examples

### With Environment Variables (Recommended)
```bash
# Set your credentials
export LANGFUSE_PUBLIC_KEY="pk_your_public_key_here"
export LANGFUSE_SECRET_KEY="sk_your_secret_key_here"
export LANGFUSE_HOST="https://cloud.langfuse.com"

# Run any example
go run examples/basic/simple_trace.go
```

### With Hardcoded Credentials (For Testing)
Edit the example files and replace the placeholder credentials:
```go
client.WithCredentials("your-public-key", "your-secret-key")
```

## Common Patterns

### 1. Client Initialization
```go
// Environment-based (recommended)
config, _ := client.LoadConfig()
langfuseClient, _ := client.New(config)

// With options
langfuseClient, _ := client.NewWithOptions(
    client.WithCredentials("pk_...", "sk_..."),
    client.WithHost("https://cloud.langfuse.com"),
    client.WithDebug(true),
)

// Always shutdown gracefully
defer langfuseClient.Shutdown(context.Background())
```

### 2. Trace Creation
```go
trace := langfuseClient.Trace("my-operation").
    WithUserID("user-123").
    WithSessionID("session-456").
    WithInput(data).
    WithMetadata(metadata)

// Do work...

trace.WithOutput(result)
if err := trace.End(); err != nil {
    log.Printf("Failed to submit trace: %v", err)
}
```

### 3. Generation Tracking
```go
generation := langfuseClient.Generation("llm-call").
    WithModel("gpt-4", map[string]interface{}{
        "temperature": 0.7,
        "max_tokens": 1000,
    }).
    WithInput(prompt).
    WithStartTime(time.Now())

// Make LLM call...

generation.WithOutput(response).
    WithEndTime(time.Now()).
    WithUsage(usage)

if err := generation.End(); err != nil {
    log.Printf("Failed to submit generation: %v", err)
}
```

## Best Practices

1. **Always call `Shutdown()`**: Use `defer langfuseClient.Shutdown(context.Background())` to ensure proper cleanup

2. **Use Environment Variables**: Store credentials in environment variables, not in code

3. **Handle Errors**: Check for errors when submitting traces, generations, and scores

4. **Use Context**: Pass context for cancellation and timeouts

5. **Flush When Needed**: Call `Flush()` if you need immediate submission

6. **Monitor Performance**: Use `GetStats()` to monitor SDK performance

## Troubleshooting

1. **Authentication Errors**: Verify your public and secret keys are correct

2. **Network Issues**: Check your internet connection and Langfuse host URL

3. **Debug Mode**: Enable debug mode to see detailed HTTP requests:
   ```go
   client.WithDebug(true)
   ```

4. **Health Check**: Use health check to verify connectivity:
   ```go
   if err := langfuseClient.HealthCheck(ctx); err != nil {
       log.Printf("Health check failed: %v", err)
   }
   ```

## Next Steps

- Check out the [Advanced Examples](../advanced/) for more complex use cases
- Review the [Middleware Examples](../middleware/) for HTTP/gRPC integrations
- Read the API documentation for detailed reference