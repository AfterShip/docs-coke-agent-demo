# HTTP Client Refactoring Implementation Tasks

This document outlines the implementation tasks for refactoring the HTTP client in pkg/langfuse to directly use resty instead of maintaining wrapper layers.

## Phase 1: Core Domain Layer - Configuration and Error Handling

- [x] 1. Create resty client configuration foundation
  - [x] 1.1 Create pkg/langfuse/api/core/client_config.go with ConfigureRestyClient function
    - Implement basic configuration (base URL, timeout, headers)
    - Add authentication setup (basic auth)
    - Implement retry configuration using resty's built-in mechanisms
    - Add debug mode configuration
  - [x] 1.2 Create pkg/langfuse/api/core/middleware.go with helper functions
    - Implement createRetryCondition function for custom retry logic
    - Implement createErrorHandler function for consistent error processing
    - Add parseHTTPError function for error classification
  - [x] 1.3 Write comprehensive unit tests for client_config.go
    - Test ConfigureRestyClient with various config combinations
    - Test authentication configuration
    - Test retry configuration
    - Test debug mode settings
  - [x] 1.4 Write unit tests for middleware.go
    - Test createRetryCondition with different error scenarios
    - Test createErrorHandler with various HTTP status codes
    - Test parseHTTPError function error mapping
  - [x] 1.5 Run compilation/build for all code affected to ensure no compilation issues
  - [x] 1.6 Run git commit to submit the changes
  - [x] 1.7 Stop and wait for human approval before continuing (mark this sub-task as completed, then stop)

## Phase 2: Application Layer - API Client Refactoring

- [x] 2. Refactor APIClient to use resty directly
  - [x] 2.1 Backup current pkg/langfuse/api/client.go implementation
    - Create backup copy of existing APIClient
    - Document current interface for compatibility reference
  - [x] 2.2 Modify APIClient struct to use *resty.Client instead of *HTTPClient
    - Update APIClient struct fields
    - Update NewAPIClient constructor to use resty.New() and ConfigureRestyClient
    - Maintain backward compatibility for existing public methods
  - [x] 2.3 Update APIClient methods to work with direct resty usage
    - Update health check functionality
    - Update timeout and configuration methods
    - Preserve existing public API surface
  - [x] 2.4 Write unit tests for refactored APIClient
    - Test NewAPIClient with various configurations
    - Test APIClient methods maintain same behavior
    - Test error handling and health checks
  - [x] 2.5 Write integration tests for APIClient
    - Test complete APIClient initialization flow
    - Test configuration application
    - Test resource client integration
  - [x] 2.6 Run compilation/build for all code affected to ensure no compilation issues
  - [x] 2.7 Run git commit to submit the changes
  - [x] 2.8 Stop and wait for human approval before continuing (mark this sub-task as completed, then stop)

## Phase 3: Infrastructure Layer - Resource Client Updates

- [ ] 3. Update all resource clients to use *resty.Client
  - [x] 3.1 Update ingestion client (pkg/langfuse/api/resources/ingestion/client.go)
    - Change Client struct to use *resty.Client field
    - Update NewClient constructor to accept *resty.Client
    - Refactor Submit, SubmitBatch, and other methods to use resty directly
    - Remove dependencies on HTTPClient and RequestOptions
  - [x] 3.2 Update health client (pkg/langfuse/api/resources/health/client.go)
    - Change Client struct to use *resty.Client field
    - Update NewClient constructor to accept *resty.Client
    - Refactor health check methods to use resty directly
  - [x] 3.3 Update traces client (pkg/langfuse/api/resources/traces/client.go)
    - Change Client struct to use *resty.Client field
    - Update NewClient constructor to accept *resty.Client
    - Refactor all HTTP methods to use resty directly
  - [x] 3.4 Update scores client (pkg/langfuse/api/resources/scores/client.go)
    - Change Client struct to use *resty.Client field
    - Update NewClient constructor to accept *resty.Client
    - Refactor all HTTP methods to use resty directly
  - [x] 3.5 Update sessions client (pkg/langfuse/api/resources/sessions/client.go)
    - Change Client struct to use *resty.Client field
    - Update NewClient constructor to accept *resty.Client
    - Refactor all HTTP methods to use resty directly
  - [x] 3.6 Update any remaining resource clients following the same pattern
    - Identify other resource clients in pkg/langfuse/api/resources/
    - Apply the same refactoring pattern to each client
  - [x] 3.7 Write unit tests for all updated resource clients
    - Test each client's methods with mock resty responses
    - Test error handling in each client
    - Test context propagation and request configuration
  - [x] 3.8 Run compilation/build for all code affected to ensure no compilation issues
  - [x] 3.9 Run git commit to submit the changes
  - [x] 3.10 Stop and wait for human approval before continuing (mark this sub-task as completed, then stop)

## Phase 4: Infrastructure Layer - Cleanup and Migration

- [x] 4. Remove deprecated wrapper layers and finalize migration
  - [x] 4.1 Remove HTTPClient wrapper (pkg/langfuse/api/core/http_client.go)
    - Delete http_client.go file
    - Remove related test files
    - Update any remaining imports
  - [x] 4.2 Remove ClientWrapper (pkg/langfuse/api/core/client_wrapper.go)
    - Delete client_wrapper.go file
    - Remove related test files
    - Update any remaining imports
  - [x] 4.3 Remove RequestOptions and related abstractions
    - Delete request_options.go file if it exists
    - Remove any other wrapper abstractions
    - Clean up imports across the codebase
  - [x] 4.4 Update existing tests to work with new resty-based implementation
    - Fix any failing tests due to interface changes
    - Update test mocks to work with resty
    - Ensure all existing functionality is preserved
  - [x] 4.5 Write integration tests for the complete refactored system
    - Test end-to-end API workflows
    - Test retry behavior with actual HTTP scenarios
    - Test error handling across different failure modes
  - [x] 4.6 Add backward compatibility tests
    - Ensure existing public APIs work identically
    - Test configuration compatibility
    - Verify feature parity with previous implementation
  - [x] 4.7 Run all existing tests to ensure no regressions
    - Run full test suite for pkg/langfuse
    - Fix any failing tests
    - Ensure test coverage is maintained
  - [x] 4.8 Run compilation/build for all code affected to ensure no compilation issues
  - [x] 4.9 Run git commit to submit the changes
  - [x] 4.10 Stop and wait for human approval before continuing (mark this sub-task as completed, then stop)

## Implementation Notes

Each phase builds incrementally on the previous phases:
- Phase 1 establishes the foundation with configuration and middleware helpers
- Phase 2 updates the core APIClient to use resty directly
- Phase 3 updates all resource clients to use the new pattern
- Phase 4 removes old abstractions and ensures system integrity

All tasks focus on code implementation and testing that can be executed by a coding agent. The implementation maintains backward compatibility while simplifying the internal architecture.