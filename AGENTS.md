# GoTemplate Project Instructions for AI Assistants

This file contains instructions for AI coding assistants working with the GoTemplate project.

## Project Information

- **Project Name**: GoTemplate
- **Author**: Moch Dieqy Dzulqaidar
- **License**: MIT License
- **Go Version**: >= go1.24.0

## Project Structure
GoTemplate uses a hexagonal architecture (ports and adapters) with the following structure:
- `cmd/`: Application entry point
  - `main.go`: Main application entrypoint
  - `seed/main.go`: Seed data command
- `internal/`: Internal implementations
  - `app.go`: Application initialization
  - `adapter/`: Adapters (inbound and outbound)
    - `inbound/`: Input adapters
      - `command/`: CLI command handlers
      - `gin/`: HTTP server using Gin framework
      - `rabbitmq/`: Message consumers for RabbitMQ
    - `outbound/`: Output adapters
      - `http/`: HTTP clients
      - `postgres/`: PostgreSQL database adapters
      - `rabbitmq/`: RabbitMQ message publishers
      - `redis/`: Redis cache adapters
  - `domain/`: Domain logic
  - `migration/`: Migration scripts
    - `postgres/`: PostgreSQL migrations
  - `model/`: Data models
  - `port/`: Port definitions
    - `inbound/`: Input port definitions
    - `outbound/`: Output port definitions
  - `seeds/`: Seed data management
    - `runner/`: Seed execution engine
      - `runner.go`: Seed runner with transaction support
      - `loader.go`: Environment-based seed loader
    - `production/`: Production seed data
      - `users.go`: Admin account
      - `clients.go`: Default API clients
      - `casbin.go`: RBAC rules
      - `routers.go`: Optional default router
    - `testing/`: Test seed data
      - `users.go`: Multiple test users
      - `clients.go`: Test API clients
      - `routers.go`: Test MikroTik routers
      - `casbin.go`: Test RBAC rules
- `design-docs/`: Design documentation
  - `images/`: Documentation images/diagrams
- `tests/`: Test files and mocks
  - `helpers/`: Test utilities
    - `postgres.go`: PostgreSQL container setup with auto-seeding
  - `mocks/`: Mock implementations for testing
- `utils/`: Utility functions
  - `activity/`: Activity tracking
  - `database/`: Database utilities
  - `google/`: Google service utilities
  - `log/`: Logging utilities
  - `rabbitmq/`: RabbitMQ utilities
  - `redis/`: Redis utilities

## Supporting Files
- `docker-compose.yml`: Docker Compose configuration
- `Dockerfile`: Docker image definition
- `go.mod` and `go.sum`: Go module definitions
- `LICENSE`: MIT license file
- `Makefile`: Build and development automation
- `README.md`: Project documentation

## Code Conventions
- Uses Go modules
- Follows Go naming standards (camelCase for private, PascalCase for public)
- Unit tests for all domain logic

## Technologies
- PostgreSQL for database
- RabbitMQ for message queue
- Redis for caching
- Gin for HTTP server

## Development Patterns
- Use Makefile for common operations (see README.md for details)
- Use Docker for external services
- Hexagonal architecture for clean separation of concerns

## Seed Data

### Overview
The project includes a comprehensive seed data system for initializing database with default data. Seeds are organized by environment (production, development, testing) and support selective seeding by entity type.

### Seed Structure

**Runner Engine** (`internal/seeds/runner/`):
- `runner.go`: Core seeder interface and execution logic with transaction support
- `loader.go`: Environment detection and configuration

**Production Seeds** (`internal/seeds/production/`):
- `users.go`: Default admin user (email: `admin@mikrotik.local`, password: `Admin@123`)
- `clients.go`: System API client
- `casbin.go`: RBAC authorization rules for admin and user roles
- `routers.go`: Optional router (controlled by env vars)

**Testing Seeds** (`internal/seeds/testing/`):
- `users.go`: 4 test users (admin, regular, inactive users)
- `clients.go`: 3 test API clients
- `routers.go`: 3 test MikroTik routers (active and inactive)
- `casbin.go`: Extended RBAC rules for testing

### Environment Variables

Configure seeding behavior with these environment variables:
- `SEED_ENV`: Environment to seed (production/development/testing) - defaults to development
- `SEED_ENTITIES`: Entities to seed (all/users/clients/routers/casbin) - defaults to all
- `ADMIN_EMAIL`: Admin email for production seed - defaults to `admin@mikrotik.local`
- `ADMIN_PASSWORD`: Admin password for production seed - defaults to `Admin@123`
- `SEED_ROUTERS`: Whether to seed routers (true/false) - defaults to false
- `SEED_ROUTER_NAME/ADDRESS/USERNAME/PASSWORD`: Router configuration when seeding

### Makefile Commands

```bash
# Run seeds with current environment (SEED_ENV)
make seed

# Run production seeds
make seed-prod

# Run development seeds
make seed-dev

# Run test seeds
make seed-test

# Clean all seed data
make seed-clean

# Clean and re-seed (refresh)
make seed-refresh

# Seed specific entities only
SEED_ENTITIES=users,clients make seed
```

### Testing Integration

Test seeds are automatically loaded when using Testcontainers:
- Integration tests using `tests/helpers/postgres.go` auto-load test data
- Seeds run in transaction for test isolation
- When using `TEST_DB_DSN` (external DB), seeds are NOT auto-loaded to avoid conflicts

### Creating New Seeders

1. Create a seeder in the appropriate environment directory:
```go
package production

import (
    "go-template/internal/seeds/runner"
    "gorm.io/gorm"
)

type MySeeder struct{}

func (s *MySeeder) Name() string {
    return "My Seeder Name"
}

func (s *MySeeder) Seed(db *gorm.DB) error {
    // Seed logic here
    return nil
}

func init() {
    runner.RegisterSeeder(&MySeeder{})
}
```

2. Seeder automatically registers via `init()` function
3. Use transaction-safe GORM operations
4. Check for existing data before inserting to support re-runs

### Key Features

- **Transaction Safety**: All seeds run in transactions, rolling back on errors
- **Idempotency**: Seeds can be run multiple times safely
- **Environment-Aware**: Different data per environment
- **Selective Seeding**: Seed specific entities only
- **Auto-Discovery**: Seeders register via `init()`, no manual registration needed
- **Logging**: Verbose output for debugging seed operations

## Code Preferences
- Avoid using global variables
- Use dependency injection
- Follow clean architecture principles
- Prioritize code readability
- Use interfaces for component interaction

## Important Notes
- Do not change the main project structure
- Always run unit tests after significant changes
- Document APIs and important functions
- Use Makefile targets for code generation

## Copyright Information
Copyright (c) 2025 Moch Dieqy Dzulqaidar
