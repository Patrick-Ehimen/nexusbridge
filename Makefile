# NexusBridge Development Makefile

.PHONY: help build test clean docker-up docker-down install-deps compile-contracts deploy-local

# Default target
help:
	@echo "NexusBridge Development Commands:"
	@echo "  install-deps     - Install all dependencies"
	@echo "  build           - Build Go binaries"
	@echo "  test            - Run all tests"
	@echo "  docker-up       - Start development environment"
	@echo "  docker-down     - Stop development environment"
	@echo "  compile-contracts - Compile smart contracts"
	@echo "  deploy-local    - Deploy contracts to local network"
	@echo "  clean           - Clean build artifacts"

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing contract dependencies..."
	cd contracts && npm install

# Build Go binaries
build:
	@echo "Building relayer service..."
	go build -o bin/relayer ./cmd/relayer
	@echo "Building API service..."
	go build -o bin/api ./cmd/api

# Run tests
test:
	@echo "Running Go tests..."
	go test -v ./...
	@echo "Running contract tests..."
	cd contracts && npm test

# Start development environment
docker-up:
	@echo "Starting development environment..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	sleep 10
	@echo "Development environment is ready!"
	@echo "- Hardhat node: http://localhost:8545"
	@echo "- PostgreSQL: localhost:5432"
	@echo "- Prometheus: http://localhost:9090"
	@echo "- Grafana: http://localhost:3000 (admin/admin)"

# Stop development environment
docker-down:
	@echo "Stopping development environment..."
	docker-compose down

# Compile smart contracts
compile-contracts:
	@echo "Compiling smart contracts..."
	cd contracts && npm run compile

# Deploy contracts to local network
deploy-local: docker-up compile-contracts
	@echo "Deploying contracts to local network..."
	cd contracts && npm run deploy:local

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf contracts/artifacts/
	rm -rf contracts/cache/
	go clean

# Run relayer service
run-relayer: build
	@echo "Starting relayer service..."
	./bin/relayer

# Run API service
run-api: build
	@echo "Starting API service..."
	./bin/api

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting contract code..."
	cd contracts && npx prettier --write "**/*.{js,sol}"

# Lint code
lint:
	@echo "Linting Go code..."
	golangci-lint run
	@echo "Linting contract code..."
	cd contracts && npx solhint "contracts/**/*.sol"