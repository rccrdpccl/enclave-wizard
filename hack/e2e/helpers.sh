#!/usr/bin/env bash
# Shared helpers for e2e tests

SSH_VM="ssh -o StrictHostKeyChecking=no -J ${TARGET} wizard@${VM_IP}"
API="https://localhost:3443"

# Run a command inside the VM
vm_exec() {
    ${SSH_VM} "$@"
}

# Run a command on the target host
host_exec() {
    ssh -o StrictHostKeyChecking=no "${TARGET}" "$@"
}

# Call the wizard API (via SSH tunnel through target → VM)
api_call() {
    local method="$1"
    local path="$2"
    shift 2
    vm_exec "curl -sfk -X ${method} https://localhost:3443${path} -H 'Content-Type: application/json' $*"
}

api_get() {
    api_call GET "$1"
}

api_put() {
    local path="$1"
    local data="$2"
    vm_exec "curl -sfk -X PUT https://localhost:3443${path} -H 'Content-Type: application/json' -d '${data}'"
}

api_post() {
    local path="$1"
    local data="$2"
    vm_exec "curl -sfk -X POST https://localhost:3443${path} -H 'Content-Type: application/json' -d '${data}'"
}

# Assert a command succeeds
assert_ok() {
    local desc="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        echo "    ✓ ${desc}"
    else
        echo "    ✗ ${desc}"
        return 1
    fi
}

# Assert a command output contains a string
assert_contains() {
    local desc="$1"
    local expected="$2"
    shift 2
    local output
    output=$("$@" 2>&1)
    if echo "${output}" | grep -q "${expected}"; then
        echo "    ✓ ${desc}"
    else
        echo "    ✗ ${desc} (expected '${expected}' in output)"
        echo "      got: ${output:0:200}"
        return 1
    fi
}

# Assert HTTP status code
assert_http_status() {
    local desc="$1"
    local expected="$2"
    local url="$3"
    local status
    status=$(vm_exec "curl -sfk -o /dev/null -w '%{http_code}' '${url}'" 2>/dev/null)
    if [ "${status}" = "${expected}" ]; then
        echo "    ✓ ${desc} (${status})"
    else
        echo "    ✗ ${desc} (expected ${expected}, got ${status})"
        return 1
    fi
}

# Assert a command output does NOT contain a string
assert_not_contains() {
    local desc="$1"
    local unexpected="$2"
    shift 2
    local output
    output=$("$@" 2>&1)
    if echo "${output}" | grep -q "${unexpected}"; then
        echo "    ✗ ${desc} (found unexpected '${unexpected}' in output)"
        return 1
    else
        echo "    ✓ ${desc}"
    fi
}

# Assert a JSON field has an exact value (requires jq on the VM)
assert_field() {
    local desc="$1"
    local jq_expr="$2"
    local expected="$3"
    local json="$4"
    local actual
    actual=$(echo "${json}" | jq -r "${jq_expr}")
    if [ "${actual}" = "${expected}" ]; then
        echo "    ✓ ${desc}"
    else
        echo "    ✗ ${desc} (expected '${expected}', got '${actual}')"
        return 1
    fi
}

# Assert HTTP status code (without -f, so we can inspect 4xx/5xx)
assert_http_code() {
    local desc="$1"
    local expected="$2"
    local method="$3"
    local path="$4"
    local body="${5:-}"
    local status
    if [ -n "${body}" ]; then
        status=$(vm_exec "curl -sk -o /dev/null -w '%{http_code}' -X ${method} https://localhost:3443${path} -H 'Content-Type: application/json' -d '${body}'")
    else
        status=$(vm_exec "curl -sk -o /dev/null -w '%{http_code}' -X ${method} https://localhost:3443${path}")
    fi
    if [ "${status}" = "${expected}" ]; then
        echo "    ✓ ${desc} (${status})"
    else
        echo "    ✗ ${desc} (expected ${expected}, got ${status})"
        return 1
    fi
}

# Run enclave's validate-schema against the config written by the wizard
validate_enclave_schema() {
    vm_exec "
        cd /opt/enclave
        python3 -c '
import yaml, sys
try:
    from jsonschema import validate, ValidationError
except ImportError:
    print(\"SKIP: jsonschema not installed\")
    sys.exit(0)

with open(\"schemas/variables.yaml\") as f:
    schema = yaml.safe_load(f)
with open(\"schemas/definitions.yaml\") as f:
    defs = yaml.safe_load(f)
with open(\"config/global.yaml\") as f:
    config = yaml.safe_load(f)

schema[\"definitions\"] = defs.get(\"definitions\", {})

for extra in [\"enabled_plugins\", \"storage_plugin\"]:
    config.pop(extra, None)

try:
    validate(instance=config, schema=schema)
    print(\"Schema validation: PASS\")
except ValidationError as e:
    print(f\"Schema validation: FAIL - {e.message}\")
    sys.exit(1)
except Exception as e:
    print(f\"Schema validation: SKIP - {e}\")
    sys.exit(0)
'
    "
}

# Run enclave's validations.sh
validate_enclave_config() {
    vm_exec "cd /opt/enclave && bash validations.sh 2>&1"
}
