# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a MikroTik PPPoE management API built with Go, following hexagonal (ports and adapters) architecture. The system provides REST APIs to manage PPPoE users, profiles, sessions, and queues on MikroTik routers through the RouterOS API.

## Development Commands

### Running the Application

Development mode (without Docker):
```sh
go run cmd/main.go http
```

Docker mode:
```sh
# Start dependencies
docker-compose up -d

# Run HTTP server
make http

# Force rebuild before running
make http BUILD=true
```

### Testing

Run unit tests:
```sh
go test -v -race ./internal/domain/... ./internal/adapter/...
```

Run integration tests:
```sh
go test -v -tags=integration ./tests/integration/...
# Or using Makefile
make test-integration
```

Run tests with coverage:
```sh
go test -coverprofile=coverage.profile -cover ./internal/domain/...
go tool cover -html coverage.profile -o coverage.html
# Or using Makefile
make test-coverage
```

Run single test:
```sh
go test -v -run TestFunctionName ./path/to/package
```

Check for intermittent test failures:
```sh
retry -d 0 -t 100 -u fail -- go test -coverprofile=coverage.profile -cover ./internal/domain/... -count=1
```

### Linting

```sh
make lint
```

### Code Generation

Generate mocks for testing:
```sh
make generate-mocks
```

### Seeding

Run seeds:
```sh
make seed                # Uses SEED_ENV from environment (default: development)
make seed-dev           # Development seeds
make seed-test          # Testing seeds
make seed-prod          # Production seeds
make seed-clean         # Clean seed data
make seed-refresh       # Clean and reseed
```

Control seeding with environment variables:
```sh
SEED_ENV=development SEED_ENTITIES=all go run cmd/seed/main.go
```

### Interactive Makefile Runner

```sh
make run
# Opens interactive menu to select Makefile targets
# Works best with fzf installed, but has fallback menu
```

## Architecture

### Hexagonal Architecture (Ports and Adapters)

The codebase follows strict hexagonal architecture with clear separation:

```
Domain (Core Business Logic)
    ↓ depends on
Ports (Interfaces)
    ↑ implemented by
Adapters (Infrastructure)
```

**Key principle**: Domain logic depends only on port interfaces, never on concrete adapters. This allows swapping infrastructure without touching business logic.

### Directory Structure

- `internal/domain/` - Core business logic (auth, client, pppoe, queue, user)
  - Each domain has its own interface and implementation
  - Domains use outbound ports to interact with external systems
  - Domains are instantiated with their dependencies via registry pattern

- `internal/port/` - Interface definitions
  - `inbound/` - How external systems interact with the app (HTTP, message, command, workflow)
  - `outbound/` - How the app interacts with external systems (database, cache, message, HTTP, MikroTik, workflow)
  - Each port category has a registry file grouping related interfaces

- `internal/adapter/` - Implementations of ports
  - `inbound/` - gin (HTTP), rabbitmq (message consumer), command (CLI), temporal (workflow)
  - `outbound/` - postgres (database), redis (cache), rabbitmq (message producer), http (external APIs), mikrotik (RouterOS API), temporal (workflow)
  - Each adapter implements the corresponding port interface

- `internal/model/` - Data structures/entities shared across layers

- `internal/migration/postgres/` - Database migrations using goose

- `tests/fixtures/` - Test data factories for all entities
  - Centralized test data generation
  - Use `fixtures.NewTestDataFactory()` in tests
  - See `tests/fixtures/README.md` for complete usage guide

- `utils/` - Shared utilities (hash, token, logging, database, etc.)

### Domain Registry Pattern

The domain registry (`internal/domain/registry.go`) wires all domain components together:

```go
type Domain interface {
    Client() client.ClientDomain
    Auth() auth.AuthDomain
    User() user.UserDomain
    Pppoe() pppoe.PppoeDomain
    Queue() queue.QueueDomain
}
```

Each domain method instantiates the domain with its required ports (database, cache, message, etc.).

### MikroTik Integration

The application manages MikroTik routers stored in the database. Each operation requires passing a `model.MikrotikRouter` instance to the `MikrotikPort` adapter methods. The `MikrotikPort` interface provides methods for:

- PPPoE Secret management (CRUD)
- PPPoE Profile management (CRUD)
- Active PPPoE session monitoring and termination
- Simple Queue management (CRUD)
- Queue statistics streaming via WebSocket

### Authentication & Authorization

Uses Casbin for role-based access control (RBAC):
- Enforcer is initialized in `postgres_outbound_adapter.InitCasbin()`
- Auth domain handles JWT token validation and policy enforcement
- Policies stored in database via gorm-adapter

### Test Data Strategy

Always use test fixtures instead of hardcoding test data:

```go
import "go-template/tests/fixtures"

func TestSomething(t *testing.T) {
    factory := fixtures.NewTestDataFactory()

    user := factory.User.ValidUser()
    router := factory.Mikrotik.ValidMikrotikRouter()
    secret := factory.Pppoe.ValidPppoeSecret()
    queue := factory.Queue.ValidPppoeQueue()

    // Use in tests...
}
```

Fixtures provide methods for:
- Valid default instances
- Multiple instances with varied data
- Customized instances (e.g., `UserWithRole("admin")`, `PppoeQueueWithPriority(1)`)

## Scaffolding New Components

The project provides Makefile targets for generating new components:

### Models
```sh
make model VAL=product
# Creates internal/model/product.go with struct, input, filter, and helper methods
```

### Domains
```sh
make domain VAL=product
# Creates internal/domain/product/ directory with domain.go
# Updates internal/domain/registry.go to wire the new domain
```

### Inbound Adapters
```sh
make inbound-http-gin VAL=product
# Creates port interface and gin HTTP adapter
# Updates registries

make inbound-message-rabbitmq VAL=product
# Creates port interface and RabbitMQ consumer adapter

make inbound-command VAL=product
# Creates port interface and CLI command adapter

make inbound-workflow-temporal VAL=product
# Creates port interface and Temporal workflow adapter
```

### Outbound Adapters
```sh
make outbound-database-postgres VAL=product
# Creates port interface and PostgreSQL adapter
# Generates mocks automatically

make outbound-http VAL=product
# Creates port interface and HTTP client adapter

make outbound-message-rabbitmq VAL=product
# Creates port interface and RabbitMQ publisher adapter

make outbound-cache-redis VAL=product
# Creates port interface and Redis cache adapter

make outbound-workflow-temporal VAL=product
# Creates port interface and Temporal workflow starter adapter
```

### Migrations
```sh
make migration-postgres VAL=create_products
# Creates numbered migration file with up/down functions
```

**Note**: All generated code uses snake_case for names with underscores (e.g., `product_category`) and properly converts to PascalCase/camelCase.

## Environment Configuration

Copy `.env.example` to `.env` before running:
```sh
cp .env.example .env
```

Key environment variables:
- `APP_MODE` - "release" or "debug"
- `SERVER_PORT` - HTTP server port (default: 8000)
- `OUTBOUND_DATABASE_DRIVER` - Database driver (postgres)
- `OUTBOUND_CACHE_DRIVER` - Cache driver (redis)
- `OUTBOUND_MESSAGE_DRIVER` - Message queue driver (rabbitmq)
- `INBOUND_HTTP_DRIVER` - HTTP framework (gin)
- Database, Redis, RabbitMQ connection settings

## Testing Guidelines

1. **Unit tests**: Test domain logic in isolation using mocks of outbound ports
   - Place in `internal/domain/*/domain_test.go` or `internal/adapter/*/adapter_test.go`
   - Use `make generate-mocks` to generate mocks from port interfaces
   - Mock ports are in `tests/mocks/port/`

2. **Integration tests**: Test full stack with real dependencies (using testcontainers)
   - Place in `tests/integration/`
   - Use `fixtures.NewTestDataFactory()` for test data
   - Tag with `//go:build integration`

3. **Test data**: Always use fixtures package for test data
   - Ensures consistency across tests
   - Easy to maintain and update
   - See `tests/fixtures/README.md` for complete guide

## Code Organization Rules

1. **Domain purity**: Domain logic must not import adapter packages or framework-specific code
2. **Dependency direction**: Domain → Ports ← Adapters (dependencies point inward)
3. **Port interfaces**: Define clear contracts between domain and infrastructure
4. **Registry pattern**: Use registries to group and instantiate related components
5. **No direct database access in domain**: Domain uses DatabasePort interface methods
6. **Adapter independence**: Each adapter should be swappable without affecting domain

## Important Patterns

### Error Handling
Use `github.com/palantir/stacktrace` for wrapping errors with context:
```go
return stacktrace.Propagate(err, "failed to create user")
```

### Logging
Use the project's logging utility (`go-template/utils/log`):
```go
log.WithContext(ctx).Info("message")
log.WithContext(ctx).Error("message", err)
```

### Activity Context
The application uses activity context for tracking requests:
```go
ctx := activity.NewContext("operation-name")
ctx = activity.WithClientID(ctx, "client-id")
```

### Mock Generation
Mocks are generated using `mockgen` from `go:generate` directives in registry files:
```go
//go:generate mockgen -source=registry_database.go -destination=./../../../tests/mocks/port/mock_database_port.go
```

## Module Name

The Go module is named `go-template`. When generating new code or importing packages, use this module name as the base import path.

## Common Development Tasks

When adding a new feature (e.g., "product management"):

1. Create model: `make model VAL=product`
2. Create migration: `make migration-postgres VAL=create_products_table`
3. Create domain: `make domain VAL=product`
4. Create database adapter: `make outbound-database-postgres VAL=product`
5. Create HTTP adapter: `make inbound-http-gin VAL=product`
6. Implement domain logic in `internal/domain/product/`
7. Implement adapter methods in `internal/adapter/outbound/postgres/product.go`
8. Implement HTTP handlers in `internal/adapter/inbound/gin/product.go`
9. Create test fixtures in `tests/fixtures/product.go`
10. Write tests using fixtures
