#!/usr/bin/env bash
# Tear down bare metal emulation infrastructure.
# Destroys VMs, removes libvirt network, stops sushy-tools.
# Safe to run when nothing exists.
set -euo pipefail

VM_PREFIX="${VM_PREFIX:-enclave-cp}"
NUM_MASTERS="${NUM_MASTERS:-3}"
NET_NAME="${NET_NAME:-enclave-bmc}"
SUSHY_CONTAINER="${SUSHY_CONTAINER:-enclave-sushy-tools}"

info()  { echo -e "\033[0;32m[INFO]\033[0m  $*" >&2; }

info "Tearing down bare metal emulation..."

# Destroy VMs
for i in $(seq 0 $((NUM_MASTERS - 1))); do
  VM_NAME="${VM_PREFIX}-${i}"
  if virsh dominfo "${VM_NAME}" &>/dev/null; then
    virsh destroy "${VM_NAME}" 2>/dev/null || true
    virsh undefine "${VM_NAME}" --remove-all-storage --nvram 2>/dev/null || true
    info "  Removed ${VM_NAME}"
  fi
done

# Stop sushy-tools
if podman ps -a --format '{{.Names}}' | grep -q "^${SUSHY_CONTAINER}$"; then
  podman stop "${SUSHY_CONTAINER}" 2>/dev/null || true
  podman rm "${SUSHY_CONTAINER}" 2>/dev/null || true
  info "  Removed ${SUSHY_CONTAINER}"
fi

# Remove network
if virsh net-info "${NET_NAME}" &>/dev/null; then
  virsh net-destroy "${NET_NAME}" 2>/dev/null || true
  virsh net-undefine "${NET_NAME}" 2>/dev/null || true
  info "  Removed network ${NET_NAME}"
fi

info "Teardown complete."
