# NexusBridge

A cross-chain interoperability protocol enabling secure asset transfers between Ethereum, Polygon, and Cosmos-based chains.

## Architecture

NexusBridge consists of three main components:

- **Smart Contracts**: Handle asset custody and event emission on each supported chain
- **Relayer Service**: Go-based service that monitors events and executes cross-chain transactions
- **API Service**: REST API for user interaction and system monitoring

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make (optional, for convenience commands)

### Development Setup

1. **Clone and setup the project:**

   ```bash
   git clone <repository-url>
   cd nexus-bridge
   cp .env.example .env
   ```

2. **Install dependencies:**

   ```bash
   make install-deps
   # or manually:
   go mod download
   cd contracts && npm install
   ```

3. **Start development environment:**

   ```bash
   make docker-up
   ```

   This will start:

   - Hardhat local Ethereum node (port 8545)
   - PostgreSQL database (port 5432)
   - Redis cache (port 6379)
   - Prometheus metrics (port 9090)
   - Grafana dashboard (port 3000)

4. **Compile and deploy contracts:**

   ```bash
   make compile-contracts
   make deploy-local
   ```

5. **Build and run services:**
   ```bash
   make build
   make run-relayer  # In one terminal
   make run-api      # In another terminal
   ```

## Project Structure

```
nexus-bridge/
├── cmd/                    # Application entry points
│   ├── api/               # REST API service
│   └── relayer/           # Relayer service
├── contracts/             # Smart contracts
│   ├── contracts/         # Solidity source files
│   ├── scripts/           # Deployment scripts
│   └── test/              # Contract tests
├── docker/                # Docker configurations
│   ├── hardhat/          # Hardhat node container
│   ├── nginx/            # Nginx proxy configs
│   └── prometheus/       # Monitoring configs
├── internal/              # Private application code
│   ├── adapters/         # Chain adapters
│   ├── api/              # API handlers
│   ├── contracts/        # Contract bindings
│   ├── models/           # Data models
│   └── relayer/          # Relayer logic
├── pkg/                   # Public packages
│   ├── crypto/           # Cryptographic utilities
│   └── types/            # Common types
├── scripts/               # Database and utility scripts
└── test/                  # Integration tests
```

## Development Commands

```bash
# Environment management
make docker-up           # Start development environment
make docker-down         # Stop development environment

# Building and testing
make build              # Build Go binaries
make test               # Run all tests
make compile-contracts  # Compile smart contracts

# Code quality
make fmt                # Format code
make lint               # Lint code

# Deployment
make deploy-local       # Deploy to local network
```

## Configuration

Copy `.env.example` to `.env` and configure:

- Database connection strings
- Blockchain RPC URLs
- Private keys (development only)
- API ports and settings

## Monitoring

- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus Metrics**: http://localhost:9090
- **API Health**: http://localhost:8080/api/v1/health

## Testing

```bash
# Run all tests
make test

# Run specific test suites
go test ./internal/...
cd contracts && npm test
```

## Contributing

1. Follow Go and Solidity best practices
2. Write tests for new functionality
3. Update documentation as needed
4. Use conventional commit messages

## Security

- Never commit private keys or sensitive data
- Use environment variables for configuration
- Follow multi-signature patterns for production
- Regular security audits recommended

## License

[Add your license here]
