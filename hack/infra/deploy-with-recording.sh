#!/usr/bin/env bash
# End-to-end deployment with recording.
# Builds RPMs, deploys wizard VM, sets up BM emulation, runs deployment,
# and downloads recorded fixtures.
#
# Usage:
#   ./hack/infra/deploy-with-recording.sh --host root@hypervisor --pull-secret /path/to/pull-secret.json
#
# Prerequisites:
#   - SSH key access to the hypervisor host
#   - A valid OpenShift pull secret (from console.redhat.com)
#   - podman, libvirt, virt-install on the host
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

TARGET=""
PULL_SECRET=""
CLUSTER_NAME="${CLUSTER_NAME:-mgmt}"
BASE_DOMAIN="${BASE_DOMAIN:-enclave-test.lab.local}"
API_VIP="${API_VIP:-192.168.223.200}"
INGRESS_VIP="${INGRESS_VIP:-192.168.223.201}"
SKIP_BUILD="${SKIP_BUILD:-false}"
SKIP_DEPLOY="${SKIP_DEPLOY:-false}"
SKIP_INFRA="${SKIP_INFRA:-false}"

while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    --pull-secret) PULL_SECRET="$2"; shift 2 ;;
    --cluster-name) CLUSTER_NAME="$2"; shift 2 ;;
    --base-domain) BASE_DOMAIN="$2"; shift 2 ;;
    --skip-build) SKIP_BUILD=true; shift ;;
    --skip-deploy) SKIP_DEPLOY=true; shift ;;
    --skip-infra) SKIP_INFRA=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "${TARGET}" ]; then
  echo "Usage: $0 --host root@hypervisor --pull-secret /path/to/pull-secret.json"
  exit 1
fi

if [ -z "${PULL_SECRET}" ] || [ ! -f "${PULL_SECRET}" ]; then
  echo "ERROR: --pull-secret must point to a valid pull secret JSON file"
  exit 1
fi

SSH="ssh -o StrictHostKeyChecking=no"
SCP="scp -o StrictHostKeyChecking=no -q"
FQDN="${CLUSTER_NAME}.${BASE_DOMAIN}"

info()  { echo -e "\033[0;32m[INFO]\033[0m  $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*"; exit 1; }

# =============================================================================
# Step 1: Build RPMs with dev tags
# =============================================================================
if [ "${SKIP_BUILD}" != "true" ]; then
  info "[1/8] Building RPMs with dev tags..."
  make -C "${REPO_DIR}" rpm TAGS=dev
else
  info "[1/8] Skipping build (--skip-build)"
fi

# =============================================================================
# Step 2: Deploy wizard VM
# =============================================================================
if [ "${SKIP_DEPLOY}" != "true" ]; then
  info "[2/8] Deploying wizard VM..."
  "${REPO_DIR}/hack/e2e/run-e2e.sh" --host "${TARGET}" --skip-teardown
else
  info "[2/8] Skipping wizard deploy (--skip-deploy)"
fi

VM_IP=$(${SSH} "${TARGET}" "cat /tmp/enclave-wizard-vm-ip")
info "  Wizard VM IP: ${VM_IP}"

# =============================================================================
# Step 3: Set up BM emulation on host
# =============================================================================
if [ "${SKIP_INFRA}" != "true" ]; then
  info "[3/8] Setting up bare metal emulation on host..."
  ${SCP} "${SCRIPT_DIR}/setup-bm-emulation.sh" "${TARGET}:/tmp/setup-bm-emulation.sh"
  ${SSH} "${TARGET}" "bash /tmp/setup-bm-emulation.sh \
    --cluster-name ${CLUSTER_NAME} \
    --base-domain ${BASE_DOMAIN} \
    --api-vip ${API_VIP} \
    --ingress-vip ${INGRESS_VIP}"
else
  info "[3/8] Skipping infra setup (--skip-infra)"
fi

# =============================================================================
# Step 4: Add BMC network NIC to wizard VM
# =============================================================================
info "[4/8] Configuring wizard VM networking..."

# Check if already has enclave-bmc NIC
HAS_BMC_NIC=$(${SSH} "${TARGET}" "virsh domiflist enclave-wizard-lz | grep -c enclave-bmc" 2>/dev/null || echo "0")
if [ "${HAS_BMC_NIC}" = "0" ]; then
  ${SSH} "${TARGET}" "virsh attach-interface enclave-wizard-lz network enclave-bmc --model virtio --config --live"
  sleep 5
fi

# Expand disk if needed
DISK_AVAIL=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "df --output=avail / | tail -1" 2>/dev/null | tr -d ' ')
if [ "${DISK_AVAIL:-0}" -lt 20000000 ]; then
  info "  Expanding VM disk to 60GB..."
  ${SSH} "${TARGET}" "virsh blockresize enclave-wizard-lz /var/lib/libvirt/images/enclave-wizard-lz.qcow2 60G"
  ${SSH} -J "${TARGET}" wizard@"${VM_IP}" "sudo growpart /dev/vda 1 2>/dev/null; sudo xfs_growfs / 2>/dev/null || sudo resize2fs /dev/vda1 2>/dev/null" || true
fi

# Configure DNS on the BMC interface
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  sudo nmcli con mod 'Wired connection 2' ipv4.dns '192.168.223.1' 2>/dev/null || true
" 2>/dev/null || true

# Get the VM's IP on the BMC network
BMC_VM_IP=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "ip -4 addr show eth1 2>/dev/null | grep -oP 'inet \K[0-9.]+'" 2>/dev/null || echo "")
info "  Wizard VM BMC IP: ${BMC_VM_IP:-unknown}"

# =============================================================================
# Step 5: Install extra deps on wizard VM
# =============================================================================
info "[5/8] Installing deployment dependencies on wizard VM..."
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  sudo dnf install -y nmstate podman httpd 2>&1 | tail -3
  sudo ansible-galaxy collection install -p /usr/share/ansible/collections containers.podman 2>&1 | tail -2
  sudo systemctl enable --now httpd 2>/dev/null || true
" 2>/dev/null

# =============================================================================
# Step 6: Build and write config
# =============================================================================
info "[6/8] Writing deployment config with pull secret..."

# Copy pull secret to host, then build config JSON there
${SCP} "${PULL_SECRET}" "${TARGET}:/tmp/enclave-pull-secret.json"

${SSH} "${TARGET}" "python3 -c \"
import json, urllib.request, ssl

ctx = ssl.create_default_context()
ctx.check_hostname = False
ctx.verify_mode = ssl.CERT_NONE

ps = json.load(open('/tmp/enclave-pull-secret.json'))

# Discover VM UUIDs from sushy-tools
name_to_uuid = {}
resp = urllib.request.urlopen('https://192.168.223.1:8100/redfish/v1/Systems', context=ctx)
for m in json.loads(resp.read())['Members']:
    sid = m['@odata.id'].split('/')[-1]
    r2 = urllib.request.urlopen(f'https://192.168.223.1:8100/redfish/v1/Systems/{sid}', context=ctx)
    info = json.loads(r2.read())
    if 'enclave-cp' in info['Name']:
        name_to_uuid[info['Name']] = sid

hosts = []
for i in range(3):
    name = f'enclave-cp-{i}'
    hosts.append({
        'name': name,
        'macAddress': f'00:60:2f:e0:c1:{i:02x}',
        'ipAddress': f'192.168.223.{10+i}',
        'redfish': '192.168.223.1:8100',
        'redfishUser': 'admin',
        'redfishPassword': 'password',
        'rootDisk': '/dev/sda',
    })

config = {
    'global': {
        'workingDir': '/opt/enclave',
        'baseDomain': '${BASE_DOMAIN}',
        'clusterName': '${CLUSTER_NAME}',
        'machineNetwork': '192.168.223.0/24',
        'apiVIP': '${API_VIP}',
        'ingressVIP': '${INGRESS_VIP}',
        'rendezvousIP': '192.168.223.10',
        'defaultDNS': '192.168.223.1',
        'defaultGateway': '192.168.223.1',
        'defaultPrefix': 24,
        'lzBmcIP': '${BMC_VM_IP}',
        'quayUser': 'admin',
        'quayPassword': 'password',
        'quayBackend': 'LocalStorage',
        'storage_plugin': 'lvms',
        'disconnected': False,
        'enabled_plugins': ['lvms'],
        'pullSecret': ps,
        'sshPubPath': '/home/wizard/.ssh/id_rsa.pub',
        'agent_hosts': hosts,
    },
    'certificates': {},
    'cloudInfra': {'discovery_hosts': []}
}
json.dump(config, open('/tmp/enclave-wizard-config.json', 'w'))

# Also create a version with extras for patching global.yaml
extras = {
    'fresh': False,
    'openshift_versions': [{'version': '4.20.21', 'default': True}],
    'bmcSystemIds': name_to_uuid,
}
json.dump(extras, open('/tmp/enclave-wizard-extras.json', 'w'))

print(f'Config created with {len(ps[\"auths\"])} registries, {len(name_to_uuid)} VM UUIDs')
\""

# Copy config to VM and write via API + patch
${SSH} "${TARGET}" "scp -o StrictHostKeyChecking=no /tmp/enclave-wizard-config.json /tmp/enclave-wizard-extras.json wizard@${VM_IP}:/tmp/"

# Reset wizard auth for fresh deploy
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  sudo rm -f /etc/enclave-wizard/password /etc/enclave-wizard/.password-changed /tmp/enclave-wizard-init-pass
  sudo systemctl restart enclave-wizard
  sleep 2
"

PASS=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "sudo cat /tmp/enclave-wizard-init-pass")
info "  Wizard password: ${PASS}"

${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  set -euo pipefail
  API='https://localhost:3443'

  # Login
  TOKEN=\$(curl -sk -X POST \"\$API/api/v1/auth/login\" \
    -H 'Content-Type: application/json' \
    -d '{\"username\":\"admin\",\"password\":\"${PASS}\"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)[\"token\"])')

  # Strip non-API fields and write config
  python3 -c '
import json
c = json.load(open(\"/tmp/enclave-wizard-config.json\"))
g = c[\"global\"]
g.pop(\"fresh\", None)
g.pop(\"openshift_versions\", None)
for h in g.get(\"agent_hosts\", []):
    h.pop(\"bmcSystemId\", None)
json.dump(c, open(\"/tmp/enclave-wizard-config-api.json\", \"w\"))
'

  HTTP=\$(curl -sk -X PUT \"\$API/api/v1/config\" \
    -H \"Authorization: Bearer \$TOKEN\" \
    -H 'Content-Type: application/json' \
    -d @/tmp/enclave-wizard-config-api.json -o /dev/null -w '%{http_code}')
  echo \"API config write: HTTP \$HTTP\"

  # Patch global.yaml with extras
  sudo python3 -c '
import yaml, json

with open(\"/opt/enclave/config/global.yaml\") as f:
    d = yaml.safe_load(f)

extras = json.load(open(\"/tmp/enclave-wizard-extras.json\"))
d[\"fresh\"] = extras[\"fresh\"]
d[\"openshift_versions\"] = extras[\"openshift_versions\"]

for h in d.get(\"agent_hosts\", []):
    if h[\"name\"] in extras[\"bmcSystemIds\"]:
        h[\"bmcSystemId\"] = extras[\"bmcSystemIds\"][h[\"name\"]]

with open(\"/opt/enclave/config/global.yaml\", \"w\") as f:
    yaml.dump(d, f, default_flow_style=False)
print(\"global.yaml patched\")
'
"

# =============================================================================
# Step 7: Trigger deployment
# =============================================================================
info "[7/8] Starting enclave deployment..."

${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  set -euo pipefail

  # Clean up any previous state
  sudo podman pod rm -f metal3-ironic 2>/dev/null || true
  sudo rm -rf /opt/enclave/ocp-cluster/
  sudo rm -f /opt/enclave/config/pull-secret.json

  API='https://localhost:3443'
  PASS=\$(sudo cat /tmp/enclave-wizard-init-pass)
  TOKEN=\$(curl -sk -X POST \"\$API/api/v1/auth/login\" \
    -H 'Content-Type: application/json' \
    -d \"{\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"\$PASS\\\"}\" | python3 -c 'import sys,json; print(json.load(sys.stdin)[\"token\"])')

  DEPLOY=\$(curl -sk -X POST \"\$API/api/v1/tasks/deploy\" \
    -H \"Authorization: Bearer \$TOKEN\")
  RUN_ID=\$(echo \"\$DEPLOY\" | python3 -c 'import sys,json; print(json.load(sys.stdin)[\"id\"])')
  echo \"Deploy started: \$RUN_ID\"

  for i in \$(seq 1 10800); do
    sleep 5
    STATUS=\$(curl -sk \"\$API/api/v1/tasks/\$RUN_ID\" -H \"Authorization: Bearer \$TOKEN\" | python3 -c 'import sys,json; print(json.load(sys.stdin)[\"status\"])')
    if [ \"\$STATUS\" != 'running' ]; then
      echo \"Deploy finished: \$STATUS (after ~\$((i*5))s)\"
      break
    fi
    [ \$((i % 120)) -eq 0 ] && echo \"  Still running (~\$((i*5/60))min)...\"
  done

  echo ''
  echo '=== Recordings ==='
  ls -lh /opt/enclave/fixtures/recordings/
"

# =============================================================================
# Step 8: Download recordings
# =============================================================================
info "[8/8] Downloading recordings..."

RECORDINGS_DIR="${REPO_DIR}/fixtures/recordings"
mkdir -p "${RECORDINGS_DIR}"

for f in $(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "ls /opt/enclave/fixtures/recordings/*.json 2>/dev/null | xargs -n1 basename"); do
  ${SSH} -J "${TARGET}" wizard@"${VM_IP}" "sudo cat /opt/enclave/fixtures/recordings/${f}" > "${RECORDINGS_DIR}/${f}"
  SIZE=$(wc -c < "${RECORDINGS_DIR}/${f}")
  info "  Saved ${f} (${SIZE} bytes)"
done

info ""
info "=== Deployment recording complete ==="
info "Recordings saved to fixtures/recordings/"
info "Wizard URL: https://$(echo "${TARGET}" | cut -d@ -f2):3443/wizard"
info "Wizard password: ${PASS}"
