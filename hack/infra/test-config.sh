#!/usr/bin/env bash
# Configure the wizard with all flavors (CaaS + VMaaS + BMaaS) enabled,
# simulating what a user would do clicking through the UI.
# Run after `make bm-emulation`, `make deploy`, and `make bm-emulation-config`.
#
# Usage:
#   ./hack/infra/test-config.sh --host root@hypervisor --pull-secret /path/to/pull-secret.json --manifest /path/to/manifest.zip
set -euo pipefail

TARGET=""
PULL_SECRET=""
MANIFEST=""
CLUSTER_NAME="${CLUSTER_NAME:-edge}"
BASE_DOMAIN="${BASE_DOMAIN:-enclave-test.lab.local}"

while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    --pull-secret) PULL_SECRET="$2"; shift 2 ;;
    --manifest) MANIFEST="$2"; shift 2 ;;
    --cluster-name) CLUSTER_NAME="$2"; shift 2 ;;
    --base-domain) BASE_DOMAIN="$2"; shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "${TARGET}" ] || [ -z "${PULL_SECRET}" ] || [ -z "${MANIFEST}" ]; then
  echo "Usage: $0 --host root@hypervisor --pull-secret /path/to/pull-secret.json --manifest /path/to/manifest.zip"
  exit 1
fi
PULL_SECRET="${PULL_SECRET/#\~/$HOME}"
MANIFEST="${MANIFEST/#\~/$HOME}"
[ ! -f "${PULL_SECRET}" ] && echo "ERROR: pull secret not found: ${PULL_SECRET}" && exit 1
[ ! -f "${MANIFEST}" ] && echo "ERROR: manifest not found: ${MANIFEST}" && exit 1

SSH="ssh -o StrictHostKeyChecking=no"
SCP="scp -o StrictHostKeyChecking=no -q"
HOST_ADDR="$(echo "${TARGET}" | cut -d@ -f2)"
API="https://${HOST_ADDR}:3443"

info()  { echo -e "\033[0;32m[INFO]\033[0m  $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*"; exit 1; }

VM_IP=$(${SSH} "${TARGET}" "cat /tmp/enclave-wizard-vm-ip 2>/dev/null" || true)
[ -z "${VM_IP}" ] && error "wizard VM not found. Run 'make deploy' first."

# Get LZ BMC IP
BMC_VM_IP=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "ip -4 addr 2>/dev/null | grep -oP 'inet 192\.168\.223\.\K[0-9]+'" 2>/dev/null || echo "")
[ -n "${BMC_VM_IP}" ] && BMC_VM_IP="192.168.223.${BMC_VM_IP}"
[ -z "${BMC_VM_IP}" ] && error "wizard VM has no BMC network IP. Run 'make bm-emulation' first."

# =============================================================================
# Step 1: Upload AAP manifest via the file upload API (same as UI FileUpload)
# =============================================================================
info "[1/3] Uploading AAP manifest..."

${SCP} "${MANIFEST}" "${TARGET}:/tmp/enclave-manifest.zip"
MANIFEST_PATH=$(${SSH} "${TARGET}" "
  curl -sk -X POST '${API}/api/v1/files' \
    -F 'file=@/tmp/enclave-manifest.zip' \
    -F 'dest=plugins' 2>/dev/null | python3 -c 'import json,sys; print(json.load(sys.stdin)[\"path\"])'
" 2>/dev/null)

if [ -z "${MANIFEST_PATH}" ]; then
  info "  File upload API not available, copying directly..."
  ${SSH} "${TARGET}" "
    ssh -o StrictHostKeyChecking=no wizard@${VM_IP} 'sudo mkdir -p /opt/enclave/config/plugins && sudo tee /opt/enclave/config/plugins/manifest.zip > /dev/null' < /tmp/enclave-manifest.zip
    ssh -o StrictHostKeyChecking=no wizard@${VM_IP} 'sudo chmod 644 /opt/enclave/config/plugins/manifest.zip'
  " 2>/dev/null
  MANIFEST_PATH="/opt/enclave/config/plugins/manifest.zip"
fi
info "  Manifest uploaded to: ${MANIFEST_PATH}"

# =============================================================================
# Step 2: Write SSH key to enclave config
# =============================================================================
info "[2/3] Setting SSH key..."

${SSH} "${TARGET}" "ssh-keygen -R ${VM_IP} 2>/dev/null || true"
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  test -f /home/wizard/.ssh/id_rsa.pub || ssh-keygen -t rsa -b 4096 -f /home/wizard/.ssh/id_rsa -N '' >/dev/null 2>&1
  sudo cp /home/wizard/.ssh/id_rsa.pub /opt/enclave/config/ssh-pub-key.pub
  sudo chmod 644 /opt/enclave/config/ssh-pub-key.pub
" 2>/dev/null

# =============================================================================
# Step 3: Write full config with all flavors enabled
# =============================================================================
info "[3/3] Writing config (CaaS + VMaaS + BMaaS)..."

# Get VM UUIDs
UUID0=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-0 2>/dev/null" || echo "unknown")
UUID1=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-1 2>/dev/null" || echo "unknown")
UUID2=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-2 2>/dev/null" || echo "unknown")

# Transfer pull secret to wizard VM
${SCP} "${PULL_SECRET}" "${TARGET}:/tmp/enclave-pull-secret.json"
${SSH} "${TARGET}" "scp -o StrictHostKeyChecking=no /tmp/enclave-pull-secret.json wizard@${VM_IP}:/tmp/pull-secret.json"

# Build and write config on the wizard VM
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  python3 -c \"
import json, subprocess as sp

ps = json.load(open('/tmp/pull-secret.json'))
ssh_key = open('/home/wizard/.ssh/id_rsa.pub').read().strip()

# All three flavors enabled = development profile with all plugins
# This is exactly what the UI sets when CaaS + VMaaS + BMaaS are selected
config = {
    'global': {
        'workingDir': '/opt/enclave',
        'lzBmcIP': '${BMC_VM_IP}',
        'disconnected': False,
        'baseDomain': '${BASE_DOMAIN}',
        'clusterName': '${CLUSTER_NAME}',
        'machineNetwork': '192.168.223.0/24',
        'apiVIP': '192.168.223.200',
        'ingressVIP': '192.168.223.201',
        'rendezvousIP': '192.168.223.10',
        'defaultDNS': '192.168.223.1',
        'defaultGateway': '192.168.223.1',
        'defaultPrefix': 24,
        'pullSecret': ps,
        'sshPubKey': ssh_key,
        'quayUser': 'admin',
        'quayPassword': 'password',
        'quayBackend': 'LocalStorage',
        'storage_plugin': 'lvms',

        # All flavors: CaaS + VMaaS + BMaaS = development profile
        'enabled_plugins': ['lvms', 'trust-manager', 'rhbk', 'authorino', 'aap', 'cnv', 'osac'],

        # OSAC settings (same as filling the OSAC Platform step)
        'osacProfile': 'development',
        'osacAapLicenseFile': '${MANIFEST_PATH}',
        'osacBYODatabase': False,

        # RHBK / Keycloak settings (same as filling the Identity Provider section)
        'rhbk_instances': 1,
        'rhbk_deploy_database': True,
        'rhbk_db_size': '5Gi',

        # Hosts
        'agent_hosts': [
            {'name': 'enclave-cp-0', 'macAddress': '00:60:2f:e0:c1:00', 'ipAddress': '192.168.223.10', 'redfish': '192.168.223.1:8100', 'redfishUser': 'admin', 'redfishPassword': 'password', 'rootDisk': '/dev/sda', 'bmcSystemId': '${UUID0}'},
            {'name': 'enclave-cp-1', 'macAddress': '00:60:2f:e0:c1:01', 'ipAddress': '192.168.223.11', 'redfish': '192.168.223.1:8100', 'redfishUser': 'admin', 'redfishPassword': 'password', 'rootDisk': '/dev/sda', 'bmcSystemId': '${UUID1}'},
            {'name': 'enclave-cp-2', 'macAddress': '00:60:2f:e0:c1:02', 'ipAddress': '192.168.223.12', 'redfish': '192.168.223.1:8100', 'redfishUser': 'admin', 'redfishPassword': 'password', 'rootDisk': '/dev/sda', 'bmcSystemId': '${UUID2}'},
        ],
    },
    'certificates': {},
    'cloudInfra': {'discovery_hosts': []},
}

# Write config — try with topology, fall back without
json.dump({**config, 'topology': {'availability_zones': []}}, open('/tmp/wizard-config.json', 'w'))
r = sp.run(['sudo', 'curl', '-sk', '-X', 'PUT', 'https://localhost:3443/api/v1/config', '-H', 'Content-Type: application/json', '-d', '@/tmp/wizard-config.json', '-o', '/tmp/resp.txt', '-w', '%{http_code}'], capture_output=True, text=True)
if r.stdout.strip() == '422':
    json.dump(config, open('/tmp/wizard-config.json', 'w'))
    r = sp.run(['sudo', 'curl', '-sk', '-X', 'PUT', 'https://localhost:3443/api/v1/config', '-H', 'Content-Type: application/json', '-d', '@/tmp/wizard-config.json', '-o', '/tmp/resp.txt', '-w', '%{http_code}'], capture_output=True, text=True)
code = r.stdout.strip()
if code in ('200', '204'):
    print('OK')
else:
    print(f'FAILED (HTTP {code})')
    print(open('/tmp/resp.txt').read())
    exit(1)
\"
"

echo ""
info "Config written with all flavors enabled:"
info "  Flavors:  CaaS + VMaaS + BMaaS"
info "  Profile:  development"
info "  Plugins:  lvms, trust-manager, rhbk, authorino, aap, cnv, osac"
info "  Manifest: ${MANIFEST_PATH}"
info ""
info "Open ${API}/wizard to review and click Deploy."
