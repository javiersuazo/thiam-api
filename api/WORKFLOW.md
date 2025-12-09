# API-First Development Workflow

This document describes how to use the OpenAPI specification for parallel frontend/backend development.

## Overview

```
1. DESIGN              2. PARALLEL DEV           3. INTEGRATE
   ──────────             ────────────              ─────────

   OpenAPI Spec    ──►    BE: Implement      ──►   FE connects
   (agreed by             handlers                 to real API
   FE + BE)
                   ──►    FE: Build UI
                          with mock server
```

## For Backend Developers

### Adding a New Feature

1. **Design First** - Add endpoints to `api/openapi.yaml`:
   ```yaml
   paths:
     /v1/your-resource:
       get:
         summary: List resources
         responses:
           "200":
             content:
               application/json:
                 schema:
                   $ref: "#/components/schemas/ResourceList"
   ```

2. **Add Schemas** - Define request/response models:
   ```yaml
   components:
     schemas:
       Resource:
         type: object
         properties:
           id:
             type: string
             format: uuid
           name:
             type: string
   ```

3. **Notify Frontend** - Once spec is committed, FE can start building

4. **Implement** - Build your handlers following clean architecture

### Validation

Validate your spec before committing:
```bash
npx @redocly/cli lint api/openapi.yaml
```

## For Frontend Developers

### Setup (One Time)

```bash
# Clone the backend repo (or add as git submodule)
git clone https://github.com/javiersuazo/thiam-api.git

# Or fetch just the spec
curl -O https://raw.githubusercontent.com/javiersuazo/thiam-api/main/api/openapi.yaml
```

### Generate TypeScript Client

```bash
# Generate types only
./api/scripts/generate-client.sh -o ./src/types

# Generate with fetch client
./api/scripts/generate-client.sh -t fetch -o ./src/api
```

### Run Mock Server

Start a mock server that returns realistic fake data:

```bash
# Default port 4010
./api/scripts/generate-client.sh --mock

# Custom port
./api/scripts/generate-client.sh --mock -p 8081
```

The mock server will:
- Return example responses from the spec
- Validate your requests against the spec
- Generate realistic fake data

### Using Generated Types

```typescript
import type { paths, components } from './types/api';

// Type-safe API response
type User = components['schemas']['User'];

// Type-safe path parameters
type GetUserParams = paths['/v1/users/{id}']['get']['parameters']['path'];

// With fetch
const response = await fetch('/v1/users');
const users: User[] = await response.json();
```

## Workflow Per Feature

### Step 1: Design Session (30 min)

FE and BE meet to:
1. Define endpoint paths
2. Agree on request/response shapes
3. Document in `api/openapi.yaml`

### Step 2: Parallel Development

| Backend | Frontend |
|---------|----------|
| Implement handlers | Generate TypeScript client |
| Write use cases | Build UI components |
| Test against spec | Use mock server for dev |
| Deploy to staging | Test with mock data |

### Step 3: Integration

1. FE switches from mock server to real API
2. Both teams verify integration
3. Ship to production

## Best Practices

### Spec Design
- Use `$ref` for reusable schemas
- Include examples in schemas
- Use consistent naming (camelCase for JSON, snake_case for query params)
- Document all error responses

### Versioning
- Use `/v1/`, `/v2/` path prefixes
- Don't break existing contracts
- Deprecate before removing

### Error Handling
- Use standard error response schema
- Include `request_id` for tracing
- Provide actionable error messages

## File Structure

```
api/
├── openapi.yaml           # Source of truth
├── WORKFLOW.md            # This file
└── scripts/
    └── generate-client.sh # Client generation helper
```

## Tools

- **Spec Editing**: [Swagger Editor](https://editor.swagger.io/)
- **Mock Server**: [Prism](https://stoplight.io/open-source/prism)
- **TypeScript Generation**: [openapi-typescript](https://github.com/drwpow/openapi-typescript)
- **Validation**: [Redocly CLI](https://redocly.com/docs/cli/)
