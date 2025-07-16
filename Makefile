# Yamcha Load Testing Tool Makefile

.PHONY: build clean test run install help deps lint fmt vet

# Variables
BINARY_NAME=yamcha
BUILD_DIR=bin
CMD_DIR=cmd/yamcha
VERSION=2.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: clean deps test build

# Build the application
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "✅ Multi-platform build completed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf results/
	@echo "✅ Clean completed"

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...
	@echo "✅ Tests completed"

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run the application with default settings
run:
	@echo "🚀 Running $(BINARY_NAME) with default settings..."
	@go run ./$(CMD_DIR)

# Run with example config
run-config:
	@echo "🚀 Running $(BINARY_NAME) with example config..."
	@go run ./$(CMD_DIR) -config config.example.yaml

# Install the binary to GOPATH/bin
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Lint the code
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run
	@echo "✅ Linting completed"

# Format the code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Formatting completed"

# Vet the code
vet:
	@echo "🔬 Vetting code..."
	@go vet ./...
	@echo "✅ Vetting completed"

# Create a release
release: clean test build-all
	@echo "📦 Creating release..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@cp README.md release/
	@cp config.example.yaml release/
	@cp config.example.json release/
	@echo "✅ Release created in release/ directory"

# Run a quick load test
demo:
	@echo "🎯 Running demo load test..."
	@go run ./$(CMD_DIR) -url http://httpbin.org/get -req 50 -rate 10 -attack steady

# Setup development environment
dev-setup:
	@echo "🛠️  Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development environment ready"

# Show help
help:
	@echo "Yamcha Load Testing Tool - Make Commands"
	@echo ""
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download and update dependencies"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  run          - Run with default settings"
	@echo "  run-config   - Run with example config file"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  release      - Create a release build"
	@echo "  demo         - Run a demo load test"
	@echo "  dev-setup    - Setup development environment"
	@echo "  help         - Show this help message"
