#!/usr/bin/env bash
set -euo pipefail

ENCLAVE_DIR="/opt/enclave"
QUADLET_DIR="/etc/containers/systemd"
SYSTEMD_DIR="/etc/systemd/system"
ORAS_VERSION="v1.2.2"
ENCLAVE_IMAGE="quay.io/edge-infrastructure/enclave:latest"

# --- Handle --remove ---
if [ "${1:-}" = "--remove" ]; then
    echo "=== Removing Enclave Wizard ==="
    systemctl stop enclave-wizard-ui 2>/dev/null || true
    systemctl stop enclave-wizard-api 2>/dev/null || true
    systemctl disable enclave-wizard-api 2>/dev/null || true
    rm -f "${QUADLET_DIR}/enclave-wizard-"*.container
    rm -f "${QUADLET_DIR}/enclave-wizard-"*.network
    rm -f "${SYSTEMD_DIR}/enclave-wizard-api.service"
    systemctl daemon-reload
    rm -f /usr/local/bin/enclave-wizard
    rm -rf "${ENCLAVE_DIR}"
    echo "Removed."
    exit 0
fi

echo "=== Provisioning Enclave Wizard ==="

# --- Install ORAS if not present ---
if ! command -v oras &>/dev/null; then
    echo "Installing ORAS..."
    curl -sL "https://github.com/oras-project/oras/releases/download/${ORAS_VERSION}/oras_${ORAS_VERSION#v}_linux_amd64.tar.gz" \
        | tar xz -C /usr/local/bin oras
fi

# --- Pull and extract enclave ---
echo "Pulling enclave from ${ENCLAVE_IMAGE}..."
mkdir -p "${ENCLAVE_DIR}"
TMPDIR=$(mktemp -d)
cd "${TMPDIR}"
oras pull "${ENCLAVE_IMAGE}"
tar xzf enclave.tar.gz -C "${ENCLAVE_DIR}"
rm -rf "${TMPDIR}"
echo "Enclave extracted to ${ENCLAVE_DIR}"

# --- Create config directory from examples ---
if [ -d "${ENCLAVE_DIR}/config" ]; then
    for f in "${ENCLAVE_DIR}/config/"*.example.yaml; do
        [ -f "${f}" ] || continue
        target="${f%.example.yaml}.yaml"
        if [ ! -f "${target}" ]; then
            cp "${f}" "${target}"
        fi
    done
    echo "Config files created from examples"
fi

# --- Install systemd service for API (native binary) ---
echo "Installing wizard-api systemd service..."
cp /tmp/quadlets/enclave-wizard-api.service "${SYSTEMD_DIR}/"

# --- Install nginx config for UI ---
echo "Installing nginx config..."
mkdir -p /etc/enclave-wizard
cp /tmp/nginx-deploy.conf /etc/enclave-wizard/nginx.conf

# --- Install quadlet for UI (container) ---
echo "Installing wizard-ui quadlet..."
mkdir -p "${QUADLET_DIR}"
cp /tmp/quadlets/enclave-wizard-ui.container "${QUADLET_DIR}/"

# --- Open firewall ports ---
echo "Opening firewall ports..."
firewall-cmd --add-port=3001/tcp --permanent 2>/dev/null || true
firewall-cmd --add-port=8080/tcp --permanent 2>/dev/null || true
firewall-cmd --reload 2>/dev/null || true

# --- Stop old services if running ---
systemctl stop enclave-wizard-ui 2>/dev/null || true
systemctl stop enclave-wizard-api 2>/dev/null || true

# --- Reload and start ---
echo "Starting services..."
systemctl daemon-reload
systemctl enable --now enclave-wizard-api
sleep 2
systemctl start enclave-wizard-ui

echo ""
echo "=== Service status ==="
systemctl is-active enclave-wizard-api && echo "  wizard-api: running (native)" || echo "  wizard-api: FAILED"
systemctl is-active enclave-wizard-ui && echo "  wizard-ui: running (container)" || echo "  wizard-ui: FAILED"
echo ""
echo "=== Provisioning complete ==="
