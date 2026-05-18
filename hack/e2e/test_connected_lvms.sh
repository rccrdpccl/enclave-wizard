# Connected deployment with LVMS storage — no Quay fields expected
set -euo pipefail
source "$(dirname "$0")/helpers.sh"

echo "  Step 1: Verify wizard services are running"
assert_http_status "API health" "401" "http://localhost:8080/api/v1/defaults"
assert_http_status "UI health" "200" "https://localhost:3443/"

echo "  Step 2: Write a connected LVMS config via the wizard API"

CONFIG='{
  "global": {
    "workingDir": "/opt/enclave",
    "baseDomain": "connected-lvms.lab.local",
    "clusterName": "edge-conn",
    "machineNetwork": "10.10.50.0/24",
    "apiVIP": "10.10.50.200",
    "ingressVIP": "10.10.50.201",
    "rendezvousIP": "10.10.50.10",
    "defaultDNS": "10.10.50.1",
    "defaultGateway": "10.10.50.1",
    "defaultPrefix": 24,
    "lzBmcIP": "10.10.50.1",
    "quayUser": "unused",
    "quayPassword": "unused",
    "quayBackend": "LocalStorage",
    "blockStorageBackend": "lvms",
    "storage_plugin": "lvms",
    "disconnected": false,
    "enabled_plugins": ["lvms"],
    "pullSecret": {"auths":{}},
    "sshPubPath": "/home/wizard/.ssh/id_rsa.pub",
    "agent_hosts": [
      {
        "name": "ctrl-01",
        "macAddress": "00:60:2f:cc:01:01",
        "ipAddress": "10.10.50.10",
        "redfish": "10.10.50.1",
        "redfishUser": "admin",
        "redfishPassword": "password",
        "rootDisk": "/dev/sda"
      },
      {
        "name": "ctrl-02",
        "macAddress": "00:60:2f:cc:01:02",
        "ipAddress": "10.10.50.11",
        "redfish": "10.10.50.1",
        "redfishUser": "admin",
        "redfishPassword": "password",
        "rootDisk": "/dev/sda"
      },
      {
        "name": "ctrl-03",
        "macAddress": "00:60:2f:cc:01:03",
        "ipAddress": "10.10.50.12",
        "redfish": "10.10.50.1",
        "redfishUser": "admin",
        "redfishPassword": "password",
        "rootDisk": "/dev/sda"
      }
    ]
  },
  "certificates": {},
  "cloudInfra": {
    "discovery_hosts": []
  }
}'

api_put "/api/v1/config" "${CONFIG}"
echo "    ✓ Config written via API"

echo "  Step 3: Read config back and verify key fields"
assert_contains "baseDomain is set" "connected-lvms.lab.local" api_get "/api/v1/config"
assert_contains "clusterName is set" "edge-conn" api_get "/api/v1/config"
assert_contains "disconnected is false" "false" api_get "/api/v1/config"
assert_contains "storage is lvms" "lvms" api_get "/api/v1/config"
assert_contains "3 agent hosts" "ctrl-03" api_get "/api/v1/config"

echo "  Step 4: Verify config files exist on disk"
assert_ok "global.yaml exists" vm_exec "test -f /opt/enclave/config/global.yaml"
assert_ok "certificates.yaml exists" vm_exec "test -f /opt/enclave/config/certificates.yaml"
assert_ok "cloud_infra.yaml exists" vm_exec "test -f /opt/enclave/config/cloud_infra.yaml"

echo "  Step 5: Verify config content in YAML files"
assert_contains "global.yaml has baseDomain" "connected-lvms.lab.local" \
    vm_exec "cat /opt/enclave/config/global.yaml"
assert_contains "global.yaml has clusterName" "edge-conn" \
    vm_exec "cat /opt/enclave/config/global.yaml"
assert_contains "global.yaml has 3 nodes" "ctrl-03" \
    vm_exec "cat /opt/enclave/config/global.yaml"
assert_contains "global.yaml has disconnected false" "false" \
    vm_exec "cat /opt/enclave/config/global.yaml"

echo "  Step 6: Validate config against API schema"
VALIDATION=$(api_post "/api/v1/config/validate" "${CONFIG}")
assert_contains "API validation passes" "true" echo "${VALIDATION}"

echo "  Step 7: Validate against enclave schema"
validate_enclave_schema

echo "  All checks passed"
