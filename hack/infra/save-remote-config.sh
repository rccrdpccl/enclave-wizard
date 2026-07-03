#!/usr/bin/env bash
# Save wizard VM config files to hack/infra/saved-config/
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAVE_DIR="${SCRIPT_DIR}/saved-config"
TARGET=""

while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    *) echo "Usage: $0 --host root@hypervisor"; exit 1 ;;
  esac
done

[ -z "${TARGET}" ] && echo "Usage: $0 --host root@hypervisor" && exit 1

SSH="ssh -o StrictHostKeyChecking=no"
VM_IP=$(${SSH} "${TARGET}" "cat /tmp/enclave-wizard-vm-ip 2>/dev/null || virsh domifaddr enclave-wizard-lz --source agent 2>/dev/null | grep -oP '192\.168\.122\.\d+' | head -1")
[ -z "${VM_IP}" ] && echo "ERROR: could not find wizard VM IP" && exit 1

echo "Saving config from ${TARGET} (VM: ${VM_IP})..."

mkdir -p "${SAVE_DIR}/plugins"

for f in global.yaml certificates.yaml cloud_infra.yaml; do
  ${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo cat /opt/enclave/config/${f}' 2>/dev/null" > "${SAVE_DIR}/${f}"
  echo "  ${f}"
done

for f in osac.yaml rhbk.yaml; do
  ${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo cat /opt/enclave/config/plugins/${f} 2>/dev/null || echo \"# not present\"' 2>/dev/null" > "${SAVE_DIR}/plugins/${f}"
  echo "  plugins/${f}"
done

echo "Saved to ${SAVE_DIR}/"
