Name:           enclave-wizard
Version:        0.0.1
Release:        1%{?dist}
Summary:        Install wizard for Red Hat Sovereign Enclave
License:        Apache-2.0
URL:            https://github.com/carbonin/enclave-wizard

Source0:        enclave-wizard
Source1:        enclave-wizard.service

Requires:       curl
Requires:       openssl

%description
Web-based install wizard for Red Hat Sovereign Enclave (RHSE).
Single binary that serves both the API and the embedded UI with TLS.
Pulls and extracts the enclave distribution via ORAS on first install.

%install
install -D -m 0755 %{SOURCE0} %{buildroot}/usr/local/bin/enclave-wizard
install -D -m 0644 %{SOURCE1} %{buildroot}/etc/systemd/system/enclave-wizard.service

%post
ORAS_VERSION="v1.2.2"
ENCLAVE_IMAGE="quay.io/edge-infrastructure/enclave:latest"
ENCLAVE_DIR="/opt/enclave"

# Install ORAS if not present
if ! /usr/local/bin/oras version &>/dev/null; then
    echo "Installing ORAS..."
    curl -sL "https://github.com/oras-project/oras/releases/download/${ORAS_VERSION}/oras_${ORAS_VERSION#v}_linux_amd64.tar.gz" \
        | tar xz -C /usr/local/bin oras
fi

# Pull and extract enclave
if [ ! -d "${ENCLAVE_DIR}/config" ]; then
    echo "Pulling enclave..."
    mkdir -p "${ENCLAVE_DIR}"
    TMPDIR=$(mktemp -d)
    cd "${TMPDIR}"
    /usr/local/bin/oras pull "${ENCLAVE_IMAGE}"
    tar xzf enclave.tar.gz -C "${ENCLAVE_DIR}"
    rm -rf "${TMPDIR}"

    # Create config from examples
    for f in "${ENCLAVE_DIR}/config/"*.example.yaml; do
        [ -f "${f}" ] || continue
        target="${f%.example.yaml}.yaml"
        [ -f "${target}" ] || cp "${f}" "${target}"
    done
fi

# Generate self-signed TLS certificate
TLS_DIR="/etc/enclave-wizard/tls"
if [ ! -f "${TLS_DIR}/server.crt" ]; then
    echo "Generating self-signed TLS certificate..."
    mkdir -p "${TLS_DIR}"
    CERT_HOST=$(hostname -f)
    openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
        -keyout "${TLS_DIR}/server.key" \
        -out "${TLS_DIR}/server.crt" \
        -subj "/CN=${CERT_HOST}" \
        -addext "subjectAltName=DNS:${CERT_HOST},DNS:localhost,IP:127.0.0.1" \
        2>/dev/null
    chmod 600 "${TLS_DIR}/server.key"
fi

# Open firewall ports
firewall-cmd --add-port=3001/tcp --permanent 2>/dev/null || true
firewall-cmd --add-port=3443/tcp --permanent 2>/dev/null || true
firewall-cmd --reload 2>/dev/null || true

# Start service
systemctl daemon-reload
systemctl enable --now enclave-wizard

echo ""
echo "Enclave Wizard installed and running."
echo "  UI:  https://$(hostname -f):3443/wizard"
echo "  (Self-signed cert — accept the browser warning)"
if [ -f /tmp/enclave-wizard-init-pass ]; then
    echo ""
    echo "  Admin password: $(cat /tmp/enclave-wizard-init-pass)"
    echo "  (You must change it on first login)"
    echo "  Check logs: journalctl -u enclave-wizard"
fi

%preun
if [ $1 -eq 0 ]; then
    systemctl stop enclave-wizard 2>/dev/null || true
    systemctl disable enclave-wizard 2>/dev/null || true
fi

%postun
if [ $1 -eq 0 ]; then
    systemctl daemon-reload
    rm -rf /opt/enclave
    firewall-cmd --remove-port=3001/tcp --permanent 2>/dev/null || true
    firewall-cmd --remove-port=3443/tcp --permanent 2>/dev/null || true
    rm -rf /etc/enclave-wizard/tls
    firewall-cmd --reload 2>/dev/null || true
fi

%files
/usr/local/bin/enclave-wizard
/etc/systemd/system/enclave-wizard.service
