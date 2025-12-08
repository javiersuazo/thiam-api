# Go Coding Rules

Enforceable coding standards for this codebase. All code must follow these rules.

## Table of Contents

1. [Error Handling](#error-handling)
2. [Interfaces](#interfaces)
3. [Context Usage](#context-usage)
4. [Naming Conventions](#naming-conventions)
5. [Package Organization](#package-organization)
6. [Testing](#testing)
7. [Logging](#logging)
8. [Configuration](#configuration)
9. [Database Operations](#database-operations)
10. [HTTP Handlers](#http-handlers)
11. [Concurrency](#concurrency)
12. [Comments and Documentation](#comments-and-documentation)

---

## Error Handling

### Never Ignore Errors

```go
// BAD
json.Unmarshal(data, &v)

// GOOD
if err := json.Unmarshal(data, &v); err != nil {
    return fmt.Errorf("unmarshal config: %w", err)
}
```

### Wrap Errors with Context

Use `%w` verb to preserve error chain:

```go
// BAD - loses original error
return fmt.Errorf("failed to get user: %v", err)

// GOOD - preserves error chain
return fmt.Errorf("UserRepo.GetByID - query: %w", err)
```

### Error Wrapping Format

Follow the pattern: `StructName.MethodName - operation: %w`

```go
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    row := r.pool.QueryRow(ctx, query, id)
    if err := row.Scan(&user); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, entity.ErrUserNotFound
        }
        return nil, fmt.Errorf("UserRepo.GetByID - scan: %w", err)
    }
    return &user, nil
}
```

### Use Sentinel Errors for Domain Errors

Define domain errors in `internal/entity/`:

```go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrInvalidEmail      = errors.New("invalid email")
)
```

Check with `errors.Is()`:

```go
// BAD
if err == entity.ErrUserNotFound {

// GOOD
if errors.Is(err, entity.ErrUserNotFound) {
```

### Map Infrastructure Errors to Domain Errors

At repository boundaries, convert infrastructure errors to domain errors:

```go
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    // ...
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, entity.ErrUserNotFound  // Domain error
    }
    return nil, fmt.Errorf("UserRepo.GetByID: %w", err)  // Infrastructure error
}
```

---

## Interfaces

### Accept Interfaces, Return Structs

```go
// GOOD - accepts interface, returns concrete type
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

// BAD - returns interface
func NewUserService(repo UserRepository) UserRepository {
    return &UserService{repo: repo}
}
```

### Define Interfaces Where They're Used

Define interfaces in the consumer package, not the implementation:

```go
// internal/usecase/contracts.go (consumer)
type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

// internal/repo/user_pg.go (implementation)
type UserRepo struct { ... }
func (r *UserRepo) GetByID(...) (*entity.User, error) { ... }
```

### Keep Interfaces Small

One or two methods is ideal. Compose if needed:

```go
// GOOD - small, focused interfaces
type UserReader interface {
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

type UserWriter interface {
    Save(ctx context.Context, user *entity.User) error
}

type UserRepository interface {
    UserReader
    UserWriter
}

// BAD - large interface
type UserRepository interface {
    GetByID(...)
    GetByEmail(...)
    GetAll(...)
    Save(...)
    Update(...)
    Delete(...)
    Count(...)
    Search(...)
}
```

### Interface Naming

- One method: method name + `-er` suffix (`Reader`, `Writer`, `Closer`)
- Multiple methods: descriptive name (`UserRepository`, `TranslationService`)

---

## Context Usage

### Always Pass Context as First Parameter

```go
// GOOD
func (uc *UserUseCase) Create(ctx context.Context, input CreateUserInput) (*entity.User, error)

// BAD
func (uc *UserUseCase) Create(input CreateUserInput, ctx context.Context) (*entity.User, error)
```

### Never Store Context in Structs

```go
// BAD
type Service struct {
    ctx context.Context
}

// GOOD - pass to each method
func (s *Service) DoWork(ctx context.Context) error
```

### Always Defer Cancel

```go
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()  // Always call, even if operation succeeds
```

### Check Context Cancellation in Long Operations

```go
for _, item := range items {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    process(item)
}
```

---

## Naming Conventions

### Package Names

- Lowercase, single word, no underscores
- Singular, not plural (`user`, not `users`)
- Avoid generic names: `util`, `common`, `helpers`, `misc`

### Variable Names

- Short, descriptive names
- `ctx` for context
- `err` for errors
- Single letter for loop variables and short-lived values

```go
// GOOD
for i, user := range users {
    if err := process(user); err != nil {
        return err
    }
}

// BAD
for index, currentUser := range users {
    if processError := process(currentUser); processError != nil {
        return processError
    }
}
```

### Receiver Names

Short (1-2 letters), consistent across methods:

```go
// GOOD
func (r *UserRepo) GetByID(...)
func (r *UserRepo) Save(...)

// BAD
func (repo *UserRepo) GetByID(...)
func (this *UserRepo) Save(...)
func (self *UserRepo) Delete(...)
```

### Constants

- Unexported: `camelCase`
- Exported: `PascalCase`
- NOT screaming snake case

```go
// GOOD
const defaultTimeout = 30 * time.Second
const MaxRetries = 3

// BAD
const DEFAULT_TIMEOUT = 30 * time.Second
const MAX_RETRIES = 3
```

### Getters

No `Get` prefix:

```go
// GOOD
func (u *User) Name() string

// BAD
func (u *User) GetName() string
```

### Don't Repeat Package Name

```go
// GOOD
package user
type User struct{}  // user.User

// BAD
package user
type UserModel struct{}  // user.UserModel
```

---

## Package Organization

### Use `internal/` for Private Code

Code in `internal/` cannot be imported by external packages:

```
internal/
├── entity/     # Domain entities
├── usecase/    # Business logic
├── repo/       # Data access
└── controller/ # Transport handlers
```

### Use `pkg/` for Reusable Infrastructure

Utilities that could be used by other projects:

```
pkg/
├── logger/
├── httpserver/
└── postgres/
```

### One File Per Major Type

```
internal/entity/
├── user.go           # User entity
├── order.go          # Order entity
└── errors.go         # Domain errors
```

### Group Related Files in Subdirectories

```
internal/repo/
├── contracts.go      # Interfaces
├── persistent/       # Database implementations
│   └── user_pg.go
└── webapi/          # External API adapters
    └── translation.go
```

---

## Testing

### Table-Driven Tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

### Capture Loop Variable for Parallel Tests

```go
for _, tt := range tests {
    tt := tt  // Capture!
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // test logic
    })
}
```

### Test Package Naming

Use `_test` suffix for black-box testing:

```go
package user_test  // Tests public API only

import "myproject/internal/entity/user"
```

### Use testify for Assertions

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
    user, err := entity.NewUser("test@example.com")
    require.NoError(t, err)  // Fail fast if error
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Use mockgen for Mocks

Generate mocks from interfaces:

```bash
mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
```

---

## Logging

### Use Structured Logging

```go
// GOOD
log.Info().
    Str("user_id", userID).
    Str("action", "created").
    Msg("user created")

// BAD
log.Info("User " + userID + " was created")
```

### Log Levels

| Level | Usage |
|-------|-------|
| `debug` | Detailed debugging info (development only) |
| `info` | Normal operations (user actions, requests) |
| `warn` | Unexpected but recoverable situations |
| `error` | Errors requiring attention |
| `fatal` | Critical errors, application will exit |

### Include Context in Logs

```go
log.Error().
    Err(err).
    Str("user_id", userID).
    Str("operation", "create_order").
    Msg("failed to create order")
```

### Never Log Sensitive Data

```go
// BAD
log.Info().Str("password", password).Msg("user login")

// GOOD
log.Info().Str("user_id", userID).Msg("user login")
```

---

## Configuration

### Use Environment Variables

All configuration via environment variables (12-factor app):

```go
type Config struct {
    Port string `env:"HTTP_PORT,required"`
}
```

### Validate Early

Fail fast on invalid configuration:

```go
func NewConfig() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("config error: %w", err)
    }
    return cfg, nil
}
```

### Use Sensible Defaults

```go
type HTTP struct {
    Port    string `env:"HTTP_PORT" envDefault:"8080"`
    Timeout int    `env:"HTTP_TIMEOUT" envDefault:"30"`
}
```

---

## Database Operations

### Always Use Context

```go
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    row := r.pool.QueryRow(ctx, query, id)  // Pass context
    // ...
}
```

### Use Query Builder for Complex Queries

```go
query, args, err := sq.
    Select("id", "email", "created_at").
    From("users").
    Where(sq.Eq{"id": id}).
    PlaceholderFormat(sq.Dollar).
    ToSql()
```

### Pre-allocate Slices

```go
const defaultCap = 64

func (r *UserRepo) GetAll(ctx context.Context) ([]entity.User, error) {
    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    users := make([]entity.User, 0, defaultCap)  // Pre-allocate
    for rows.Next() {
        var u entity.User
        if err := rows.Scan(&u.ID, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}
```

### Close Resources

```go
rows, err := r.pool.Query(ctx, query)
if err != nil {
    return nil, err
}
defer rows.Close()  // Always close
```

---

## HTTP Handlers

### Validate All Input

```go
func (h *Handler) CreateUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(errorResponse("invalid request body"))
    }

    if err := h.validator.Struct(req); err != nil {
        return c.Status(fiber.StatusUnprocessableEntity).JSON(validationErrorResponse(err))
    }

    // Process valid request
}
```

### Map Domain Errors to HTTP Status

```go
func (h *Handler) GetUser(c *fiber.Ctx) error {
    user, err := h.useCase.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, entity.ErrUserNotFound) {
            return c.Status(fiber.StatusNotFound).JSON(errorResponse("user not found"))
        }
        h.log.Error().Err(err).Msg("failed to get user")
        return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("internal error"))
    }
    return c.JSON(user)
}
```

### Never Expose Internal Errors

```go
// BAD - exposes internals
return c.Status(500).JSON(fiber.Map{"error": err.Error()})

// GOOD - generic message, log details
h.log.Error().Err(err).Msg("database query failed")
return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
```

---

## Concurrency

### Use errgroup for Coordinated Goroutines

```go
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    return fetchUsers(ctx)
})

g.Go(func() error {
    return fetchOrders(ctx)
})

if err := g.Wait(); err != nil {
    return err
}
```

### Use Worker Pools for Controlled Concurrency

```go
func worker(jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        results <- process(job)
    }
}

jobs := make(chan Job, 100)
results := make(chan Result, 100)

for w := 0; w < numWorkers; w++ {
    go worker(jobs, results)
}
```

### Implement Graceful Shutdown

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

<-ctx.Done()

shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

server.Shutdown(shutdownCtx)
```

---

## Comments and Documentation

### Minimal Comments

Code should be self-documenting. Only add comments when logic is non-obvious:

```go
// GOOD - explains WHY
// Retry up to 3 times because external API has intermittent failures
for attempts := 0; attempts < 3; attempts++ {
    if err := callAPI(); err == nil {
        break
    }
}

// BAD - explains WHAT (obvious from code)
// Loop 3 times
for i := 0; i < 3; i++ {
```

### Document Exported Types

Public types need documentation:

```go
// User represents a registered user in the system.
type User struct {
    ID    uuid.UUID
    Email string
}

// NewUser creates a new User with the given email.
// Returns ErrInvalidEmail if email format is invalid.
func NewUser(email string) (*User, error) {
```

### No TODO Comments in Main Code

Use issue tracker instead of TODO comments.
