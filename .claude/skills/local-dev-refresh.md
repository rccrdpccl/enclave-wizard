---
name: local-dev-refresh
description: Rebuild and restart the enclave-wizard locally with enclave-mock for development. Use when testing UI or backend changes locally, refreshing the running wizard, or starting the wizard in mock mode.
---

# Local Dev Refresh

## Overview

The wizard embeds the UI at Go build time. Both the UI (vite build) and the Go binary must be rebuilt when either changes. The wizard runs against `enclave-mock/` which is generated from the real enclave repo with noop'd Ansible tasks.

## Full Refresh Steps

### 1. Generate enclave-mock (if not present)

```bash
test -d enclave-mock || python3 hack/generate-enclave-mock.py
```

To regenerate from a specific branch:
```bash
make clean-enclave-mock
make enclave-mock ENCLAVE_MOCK_BRANCH=<branch>
```

### 2. Kill any running wizard on the dev port

```bash
pkill -9 -f 'enclave-wizard.*5443' 2>/dev/null; sleep 1
```

### 3. Rebuild the UI

```bash
podman run --rm -v $(pwd)/ui:/app:z -w /app node:22-alpine \
  sh -c "corepack enable && yarn install && \
  yarn workspace @enclave-wizard-ui/wizard run -T vite build"
```

### 4. Rebuild the Go binary

```bash
rm -f ./enclave-wizard
podman run --rm -v $(pwd):/app:z -w /app golang:latest \
  go build -ldflags='-w -s' -o enclave-wizard .
```

### 5. Start the wizard

```bash
rm -f /tmp/enclave-wizard-dev.pass
./enclave-wizard \
  --enclave-dir enclave-mock \
  --tls-cert hack/tls/server.crt \
  --tls-key hack/tls/server.key \
  --no-auth \
  --password-file /tmp/enclave-wizard-dev.pass \
  --https-port 5443 --http-port 5001 \
  > /tmp/wizard-dev.log 2>&1 &
sleep 3
grep "Initial admin password" /tmp/wizard-dev.log
```

### 6. Generate TLS certs (first time only)

If `hack/tls/server.crt` doesn't exist:

```bash
mkdir -p hack/tls
openssl req -x509 -newkey rsa:2048 \
  -keyout hack/tls/server.key -out hack/tls/server.crt \
  -days 365 -nodes -subj "/CN=localhost"
```

## Quick Refresh (backend-only changes)

When only Go code changed (no UI changes), skip step 3:

```bash
pkill -9 -f 'enclave-wizard.*5443' 2>/dev/null; sleep 1
rm -f ./enclave-wizard
podman run --rm -v $(pwd):/app:z -w /app golang:latest \
  go build -ldflags='-w -s' -o enclave-wizard .
rm -f /tmp/enclave-wizard-dev.pass
./enclave-wizard --enclave-dir enclave-mock \
  --tls-cert hack/tls/server.crt --tls-key hack/tls/server.key \
  --no-auth --password-file /tmp/enclave-wizard-dev.pass \
  --https-port 5443 --http-port 5001 > /tmp/wizard-dev.log 2>&1 &
sleep 3
grep "Initial admin password" /tmp/wizard-dev.log
```

## Important Notes

- The UI is embedded in the Go binary. If you change UI code, you MUST rebuild the UI (step 3) AND the Go binary (step 4).
- Port 5443 is used for dev to avoid conflicts with any production instance on 3443 or other instances on 4443.
- The `--no-auth` flag bypasses bearer token checks on API calls, but the login screen still appears. Use the generated password.
- `ansible-runner` is not available locally. Schema validation will be skipped on config writes. Use the container-based run (`Containerfile` with ansible-runner) for full validation testing.
- The `enclave-mock/` directory is gitignored.
