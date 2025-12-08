# Architecture Documentation

This document describes the architectural patterns, principles, and guidelines for this codebase.

## Overview

This is an **API-first backend** following Clean Architecture principles. The primary interface is REST/gRPC APIs for synchronous operations. Event-driven patterns complement the API for asynchronous side-effects (notifications, analytics).

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Clients                                    │
│                    (Web, Mobile, External Services)                     │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Transport Layer                                 │
│              HTTP (Fiber) │ gRPC │ AMQP RPC │ NATS RPC                 │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Application Layer                               │
│                          (Use Cases)                                    │
│              Orchestrates domain logic, manages transactions            │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Domain Layer                                   │
│                          (Entities)                                     │
│           Core business logic, rules, invariants, domain events         │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       Infrastructure Layer                              │
│          Repositories │ External APIs │ Message Queues │ Cache         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Core Principles

### SOLID Principles

| Principle | Application in This Codebase |
|-----------|------------------------------|
| **Single Responsibility** | Each package has one reason to change. `usecase/` handles business logic, `repo/` handles persistence, `controller/` handles transport. |
| **Open/Closed** | Add new features by implementing interfaces, not modifying existing code. New transport? Implement the handler interface. |
| **Liskov Substitution** | All interface implementations are interchangeable. `PostgresRepo` and `MockRepo` both satisfy `Repository` interface. |
| **Interface Segregation** | Small, focused interfaces. Clients depend only on methods they use. |
| **Dependency Inversion** | High-level modules (use cases) depend on abstractions (interfaces), not concrete implementations. |

### DDD Concepts

| Concept | Location | Purpose |
|---------|----------|---------|
| **Entities** | `internal/entity/` | Domain objects with identity and business rules |
| **Value Objects** | `internal/entity/` | Immutable objects defined by attributes, not identity |
| **Aggregates** | `internal/entity/` | Cluster of entities treated as a single unit |
| **Repositories** | `internal/repo/` | Persistence abstraction for aggregates |
| **Use Cases** | `internal/usecase/` | Application services orchestrating domain logic |
| **Domain Events** | `internal/entity/events/` | Immutable facts representing what happened |

### Dependency Rule

**Dependencies point inward. Inner layers know nothing about outer layers.**

```
Controllers → UseCases → Entities
     ↓            ↓
  Interfaces   Interfaces
     ↓            ↓
Repositories  External Services
```

## Package Structure

```
.
├── cmd/app/                    # Application entry point
│   └── main.go
│
├── config/                     # Configuration management
│   └── config.go              # Env-based config structs
│
├── internal/                   # Private application code
│   ├── app/                   # Application bootstrap
│   │   ├── app.go            # Main initialization
│   │   └── migrate.go        # Database migrations
│   │
│   ├── entity/                # Domain layer
│   │   ├── translation.go    # Domain entities
│   │   └── events/           # Domain events (future)
│   │
│   ├── usecase/               # Application layer
│   │   ├── contracts.go      # Use case interfaces
│   │   └── translation/      # Use case implementations
│   │
│   ├── repo/                  # Repository layer
│   │   ├── contracts.go      # Repository interfaces
│   │   ├── persistent/       # Database implementations
│   │   └── webapi/           # External API adapters
│   │
│   └── controller/            # Transport layer
│       ├── http/             # REST API (Fiber)
│       │   ├── router.go     # Route definitions
│       │   ├── middleware/   # HTTP middleware
│       │   └── v1/           # API v1 handlers
│       ├── grpc/             # gRPC handlers
│       └── amqp/             # RabbitMQ handlers
│
├── pkg/                       # Shared infrastructure
│   ├── logger/               # Logging abstraction
│   ├── httpserver/           # HTTP server wrapper
│   ├── grpcserver/           # gRPC server wrapper
│   ├── postgres/             # Database connection
│   ├── rabbitmq/             # RabbitMQ client
│   └── nats/                 # NATS client
│
├── migrations/                # Database migrations
├── docs/                      # Documentation & Swagger
└── integration-test/          # Integration tests
```

## Layer Responsibilities

### Transport Layer (`internal/controller/`)

- Parse and validate HTTP/gRPC requests
- Convert requests to use case inputs
- Call use cases
- Convert use case outputs to responses
- Handle transport-specific errors (HTTP status codes)
- Authentication and authorization middleware

**Does NOT contain business logic.**

### Application Layer (`internal/usecase/`)

- Orchestrate domain operations
- Manage transactions
- Coordinate multiple repositories
- Enforce business workflows
- Publish domain events (future)

**Does NOT know about HTTP, gRPC, or any transport.**

### Domain Layer (`internal/entity/`)

- Define domain entities and value objects
- Implement business rules and invariants
- Raise domain events
- Define domain errors

**Has NO external dependencies.**

### Infrastructure Layer (`internal/repo/`, `pkg/`)

- Implement repository interfaces
- Handle database operations
- Integrate with external services
- Manage connections and pools

**Implements interfaces defined by inner layers.**

## Interface Patterns

### Repository Interfaces

Defined where they're used, in `internal/repo/contracts.go`:

```go
type TranslationRepo interface {
    Store(ctx context.Context, t entity.Translation) error
    GetHistory(ctx context.Context) ([]entity.Translation, error)
}

type TranslationWebAPI interface {
    Translate(original, source, destination string) (string, error)
}
```

### Use Case Interfaces

Defined in `internal/usecase/contracts.go`:

```go
type Translation interface {
    Translate(ctx context.Context, t entity.Translation) (entity.Translation, error)
    History(ctx context.Context) (entity.TranslationHistory, error)
}
```

### Accept Interfaces, Return Structs

```go
func New(repo repo.TranslationRepo, webAPI repo.TranslationWebAPI) *UseCase {
    return &UseCase{
        repo:   repo,
        webAPI: webAPI,
    }
}
```

## API Design

### REST Conventions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/resources` | List resources |
| GET | `/v1/resources/:id` | Get resource by ID |
| POST | `/v1/resources` | Create resource |
| PUT | `/v1/resources/:id` | Full update |
| PATCH | `/v1/resources/:id` | Partial update |
| DELETE | `/v1/resources/:id` | Delete resource |

### Versioning

API versioning via URL path: `/v1/`, `/v2/`

### Health Endpoints

- `GET /healthz` - Basic health check
- `GET /healthz/db` - Database health with migration status

## Configuration

Environment-based configuration using `caarlos0/env`:

```go
type Config struct {
    App     App
    HTTP    HTTP
    PG      PG
    // ...
}

type HTTP struct {
    Port           string `env:"HTTP_PORT,required"`
    UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
}
```

All configuration is loaded from environment variables, following 12-factor app principles.

## Multi-Transport Support

The same business logic is exposed via multiple transports:

- **HTTP/REST** - Primary API for web/mobile clients
- **gRPC** - High-performance internal communication
- **RabbitMQ RPC** - Async request/response patterns
- **NATS RPC** - Lightweight messaging

Each transport has its own controller implementation but shares the same use cases.

## Graceful Shutdown

All servers implement proper lifecycle management:

1. Receive shutdown signal (SIGTERM/SIGINT)
2. Stop accepting new connections
3. Wait for in-flight requests to complete
4. Close database connections
5. Exit cleanly

```go
interrupt := make(chan os.Signal, 1)
signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

select {
case s := <-interrupt:
    log.Info("app - Run - signal: " + s.String())
case err = <-httpServer.Notify():
    log.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
}

// Graceful shutdown
err = httpServer.Shutdown()
```
