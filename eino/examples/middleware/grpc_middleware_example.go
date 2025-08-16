package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/middleware"
)

// Mock protobuf definitions (normally generated from .proto files)
type UserRequest struct {
	ID string `json:"id"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Created bool   `json:"created"`
}

// UserService implements the gRPC service
type UserService struct {
	langfuse *client.Langfuse
}

// GetUser implements the GetUser RPC method
func (s *UserService) GetUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	// Extract trace from gRPC middleware context
	trace := middleware.TraceFromContext(ctx)
	if trace != nil {
		trace.WithInput(map[string]interface{}{
			"user_id": req.ID,
			"method":  "GetUser",
		}).WithMetadata(map[string]interface{}{
			"service":   "UserService",
			"operation": "get_user",
			"grpc":      true,
		})
	}

	// Validate request
	if req.ID == "" {
		if trace != nil {
			trace.WithLevel("ERROR").
				WithStatusMessage("User ID is required").
				WithOutput("validation_error")
		}
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Create span for user lookup
	if trace != nil {
		lookupSpan := trace.Span("user-lookup").
			WithInput(map[string]interface{}{
				"lookup_id": req.ID,
				"method":    "database_query",
			}).
			WithStartTime(time.Now()).
			WithMetadata(map[string]interface{}{
				"database": "postgresql",
				"table":    "users",
			})

		// Simulate database lookup
		time.Sleep(75 * time.Millisecond)

		lookupSpan.WithOutput(map[string]interface{}{
			"found":     true,
			"query_time": "75ms",
		}).WithEndTime(time.Now())

		if err := lookupSpan.End(); err != nil {
			log.Printf("Failed to submit lookup span: %v", err)
		}
	}

	// Create generation for AI-enhanced user data (simulating AI feature)
	if trace != nil {
		generation := trace.Generation("user-enhancement").
			WithModel("gpt-3.5-turbo", map[string]interface{}{
				"temperature": 0.1,
				"purpose":     "user_data_enhancement",
			}).
			WithInput(fmt.Sprintf("Enhance user profile for ID: %s", req.ID)).
			WithStartTime(time.Now())

		// Simulate AI processing
		time.Sleep(120 * time.Millisecond)

		enhancement := fmt.Sprintf("Enhanced profile data for user %s", req.ID)
		generation.WithOutput(enhancement).
			WithEndTime(time.Now()).
			WithUsage(&client.Usage{
				Input:  intPtr(20),
				Output: intPtr(15),
				Total:  intPtr(35),
			})

		if err := generation.End(); err != nil {
			log.Printf("Failed to submit generation: %v", err)
		}
	}

	// Return user data
	response := &UserResponse{
		ID:    req.ID,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	if trace != nil {
		trace.WithOutput(map[string]interface{}{
			"user_found":    true,
			"response_size": len(response.Name) + len(response.Email),
		})
	}

	return response, nil
}

// CreateUser implements the CreateUser RPC method
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	trace := middleware.TraceFromContext(ctx)
	if trace != nil {
		trace.WithInput(map[string]interface{}{
			"name":   req.Name,
			"email":  req.Email,
			"method": "CreateUser",
		}).WithMetadata(map[string]interface{}{
			"service":   "UserService",
			"operation": "create_user",
			"grpc":      true,
		})
	}

	// Validation
	if req.Name == "" || req.Email == "" {
		if trace != nil {
			trace.WithLevel("ERROR").
				WithStatusMessage("Name and email are required").
				WithOutput("validation_error")
		}
		return nil, status.Error(codes.InvalidArgument, "name and email are required")
	}

	// Create validation span
	if trace != nil {
		validationSpan := trace.Span("input-validation").
			WithInput(map[string]interface{}{
				"name":  req.Name,
				"email": req.Email,
			}).
			WithStartTime(time.Now())

		// Simulate validation
		time.Sleep(25 * time.Millisecond)

		validationSpan.WithOutput(map[string]interface{}{
			"valid":           true,
			"email_format_ok": true,
			"name_length_ok":  len(req.Name) > 0,
		}).WithEndTime(time.Now())

		if err := validationSpan.End(); err != nil {
			log.Printf("Failed to submit validation span: %v", err)
		}
	}

	// Create database insertion span
	if trace != nil {
		dbSpan := trace.Span("database-insert").
			WithInput(map[string]interface{}{
				"table":  "users",
				"fields": []string{"name", "email"},
			}).
			WithStartTime(time.Now()).
			WithMetadata(map[string]interface{}{
				"database":   "postgresql",
				"connection": "primary",
			})

		// Simulate database insert
		time.Sleep(85 * time.Millisecond)

		newUserID := fmt.Sprintf("user_%d", time.Now().Unix())
		dbSpan.WithOutput(map[string]interface{}{
			"user_id":       newUserID,
			"rows_affected": 1,
			"insert_time":   "85ms",
		}).WithEndTime(time.Now())

		if err := dbSpan.End(); err != nil {
			log.Printf("Failed to submit database span: %v", err)
		}
	}

	// Generate welcome message using AI
	if trace != nil {
		generation := trace.Generation("welcome-message").
			WithModel("gpt-3.5-turbo", map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  100,
			}).
			WithInput(fmt.Sprintf("Generate a welcome message for new user: %s", req.Name)).
			WithStartTime(time.Now())

		// Simulate AI processing
		time.Sleep(150 * time.Millisecond)

		welcomeMessage := fmt.Sprintf("Welcome to our platform, %s! We're excited to have you aboard.", req.Name)
		generation.WithOutput(welcomeMessage).
			WithEndTime(time.Now()).
			WithUsage(&client.Usage{
				Input:  intPtr(25),
				Output: intPtr(20),
				Total:  intPtr(45),
			})

		if err := generation.End(); err != nil {
			log.Printf("Failed to submit welcome generation: %v", err)
		}
	}

	response := &CreateUserResponse{
		ID:      fmt.Sprintf("user_%d", time.Now().Unix()),
		Name:    req.Name,
		Email:   req.Email,
		Created: true,
	}

	if trace != nil {
		trace.WithOutput(map[string]interface{}{
			"user_created": true,
			"user_id":      response.ID,
		})
	}

	return response, nil
}

// Custom gRPC interceptor for additional business logic
func BusinessLogicInterceptor(langfuse *client.Langfuse) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get existing trace from Langfuse middleware
		trace := middleware.TraceFromContext(ctx)
		
		if trace != nil {
			// Add business logic span
			businessSpan := trace.Span("grpc-business-logic").
				WithInput(map[string]interface{}{
					"method":     info.FullMethod,
					"request":    fmt.Sprintf("%T", req),
					"timestamp":  time.Now(),
				}).
				WithStartTime(time.Now()).
				WithMetadata(map[string]interface{}{
					"interceptor": "business_logic",
					"grpc_method": info.FullMethod,
				})

			// Process request
			resp, err := handler(ctx, req)

			// Complete business span
			output := map[string]interface{}{
				"success": err == nil,
			}
			if err != nil {
				output["error"] = err.Error()
				businessSpan.WithLevel("ERROR").
					WithStatusMessage(err.Error())
			}
			if resp != nil {
				output["response_type"] = fmt.Sprintf("%T", resp)
			}

			businessSpan.WithOutput(output).WithEndTime(time.Now())

			if submitErr := businessSpan.End(); submitErr != nil {
				log.Printf("Failed to submit business logic span: %v", submitErr)
			}

			return resp, err
		}

		return handler(ctx, req)
	}
}

func main() {
	// Initialize Langfuse client
	langfuseClient, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
		client.WithFlushSettings(5, 10*time.Second),
	)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	// Create gRPC server with interceptors
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc.ChainUnaryInterceptor(
			middleware.UnaryServerInterceptor(langfuseClient), // Langfuse tracing
			BusinessLogicInterceptor(langfuseClient),          // Custom business logic
			loggingInterceptor,                                // Request logging
		)),
	)

	// Register service
	userService := &UserService{langfuse: langfuseClient}
	// In a real application, you would register the service with the generated protobuf code:
	// pb.RegisterUserServiceServer(server, userService)
	
	// For this example, we'll create a simple mock service registration
	fmt.Println("UserService registered (mock implementation)")

	// Listen on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fmt.Println("=== gRPC Middleware Example ===")
	fmt.Println("Server starting on :50051")
	fmt.Println("\nThis example demonstrates gRPC interceptors with Langfuse tracing.")
	fmt.Println("In a real application, you would:")
	fmt.Println("1. Define your service in .proto files")
	fmt.Println("2. Generate Go code using protoc")
	fmt.Println("3. Register your service implementation")
	fmt.Println("4. Create gRPC clients to test the service")
	fmt.Println("\nExample service methods:")
	fmt.Println("- GetUser(UserRequest) returns (UserResponse)")
	fmt.Println("- CreateUser(CreateUserRequest) returns (CreateUserResponse)")
	fmt.Println("\nPress Ctrl+C to stop the server")

	// Start server
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Simulate some gRPC calls for demonstration
	go simulateGRPCCalls(userService, langfuseClient)

	// Keep server running
	select {}
}

// Simple logging interceptor for demonstration
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	
	resp, err := handler(ctx, req)
	
	status := "OK"
	if err != nil {
		status = "ERROR"
	}
	
	log.Printf("gRPC %s %s %v", info.FullMethod, status, time.Since(start))
	
	return resp, err
}

// Simulate gRPC calls for demonstration
func simulateGRPCCalls(service *UserService, langfuse *client.Langfuse) {
	time.Sleep(2 * time.Second) // Wait for server to start

	fmt.Println("\n=== Simulating gRPC Calls ===")

	// Create a context with Langfuse tracing enabled
	ctx := context.Background()

	// Simulate GetUser call
	fmt.Println("1. Simulating GetUser call...")
	trace1 := langfuse.Trace("grpc-get-user-simulation")
	ctx1 := middleware.ContextWithTrace(ctx, trace1)
	
	getUserReq := &UserRequest{ID: "user123"}
	if resp, err := service.GetUser(ctx1, getUserReq); err != nil {
		log.Printf("GetUser failed: %v", err)
	} else {
		fmt.Printf("   ✅ GetUser succeeded: %+v\n", resp)
	}
	trace1.End()

	time.Sleep(1 * time.Second)

	// Simulate CreateUser call
	fmt.Println("2. Simulating CreateUser call...")
	trace2 := langfuse.Trace("grpc-create-user-simulation")
	ctx2 := middleware.ContextWithTrace(ctx, trace2)
	
	createUserReq := &CreateUserRequest{
		Name:  "Alice Johnson", 
		Email: "alice@example.com",
	}
	if resp, err := service.CreateUser(ctx2, createUserReq); err != nil {
		log.Printf("CreateUser failed: %v", err)
	} else {
		fmt.Printf("   ✅ CreateUser succeeded: %+v\n", resp)
	}
	trace2.End()

	time.Sleep(1 * time.Second)

	// Simulate error case
	fmt.Println("3. Simulating error case...")
	trace3 := langfuse.Trace("grpc-error-simulation")
	ctx3 := middleware.ContextWithTrace(ctx, trace3)
	
	invalidReq := &UserRequest{ID: ""} // Empty ID should cause error
	if resp, err := service.GetUser(ctx3, invalidReq); err != nil {
		fmt.Printf("   ✅ Expected error occurred: %v\n", err)
	} else {
		fmt.Printf("   ❌ Unexpected success: %+v\n", resp)
	}
	trace3.End()

	fmt.Println("\n=== Simulation Complete ===")

	// Flush traces
	if err := langfuse.Flush(context.Background()); err != nil {
		log.Printf("Failed to flush traces: %v", err)
	}

	// Show statistics
	time.Sleep(1 * time.Second)
	stats := langfuse.GetStats()
	fmt.Printf("\nLangfuse Statistics:\n")
	fmt.Printf("- Traces Created: %d\n", stats.TracesCreated)
	fmt.Printf("- Generations Created: %d\n", stats.GenerationsCreated)
	fmt.Printf("- Events Submitted: %d\n", stats.EventsSubmitted)
}

// Helper function
func intPtr(v int) *int {
	return &v
}