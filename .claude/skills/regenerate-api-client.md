---
name: regenerate-api-client
description: Regenerate the TypeScript API client from OpenAPI spec after Go model changes. Use when struct tags, endpoints, or request/response types change.
---

# Regenerate API Client

Run this whenever Go model structs or API endpoints change.

## Steps

### 1. Start the backend

```bash
make dev &
```

Wait for "Listening on ..." message.

### 2. Download the updated OpenAPI spec

```bash
curl -sk https://localhost:3443/openapi.yaml -o ui/packages/api-client/api/openapi.yaml
```

### 3. Stop the backend

```bash
kill %1
```

### 4. Regenerate the TypeScript client

```bash
cd ui && make api-client
```

This runs `openapi-generator-cli` configured in `ui/openapitools.json`. Generated code lands in `ui/packages/api-client/src/`.

### 5. Verify

```bash
# Check TypeScript compiles (needs distrobox)
distrobox enter osac
cd ui && npx tsc --noEmit
```

## When to Run

- After adding/modifying fields in `internal/models/*.go`
- After adding/modifying API endpoints in `internal/api/*.go`
- After changing struct tags (json, doc, pattern, enum, etc.)

You do NOT need to regen for frontend-only changes.
