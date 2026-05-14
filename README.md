# Enclave Configuration Wizard API

Backend API for the Enclave Configuration Wizard — a web-based tool that runs on the Landing Zone (LZ) and manages Enclave deployment config files (`config/global.yaml`, `config/certificates.yaml`, `config/cloud_infra.yaml`).

Built with [Huma](https://huma.rocks/) on the Go stdlib router. The API generates an OpenAPI 3.1 spec from Go types, which can be used to generate typed clients for any frontend framework.

## Quick start

```bash
make run
# or: go run . --port 8080 --enclave-dir ../enclave
```

The server starts on `http://localhost:8080` with interactive docs at `/docs`.

### CLI flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port`, `-p` | `8080` | Port to listen on |
| `--enclave-dir` | `../enclave` | Path to the Enclave repository root |

## OpenAPI schema

With the server running:

```bash
# OpenAPI 3.1 spec (YAML)
curl http://localhost:8080/openapi.yaml -o openapi.yaml

# OpenAPI 3.1 spec (JSON)
curl http://localhost:8080/openapi.json -o openapi.json
```

Use the spec to generate a typed client, for example with [openapi-typescript](https://openapi-ts.dev/):

```bash
npx openapi-typescript openapi.yaml -o src/api/schema.d.ts
```

## API endpoints

### Full config

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/config` | Load existing config from disk |
| PUT | `/api/v1/config` | Write config to disk |
| POST | `/api/v1/config/validate` | Validate config against Enclave schemas |
| POST | `/api/v1/config/preview` | Preview rendered YAML without writing |

### Config sections

Each section can be read or updated independently. A PUT to a section endpoint merges the payload into the existing config and writes the result to disk — other sections are left untouched.

| Method | Path | Description |
|--------|------|-------------|
| GET/PUT | `/api/v1/config/lz` | Landing zone (BMC IP, working dir, disconnected mode) |
| GET/PUT | `/api/v1/config/cluster` | Management cluster install (domain, VIPs, control-plane hosts, pull secret) |
| GET/PUT | `/api/v1/config/network` | Host network (DNS, gateway, prefix) |
| GET/PUT | `/api/v1/config/quay` | Quay registry (credentials, backend, RGW config) |
| GET/PUT | `/api/v1/config/storage` | Block storage (LVMS/ODF backend, external config) |
| GET/PUT | `/api/v1/config/plugins` | Enabled plugins and their configuration |
| GET/PUT | `/api/v1/config/certificates` | TLS certificates (API server, ingress, CA, Ironic) |
| GET/PUT | `/api/v1/config/hosts` | Discovery hosts (cloud infrastructure) |

### Defaults & plugins

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/defaults` | Get default values from `defaults/*.yaml` |
| GET | `/api/v1/plugins` | List available plugin descriptors |
| POST | `/api/v1/plugins/validate` | Check if a plugin combination is valid |

## Development

```bash
make build    # compile binary
make test     # run tests
make lint     # go vet
make tidy     # go mod tidy
```
