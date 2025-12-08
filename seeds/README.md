# Database Seeds

This directory contains SQL seed files for populating the database with initial data.

## Structure

```
seeds/
├── development/    # Seed data for local development
│   └── 001_sample_data.sql
├── test/           # Fixtures for integration tests
│   └── 001_test_fixtures.sql
└── README.md
```

## Usage

```bash
# Load development seed data
make seed-dev

# Load test fixtures
make seed-test
```

## Naming Convention

Files are executed in alphabetical order. Use numeric prefixes to control execution order:

- `001_users.sql`
- `002_products.sql`
- `003_orders.sql`

## Guidelines

1. **Idempotency**: Seeds should be safe to run multiple times. Use `ON CONFLICT DO NOTHING` or check for existing data.
2. **Dependencies**: Ensure seed files respect foreign key constraints by ordering appropriately.
3. **No production data**: Never include real user data or sensitive information.
4. **Keep it minimal**: Only include data necessary for development/testing.
