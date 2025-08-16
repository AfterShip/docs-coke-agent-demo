package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"eino/pkg/langfuse/client"
)

// GRPCInterceptorConfig contains configuration options for gRPC interceptors
type GRPCInterceptorConfig struct {
	// Client is the Langfuse client instance to use for tracing
	Client *client.Langfuse
	
	// TraceNameFunc allows customizing trace names based on the gRPC method
	// Default: uses the full method name
	TraceNameFunc func(string) string
	
	// ShouldTraceFunc determines whether a request should be traced
	// Default: traces all requests except health checks
	ShouldTraceFunc func(string) bool
	
	// UserIDExtractor extracts user ID from gRPC metadata
	// Default: looks for "user-id" metadata key
	UserIDExtractor func(metadata.MD) string
	
	// SessionIDExtractor extracts session ID from gRPC metadata
	// Default: looks for "session-id" metadata key
	SessionIDExtractor func(metadata.MD) string
	
	// MetadataExtractor extracts additional metadata from gRPC context
	// Default: includes method name, peer info, and selected metadata
	MetadataExtractor func(context.Context, string) map[string]interface{}
	
	// TagExtractor extracts tags from the gRPC method
	// Default: includes "grpc" and service name
	TagExtractor func(string) []string
	
	// CaptureRequest determines whether to capture request message as input
	// Default: false (can be memory intensive and may contain sensitive data)
	CaptureRequest bool
	
	// CaptureResponse determines whether to capture response message as output
	// Default: false (can be memory intensive and may contain sensitive data)
	CaptureResponse bool
}

// DefaultGRPCInterceptorConfig returns a default configuration
func DefaultGRPCInterceptorConfig(langfuseClient *client.Langfuse) *GRPCInterceptorConfig {
	return &GRPCInterceptorConfig{
		Client:                 langfuseClient,
		TraceNameFunc:          defaultGRPCTraceNameFunc,
		ShouldTraceFunc:        defaultGRPCShouldTraceFunc,
		UserIDExtractor:        defaultGRPCUserIDExtractor,
		SessionIDExtractor:     defaultGRPCSessionIDExtractor,
		MetadataExtractor:      defaultGRPCMetadataExtractor,
		TagExtractor:           defaultGRPCTagExtractor,
		CaptureRequest:         false,
		CaptureResponse:        false,
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that traces requests
func UnaryServerInterceptor(config *GRPCInterceptorConfig) grpc.UnaryServerInterceptor {
	if config == nil {
		panic("GRPCInterceptorConfig cannot be nil")
	}
	if config.Client == nil {
		panic("GRPCInterceptorConfig must have a Langfuse client")
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if this method should be traced
		if !config.ShouldTraceFunc(info.FullMethod) {
			return handler(ctx, req)
		}

		// Start timing
		startTime := time.Now()

		// Create trace
		traceName := config.TraceNameFunc(info.FullMethod)
		traceBuilder := config.Client.Trace(traceName)

		// Extract metadata from gRPC context
		md, _ := metadata.FromIncomingContext(ctx)

		// Set trace metadata
		if metadata := config.MetadataExtractor(ctx, info.FullMethod); metadata != nil {
			traceBuilder.WithMetadata(metadata)
		}

		// Set trace tags
		if tags := config.TagExtractor(info.FullMethod); tags != nil {
			traceBuilder.WithTags(tags...)
		}

		// Set user ID if available
		if userID := config.UserIDExtractor(md); userID != "" {
			traceBuilder.WithUser(userID)
		}

		// Set session ID if available
		if sessionID := config.SessionIDExtractor(md); sessionID != "" {
			traceBuilder.WithSession(sessionID)
		}

		// Capture request input if configured
		if config.CaptureRequest {
			traceBuilder.WithInput(req)
		} else {
			// At minimum, capture method info
			traceBuilder.WithInput(map[string]interface{}{
				"method":     info.FullMethod,
				"service":    extractServiceName(info.FullMethod),
				"rpc":        extractRPCName(info.FullMethod),
			})
		}

		// Create span for the gRPC operation
		spanBuilder := traceBuilder.Span("grpc_unary").
			WithStartTime(startTime)

		// Add trace and span to context
		ctx = context.WithValue(ctx, traceBuilderContextKey, traceBuilder)
		ctx = context.WithValue(ctx, spanBuilderContextKey, spanBuilder)

		// Execute the handler
		resp, err := handler(ctx, req)

		// Complete timing
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// Determine status and level
		grpcStatus := status.Code(err)
		level := "DEFAULT"
		statusMessage := ""

		if err != nil {
			if grpcStatus == codes.Canceled || grpcStatus == codes.DeadlineExceeded {
				level = "WARNING"
			} else if grpcStatus != codes.OK {
				level = "ERROR"
			}
			statusMessage = status.Convert(err).Message()
		}

		// Complete the span
		spanOutput := map[string]interface{}{
			"status_code": grpcStatus.String(),
			"duration_ms": float64(duration.Nanoseconds()) / 1e6,
		}

		if config.CaptureResponse && resp != nil {
			spanOutput["response"] = resp
		}

		spanBuilder.
			WithEndTime(endTime).
			WithOutput(spanOutput).
			WithLevel(level)

		if statusMessage != "" {
			spanBuilder.WithStatusMessage(statusMessage)
		}

		if spanErr := spanBuilder.End(); spanErr != nil {
			// Log error but don't affect the actual RPC
		}

		// Complete the trace
		traceOutput := map[string]interface{}{
			"status_code": grpcStatus.String(),
			"duration_ms": float64(duration.Nanoseconds()) / 1e6,
			"success":     err == nil,
		}

		if config.CaptureResponse && resp != nil {
			traceOutput["response"] = resp
		}

		if err != nil {
			traceOutput["error"] = err.Error()
		}

		traceBuilder.WithOutput(traceOutput)

		if traceErr := traceBuilder.End(); traceErr != nil {
			// Log error but don't affect the actual RPC
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that traces requests
func StreamServerInterceptor(config *GRPCInterceptorConfig) grpc.StreamServerInterceptor {
	if config == nil {
		panic("GRPCInterceptorConfig cannot be nil")
	}
	if config.Client == nil {
		panic("GRPCInterceptorConfig must have a Langfuse client")
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Check if this method should be traced
		if !config.ShouldTraceFunc(info.FullMethod) {
			return handler(srv, ss)
		}

		// Start timing
		startTime := time.Now()

		// Create trace
		traceName := config.TraceNameFunc(info.FullMethod)
		traceBuilder := config.Client.Trace(traceName)

		// Extract metadata from gRPC context
		ctx := ss.Context()
		md, _ := metadata.FromIncomingContext(ctx)

		// Set trace metadata
		if metadata := config.MetadataExtractor(ctx, info.FullMethod); metadata != nil {
			traceBuilder.WithMetadata(metadata)
		}

		// Set trace tags (add streaming tag)
		tags := config.TagExtractor(info.FullMethod)
		tags = append(tags, "streaming")
		traceBuilder.WithTags(tags...)

		// Set user ID if available
		if userID := config.UserIDExtractor(md); userID != "" {
			traceBuilder.WithUser(userID)
		}

		// Set session ID if available
		if sessionID := config.SessionIDExtractor(md); sessionID != "" {
			traceBuilder.WithSession(sessionID)
		}

		// Set input (for streaming, we don't capture the actual stream)
		traceBuilder.WithInput(map[string]interface{}{
			"method":       info.FullMethod,
			"service":      extractServiceName(info.FullMethod),
			"rpc":          extractRPCName(info.FullMethod),
			"stream_type":  getStreamType(info),
		})

		// Create span for the gRPC streaming operation
		spanBuilder := traceBuilder.Span("grpc_stream").
			WithStartTime(startTime)

		// Wrap the server stream to track streaming metrics
		wrappedStream := &tracedServerStream{
			ServerStream: ss,
			traceBuilder: traceBuilder,
			spanBuilder:  spanBuilder,
			messagesSent: 0,
			messagesReceived: 0,
		}

		// Add trace and span to context
		ctx = context.WithValue(ctx, traceBuilderContextKey, traceBuilder)
		ctx = context.WithValue(ctx, spanBuilderContextKey, spanBuilder)
		wrappedStream.ctx = ctx

		// Execute the handler
		err := handler(srv, wrappedStream)

		// Complete timing
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// Determine status and level
		grpcStatus := status.Code(err)
		level := "DEFAULT"
		statusMessage := ""

		if err != nil {
			if grpcStatus == codes.Canceled || grpcStatus == codes.DeadlineExceeded {
				level = "WARNING"
			} else if grpcStatus != codes.OK {
				level = "ERROR"
			}
			statusMessage = status.Convert(err).Message()
		}

		// Complete the span
		spanOutput := map[string]interface{}{
			"status_code":         grpcStatus.String(),
			"duration_ms":         float64(duration.Nanoseconds()) / 1e6,
			"messages_sent":       wrappedStream.messagesSent,
			"messages_received":   wrappedStream.messagesReceived,
		}

		spanBuilder.
			WithEndTime(endTime).
			WithOutput(spanOutput).
			WithLevel(level)

		if statusMessage != "" {
			spanBuilder.WithStatusMessage(statusMessage)
		}

		if spanErr := spanBuilder.End(); spanErr != nil {
			// Log error but don't affect the actual RPC
		}

		// Complete the trace
		traceOutput := map[string]interface{}{
			"status_code":         grpcStatus.String(),
			"duration_ms":         float64(duration.Nanoseconds()) / 1e6,
			"success":             err == nil,
			"messages_sent":       wrappedStream.messagesSent,
			"messages_received":   wrappedStream.messagesReceived,
		}

		if err != nil {
			traceOutput["error"] = err.Error()
		}

		traceBuilder.WithOutput(traceOutput)

		if traceErr := traceBuilder.End(); traceErr != nil {
			// Log error but don't affect the actual RPC
		}

		return err
	}
}

// tracedServerStream wraps grpc.ServerStream to track streaming metrics
type tracedServerStream struct {
	grpc.ServerStream
	traceBuilder     *client.TraceBuilder
	spanBuilder      *client.SpanBuilder
	ctx              context.Context
	messagesSent     int64
	messagesReceived int64
}

func (s *tracedServerStream) Context() context.Context {
	return s.ctx
}

func (s *tracedServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		s.messagesSent++
	}
	return err
}

func (s *tracedServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.messagesReceived++
	}
	return err
}

// UnaryClientInterceptor returns a gRPC unary client interceptor that traces outgoing requests
func UnaryClientInterceptor(config *GRPCInterceptorConfig) grpc.UnaryClientInterceptor {
	if config == nil {
		panic("GRPCInterceptorConfig cannot be nil")
	}
	if config.Client == nil {
		panic("GRPCInterceptorConfig must have a Langfuse client")
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Check if this method should be traced
		if !config.ShouldTraceFunc(method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Start timing
		startTime := time.Now()

		// Create trace (or get from context if already exists)
		var traceBuilder *client.TraceBuilder
		if existingTrace := GetTraceFromContext(ctx); existingTrace != nil {
			traceBuilder = existingTrace
		} else {
			traceName := config.TraceNameFunc(method) + "_client"
			traceBuilder = config.Client.Trace(traceName)
		}

		// Create span for the outgoing gRPC call
		spanBuilder := traceBuilder.Span("grpc_client").
			WithStartTime(startTime)

		// Set span metadata
		spanMetadata := map[string]interface{}{
			"method":     method,
			"service":    extractServiceName(method),
			"rpc":        extractRPCName(method),
			"target":     cc.Target(),
		}
		spanBuilder.WithMetadata(spanMetadata)

		// Set span tags
		tags := []string{"grpc", "client"}
		if serviceName := extractServiceName(method); serviceName != "" {
			tags = append(tags, serviceName)
		}
		spanBuilder.WithTags(tags...)

		// Capture request input if configured
		if config.CaptureRequest {
			spanBuilder.WithInput(req)
		} else {
			spanBuilder.WithInput(map[string]interface{}{
				"method":  method,
				"service": extractServiceName(method),
				"rpc":     extractRPCName(method),
			})
		}

		// Execute the call
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Complete timing
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// Determine status and level
		grpcStatus := status.Code(err)
		level := "DEFAULT"
		statusMessage := ""

		if err != nil {
			if grpcStatus == codes.Canceled || grpcStatus == codes.DeadlineExceeded {
				level = "WARNING"
			} else if grpcStatus != codes.OK {
				level = "ERROR"
			}
			statusMessage = status.Convert(err).Message()
		}

		// Complete the span
		spanOutput := map[string]interface{}{
			"status_code": grpcStatus.String(),
			"duration_ms": float64(duration.Nanoseconds()) / 1e6,
		}

		if config.CaptureResponse && reply != nil {
			spanOutput["response"] = reply
		}

		if err != nil {
			spanOutput["error"] = err.Error()
		}

		spanBuilder.
			WithEndTime(endTime).
			WithOutput(spanOutput).
			WithLevel(level)

		if statusMessage != "" {
			spanBuilder.WithStatusMessage(statusMessage)
		}

		if spanErr := spanBuilder.End(); spanErr != nil {
			// Log error but don't affect the actual RPC
		}

		return err
	}
}

// Default extractor functions

func defaultGRPCTraceNameFunc(fullMethod string) string {
	return fullMethod
}

func defaultGRPCShouldTraceFunc(fullMethod string) bool {
	// Skip health check methods
	return !strings.Contains(strings.ToLower(fullMethod), "health")
}

func defaultGRPCUserIDExtractor(md metadata.MD) string {
	if values := md.Get("user-id"); len(values) > 0 {
		return values[0]
	}
	if values := md.Get("x-user-id"); len(values) > 0 {
		return values[0]
	}
	return ""
}

func defaultGRPCSessionIDExtractor(md metadata.MD) string {
	if values := md.Get("session-id"); len(values) > 0 {
		return values[0]
	}
	if values := md.Get("x-session-id"); len(values) > 0 {
		return values[0]
	}
	return ""
}

func defaultGRPCMetadataExtractor(ctx context.Context, fullMethod string) map[string]interface{} {
	metadata := map[string]interface{}{
		"method":  fullMethod,
		"service": extractServiceName(fullMethod),
		"rpc":     extractRPCName(fullMethod),
	}

	// Add peer information if available
	if peerInfo, ok := peer.FromContext(ctx); ok {
		metadata["peer_addr"] = peerInfo.Addr.String()
	}

	return metadata
}

func defaultGRPCTagExtractor(fullMethod string) []string {
	tags := []string{"grpc"}
	
	if serviceName := extractServiceName(fullMethod); serviceName != "" {
		tags = append(tags, serviceName)
	}
	
	return tags
}

// Helper functions

func extractServiceName(fullMethod string) string {
	// fullMethod format: /package.service/method
	if strings.HasPrefix(fullMethod, "/") {
		fullMethod = fullMethod[1:]
	}
	
	parts := strings.Split(fullMethod, "/")
	if len(parts) >= 1 {
		serviceParts := strings.Split(parts[0], ".")
		if len(serviceParts) > 0 {
			return serviceParts[len(serviceParts)-1]
		}
	}
	
	return ""
}

func extractRPCName(fullMethod string) string {
	// fullMethod format: /package.service/method
	parts := strings.Split(fullMethod, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	
	return ""
}

func getStreamType(info *grpc.StreamServerInfo) string {
	if info.IsClientStream && info.IsServerStream {
		return "bidirectional"
	} else if info.IsClientStream {
		return "client_streaming"
	} else if info.IsServerStream {
		return "server_streaming"
	}
	return "unknown"
}