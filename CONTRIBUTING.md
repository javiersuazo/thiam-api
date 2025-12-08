# Contributing

## Development Setup

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make
- Protocol Buffers compiler (`protoc`)

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd backend

# One-command setup (installs tools, copies .env, installs pre-commit hooks)
make setup

# Start infrastructure (PostgreSQL, RabbitMQ, NATS)
make compose-up

# Run with hot-reload
make dev
```

### Environment Diagnosis

Run `make doctor` to check if all required tools are installed:

```bash
make doctor
```

## Development Workflow

### Running the Application

```bash
# Development mode with hot-reload
make dev

# Production-like run
make run

# Run with all infrastructure
make compose-up-all
```

### Code Quality

```bash
# Format code
make format

# Run linter
make linter-golangci

# Run tests
make test

# Run all pre-commit checks
make pre-commit
```

### Database

```bash
# Start database
make compose-up

# Run migrations
make migrate-up

# Create new migration
make migrate-create "add_users_table"

# Reset database to clean state
make reset-db
```

### Code Generation

```bash
# Generate Swagger docs
make swag-v1

# Generate gRPC code
make proto-v1

# Generate mocks
make mock
```

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/). Commit messages are validated by pre-commit hooks.

Format: `<type>(<scope>): <description>`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Maintenance tasks
- `revert`: Reverting changes

Examples:
```
feat(auth): add JWT token refresh endpoint
fix(db): resolve connection pool leak
docs(readme): update installation instructions
```

## Project Structure

```
.
├── cmd/app/              # Application entry point
├── internal/             # Private application code
│   ├── app/              # Application bootstrap
│   ├── entity/           # Domain models
│   ├── controller/       # HTTP, gRPC, message queue handlers
│   ├── usecase/          # Business logic
│   └── repo/             # Data access layer
├── pkg/                  # Shared utilities
├── config/               # Configuration
├── migrations/           # Database migrations
├── docs/                 # API documentation
└── integration-test/     # Integration tests
```

## Testing

```bash
# Unit tests
make test

# Integration tests (requires Docker)
make compose-up-integration-test

# Watch mode (re-runs on file changes)
make test-watch
```

## IDE Setup

### VS Code

The repository includes VS Code configuration in `.vscode/`. Install recommended extensions when prompted.

### GoLand

Import the project and it will automatically detect the Go module.

## Useful Commands

| Command | Description |
|---------|-------------|
| `make setup` | One-command dev environment setup |
| `make dev` | Run with hot-reload |
| `make doctor` | Diagnose development environment |
| `make test` | Run unit tests |
| `make linter-golangci` | Run linter |
| `make pre-commit` | Run all quality checks |
| `make clean` | Clean generated files |
| `make reset-db` | Reset database |
| `make logs` | Tail application logs |
| `make help` | Show all available commands |
