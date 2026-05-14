Name:           enclave-wizard
Version:        0.0.1
Release:        1%{?dist}
Summary:        Install wizard for Red Hat Sovereign Enclave
License:        Apache-2.0
URL:            https://github.com/carbonin/enclave-wizard

Source0:        enclave-wizard
Source1:        wizard-ui.tar
Source2:        enclave-wizard-api.service
Source3:        enclave-wizard-ui.container
Source4:        nginx-deploy.conf

Requires:       podman
Requires:       curl
Requires:       openssl

%description
Web-based install wizard for Red Hat Sovereign Enclave (RHSE).
Installs and configures the wizard API (native Go binary) and
wizard UI (nginx container via podman quadlet). Pulls and extracts
the enclave distribution via ORAS on first install.

%install
install -D -m 0755 %{SOURCE0} %{buildroot}/usr/local/bin/enclave-wizard
install -D -m 0644 %{SOURCE1} %{buildroot}/usr/share/enclave-wizard/wizard-ui.tar
install -D -m 0644 %{SOURCE2} %{buildroot}/etc/systemd/system/enclave-wizard-api.service
install -D -m 0644 %{SOURCE3} %{buildroot}/etc/containers/systemd/enclave-wizard-ui.container
install -D -m 0644 %{SOURCE4} %{buildroot}/etc/enclave-wizard/nginx.conf

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

# Load UI container image
echo "Loading wizard-ui container image..."
podman load -i /usr/share/enclave-wizard/wizard-ui.tar 2>/dev/null || true

# Open firewall ports
firewall-cmd --add-port=3001/tcp --permanent 2>/dev/null || true
firewall-cmd --add-port=8080/tcp --permanent 2>/dev/null || true
firewall-cmd --reload 2>/dev/null || true

# Start services
systemctl daemon-reload
systemctl enable --now enclave-wizard-api
sleep 2
systemctl start enclave-wizard-ui

echo ""
echo "Enclave Wizard installed and running."
echo "  UI:  http://$(hostname):3001/wizard"
echo "  API: http://localhost:8080/api/v1/defaults"

%preun
if [ $1 -eq 0 ]; then
    systemctl stop enclave-wizard-ui 2>/dev/null || true
    systemctl stop enclave-wizard-api 2>/dev/null || true
    systemctl disable enclave-wizard-api 2>/dev/null || true
fi

%postun
if [ $1 -eq 0 ]; then
    systemctl daemon-reload
    podman rmi -f localhost/enclave-wizard-ui:dev 2>/dev/null || true
    rm -rf /opt/enclave
    firewall-cmd --remove-port=3001/tcp --permanent 2>/dev/null || true
    firewall-cmd --remove-port=8080/tcp --permanent 2>/dev/null || true
    firewall-cmd --reload 2>/dev/null || true
fi

%files
/usr/local/bin/enclave-wizard
/usr/share/enclave-wizard/wizard-ui.tar
/etc/systemd/system/enclave-wizard-api.service
/etc/containers/systemd/enclave-wizard-ui.container
%config(noreplace) /etc/enclave-wizard/nginx.conf
