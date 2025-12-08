# Architecture Documentation

This document describes the architectural patterns, principles, and guidelines for this codebase.

## Core Principles

This project follows **Domain-Driven Design (DDD)** and **SOLID** principles rigorously.

### SOLID Principles

| Principle | Application |
|-----------|-------------|
| **Single Responsibility** | Each struct/package has one reason to change. UseCases handle business logic, Repositories handle persistence, Controllers handle transport. |
| **Open/Closed** | Extend behavior through interfaces, not modification. Add new notification channels by implementing `Notifier`, not changing existing code. |
| **Liskov Substitution** | All interface implementations are interchangeable. `PostgresUserRepo` and `MockUserRepo` both satisfy `UserRepository`. |
| **Interface Segregation** | Small, focused interfaces. `UserReader` and `UserWriter` instead of one large `UserRepository`. |
| **Dependency Inversion** | High-level modules depend on abstractions. UseCases depend on repository interfaces, not concrete implementations. |

### DDD Concepts

| Concept | Location | Description |
|---------|----------|-------------|
| **Entities** | `internal/entity/` | Domain objects with identity |
| **Value Objects** | `internal/entity/` | Immutable objects without identity |
| **Aggregates** | `internal/entity/` | Cluster of entities treated as a unit |
| **Repositories** | `internal/repo/` | Persistence abstraction for aggregates |
| **Use Cases** | `internal/usecase/` | Application services orchestrating domain logic |
| **Domain Events** | `internal/entity/events/` | Immutable facts about what happened |

## Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Controllers (HTTP/gRPC/MQ)              │
│  Handles transport concerns: serialization, validation,     │
│  authentication, routing                                    │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Use Cases                            │
│  Orchestrates business logic, coordinates domain objects,   │
│  manages transactions                                       │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Domain (Entities)                       │
│  Core business logic, aggregates, domain events,            │
│  business rules and invariants                              │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Infrastructure                         │
│  Repositories, external services, message queues,           │
│  database connections                                       │
└─────────────────────────────────────────────────────────────┘
```

### Dependency Rule

Dependencies point inward. Inner layers know nothing about outer layers.

```
Controllers → UseCases → Entities ← Repositories
                 ↓
            Interfaces (contracts.go)
```

## Package Structure

```
internal/
├── app/                    # Application bootstrap and wiring
│   ├── app.go             # Main application setup
│   └── migrate.go         # Database migration (build tag: migrate)
│
├── entity/                 # Domain layer
│   ├── user.go            # User aggregate
│   ├── translation.go     # Translation entity
│   └── events/            # Domain events (to be added)
│
├── usecase/               # Application layer
│   ├── contracts.go       # Use case interfaces
│   ├── translation.go     # Translation use case
│   └── *_test.go         # Unit tests with mocks
│
├── repo/                  # Repository implementations
│   ├── contracts.go       # Repository interfaces
│   ├── translation_pg.go  # PostgreSQL implementation
│   └── webapi/           # External API adapters
│
└── controller/            # Transport layer
    ├── http/             # REST API (Fiber)
    │   ├── router.go     # Route definitions
    │   └── v1/           # API version 1 handlers
    ├── grpc/             # gRPC handlers
    └── amqp/             # RabbitMQ handlers
```

## Interface Patterns

### Repository Interfaces

Defined in `internal/repo/contracts.go`:

```go
type UserReader interface {
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type UserWriter interface {
    Save(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}

type UserRepository interface {
    UserReader
    UserWriter
}
```

### Use Case Interfaces

Defined in `internal/usecase/contracts.go`:

```go
type CreateUserInput struct {
    Email    string
    Password string
}

type UserCreator interface {
    Create(ctx context.Context, input CreateUserInput) (*entity.User, error)
}
```

## Error Handling

### Domain Errors

Domain errors are defined in the entity package:

```go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrInvalidEmail      = errors.New("invalid email format")
)
```

### Error Wrapping

Use `fmt.Errorf` with `%w` for error wrapping:

```go
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    user, err := r.pool.QueryRow(ctx, query, id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, entity.ErrUserNotFound
        }
        return nil, fmt.Errorf("UserRepo.GetByID: %w", err)
    }
    return user, nil
}
```

## Context Usage

- Always pass `context.Context` as the first parameter
- Use context for cancellation and timeouts
- Never store request-scoped values in context (use explicit parameters)

```go
func (uc *UserUseCase) Create(ctx context.Context, input CreateUserInput) (*entity.User, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    // ...
}
```

## Testing Strategy

| Layer | Test Type | Dependencies |
|-------|-----------|--------------|
| Entity | Unit | None |
| UseCase | Unit | Mocked repositories |
| Repository | Integration | Real database |
| Controller | Integration | Full stack or mocked use cases |

### Mock Generation

```bash
make mock
```

Generates mocks in `internal/usecase/mocks_*_test.go`.

## Database Patterns

### Transactions

Use the unit of work pattern for operations spanning multiple aggregates:

```go
type UnitOfWork interface {
    Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
    Commit() error
    Rollback() error
    UserRepo() UserRepository
    OrderRepo() OrderRepository
}
```

### Query Building

Use Squirrel for type-safe query building:

```go
query, args, err := sq.
    Select("id", "email", "created_at").
    From("users").
    Where(sq.Eq{"id": id}).
    PlaceholderFormat(sq.Dollar).
    ToSql()
```

## Configuration

Environment-based configuration via `caarlos0/env`:

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

## Logging

Use structured logging with `zerolog`:

```go
log.Info().
    Str("user_id", userID.String()).
    Str("action", "created").
    Msg("User created successfully")
```

Log levels:
- `debug`: Detailed information for debugging
- `info`: General operational information
- `warn`: Unexpected situations that aren't errors
- `error`: Errors that need attention
- `fatal`: Critical errors causing shutdown

## API Design

### REST Endpoints

Follow REST conventions:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/users` | List users |
| GET | `/v1/users/:id` | Get user by ID |
| POST | `/v1/users` | Create user |
| PUT | `/v1/users/:id` | Update user |
| DELETE | `/v1/users/:id` | Delete user |

### Response Format

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100
  }
}
```

Error responses:

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID xyz not found"
  }
}
```

## Code Style

- No comments unless logic is non-obvious
- Function names should be self-documenting
- Keep functions under 30 lines when possible
- Use early returns to reduce nesting
- Group related code together

```go
// Good: self-documenting
func (u *User) CanPlaceOrder() bool {
    return u.IsActive && u.EmailVerified && !u.IsBanned
}

// Avoid: unnecessary comment
// CanPlaceOrder checks if user can place an order
func (u *User) CanPlaceOrder() bool {
    return u.IsActive && u.EmailVerified && !u.IsBanned
}
```
