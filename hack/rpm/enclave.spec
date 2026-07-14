Name:           enclave
Version:        0.0.1
Release:        1%{?dist}
Summary:        Red Hat Sovereign Enclave
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
Red Hat Sovereign Enclave repo. Run setup_env.sh and setup_ansible.sh
after install to configure the environment.

%install
mkdir -p %{buildroot}/opt/enclave
tar xzf %{SOURCE0} --strip-components=1 -C %{buildroot}/opt/enclave

%post
ENCLAVE_DIR="/opt/enclave"
HOME_DIR="/home/enclave"

# Create working directory and home
mkdir -p "${HOME_DIR}"
mkdir -p "${HOME_DIR}/.ssh"

# Generate SSH key if missing
if [ ! -f "${HOME_DIR}/.ssh/id_rsa" ]; then
    ssh-keygen -t rsa -b 4096 -f "${HOME_DIR}/.ssh/id_rsa" -N "" -q
fi

# Create config from examples if not present
for f in "${ENCLAVE_DIR}/config/"*.example.yaml; do
    [ -f "${f}" ] || continue
    target="${f%.example.yaml}.yaml"
    [ -f "${target}" ] || cp "${f}" "${target}"
done

# Set workingDir in global.yaml
if [ -f "${ENCLAVE_DIR}/config/global.yaml" ]; then
    sed -i "s|workingDir:.*|workingDir: ${HOME_DIR}|" "${ENCLAVE_DIR}/config/global.yaml"
fi

# Run ansible setup (uv + ansible + collections — no dnf calls)
cd "${ENCLAVE_DIR}"
export HOME="${HOME_DIR}"
export PATH="${HOME}/.local/bin:${PATH}"

# Install AWS CLI if missing
if ! command -v aws &>/dev/null; then
    echo "Installing AWS CLI..."
    AWSCLI_TMP=$(mktemp -d)
    curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-2.34.53.zip" -o "${AWSCLI_TMP}/awscliv2.zip"
    unzip -q "${AWSCLI_TMP}/awscliv2.zip" -d "${AWSCLI_TMP}"
    "${AWSCLI_TMP}/aws/install" --update 2>&1 | tail -1
    rm -rf "${AWSCLI_TMP}"
fi

# Enable httpd
systemctl enable --now httpd 2>/dev/null || true
mkdir -p /var/www/html
chmod 755 /var/www/html

echo "Running setup_ansible.sh (installs uv, ansible, collections)..."
bash ./setup_ansible.sh 2>&1 | tail -10

# Add ansible-runner to the same environment
echo "Adding ansible-runner..."
"${HOME}/.local/bin/uv" tool install --force "${ENCLAVE_DIR}" \
    --with-executables-from ansible-core \
    --with-executables-from ansible-runner \
    --with ansible-runner 2>&1 | tail -5

# Ensure ansible finds collections
if ! grep -q collections_path "${ENCLAVE_DIR}/ansible.cfg" 2>/dev/null; then
    sed -i "/^\[defaults\]/a collections_path=${HOME}/.ansible/collections" "${ENCLAVE_DIR}/ansible.cfg"
fi

echo "Enclave installed at ${ENCLAVE_DIR}, HOME=${HOME_DIR}"

%preun
# Nothing to stop

%postun
if [ $1 -eq 0 ]; then
    rm -rf /opt/enclave
fi

%files
/opt/enclave
