# Makefile for Eino Listing Agent

# Variables
APP_NAME=eino-listing-agent
CMD_PATH=./cmd/agent
BUILD_DIR=./bin
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

.PHONY: all build clean test coverage deps fmt vet start run dev help

# Default target
all: deps fmt vet test build

# Quick start the service
start: deps
	@echo "🚀 Starting $(APP_NAME)..."
	$(GOCMD) run $(CMD_PATH)

# Alternative alias for start
run: start

# Development mode with auto-restart (requires air to be installed)
dev:
	@if command -v air >/dev/null 2>&1; then \
		echo "🔄 Starting in development mode with hot reload..."; \
		air; \
	else \
		echo "📦 Installing air for hot reload..."; \
		$(GOGET) -u github.com/air-verse/air@latest; \
		echo "🔄 Starting in development mode with hot reload..."; \
		air; \
	fi

# Build the application
build: deps
	@echo "🔨 Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "✅ Build completed: $(BUILD_DIR)/$(APP_NAME)"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "🔧 Formatting code..."
	$(GOFMT) ./...

# Vet code
vet:
	@echo "🧹 Running go vet..."
	$(GOVET) ./...

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v -race ./...

# Run tests with coverage
coverage:
	@echo "🧪 Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "📊 Coverage report saved to $(COVERAGE_HTML)"

# Run full test suite (equivalent to scripts/test.sh)
test-full: deps fmt vet coverage
	@echo "🎉 Full test suite completed!"

# Clean build artifacts
clean:
	@echo "🧽 Cleaning..."
	$(GOCLEAN)
	@rm -f $(BUILD_DIR)/$(APP_NAME)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "✅ Clean completed"

# Install the application globally
install: build
	@echo "📥 Installing $(APP_NAME) globally..."
	$(GOCMD) install $(CMD_PATH)/main.go

# Run benchmarks
bench:
	@echo "🏃 Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Check for security issues (requires gosec)
security:
	@if command -v gosec >/dev/null 2>&1; then \
		echo "🔒 Running security scan..."; \
		gosec ./...; \
	else \
		echo "📦 Installing gosec..."; \
		$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		echo "🔒 Running security scan..."; \
		gosec ./...; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  start      - Quick start the service (default: go run)"
	@echo "  run        - Alias for start"
	@echo "  dev        - Start with hot reload (installs air if needed)"
	@echo "  build      - Build the application"
	@echo "  test       - Run unit tests"
	@echo "  test-full  - Run complete test suite with coverage"
	@echo "  coverage   - Run tests with coverage report"
	@echo "  bench      - Run benchmarks"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  deps       - Install dependencies"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install application globally"
	@echo "  security   - Run security scan (installs gosec if needed)"
	@echo "  help       - Show this help message"