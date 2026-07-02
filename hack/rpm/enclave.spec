Name:           enclave
Version:        0.0.1
Release:        1%{?dist}
Summary:        Red Hat Sovereign Enclave distribution and runtime dependencies
License:        Apache-2.0
URL:            https://github.com/rh-ecosystem-edge/enclave
BuildArch:      noarch

Source0:        enclave-repo.tar.gz

Requires:       bind-utils
Requires:       curl
Requires:       httpd
Requires:       httpd-tools
Requires:       ipcalc
Requires:       jq
Requires:       lsof
Requires:       make
Requires:       nmstate
Requires:       openssl
Requires:       podman
Requires:       python3
Requires:       rsync
Requires:       skopeo
Requires:       tar
Requires:       unzip

%description
Red Hat Sovereign Enclave distribution with all runtime dependencies
for running enclave playbooks (ansible-core, ansible collections,
Python packages). The enclave repo is pre-packaged at build time.

%install
mkdir -p %{buildroot}/opt/enclave
tar xzf %{SOURCE0} --strip-components=1 -C %{buildroot}/opt/enclave

%post
ENCLAVE_DIR="/opt/enclave"
export HOME="${ENCLAVE_DIR}"
export PATH="${HOME}/.local/bin:${PATH}"
export XDG_DATA_HOME="${HOME}/.local/share"
export XDG_CACHE_HOME="${HOME}/.cache"

# Create config from examples if not present
for f in "${ENCLAVE_DIR}/config/"*.example.yaml; do
    [ -f "${f}" ] || continue
    target="${f%.example.yaml}.yaml"
    [ -f "${target}" ] || cp "${f}" "${target}"
done

# Run enclave setup scripts
cd "${ENCLAVE_DIR}"
echo "Running enclave environment setup..."
bash ./setup_env.sh 2>&1 | tail -5
echo "Running enclave ansible setup..."
bash ./setup_ansible.sh 2>&1 | tail -5

# Re-install with ansible-runner in the same venv so it shares all deps (kubernetes, etc.)
echo "Adding ansible-runner to enclave environment..."
"${HOME}/.local/bin/uv" tool install --force "${ENCLAVE_DIR}" \
    --with-executables-from ansible-core \
    --with-executables-from ansible-runner \
    --with ansible-runner 2>&1 | tail -5

echo "Enclave installed at ${ENCLAVE_DIR}"

%preun
# Nothing to stop

%postun
if [ $1 -eq 0 ]; then
    rm -rf /opt/enclave
fi

%files
/opt/enclave
