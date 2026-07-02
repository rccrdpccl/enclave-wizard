#!/usr/bin/env bash
# Write BM emulation infrastructure settings to the wizard via its API.
# Merges with any existing config (preserves flavor/plugin selections from UI).
# Run after `make bm-emulation` and `make deploy`.
#
# Usage:
#   ./hack/infra/bm-emulation-config.sh --host root@hypervisor --pull-secret /path/to/pull-secret.json
set -euo pipefail

TARGET=""
PULL_SECRET=""
CLUSTER_NAME="${CLUSTER_NAME:-edge}"
BASE_DOMAIN="${BASE_DOMAIN:-enclave-test.lab.local}"

while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    --pull-secret) PULL_SECRET="$2"; shift 2 ;;
    --cluster-name) CLUSTER_NAME="$2"; shift 2 ;;
    --base-domain) BASE_DOMAIN="$2"; shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "${TARGET}" ]; then
  echo "Usage: $0 --host root@hypervisor --pull-secret /path/to/pull-secret.json"
  exit 1
fi
PULL_SECRET="${PULL_SECRET/#\~/$HOME}"
if [ -z "${PULL_SECRET}" ] || [ ! -f "${PULL_SECRET}" ]; then
  echo "ERROR: --pull-secret must point to a valid pull secret JSON file"
  exit 1
fi

SSH="ssh -o StrictHostKeyChecking=no"
SCP="scp -o StrictHostKeyChecking=no -q"

info()  { echo -e "\033[0;32m[INFO]\033[0m  $*"; }

VM_IP=$(${SSH} "${TARGET}" "cat /tmp/enclave-wizard-vm-ip 2>/dev/null" || true)
if [ -z "${VM_IP}" ]; then
  echo "ERROR: wizard VM not found. Run 'make deploy TARGET=${TARGET}' first."
  exit 1
fi

# Get LZ BMC IP (check any interface on the BMC subnet)
BMC_VM_IP=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "ip -4 addr 2>/dev/null | grep -oP 'inet 192\.168\.223\.\K[0-9]+'" 2>/dev/null || echo "")
[ -n "${BMC_VM_IP}" ] && BMC_VM_IP="192.168.223.${BMC_VM_IP}"
if [ -z "${BMC_VM_IP}" ]; then
  echo "ERROR: wizard VM has no BMC network IP. Run 'make bm-emulation' first."
  exit 1
fi

info "Writing BM emulation config to wizard..."

# Ensure wizard VM has an SSH key
${SSH} "${TARGET}" "ssh-keygen -R ${VM_IP} 2>/dev/null || true"
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  test -f /home/wizard/.ssh/id_rsa.pub || ssh-keygen -t rsa -b 4096 -f /home/wizard/.ssh/id_rsa -N '' >/dev/null 2>&1
" 2>/dev/null

# Transfer pull secret
${SCP} "${PULL_SECRET}" "${TARGET}:/tmp/enclave-pull-secret.json"

# Write SSH pub key directly to enclave config (bypass RPM default)
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  sudo cp /home/wizard/.ssh/id_rsa.pub /opt/enclave/config/ssh-pub-key.pub
  sudo chmod 644 /opt/enclave/config/ssh-pub-key.pub
" 2>/dev/null

# Read SSH key from wizard VM
SSH_KEY=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "cat /home/wizard/.ssh/id_rsa.pub" 2>/dev/null)
if [ -z "${SSH_KEY}" ]; then
  echo "ERROR: could not read SSH public key from wizard VM"
  exit 1
fi

# Get VM UUIDs from hypervisor
UUID0=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-0 2>/dev/null" || echo "unknown")
UUID1=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-1 2>/dev/null" || echo "unknown")
UUID2=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-2 2>/dev/null" || echo "unknown")

# Transfer pull secret to wizard VM
${SSH} "${TARGET}" "scp -o StrictHostKeyChecking=no /tmp/enclave-pull-secret.json wizard@${VM_IP}:/tmp/pull-secret.json"

# Build and write config directly on the wizard VM
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  python3 -c \"
import json, subprocess as sp

ps = json.load(open('/tmp/pull-secret.json'))
ssh_key = open('/home/wizard/.ssh/id_rsa.pub').read().strip()

# Read existing config to preserve UI selections (flavors, plugins, OSAC)
try:
    r = sp.run(['sudo', 'curl', '-sk', 'https://localhost:3443/api/v1/config'], capture_output=True, text=True)
    existing = json.loads(r.stdout) if r.returncode == 0 else {}
except:
    existing = {}
existing_global = existing.get('global', {})

# Infrastructure settings only — never touch enabled_plugins or plugin-specific fields
infra = {
    'workingDir': '/home/enclave',
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
    'agent_hosts': [
        {'name': 'enclave-cp-0', 'macAddress': '00:60:2f:e0:c1:00', 'ipAddress': '192.168.223.10', 'redfish': '192.168.223.1:8100', 'redfishUser': 'admin', 'redfishPassword': 'password', 'rootDisk': '/dev/sda', 'bmcSystemId': '${UUID0}'},
        {'name': 'enclave-cp-1', 'macAddress': '00:60:2f:e0:c1:01', 'ipAddress': '192.168.223.11', 'redfish': '192.168.223.1:8100', 'redfishUser': 'admin', 'redfishPassword': 'password', 'rootDisk': '/dev/sda', 'bmcSystemId': '${UUID1}'},
        {'name': 'enclave-cp-2', 'macAddress': '00:60:2f:e0:c1:02', 'ipAddress': '192.168.223.12', 'redfish': '192.168.223.1:8100', 'redfishUser': 'admin', 'redfishPassword': 'password', 'rootDisk': '/dev/sda', 'bmcSystemId': '${UUID2}'},
    ],
}

# Merge: existing keeps plugin selections, infra overwrites network/host settings
# Set defaults for fields the existing config might not have
if 'storage_plugin' not in existing_global:
    infra['storage_plugin'] = 'lvms'
if 'enabled_plugins' not in existing_global:
    infra['enabled_plugins'] = ['lvms']

merged_global = {**existing_global, **infra}
config = {
    'global': merged_global,
    'certificates': existing.get('certificates', {}),
    'cloudInfra': existing.get('cloudInfra', {'discovery_hosts': []}),
}

# Try with topology, then without
json.dump({**config, 'topology': existing.get('topology', {'availability_zones': []})}, open('/tmp/wizard-config.json', 'w'))
r = sp.run(['sudo', 'curl', '-sk', '-X', 'PUT', 'https://localhost:3443/api/v1/config', '-H', 'Content-Type: application/json', '-d', '@/tmp/wizard-config.json', '-o', '/tmp/resp.txt', '-w', '%{http_code}'], capture_output=True, text=True)
if r.stdout.strip() == '422':
    json.dump(config, open('/tmp/wizard-config.json', 'w'))
    r = sp.run(['sudo', 'curl', '-sk', '-X', 'PUT', 'https://localhost:3443/api/v1/config', '-H', 'Content-Type: application/json', '-d', '@/tmp/wizard-config.json', '-o', '/tmp/resp.txt', '-w', '%{http_code}'], capture_output=True, text=True)
code = r.stdout.strip()
if code in ('200', '204'):
    plugins = merged_global.get('enabled_plugins', ['lvms'])
    print(f'OK (plugins: {plugins})')
else:
    print(f'FAILED (HTTP {code})')
    print(open('/tmp/resp.txt').read())
    exit(1)
\"
"

info "Config written. Wizard is ready at https://$(echo "${TARGET}" | cut -d@ -f2):3443/wizard"
