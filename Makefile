.PHONY: help run dev test build clean docker-up docker-down install

# Default target
help:
	@echo "Pulseguard - Makefile Commands"
	@echo ""
	@echo "Development:"
	@echo "  make run         - Run the unified Pulseguard application (API + Worker)"
	@echo "  make dev         - Start Docker services and run the application"
	@echo ""
	@echo "Database & Redis:"
	@echo "  make docker-up   - Start Postgres and Redis containers"
	@echo "  make docker-down - Stop and remove containers"
	@echo ""
	@echo "Build & Test:"
	@echo "  make install     - Install dependencies"
	@echo "  make test        - Run all tests"
	@echo "  make build       - Build the Pulseguard binary"
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

# Run the unified Pulseguard application
run:
	@echo "Starting Pulseguard (API + Worker + Bootstrap)..."
	go run ./cmd/api

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

# Build the unified binary
build:
	@echo "Building Pulseguard binary..."
	@mkdir -p bin
	go build -o bin/pulseguard ./cmd/api
	@echo "✓ Binary built: ./bin/pulseguard"

# Clean built binaries
clean:
	@echo "Cleaning..."
	rm -rf bin/
	@echo "✓ Cleaned"

# Development mode - start everything
dev: docker-up
	@echo ""
	@echo "✓ Docker services started"
	@echo ""
	@echo "Starting Pulseguard application..."
	@echo ""
	@$(MAKE) run
