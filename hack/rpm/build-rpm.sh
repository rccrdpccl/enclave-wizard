#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
OUT_DIR="${REPO_DIR}/out"

echo "=== Building Enclave Wizard RPM ==="

# --- Build artifacts ---
echo "[1/3] Building artifacts..."
make -C "${REPO_DIR}" build-linux
make -C "${REPO_DIR}" build-ui

# Save UI container image as tarball
echo "  Saving UI container image..."
rm -f "${REPO_DIR}/wizard-ui.tar"
podman save enclave-wizard-ui:dev -o "${REPO_DIR}/wizard-ui.tar"

# --- Build RPM in container ---
echo "[2/3] Building RPM..."
mkdir -p "${OUT_DIR}"

podman run --rm \
    -v "${REPO_DIR}:/src:z" \
    -v "${OUT_DIR}:/out:z" \
    -w /src \
    fedora:latest \
    bash -c '
        dnf install -y rpm-build 2>/dev/null

        RPMBUILD_DIR=$(mktemp -d)
        mkdir -p ${RPMBUILD_DIR}/{SOURCES,SPECS,RPMS,BUILD,SRPMS}

        cp /src/enclave-wizard          ${RPMBUILD_DIR}/SOURCES/enclave-wizard
        cp /src/wizard-ui.tar           ${RPMBUILD_DIR}/SOURCES/wizard-ui.tar
        cp /src/hack/quadlets/enclave-wizard-api.service   ${RPMBUILD_DIR}/SOURCES/
        cp /src/hack/quadlets/enclave-wizard-ui.container  ${RPMBUILD_DIR}/SOURCES/
        cp /src/hack/nginx-deploy.conf  ${RPMBUILD_DIR}/SOURCES/nginx-deploy.conf
        cp /src/hack/rpm/enclave-wizard.spec ${RPMBUILD_DIR}/SPECS/

        rpmbuild -bb \
            --define "_topdir ${RPMBUILD_DIR}" \
            ${RPMBUILD_DIR}/SPECS/enclave-wizard.spec

        cp ${RPMBUILD_DIR}/RPMS/*/*.rpm /out/
        rm -rf ${RPMBUILD_DIR}
    '

# --- Generate checksum ---
echo "[3/4] Generating checksum..."
RPM_FILE=$(ls -t "${OUT_DIR}/"*.rpm | head -1)
sha256sum "${RPM_FILE}" > "${RPM_FILE}.sha256"

echo "[4/4] Done."
echo ""
ls -lh "${OUT_DIR}/"enclave-wizard*
echo ""
echo "RPM:      $(basename ${RPM_FILE})"
echo "Checksum: $(basename ${RPM_FILE}).sha256"
echo ""
echo "Verify with: sha256sum -c $(basename ${RPM_FILE}).sha256"
