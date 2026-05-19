#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
OUT_DIR="${REPO_DIR}/out"

ENCLAVE_REPO="${ENCLAVE_REPO:-https://github.com/rh-ecosystem-edge/enclave.git}"
ENCLAVE_BRANCH="${ENCLAVE_BRANCH:-main}"

echo "=== Building Enclave Wizard RPMs ==="
echo "  Enclave repo:   ${ENCLAVE_REPO}"
echo "  Enclave branch: ${ENCLAVE_BRANCH}"

# --- Build wizard binary (includes embedded UI) ---
echo ""
echo "[1/5] Building wizard binary with embedded UI..."
make -C "${REPO_DIR}" build-linux

# --- Clone enclave repo ---
echo "[2/5] Cloning enclave repo (${ENCLAVE_BRANCH})..."
ENCLAVE_TMP=$(mktemp -d)
git clone --depth 1 --branch "${ENCLAVE_BRANCH}" "${ENCLAVE_REPO}" "${ENCLAVE_TMP}/enclave"
tar czf "${REPO_DIR}/enclave-repo.tar.gz" -C "${ENCLAVE_TMP}" enclave
rm -rf "${ENCLAVE_TMP}"

# --- Build enclave RPM ---
echo "[3/5] Building enclave RPM..."
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

        cp /src/enclave-repo.tar.gz  ${RPMBUILD_DIR}/SOURCES/enclave-repo.tar.gz
        cp /src/hack/rpm/enclave.spec ${RPMBUILD_DIR}/SPECS/

        rpmbuild -bb \
            --define "_topdir ${RPMBUILD_DIR}" \
            ${RPMBUILD_DIR}/SPECS/enclave.spec

        cp ${RPMBUILD_DIR}/RPMS/*/*.rpm /out/
        rm -rf ${RPMBUILD_DIR}
    '

rm -f "${REPO_DIR}/enclave-repo.tar.gz"

# --- Build enclave-wizard RPM ---
echo "[4/5] Building enclave-wizard RPM..."

podman run --rm \
    -v "${REPO_DIR}:/src:z" \
    -v "${OUT_DIR}:/out:z" \
    -w /src \
    fedora:latest \
    bash -c '
        dnf install -y rpm-build 2>/dev/null

        RPMBUILD_DIR=$(mktemp -d)
        mkdir -p ${RPMBUILD_DIR}/{SOURCES,SPECS,RPMS,BUILD,SRPMS}

        cp /src/enclave-wizard                          ${RPMBUILD_DIR}/SOURCES/enclave-wizard
        cp /src/hack/quadlets/enclave-wizard.service     ${RPMBUILD_DIR}/SOURCES/enclave-wizard.service
        cp /src/hack/rpm/enclave-wizard.spec             ${RPMBUILD_DIR}/SPECS/

        rpmbuild -bb \
            --define "_topdir ${RPMBUILD_DIR}" \
            ${RPMBUILD_DIR}/SPECS/enclave-wizard.spec

        cp ${RPMBUILD_DIR}/RPMS/*/*.rpm /out/
        rm -rf ${RPMBUILD_DIR}
    '

# --- Generate checksums ---
echo "[5/5] Done."
for rpm in "${OUT_DIR}/"*.rpm; do
    sha256sum "${rpm}" > "${rpm}.sha256"
done

echo ""
ls -lh "${OUT_DIR}/"*.rpm
echo ""
echo "RPMs built:"
for rpm in "${OUT_DIR}/"*.rpm; do
    echo "  $(basename ${rpm})"
done
