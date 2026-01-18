# GoAutomation Hub - REST API Automation Platform

A powerful workflow automation platform built with Go, featuring a flexible task execution engine and REST API for orchestrating automated workflows.

## Features

- 🏗️ Standard Go Project Layout (cmd, internal, pkg)
- 🚀 Gin-Gonic HTTP framework for high-performance API
- 🐘 PostgreSQL database with JSONB support
- 🐳 Docker Compose for easy deployment
- ✅ Comprehensive test coverage (>80% on core components)
- 📊 Structured logging with Go's native slog package

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (handled by Docker Compose)

## Quick Start with Docker Compose

1. **Clone the repository**
   ```bash
   git clone https://github.com/davioliveira/rest_api_automation_hub_go.git
   cd rest_api_automation_hub_go
   ```

2. **Configure environment variables**

   The `.env` file is already set up with default values. Update as needed:
   ```bash
   # API Server Configuration
   PORT=8080

   # Database Configuration
   DB_HOST=postgres
   DB_PORT=5432
   DB_USER=automation_hub
   DB_PASSWORD=changeme_secure_password
   DB_NAME=automation_hub_db

   # Logging Configuration
   LOG_LEVEL=info
   ```

3. **Start the application**
   ```bash
   docker-compose up --build
   ```

4. **Verify the health endpoint**
   ```bash
   curl http://localhost:8080/health
   ```

   Expected response:
   ```json
   {"status":"healthy"}
   ```

## Development Setup

### Running Locally (without Docker)

1. **Initialize dependencies**
   ```bash
   go mod download
   ```

2. **Run the application**
   ```bash
   export PORT=8080
   go run cmd/api/main.go
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Generate detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Building the Binary

```bash
go build -o bin/api ./cmd/api
./bin/api
```

## Project Structure

```
.
├── cmd/
│   └── api/
│       ├── main.go           # Application entry point
│       └── main_test.go      # Unit tests for main package
├── internal/                 # Private application code
├── pkg/                      # Public library code
├── docs/                     # Documentation
├── .env                      # Environment configuration
├── Dockerfile                # Multi-stage Docker build
├── docker-compose.yml        # Docker Compose orchestration
├── go.mod                    # Go module dependencies
└── README.md                 # This file
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `postgres` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `automation_hub` |
| `DB_PASSWORD` | Database password | `changeme_secure_password` |
| `DB_NAME` | Database name | `automation_hub_db` |
| `LOG_LEVEL` | Logging level (info, debug, error) | `info` |

## API Endpoints

### Health Check

**GET** `/health`

Returns the health status of the API server.

**Response:**
```json
{
  "status": "healthy"
}
```

## Docker Configuration

### Dockerfile

The project uses a multi-stage Docker build:
- **Builder stage**: Compiles the Go application
- **Runtime stage**: Minimal Alpine Linux image with the compiled binary

### Docker Compose Services

- **app**: The main API application
- **postgres**: PostgreSQL 15 database with persistent volume

## Testing Standards

- Minimum 80% test coverage for core components
- Go's native testing framework
- Tests co-located with source files
- Mock external services (database, HTTP clients)

## License

MIT License

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting pull requests.
