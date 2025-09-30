# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Build and Run:**
```bash
go mod tidy              # Download and organize dependencies
go run cmd/main.go       # Run the application
go build -o app cmd/main.go  # Build binary
```

**Testing:**
```bash
go test ./...                    # Run all tests
go test ./internal/infrastructure/cache/  # Run specific package tests
go test -v ./tests/integration/  # Run integration tests with verbose output
go test -v ./tests/e2e/         # Run end-to-end tests
```

**Code Quality:**
```bash
go vet ./...     # Static analysis
go fmt ./...     # Format code
```

## Architecture Overview

This is a **Clean Architecture** Go application following Domain-Driven Design principles with clear separation of concerns across layers.

### Core Architecture Layers

**Domain Layer** (`internal/domain/`):
- `entity/` - Core business entities (User, Customer, Order, Product)
- `interfaces/` - Repository and service contracts
- `valueobjects/` - Domain value objects (Money, etc.)
- `dto/` - Data transfer objects for requests/responses
- `errors/` - Domain-specific error types
- `constants/` - Business constants (roles, status, errors)

**Application Layer** (`internal/services/`):
- Business logic implementation
- Orchestrates domain entities and repositories
- Implements service interfaces from domain layer

**Infrastructure Layer** (`internal/infrastructure/`):
- External service integrations (AWS, Azure, Stripe)
- Messaging systems (Kafka, RabbitMQ)
- Caching (Redis, Memory)
- HTTP clients and middleware
- Metrics and monitoring (Prometheus)

**Delivery Layer** (`internal/delivery/`):
- `handlers/` - HTTP request handlers
- `listeners/` - Event listeners (Discord, Kafka, RX)
- `router/` - HTTP routing configuration

**Repository Layer** (`internal/repositories/`):
- Data persistence implementations
- Multiple database support (Postgres, MongoDB, MySQL, Elasticsearch)
- In-memory implementations for testing

### Key Patterns

**Dependency Injection:** Uses `go.uber.org/dig` for dependency container management.

**Configuration:** Environment-based config using `github.com/kelseyhightower/envconfig` with singleton pattern in `config/config.go`.

**Worker System:** Built-in async job processing with:
- `internal/workers/dispatcher.go` - Job distribution and worker coordination
- `internal/workers/pool.go` - Worker pool management
- `internal/workers/jobs/` - Job implementations

**Repository Pattern:** All data access goes through repository interfaces defined in `internal/domain/interfaces/repositories.go`.

**Service Pattern:** Business logic encapsulated in services implementing interfaces from `internal/domain/interfaces/services.go`.

### Testing Structure

- `tests/integration/` - Integration tests for API endpoints and database interactions
- `tests/e2e/` - End-to-end test flows
- `tests/testdata/mocks/` - Test mocks and fixtures
- Unit tests alongside implementation files (e.g., `cache/memory_test.go`)

### Entry Point

Application starts in `cmd/main.go` which calls `internal/app/start.go`. The start function is currently empty and needs implementation for dependency injection and service initialization.

## Development Notes

- Repository implementations should satisfy interfaces in `internal/domain/interfaces/`
- New entities go in `internal/domain/entity/` with behavior methods
- Business logic belongs in `internal/services/`, not in handlers or repositories
- Use the existing worker system for async operations
- Configuration follows environment variable pattern with defaults
- All external dependencies are abstracted through interfaces for testability