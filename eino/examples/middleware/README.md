# Middleware Integration Examples

This directory demonstrates how to integrate Langfuse tracing with HTTP and gRPC middleware, providing automatic observability for web services and microservices.

## Prerequisites

1. **Dependencies**: Install required packages:
   ```bash
   go mod tidy
   ```

2. **Langfuse Configuration**:
   ```bash
   export LANGFUSE_PUBLIC_KEY="pk_your_public_key_here"
   export LANGFUSE_SECRET_KEY="sk_your_secret_key_here"  
   export LANGFUSE_HOST="https://cloud.langfuse.com"
   ```

3. **Additional Dependencies** (for gRPC example):
   ```bash
   go get google.golang.org/grpc
   ```

## Examples Overview

### 1. HTTP Middleware (`http_middleware_example.go`)
Complete HTTP server implementation with Langfuse middleware integration.

```bash
go run examples/middleware/http_middleware_example.go
```

**Features:**
- Automatic request/response tracing
- Custom business logic spans
- Error handling and status code tracking
- Multiple middleware layers
- Context propagation
- Response size and timing metrics

**Test the server:**
```bash
# Get user
curl "http://localhost:8080/users?id=123"

# Create user  
curl -X POST "http://localhost:8080/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# Health check
curl "http://localhost:8080/health"
```

### 2. gRPC Middleware (`grpc_middleware_example.go`)
gRPC server with interceptor-based tracing and service method instrumentation.

```bash
go run examples/middleware/grpc_middleware_example.go
```

**Features:**
- Unary server interceptor integration
- Service method tracing
- Request/response payload capture
- Error handling for gRPC status codes
- Custom business logic interceptors
- Simulated AI/LLM operations

## Architecture Patterns

### HTTP Middleware Stack
```
Request
  ↓
Logging Middleware
  ↓
Business Logic Middleware  
  ↓
Langfuse HTTP Middleware
  ↓
Application Handler
  ↓
Response
```

### gRPC Interceptor Chain
```
gRPC Request
  ↓
Logging Interceptor
  ↓
Business Logic Interceptor
  ↓
Langfuse gRPC Interceptor
  ↓
Service Method
  ↓
gRPC Response
```

## Key Integration Concepts

### 1. Automatic Trace Creation
Middleware automatically creates traces for each request:

```go
// HTTP
handler := middleware.HTTP(langfuseClient)(yourHandler)

// gRPC
server := grpc.NewServer(
    grpc.UnaryInterceptor(middleware.UnaryServerInterceptor(langfuseClient)),
)
```

### 2. Context Propagation
Traces are passed through request context:

```go
// Extract trace from context
trace := middleware.TraceFromContext(r.Context())

// Add custom spans to existing trace
span := trace.Span("custom-operation")
```

### 3. Layered Middleware
Multiple middleware can enhance the same trace:

```go
// Chain middleware
handler = langfuseMiddleware(businessLogicMiddleware(baseHandler))

// Each layer adds spans to the same trace
trace.Span("langfuse-layer")
trace.Span("business-logic-layer")
```

## Production Implementation

### HTTP Server Configuration
```go
func setupHTTPServer(langfuse *client.Langfuse) *http.Server {
    mux := http.NewServeMux()
    
    // Register routes
    mux.HandleFunc("/api/users", userHandler)
    mux.HandleFunc("/api/health", healthHandler)
    
    // Apply middleware layers
    handler := middleware.HTTP(langfuse)(mux)
    handler = rateLimitMiddleware(handler)
    handler = authMiddleware(handler)
    handler = loggingMiddleware(handler)
    
    return &http.Server{
        Addr:    ":8080",
        Handler: handler,
        Timeout: 30 * time.Second,
    }
}
```

### gRPC Server Configuration
```go
func setupGRPCServer(langfuse *client.Langfuse) *grpc.Server {
    server := grpc.NewServer(
        grpc.UnaryInterceptor(grpc.ChainUnaryInterceptor(
            middleware.UnaryServerInterceptor(langfuse),
            authInterceptor,
            rateLimitInterceptor,
            loggingInterceptor,
        )),
        grpc.StreamInterceptor(grpc.ChainStreamInterceptor(
            middleware.StreamServerInterceptor(langfuse),
            authStreamInterceptor,
        )),
    )
    
    // Register services
    pb.RegisterUserServiceServer(server, userService)
    pb.RegisterOrderServiceServer(server, orderService)
    
    return server
}
```

## Advanced Usage Patterns

### 1. Custom Span Creation
Add business-specific spans within traced requests:

```go
func (h *Handler) ProcessOrder(w http.ResponseWriter, r *http.Request) {
    trace := middleware.TraceFromContext(r.Context())
    
    // Validation span
    validationSpan := trace.Span("order-validation")
    if err := validateOrder(order); err != nil {
        validationSpan.WithLevel("ERROR")
        return
    }
    validationSpan.End()
    
    // Processing span
    processingSpan := trace.Span("order-processing")
    result := processOrder(order)
    processingSpan.WithOutput(result).End()
}
```

### 2. Error Handling
Capture and trace errors appropriately:

```go
func businessLogicMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        trace := middleware.TraceFromContext(r.Context())
        
        defer func() {
            if err := recover(); err != nil {
                if trace != nil {
                    trace.WithLevel("ERROR").
                        WithStatusMessage(fmt.Sprintf("Panic: %v", err))
                }
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

### 3. Performance Monitoring
Track performance metrics:

```go
func performanceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        trace := middleware.TraceFromContext(r.Context())
        
        // Wrap response writer to capture metrics
        rw := &metricsResponseWriter{ResponseWriter: w}
        next.ServeHTTP(rw, r)
        
        if trace != nil {
            trace.WithMetadata(map[string]interface{}{
                "response_time_ms": time.Since(start).Milliseconds(),
                "response_size":    rw.bytesWritten,
                "status_code":      rw.statusCode,
            })
        }
    })
}
```

## Best Practices

### 1. Middleware Ordering
Apply middleware in logical order:
```go
// Recommended order (outer to inner):
handler = recoveryMiddleware(        // Panic recovery
    loggingMiddleware(              // Request logging
        authMiddleware(             // Authentication
            rateLimitMiddleware(    // Rate limiting
                langfuseMiddleware( // Tracing
                    businessHandler // Business logic
                )))))
```

### 2. Context Management
Always use context for trace propagation:
```go
// Extract trace from context
trace := middleware.TraceFromContext(ctx)
if trace == nil {
    // Handle case where tracing is disabled
    return
}

// Pass context to downstream calls
result, err := downstreamService.Call(ctx, request)
```

### 3. Resource Cleanup
Ensure proper cleanup in middleware:
```go
func resourceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        resource := acquireResource()
        defer resource.Release()
        
        // Add resource info to trace
        if trace := middleware.TraceFromContext(r.Context()); trace != nil {
            trace.WithMetadata(map[string]interface{}{
                "resource_id": resource.ID(),
            })
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### 4. Error Propagation
Handle errors gracefully in middleware:
```go
func errorHandlingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                if trace := middleware.TraceFromContext(r.Context()); trace != nil {
                    trace.WithLevel("ERROR").
                        WithStatusMessage(fmt.Sprintf("Panic recovered: %v", err)).
                        WithOutput(map[string]interface{}{
                            "error_type": "panic",
                            "error":      err,
                        })
                }
                
                log.Printf("Panic recovered: %v", err)
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

## Integration with Popular Frameworks

### Gin Framework
```go
import "github.com/gin-gonic/gin"

func setupGinServer(langfuse *client.Langfuse) *gin.Engine {
    r := gin.New()
    
    // Langfuse middleware for Gin
    r.Use(func(c *gin.Context) {
        // Adapt HTTP middleware to Gin
        middleware.HTTP(langfuse)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Next()
        }))(c.Writer, c.Request)
    })
    
    r.GET("/users/:id", getUserHandler)
    return r
}
```

### Echo Framework
```go
import "github.com/labstack/echo/v4"

func setupEchoServer(langfuse *client.Langfuse) *echo.Echo {
    e := echo.New()
    
    // Langfuse middleware for Echo
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Adapt HTTP middleware to Echo
            middleware.HTTP(langfuse)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                c.SetRequest(r)
                next(c)
            }))(c.Response(), c.Request())
            return nil
        }
    })
    
    return e
}
```

## Testing Middleware

### Unit Testing
```go
func TestHTTPMiddleware(t *testing.T) {
    // Create test Langfuse client
    langfuse := createTestLangfuseClient()
    
    // Create test handler
    handler := middleware.HTTP(langfuse)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        trace := middleware.TraceFromContext(r.Context())
        assert.NotNil(t, trace)
        w.WriteHeader(http.StatusOK)
    }))
    
    // Test request
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Integration Testing
```go
func TestFullIntegration(t *testing.T) {
    langfuse := createTestLangfuseClient()
    server := setupHTTPServer(langfuse)
    
    // Start test server
    listener, err := net.Listen("tcp", ":0")
    require.NoError(t, err)
    defer listener.Close()
    
    go server.Serve(listener)
    
    // Test client requests
    client := &http.Client{}
    resp, err := client.Get(fmt.Sprintf("http://%s/health", listener.Addr()))
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

These middleware examples provide the foundation for integrating Langfuse tracing into production web services and microservices architectures.