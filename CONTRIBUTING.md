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

## Database Migrations

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Migration Commands

| Command | Description |
|---------|-------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-down-all` | Rollback all migrations (requires confirmation) |
| `make migrate-status` | Show current migration version |
| `make migrate-create name` | Create new migration files |
| `make migrate-force version=XX` | Force set migration version (for fixing dirty state) |
| `make migrate-validate` | Test migrations up/down/up cycle |

### Creating Migrations

```bash
# Create a new migration
make migrate-create add_users_table

# This creates two files:
# migrations/20241208123456_add_users_table.up.sql
# migrations/20241208123456_add_users_table.down.sql
```

### Migration File Structure

```sql
-- migrations/20241208123456_add_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- migrations/20241208123456_add_users_table.down.sql
DROP TABLE IF EXISTS users;
```

### Zero-Downtime Migration Patterns

When deploying to production, follow these patterns to avoid downtime:

**Safe Operations (can be done in single migration):**
- `ADD COLUMN` (nullable, without default)
- `CREATE INDEX CONCURRENTLY`
- `ADD CONSTRAINT ... NOT VALID` followed by `VALIDATE CONSTRAINT`
- `CREATE TABLE`

**Unsafe Operations (require multi-step deployment):**

| Operation | Safe Approach |
|-----------|---------------|
| Drop column | 1. Deploy code that doesn't use column 2. Then drop column |
| Rename column | 1. Add new column 2. Migrate data 3. Deploy code using new column 4. Drop old column |
| Add NOT NULL | 1. Add nullable column 2. Backfill data 3. Add NOT NULL constraint |
| Change column type | 1. Add new column 2. Migrate data 3. Update code 4. Drop old column |

**Example: Adding NOT NULL column safely**

```sql
-- Step 1: Add nullable column (migration 1)
ALTER TABLE users ADD COLUMN name VARCHAR(255);

-- Step 2: Backfill data (migration 2 or script)
UPDATE users SET name = 'Unknown' WHERE name IS NULL;

-- Step 3: Add NOT NULL constraint (migration 3)
ALTER TABLE users ALTER COLUMN name SET NOT NULL;
```

### Rollback Procedures

1. **Check current version:**
   ```bash
   make migrate-status
   ```

2. **Rollback one migration:**
   ```bash
   make migrate-down
   ```

3. **If migration is dirty (failed mid-way):**
   ```bash
   # Check the schema_migrations table
   psql $PG_URL -c "SELECT * FROM schema_migrations;"

   # Force to last known good version
   make migrate-force version=20241208123456
   ```

### Seed Data

```bash
# Load development sample data
make seed-dev

# Load test fixtures
make seed-test
```

Seed files are located in `seeds/development/` and `seeds/test/`.

### Health Check Endpoint

The `/healthz/db` endpoint returns the current migration status:

```bash
curl http://localhost:8080/healthz/db
```

Response:
```json
{
  "status": "healthy",
  "migration_version": 20210221023242,
  "dirty": false
}
```

### Configuration

Migration behavior can be configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `MIGRATION_ENABLED` | `true` | Enable/disable auto-migration on startup |
| `MIGRATION_RETRY_ATTEMPTS` | `20` | Number of DB connection retry attempts |
| `MIGRATION_RETRY_INTERVAL_SEC` | `1` | Seconds between retry attempts |

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
