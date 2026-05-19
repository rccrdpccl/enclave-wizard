Name:           enclave
Version:        0.0.1
Release:        1%{?dist}
Summary:        Red Hat Sovereign Enclave distribution and runtime dependencies
License:        Apache-2.0
URL:            https://github.com/rh-ecosystem-edge/enclave
BuildArch:      noarch

Source0:        enclave-repo.tar.gz

Requires:       ansible-core
Requires:       python3-pip
Requires:       curl
Requires:       openssl
Requires:       jq
Requires:       ipcalc
Requires:       bind-utils
Requires:       git-core

%description
Red Hat Sovereign Enclave distribution with all runtime dependencies
for running enclave playbooks (ansible-core, ansible collections,
Python packages). The enclave repo is pre-packaged at build time.

%install
mkdir -p %{buildroot}/opt/enclave
tar xzf %{SOURCE0} --strip-components=1 -C %{buildroot}/opt/enclave

%post
ENCLAVE_DIR="/opt/enclave"

# Create config from examples if not present
for f in "${ENCLAVE_DIR}/config/"*.example.yaml; do
    [ -f "${f}" ] || continue
    target="${f%.example.yaml}.yaml"
    [ -f "${target}" ] || cp "${f}" "${target}"
done

# Install ansible-runner and collections
echo "Installing ansible-runner..."
pip3 install ansible-runner 2>&1 | tail -3

echo "Installing Ansible collections..."
ansible-galaxy collection install ansible.utils ansible.posix community.general 2>&1 | tail -3

# Install Python dependencies if requirements.txt exists
if [ -f "${ENCLAVE_DIR}/requirements.txt" ]; then
    echo "Installing Python dependencies..."
    pip3 install -r "${ENCLAVE_DIR}/requirements.txt" 2>&1 | tail -3
fi

echo "Enclave installed at ${ENCLAVE_DIR}"

%preun
# Nothing to stop

%postun
if [ $1 -eq 0 ]; then
    rm -rf /opt/enclave
fi

%files
/opt/enclave
