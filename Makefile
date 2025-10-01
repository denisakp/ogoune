.PHONY: help dev api worker test build clean docker-up docker-down install

# Default target
help:
	@echo "Pulseguard - Makefile Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev         - Start all services (Postgres, Redis, API, Worker)"
	@echo "  make api         - Run API server only"
	@echo "  make worker      - Run worker only"
	@echo ""
	@echo "Database & Redis:"
	@echo "  make docker-up   - Start Postgres and Redis containers"
	@echo "  make docker-down - Stop and remove containers"
	@echo ""
	@echo "Build & Test:"
	@echo "  make install     - Install dependencies"
	@echo "  make test        - Run all tests"
	@echo "  make build       - Build binaries"
	@echo "  make clean       - Remove built binaries"

# Install dependencies
install:
	@echo "Installing dependencies..."
	go get github.com/joho/godotenv
	go mod download
	go mod tidy
	@echo "✓ Dependencies installed"

# Start Docker services
docker-up:
	@echo "Starting Postgres and Redis..."
	@docker run --name pulse-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=pulse -p 5432:5432 -d postgres:15 2>/dev/null || docker start pulse-postgres
	@docker run --name pulse-redis -p 6379:6379 -d redis:7 2>/dev/null || docker start pulse-redis
	@echo "✓ Postgres running on :5432"
	@echo "✓ Redis running on :6379"

# Stop Docker services
docker-down:
	@echo "Stopping containers..."
	@docker stop pulse-postgres pulse-redis 2>/dev/null || true
	@echo "✓ Containers stopped"

# Run API server
api:
	@echo "Starting API server..."
	go run ./cmd/api

# Run worker
worker:
	@echo "Starting worker..."
	go run ./cmd/worker

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

# Build binaries
build:
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o bin/pulseguard-api ./cmd/api
	go build -o bin/pulseguard-worker ./cmd/worker
	@echo "✓ Binaries built in ./bin/"

# Clean built binaries
clean:
	@echo "Cleaning..."
	rm -rf bin/
	@echo "✓ Cleaned"

# Development mode - start everything
dev: docker-up
	@echo ""
	@echo "Services started. Run in separate terminals:"
	@echo "  Terminal 1: make api"
	@echo "  Terminal 2: make worker"
	@echo ""
	@echo "Or use your IDE to run them separately"
