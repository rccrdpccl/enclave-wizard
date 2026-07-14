#!/usr/bin/env bash
# Tear down bare metal emulation infrastructure on a remote hypervisor.
# Removes CP VMs, libvirt network, and sushy-tools.
# Does NOT remove the wizard VM — use `make teardown TARGET=...` for that.
#
# Usage:
#   ./hack/infra/bm-emulation-cleanup.sh --host root@hypervisor
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TARGET=""

while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "${TARGET}" ]; then
  echo "Usage: $0 --host root@hypervisor"
  exit 1
fi

SSH="ssh -o StrictHostKeyChecking=no"
SCP="scp -o StrictHostKeyChecking=no -q"

info()  { echo -e "\033[0;32m[INFO]\033[0m  $*"; }

info "Tearing down bare metal emulation..."
${SCP} "${SCRIPT_DIR}/teardown-bm-emulation.sh" "${TARGET}:/tmp/teardown-bm-emulation.sh"
${SSH} "${TARGET}" "bash /tmp/teardown-bm-emulation.sh"
${SSH} "${TARGET}" "rm -f /tmp/setup-bm-emulation.sh /tmp/teardown-bm-emulation.sh" 2>/dev/null || true

info ""
info "BM emulation removed (CP VMs, network, sushy-tools)."
info "Wizard VM is still running — use 'make teardown TARGET=${TARGET}' to remove it."
