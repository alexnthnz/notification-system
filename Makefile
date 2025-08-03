# Notification Service Makefile

.PHONY: help build test clean proto run docker-up docker-down lint

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build commands
build: ## Build all services
	@echo "Building all services..."
	@go build -o bin/api ./cmd/api
	@go build -o bin/email-service ./cmd/email-service
	@go build -o bin/sms-service ./cmd/sms-service
	@go build -o bin/push-service ./cmd/push-service
	@echo "✅ Build completed"

build-docker: ## Build Docker images
	@echo "Building Docker images..."
	@docker-compose build
	@echo "✅ Docker build completed"

# Development commands
run-api: ## Run API service locally
	@echo "Starting API service..."
	@go run ./cmd/api

run-email: ## Run email service locally
	@echo "Starting email service..."
	@go run ./cmd/email-service

run-sms: ## Run SMS service locally
	@echo "Starting SMS service..."
	@go run ./cmd/sms-service

run-push: ## Run push service locally
	@echo "Starting push service..."
	@go run ./cmd/push-service

# Protobuf generation
proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@./scripts/generate-proto.sh
	@echo "✅ Protobuf generation completed"

# Testing and validation
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

lint: ## Run linter
	@echo "Running linter..."
	@go vet ./...
	@go fmt ./...

# Docker commands
docker-up: ## Start all services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "✅ Services started"
	@echo "REST API: http://localhost:8080"
	@echo "gRPC API: localhost:9090"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000"

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down
	@echo "✅ Services stopped"

docker-logs: ## Show Docker logs
	@docker-compose logs -f

# Database commands
db-migrate: ## Run database migrations (placeholder)
	@echo "Running database migrations..."
	@echo "⚠️  Database migrations not implemented yet"

# gRPC tools
grpc-test: ## Test gRPC API with example client
	@echo "Testing gRPC API..."
	@go run ./examples/grpc-client

grpc-health: ## Check gRPC server health using grpcurl
	@echo "Checking gRPC server health..."
	@grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check || echo "Install grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"

grpc-list: ## List gRPC services using grpcurl
	@echo "Listing gRPC services..."
	@grpcurl -plaintext localhost:9090 list || echo "Install grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"

# Development utilities
dev-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@echo "✅ Development dependencies installed"

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "✅ Cleanup completed"

# Installation and setup
install: dev-deps proto build ## Full installation (dependencies, proto, build)
	@echo "✅ Installation completed"

# Go module management
mod-tidy: ## Tidy Go modules
	@go mod tidy

mod-download: ## Download Go modules
	@go mod download

# Example usage
example-rest: ## Run REST API examples
	@echo "Testing REST API..."
	@echo "Creating email notification..."
	@curl -X POST http://localhost:8080/api/v1/notifications \
		-H "Content-Type: application/json" \
		-d '{"user_id":"123","channel":"email","recipient":"test@example.com","subject":"Test","body":"Hello World"}'

example-grpc: grpc-test ## Run gRPC examples (alias for grpc-test)