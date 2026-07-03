#!/usr/bin/env bash
# Set up bare metal emulation infrastructure on a remote hypervisor.
# Creates libvirt network (with DNS), sushy-tools BMC emulator, and CP VMs.
# Does NOT deploy the wizard — use `make deploy TARGET=...` for that.
#
# Usage:
#   ./hack/infra/bm-emulation.sh --host root@hypervisor
#
# Prerequisites:
#   - SSH key access to the hypervisor host
#   - podman, libvirt, virt-install on the host
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TARGET=""
CLUSTER_NAME="${CLUSTER_NAME:-edge}"
BASE_DOMAIN="${BASE_DOMAIN:-enclave-test.lab.local}"
API_VIP="${API_VIP:-192.168.223.200}"
INGRESS_VIP="${INGRESS_VIP:-192.168.223.201}"

while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    --cluster-name) CLUSTER_NAME="$2"; shift 2 ;;
    --base-domain) BASE_DOMAIN="$2"; shift 2 ;;
    --api-vip) API_VIP="$2"; shift 2 ;;
    --ingress-vip) INGRESS_VIP="$2"; shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "${TARGET}" ]; then
  echo "Usage: $0 --host root@hypervisor"
  echo "  --host           SSH target (e.g. root@my-hypervisor)"
  echo "  --cluster-name   Cluster name (default: edge)"
  echo "  --base-domain    Base domain (default: enclave-test.lab.local)"
  exit 1
fi

SSH="ssh -o StrictHostKeyChecking=no"
SCP="scp -o StrictHostKeyChecking=no -q"
FQDN="${CLUSTER_NAME}.${BASE_DOMAIN}"
HOST_ADDR="$(echo "${TARGET}" | cut -d@ -f2)"

info()  { echo -e "\033[0;32m[INFO]\033[0m  $*"; }

# =============================================================================
# Step 1: Set up BM emulation (network, sushy-tools, VMs)
# =============================================================================
info "[1/2] Setting up bare metal emulation on ${HOST_ADDR}..."
${SCP} "${SCRIPT_DIR}/setup-bm-emulation.sh" "${TARGET}:/tmp/setup-bm-emulation.sh"
${SSH} "${TARGET}" "bash /tmp/setup-bm-emulation.sh \
  --cluster-name ${CLUSTER_NAME} \
  --base-domain ${BASE_DOMAIN} \
  --api-vip ${API_VIP} \
  --ingress-vip ${INGRESS_VIP}"

# =============================================================================
# Step 2: Collect info and print table
# =============================================================================
info "[2/2] Collecting infrastructure info..."

VM_UUIDS=""
for i in 0 1 2; do
  UUID=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-${i} 2>/dev/null" || echo "unknown")
  VM_UUIDS="${VM_UUIDS}enclave-cp-${i}|${UUID}
"
done

echo ""
echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                        BM Emulation Ready                                  ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
echo "║  Network                                                                   ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
printf "║  %-74s ║\n" "Machine Network:  192.168.223.0/24"
printf "║  %-74s ║\n" "Gateway/DNS:      192.168.223.1"
printf "║  %-74s ║\n" "API VIP:          ${API_VIP}"
printf "║  %-74s ║\n" "Ingress VIP:      ${INGRESS_VIP}"
printf "║  %-74s ║\n" "Rendezvous IP:    192.168.223.10"
echo "║                                                                            ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
echo "║  Cluster                                                                   ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
printf "║  %-74s ║\n" "Base Domain:   ${BASE_DOMAIN}"
printf "║  %-74s ║\n" "Cluster Name:  ${CLUSTER_NAME}"
printf "║  %-74s ║\n" "API:           api.${FQDN}"
printf "║  %-74s ║\n" "Console:       console-openshift-console.apps.${FQDN}"
echo "║                                                                            ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
echo "║  BMC / Redfish                                                             ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
printf "║  %-74s ║\n" "Endpoint:  192.168.223.1:8100"
printf "║  %-74s ║\n" "User:      admin"
printf "║  %-74s ║\n" "Password:  password"
echo "║                                                                            ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
echo "║  Control Plane VMs                                                         ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo ""
printf "  %-15s %-20s %-17s %s\n" "Name" "MAC" "IP" "Disk"
printf "  %-15s %-20s %-17s %s\n" "───────────────" "────────────────────" "─────────────────" "────────"

for i in 0 1 2; do
  VM_NAME="enclave-cp-${i}"
  VM_MAC="00:60:2f:e0:c1:$(printf '%02x' "$i")"
  VM_IP_ADDR="192.168.223.$((10 + i))"
  UUID=$(echo "${VM_UUIDS}" | grep "^${VM_NAME}|" | cut -d'|' -f2 | tr -d '[:space:]')
  printf "  %-15s %-20s %-17s %s\n" "${VM_NAME}" "${VM_MAC}" "${VM_IP_ADDR}" "/dev/sda"
  printf "  %-15s BMC: 192.168.223.1:8100  System ID: %s\n" "" "${UUID}"
done

PARAMS_FILE="${SCRIPT_DIR}/../../demo-params.json"
cat > "${PARAMS_FILE}" <<JSONEOF
{
  "infra": {
    "machineNetwork": "192.168.223.0/24",
    "gateway": "192.168.223.1",
    "apiVIP": "${API_VIP}",
    "ingressVIP": "${INGRESS_VIP}",
    "rendezvousIP": "192.168.223.10",
    "baseDomain": "${BASE_DOMAIN}",
    "clusterName": "${CLUSTER_NAME}",
    "defaultPrefix": 24,
    "bmc": {
      "endpoint": "192.168.223.1:8100",
      "user": "admin",
      "password": "password"
    },
    "hosts": [
      {"name": "enclave-cp-0", "mac": "00:60:2f:e0:c1:00", "ip": "192.168.223.10", "disk": "/dev/sda", "uuid": "$(echo "${VM_UUIDS}" | grep "^enclave-cp-0|" | cut -d'|' -f2 | tr -d '[:space:]')"},
      {"name": "enclave-cp-1", "mac": "00:60:2f:e0:c1:01", "ip": "192.168.223.11", "disk": "/dev/sda", "uuid": "$(echo "${VM_UUIDS}" | grep "^enclave-cp-1|" | cut -d'|' -f2 | tr -d '[:space:]')"},
      {"name": "enclave-cp-2", "mac": "00:60:2f:e0:c1:02", "ip": "192.168.223.12", "disk": "/dev/sda", "uuid": "$(echo "${VM_UUIDS}" | grep "^enclave-cp-2|" | cut -d'|' -f2 | tr -d '[:space:]')"}
    ]
  }
}
JSONEOF
info "Params written to $(realpath "${PARAMS_FILE}")"

echo ""
info "Next: deploy the wizard with 'make deploy TARGET=${TARGET}'"
info "Then open the wizard UI and fill in the config from the table above."
info ""
info "To tear down:  make bm-emulation-cleanup TARGET=${TARGET}"
