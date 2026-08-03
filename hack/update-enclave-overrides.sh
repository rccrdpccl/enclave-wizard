#!/usr/bin/env bash
# Sync plugin defaults from the enclave repo into hack/enclave/ overrides.
# Copies upstream defaults and applies wizard-specific patches (pinned images, etc.).
#
# Usage:
#   ./hack/update-enclave-overrides.sh [--enclave-dir /path/to/enclave]
#
# If --enclave-dir is not given, uses ../enclave relative to this repo.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENCLAVE_DIR="${ENCLAVE_DIR:-$(cd "${REPO_DIR}/../enclave" 2>/dev/null && pwd || echo "")}"
OVERRIDES_DIR="${REPO_DIR}/hack/enclave"

while [ $# -gt 0 ]; do
  case "$1" in
    --enclave-dir) ENCLAVE_DIR="$2"; shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "${ENCLAVE_DIR}" ] || [ ! -d "${ENCLAVE_DIR}/plugins" ]; then
  echo "ERROR: enclave repo not found at ${ENCLAVE_DIR}"
  echo "Usage: $0 --enclave-dir /path/to/enclave"
  exit 1
fi

info() { echo -e "\033[0;32m[INFO]\033[0m  $*"; }

info "Syncing plugin defaults from ${ENCLAVE_DIR}..."

# Sync each plugin's defaults.yaml from the enclave repo
for plugin_dir in "${ENCLAVE_DIR}"/plugins/*/; do
  plugin_name=$(basename "${plugin_dir}")
  [ "${plugin_name}" = "example" ] && continue

  src="${plugin_dir}/defaults.yaml"
  [ ! -f "${src}" ] && continue

  dest_dir="${OVERRIDES_DIR}/plugins/${plugin_name}"
  dest="${dest_dir}/defaults.yaml"

  mkdir -p "${dest_dir}"
  cp "${src}" "${dest}"
  info "  ${plugin_name}/defaults.yaml synced"
done

# Apply wizard-specific patches
# These override upstream defaults with pinned versions for development.
OSAC_DEFAULTS="${OVERRIDES_DIR}/plugins/osac/defaults.yaml"
if [ -f "${OSAC_DEFAULTS}" ]; then
  python3 -c "
import yaml, sys

with open('${OSAC_DEFAULTS}') as f:
    d = yaml.safe_load(f)

# Pin OSAC images for wizard development
d['osac_images'] = {
    'operator': 'ghcr.io/osac-project/osac-operator',
    'operator_tag': 'sha-0c69537',
    'fulfillment_service': 'ghcr.io/osac-project/fulfillment-service:v0.0.64',
    'envoy': 'docker.io/envoyproxy/envoy:v1.33.0',
    'aap_bootstrap': 'ghcr.io/osac-project/osac-aap:sha-3453b54',
    'cli': 'quay.io/openshift/origin-cli:4.20.0',
}

# Use sclorg postgres for dev (smaller, doesn't need RHEL subscription)
d['osac_db_image'] = 'quay.io/sclorg/postgresql-15-c9s:latest'

with open('${OSAC_DEFAULTS}', 'w') as f:
    yaml.dump(d, f, default_flow_style=False, sort_keys=False)
"
  info "  osac/defaults.yaml patched with wizard overrides"
fi

info "Done. Review changes with: git diff hack/enclave/"
