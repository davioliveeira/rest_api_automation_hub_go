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
   {
     "status": "healthy",
     "database": "connected"
   }
   ```

## Development Setup

### Running Locally (without Docker)

1. **Initialize dependencies**
   ```bash
   go mod download
   ```

2. **Set up environment variables**
   ```bash
   export PORT=8080
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=goautomation
   ```

3. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```
   
   **Note:** Make sure PostgreSQL is running locally or use Docker Compose to start the database.

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
├── cmd/
│   └── api/
│       ├── main.go           # Application entry point
│       ├── handlers.go       # REST API handlers
│       └── *_test.go         # Test files
├── internal/                 # Private application code
│   ├── engine/              # Workflow execution engine
│   ├── repository/          # Database layer (GORM)
│   │   ├── db.go            # Database connection
│   │   ├── models.go        # Database models
│   │   ├── workflow_repository.go  # Repository implementation
│   │   └── converter.go     # Model converters
│   └── tasks/               # Task executors
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

Returns the health status of the API server and database connection.

**Response:**
```json
{
  "status": "healthy",
  "database": "connected"
}
```

### Workflow Management

#### Create Workflow

**POST** `/workflows`

Creates a new workflow definition.

**Request Body:**
```json
{
  "name": "my-workflow",
  "definition": {
    "name": "my-workflow",
    "tasks": [
      {
        "id": "fetch",
        "type": "http_request",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/data"
        }
      }
    ]
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-workflow",
  "definition": {...},
  "created_at": "2026-01-17T22:00:00Z",
  "updated_at": "2026-01-17T22:00:00Z"
}
```

#### Get Workflow by ID

**GET** `/workflows/:id`

Retrieves a workflow by its UUID.

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-workflow",
  "definition": {...},
  "created_at": "2026-01-17T22:00:00Z",
  "updated_at": "2026-01-17T22:00:00Z"
}
```

#### List All Workflows

**GET** `/workflows`

Retrieves all workflows.

**Response:** `200 OK`
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-workflow",
    "definition": {...},
    "created_at": "2026-01-17T22:00:00Z",
    "updated_at": "2026-01-17T22:00:00Z"
  }
]
```

#### Update Workflow

**PUT** `/workflows/:id`

Updates an existing workflow.

**Request Body:** Same as POST `/workflows`

**Response:** `200 OK` (same structure as GET)

#### Delete Workflow

**DELETE** `/workflows/:id`

Deletes a workflow by its UUID.

**Response:** `204 No Content`

## Docker Configuration

### Dockerfile

The project uses a multi-stage Docker build:
- **Builder stage**: Compiles the Go application
- **Runtime stage**: Minimal Alpine Linux image with the compiled binary

### Docker Compose Services

- **app**: The main API application
- **postgres**: PostgreSQL 15 database with persistent volume

## Database

The application uses PostgreSQL with GORM as the ORM. Database migrations are automatically run on startup using GORM's AutoMigrate feature.

### Database Schema

The `workflows` table stores workflow definitions:

- `id` (UUID): Primary key
- `name` (VARCHAR): Unique workflow name
- `definition` (JSONB): Workflow definition JSON
- `created_at` (TIMESTAMP): Creation timestamp
- `updated_at` (TIMESTAMP): Last update timestamp

### Migration

Migrations are handled automatically by GORM's `AutoMigrate()` function, which creates/updates the database schema on application startup.

## Testing Standards

- Minimum 80% test coverage for core components
- Go's native testing framework
- Tests co-located with source files
- Mock external services (database, HTTP clients)
- Integration tests for API endpoints

## License

MIT License

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting pull requests.
