#!/usr/bin/env bash
# End-to-end deployment with recording.
# Builds RPMs, deploys wizard VM, sets up BM emulation, runs deployment,
# and downloads recorded fixtures.
#
# Usage:
#   ./hack/infra/deploy-with-recording.sh --host root@hypervisor --pull-secret /path/to/pull-secret.json
#   ./hack/infra/deploy-with-recording.sh --host root@hypervisor --pull-secret /path/to/pull-secret.json --interactive
#
# Flags:
#   --interactive   Open a headed browser and pause at Review so you can deploy manually
#
# Prerequisites:
#   - SSH key access to the hypervisor host
#   - A valid OpenShift pull secret (from console.redhat.com)
#   - podman, libvirt, virt-install on the host
#   - Node.js with npx and Playwright installed (for filling the wizard)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

TARGET=""
PULL_SECRET=""
CLUSTER_NAME="${CLUSTER_NAME:-edge}"
BASE_DOMAIN="${BASE_DOMAIN:-enclave-test.lab.local}"
API_VIP="${API_VIP:-192.168.223.200}"
INGRESS_VIP="${INGRESS_VIP:-192.168.223.201}"
SKIP_BUILD="${SKIP_BUILD:-false}"
SKIP_DEPLOY="${SKIP_DEPLOY:-false}"
SKIP_INFRA="${SKIP_INFRA:-false}"
INTERACTIVE="${INTERACTIVE:-false}"

while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    --pull-secret) PULL_SECRET="$2"; shift 2 ;;
    --cluster-name) CLUSTER_NAME="$2"; shift 2 ;;
    --base-domain) BASE_DOMAIN="$2"; shift 2 ;;
    --skip-build) SKIP_BUILD=true; shift ;;
    --skip-deploy) SKIP_DEPLOY=true; shift ;;
    --skip-infra) SKIP_INFRA=true; shift ;;
    --interactive) INTERACTIVE=true; shift ;;
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
  "${REPO_DIR}/hack/e2e/run-e2e.sh" --host "${TARGET}" --skip-teardown || \
    info "  E2E tests had failures (non-fatal — wizard VM is deployed)"
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

# Attach enclave-bmc NIC if not already present
${SSH} "${TARGET}" "
  if ! virsh domiflist enclave-wizard-lz | grep -q enclave-bmc; then
    virsh attach-interface enclave-wizard-lz network enclave-bmc --model virtio --config --live
    echo 'BMC NIC attached'
  else
    echo 'BMC NIC already attached'
  fi
"

# Expand disk if needed
DISK_AVAIL=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "df --output=avail / | tail -1" 2>/dev/null | tr -d ' ')
if [ "${DISK_AVAIL:-0}" -lt 20000000 ]; then
  info "  Expanding VM disk to 60GB..."
  ${SSH} "${TARGET}" "virsh blockresize enclave-wizard-lz /var/lib/libvirt/images/enclave-wizard-lz.qcow2 60G"
  ${SSH} -J "${TARGET}" wizard@"${VM_IP}" "sudo growpart /dev/vda 1 2>/dev/null; sudo xfs_growfs / 2>/dev/null || sudo resize2fs /dev/vda1 2>/dev/null" || true
fi

# Activate the BMC NIC inside the VM and configure DNS
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  # Find the BMC interface (second ethernet device, not eth0)
  BMC_DEV=\$(nmcli -t -f DEVICE,TYPE dev status | grep ethernet | grep -v eth0 | head -1 | cut -d: -f1)
  if [ -n \"\$BMC_DEV\" ]; then
    sudo nmcli dev connect \"\$BMC_DEV\" 2>/dev/null || true
    sleep 3
    CON=\$(nmcli -t -f NAME,DEVICE con show --active | grep \"\$BMC_DEV\" | cut -d: -f1)
    if [ -n \"\$CON\" ]; then
      sudo nmcli con mod \"\$CON\" ipv4.dns '192.168.223.1' ipv4.dns-priority -10
      sudo nmcli con up \"\$CON\"
    fi
  fi
" 2>/dev/null || true

# Get the VM's IP on the BMC network (retry — DHCP may need time)
BMC_VM_IP=""
for attempt in $(seq 1 18); do
  BMC_VM_IP=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
    ip -4 addr show 2>/dev/null | grep -A2 'state UP' | grep -v '192.168.122' | grep -oP 'inet \K[0-9.]+' | head -1
  " 2>/dev/null || echo "")
  [ -n "${BMC_VM_IP}" ] && break
  info "  Waiting for BMC NIC IP (attempt ${attempt}/18)..."
  sleep 5
done
info "  Wizard VM BMC IP: ${BMC_VM_IP:-unknown}"
[ -z "${BMC_VM_IP}" ] && error "Wizard VM did not get a BMC network IP after 90s"

# =============================================================================
# Step 5: Install extra deps on wizard VM
# =============================================================================
info "[5/8] Verifying deployment dependencies on wizard VM..."
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  sudo systemctl enable --now httpd 2>/dev/null || true
" 2>/dev/null

# =============================================================================
# Step 6: Prepare demo-params.json and wizard auth
# =============================================================================
info "[6/8] Updating demo-params.json..."

WIZARD_HOST="$(echo "${TARGET}" | cut -d@ -f2)"
WIZARD_URL="https://${WIZARD_HOST}:3443"
PARAMS_FILE="${REPO_DIR}/demo-params.json"

# Copy pull secret to hypervisor (Ansible needs it during deploy)
${SCP} "${PULL_SECRET}" "${TARGET}:/tmp/enclave-pull-secret.json"
${SSH} "${TARGET}" "scp -o StrictHostKeyChecking=no /tmp/enclave-pull-secret.json wizard@${VM_IP}:/tmp/"

# Discover VM UUIDs from sushy-tools on the hypervisor
VM_UUIDS=$(${SSH} "${TARGET}" "python3 -c \"
import json, urllib.request, ssl
ctx = ssl.create_default_context()
ctx.check_hostname = False
ctx.verify_mode = ssl.CERT_NONE
result = []
resp = urllib.request.urlopen('https://192.168.223.1:8100/redfish/v1/Systems', context=ctx)
for m in json.loads(resp.read())['Members']:
    sid = m['@odata.id'].split('/')[-1]
    r2 = urllib.request.urlopen(f'https://192.168.223.1:8100/redfish/v1/Systems/{sid}', context=ctx)
    info = json.loads(r2.read())
    if 'enclave-cp' in info['Name']:
        result.append({'name': info['Name'], 'uuid': sid})
print(json.dumps(sorted(result, key=lambda x: x['name'])))
\"")

# Reset wizard auth for fresh deploy
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  sudo rm -f /etc/enclave-wizard/password /etc/enclave-wizard/.password-changed /tmp/enclave-wizard-init-pass
  sudo systemctl restart enclave-wizard
  sleep 2
"

PASS=$(${SSH} -J "${TARGET}" wizard@"${VM_IP}" "sudo cat /tmp/enclave-wizard-init-pass")
info "  Wizard password: ${PASS}"

# Write demo-params.json locally
python3 -c "
import json, os

vm_info = json.loads('${VM_UUIDS}')
uuid_by_name = {v['name']: v['uuid'] for v in vm_info}

hosts = []
for i in range(3):
    name = f'enclave-cp-{i}'
    hosts.append({
        'name': name,
        'mac': f'00:60:2f:e0:c1:{i:02x}',
        'ip': f'192.168.223.{10+i}',
        'disk': '/dev/sda',
        'uuid': uuid_by_name.get(name, ''),
    })

data = {}
path = '${PARAMS_FILE}'
if os.path.exists(path):
    with open(path) as f:
        data = json.load(f)

data['infra'] = {
    'machineNetwork': '192.168.223.0/24',
    'gateway': '192.168.223.1',
    'apiVIP': '${API_VIP}',
    'ingressVIP': '${INGRESS_VIP}',
    'rendezvousIP': '192.168.223.10',
    'baseDomain': '${BASE_DOMAIN}',
    'clusterName': '${CLUSTER_NAME}',
    'defaultPrefix': 24,
    'bmc': {
        'endpoint': '192.168.223.1:8100',
        'user': 'admin',
        'password': 'password',
    },
    'hosts': hosts,
}
data['wizard'] = {
    'url': '${WIZARD_URL}/wizard',
    'password': '${PASS}',
    'vmIP': '${VM_IP}',
    'lzBmcIP': '${BMC_VM_IP}',
    'target': '${TARGET}',
    'sshJump': 'ssh -J ${TARGET} wizard@${VM_IP}',
}

with open(path, 'w') as f:
    json.dump(data, f, indent=2)
print(f'  demo-params.json updated with {len(hosts)} hosts, {len(uuid_by_name)} UUIDs')
"

# =============================================================================
# Step 7: Fill wizard via Playwright and trigger deployment
# =============================================================================
E2E_DIR="${REPO_DIR}/ui/apps/wizard/e2e"

run_playwright() {
  local extra_env="$1"
  shift
  if command -v npx &>/dev/null; then
    eval "${extra_env}" npx "$@"
  else
    distrobox enter osac -- bash -c "export PATH=~/.local/bin:\$PATH && cd ${E2E_DIR} && ${extra_env} npx $*"
  fi
}

if [ "${INTERACTIVE}" = "true" ]; then
  info "[7/8] Filling wizard via Playwright (interactive — headed browser)..."
  cd "${E2E_DIR}"
  run_playwright \
    "WIZARD_URL=${WIZARD_URL} WIZARD_PASSWORD=${PASS} PULL_SECRET=${PULL_SECRET}" \
    playwright test --config playwright.config.ts \
      --headed \
      --grep "fill wizard from demo-params" \
      --reporter=list
  info ""
  info "=== Interactive mode — deploy manually in the browser ==="
  info "Wizard URL: ${WIZARD_URL}/wizard"
  info "Wizard password: ${PASS}"
  exit 0
fi

info "[7/8] Filling wizard via Playwright and starting deployment..."
cd "${E2E_DIR}"
run_playwright \
  "WIZARD_URL=${WIZARD_URL} WIZARD_PASSWORD=${PASS} PULL_SECRET=${PULL_SECRET} SKIP_PAUSE=1" \
  playwright test --config playwright.config.ts \
    --grep "fill wizard from demo-params" \
    --reporter=list
info "  Config written via Playwright"

# Clean up stale state and trigger deploy via API
${SSH} -J "${TARGET}" wizard@"${VM_IP}" "
  set -euo pipefail

  sudo podman pod rm -f metal3-ironic 2>/dev/null || true
  sudo rm -rf /opt/enclave/ocp-cluster/
  sudo rm -f /opt/enclave/config/pull-secret.json

  API='https://localhost:3443'
  TOKEN=\$(curl -sk -X POST \"\$API/api/v1/auth/login\" \
    -H 'Content-Type: application/json' \
    -d '{\"username\":\"admin\",\"password\":\"pleaseletmein\"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)[\"token\"])')

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
info "Wizard URL: ${WIZARD_URL}/wizard"
info "Wizard password: ${PASS}"
