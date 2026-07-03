#!/usr/bin/env bash
# Sanity check: verify BM emulation environment is ready for deployment.
# Run after make bm-emulation + make deploy + make restore-remote-config.
#
# Usage: ./hack/infra/sanity-check.sh --host root@hypervisor
set -euo pipefail

TARGET=""
while [ $# -gt 0 ]; do
  case "$1" in
    --host) TARGET="$2"; shift 2 ;;
    *) echo "Usage: $0 --host root@hypervisor"; exit 1 ;;
  esac
done
[ -z "${TARGET}" ] && echo "Usage: $0 --host root@hypervisor" && exit 1

SSH="ssh -o StrictHostKeyChecking=no"
PASS=0
FAIL=0

ok()   { echo -e "  \033[0;32m✓\033[0m $*"; PASS=$((PASS + 1)); }
fail() { echo -e "  \033[0;31m✗\033[0m $*"; FAIL=$((FAIL + 1)); }

echo "=== Infrastructure ==="

# Wizard VM
VM_IP=$(${SSH} "${TARGET}" "virsh domifaddr enclave-wizard-lz --source agent 2>/dev/null | grep -oP '192\.168\.122\.\d+' | head -1" || echo "")
[ -n "${VM_IP}" ] && ok "Wizard VM running: ${VM_IP}" || fail "Wizard VM not found"

# CP VMs
for i in 0 1 2; do
  ${SSH} "${TARGET}" "virsh dominfo enclave-cp-${i}" &>/dev/null && ok "enclave-cp-${i} exists" || fail "enclave-cp-${i} missing"
done

# Sushy-tools
SUSHY=$(${SSH} "${TARGET}" "curl -sk -o /dev/null -w '%{http_code}' https://192.168.223.1:8100/redfish/v1/ 2>/dev/null" || echo "000")
[ "${SUSHY}" = "200" ] && ok "Sushy-tools responding" || fail "Sushy-tools not responding (HTTP ${SUSHY})"

# BMC network
${SSH} "${TARGET}" "virsh net-info enclave-bmc" &>/dev/null && ok "enclave-bmc network active" || fail "enclave-bmc network missing"

echo ""
echo "=== DNS Resolution (from wizard VM) ==="

if [ -n "${VM_IP}" ]; then
  for host in \
    "api.edge.enclave-test.lab.local" \
    "console-openshift-console.apps.edge.enclave-test.lab.local" \
    "registry-quay-quay-enterprise.apps.edge.enclave-test.lab.local" \
    "oauth-openshift.apps.edge.enclave-test.lab.local" \
    "mirror.enclave-test.lab.local"; do
    RESOLVED=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'getent hosts ${host} 2>/dev/null | awk \"{print \\\$1}\"'" 2>/dev/null || echo "")
    [ -n "${RESOLVED}" ] && ok "${host} → ${RESOLVED}" || fail "${host} CANNOT RESOLVE"
  done

  # Wildcard test
  WILD=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'getent hosts random-test.apps.edge.enclave-test.lab.local 2>/dev/null | awk \"{print \\\$1}\"'" 2>/dev/null || echo "")
  [ -n "${WILD}" ] && ok "*.apps wildcard → ${WILD}" || fail "*.apps wildcard NOT RESOLVING"
fi

echo ""
echo "=== Wizard Config ==="

if [ -n "${VM_IP}" ]; then
  # lzBmcIP
  LZ_BMC=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo grep lzBmcIP /opt/enclave/config/global.yaml | awk \"{print \\\$2}\"'" 2>/dev/null || echo "")
  [ "${LZ_BMC}" = "${VM_IP}" ] && ok "lzBmcIP: ${LZ_BMC} (matches VM IP)" || fail "lzBmcIP: ${LZ_BMC} (expected ${VM_IP})"

  # disconnected
  DISCO=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo grep \"^disconnected:\" /opt/enclave/config/global.yaml | awk \"{print \\\$2}\"'" 2>/dev/null || echo "")
  [ "${DISCO}" = "false" ] && ok "disconnected: false" || fail "disconnected: ${DISCO} (expected false)"

  # bmcSystemIds match real VMs
  for i in 0 1 2; do
    REAL_UUID=$(${SSH} "${TARGET}" "virsh domuuid enclave-cp-${i} 2>/dev/null" || echo "")
    CFG_UUID=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo python3 -c \"
import yaml
with open(\\\"/opt/enclave/config/global.yaml\\\") as f:
    d = yaml.safe_load(f)
hosts = d.get(\\\"agent_hosts\\\", [])
print(hosts[${i}].get(\\\"bmcSystemId\\\", \\\"\\\") if len(hosts) > ${i} else \\\"\\\")
\"'" 2>/dev/null || echo "")
    if [ -n "${REAL_UUID}" ] && [ "${REAL_UUID}" = "${CFG_UUID}" ]; then
      ok "enclave-cp-${i} UUID: ${REAL_UUID:0:8}... (matches)"
    else
      fail "enclave-cp-${i} UUID mismatch: config=${CFG_UUID:0:8}... actual=${REAL_UUID:0:8}..."
    fi

    # Verify UUID exists in sushy-tools
    SUSHY_CODE=$(${SSH} "${TARGET}" "curl -sk -o /dev/null -w '%{http_code}' https://192.168.223.1:8100/redfish/v1/Systems/${REAL_UUID} 2>/dev/null" || echo "000")
    [ "${SUSHY_CODE}" = "200" ] && ok "  sushy-tools knows ${REAL_UUID:0:8}..." || fail "  sushy-tools: HTTP ${SUSHY_CODE} for ${REAL_UUID:0:8}..."
  done

  # SSH key
  KEY_MATCH=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'diff <(sudo cat /opt/enclave/config/ssh-pub-key.pub 2>/dev/null) <(cat /home/wizard/.ssh/id_rsa.pub 2>/dev/null) >/dev/null 2>&1 && echo yes || echo no'" 2>/dev/null || echo "no")
  [ "${KEY_MATCH}" = "yes" ] && ok "SSH key matches" || fail "SSH key mismatch"

  # httpd
  HTTPD=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo systemctl is-active httpd'" 2>/dev/null || echo "inactive")
  [ "${HTTPD}" = "active" ] && ok "httpd running" || fail "httpd ${HTTPD}"

  # wizard service
  WIZARD=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo systemctl is-active enclave-wizard'" 2>/dev/null || echo "inactive")
  [ "${WIZARD}" = "active" ] && ok "enclave-wizard running" || fail "enclave-wizard ${WIZARD}"

  # OSAC config
  OSAC=$(${SSH} "${TARGET}" "${SSH} wizard@${VM_IP} 'sudo cat /opt/enclave/config/plugins/osac.yaml 2>/dev/null | head -1'" 2>/dev/null || echo "")
  [ -n "${OSAC}" ] && [ "${OSAC}" != "# not present" ] && ok "osac.yaml present" || ok "osac.yaml not present (optional)"
fi

echo ""
echo "=== Results: ${PASS} passed, ${FAIL} failed ==="
[ ${FAIL} -eq 0 ] && echo -e "\033[0;32mAll checks passed.\033[0m" || echo -e "\033[0;31m${FAIL} check(s) failed.\033[0m"
exit ${FAIL}
