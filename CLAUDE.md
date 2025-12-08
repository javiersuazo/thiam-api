# Claude Code Instructions

Project-specific instructions for Claude Code when working on this codebase.

## Documentation

Before writing code, review these documents:

- **Architecture**: `docs/ARCHITECTURE.md` - Layer structure, DDD/SOLID principles, package organization
- **Coding Rules**: `docs/CODING_RULES.md` - Error handling, interfaces, context, naming, testing
- **Event Patterns**: `docs/EVENT_DRIVEN.md` - When to use events vs API, outbox pattern

## Key Rules

### Error Handling
- Wrap errors with context: `fmt.Errorf("StructName.Method - operation: %w", err)`
- Use sentinel errors for domain errors in `internal/entity/`
- Check with `errors.Is()`, not `==`
- Map infrastructure errors to domain errors at repository boundaries

### Interfaces
- Accept interfaces, return structs
- Define interfaces where they're used (consumer package)
- Keep interfaces small (1-2 methods ideal)

### Context
- Always pass `context.Context` as first parameter
- Never store context in structs
- Always `defer cancel()` after `context.WithTimeout/Cancel`

### Architecture
- Dependencies point inward: controller → usecase → entity
- Business logic in `internal/usecase/`, not controllers
- Domain entities in `internal/entity/` have no external dependencies
- Repositories implement interfaces defined in `internal/repo/contracts.go`

### Testing
- Use table-driven tests
- Use `t.Run()` for subtests
- Capture loop variable for parallel tests: `tt := tt`
- Use testify for assertions

### Code Style
- Minimal comments - code should be self-documenting
- No TODO comments - use issue tracker
- Follow existing patterns in the codebase

## Project Structure

```
internal/
├── entity/      # Domain layer (no dependencies)
├── usecase/     # Application layer (business logic)
├── repo/        # Repository implementations
└── controller/  # Transport layer (HTTP, gRPC, AMQP)
```

## Common Commands

```bash
make dev          # Run with hot-reload
make test         # Run unit tests
make linter-golangci  # Run linter
make pre-commit   # Run all quality checks
```
