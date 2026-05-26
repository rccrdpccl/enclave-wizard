#!/usr/bin/env bash
# Set up bare metal emulation on a hypervisor host for enclave deployment testing.
# Creates a libvirt network, sushy-tools (Redfish BMC emulator), and UEFI VMs.
# Idempotent — safe to re-run.
#
# Usage: ./setup-bm-emulation.sh [options]
#   --cluster-name NAME    Cluster name (default: mgmt)
#   --base-domain DOMAIN   Base domain (default: enclave-test.lab.local)
#   --api-vip IP           API VIP (default: 192.168.223.200)
#   --ingress-vip IP       Ingress VIP (default: 192.168.223.201)
#   --num-masters N        Number of control plane VMs (default: 3)
#   --teardown             Tear down everything instead of setting up
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-mgmt}"
BASE_DOMAIN="${BASE_DOMAIN:-enclave-test.lab.local}"
API_VIP="${API_VIP:-192.168.223.200}"
INGRESS_VIP="${INGRESS_VIP:-192.168.223.201}"
NUM_MASTERS="${NUM_MASTERS:-3}"

NET_NAME="enclave-bmc"
NET_BRIDGE="enclave-bmc"
NET_CIDR="192.168.223.0/24"
NET_GATEWAY="192.168.223.1"
BMC_PORT=8100
SUSHY_CONTAINER="enclave-sushy-tools"
SUSHY_DIR="/root/sushy-enclave"

VM_PREFIX="enclave-cp"
VM_RAM=16384
VM_CPUS=8
VM_DISK_GB=120
VM_MAC_PREFIX="00:60:2f:e0:c1"

while [ $# -gt 0 ]; do
  case "$1" in
    --cluster-name) CLUSTER_NAME="$2"; shift 2 ;;
    --base-domain) BASE_DOMAIN="$2"; shift 2 ;;
    --api-vip) API_VIP="$2"; shift 2 ;;
    --ingress-vip) INGRESS_VIP="$2"; shift 2 ;;
    --num-masters) NUM_MASTERS="$2"; shift 2 ;;
    --teardown) exec "$(dirname "$0")/teardown-bm-emulation.sh"; ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

FQDN="${CLUSTER_NAME}.${BASE_DOMAIN}"

info()  { echo -e "\033[0;32m[INFO]\033[0m  $*" >&2; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*" >&2; }

# --- Step 1: Libvirt network ---
info "Setting up libvirt network '${NET_NAME}'..."

if virsh net-info "${NET_NAME}" &>/dev/null; then
  info "Network '${NET_NAME}' already exists, updating DNS..."
  virsh net-destroy "${NET_NAME}" 2>/dev/null || true
  virsh net-undefine "${NET_NAME}" 2>/dev/null || true
fi

NET_XML=$(mktemp)
cat > "${NET_XML}" <<EOF
<network xmlns:dnsmasq="http://libvirt.org/schemas/network/dnsmasq/1.0">
  <name>${NET_NAME}</name>
  <forward mode='nat'>
    <nat>
      <port start='1024' end='65535'/>
    </nat>
  </forward>
  <bridge name='${NET_BRIDGE}' stp='on' delay='0'/>
  <dns>
    <host ip='${API_VIP}'>
      <hostname>api.${FQDN}</hostname>
      <hostname>api-int.${FQDN}</hostname>
    </host>
    <host ip='${INGRESS_VIP}'>
      <hostname>console-openshift-console.apps.${FQDN}</hostname>
      <hostname>oauth-openshift.apps.${FQDN}</hostname>
    </host>
  </dns>
  <ip address='${NET_GATEWAY}' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.223.10' end='192.168.223.254'/>
    </dhcp>
  </ip>
  <dnsmasq:options>
    <dnsmasq:option value="address=/apps.${FQDN}/${INGRESS_VIP}"/>
  </dnsmasq:options>
</network>
EOF

virsh net-define "${NET_XML}"
virsh net-start "${NET_NAME}"
virsh net-autostart "${NET_NAME}"
rm -f "${NET_XML}"
info "Network '${NET_NAME}' created with DNS for ${FQDN}"

# --- Step 2: Sushy-tools with SSL ---
info "Setting up sushy-tools BMC emulator..."

mkdir -p "${SUSHY_DIR}"

if [ ! -f "${SUSHY_DIR}/sushy.crt" ] || [ ! -f "${SUSHY_DIR}/sushy.key" ]; then
  openssl req -new -x509 -nodes -days 3650 \
    -subj "/C=US/ST=State/L=City/O=Enclave-Lab/CN=sushy-tools" \
    -keyout "${SUSHY_DIR}/sushy.key" \
    -out "${SUSHY_DIR}/sushy.crt" \
    -addext "subjectAltName=IP:${NET_GATEWAY},IP:127.0.0.1" \
    > /dev/null 2>&1
  info "SSL certificate generated"
fi

cat > "${SUSHY_DIR}/conf.py" <<EOF
SUSHY_EMULATOR_LIBVIRT_URI = "qemu:///system"
SUSHY_EMULATOR_LISTEN_IP = "${NET_GATEWAY}"
SUSHY_EMULATOR_LISTEN_PORT = ${BMC_PORT}
SUSHY_EMULATOR_SSL_CERT = "/etc/sushy/sushy.crt"
SUSHY_EMULATOR_SSL_KEY = "/etc/sushy/sushy.key"
EOF

podman stop "${SUSHY_CONTAINER}" 2>/dev/null || true
podman rm "${SUSHY_CONTAINER}" 2>/dev/null || true

podman run -d --name "${SUSHY_CONTAINER}" --replace \
  --privileged \
  --network host \
  -v /var/run/libvirt:/var/run/libvirt:z \
  -v "${SUSHY_DIR}/conf.py:/etc/sushy/sushy-emulator.conf:z" \
  -v "${SUSHY_DIR}/sushy.crt:/etc/sushy/sushy.crt:z" \
  -v "${SUSHY_DIR}/sushy.key:/etc/sushy/sushy.key:z" \
  quay.io/metal3-io/sushy-tools:latest \
  sushy-emulator --config /etc/sushy/sushy-emulator.conf

# Wait for sushy-tools to start
for i in $(seq 1 15); do
  curl -sk "https://${NET_GATEWAY}:${BMC_PORT}/redfish/v1/" &>/dev/null && break
  sleep 2
done
info "Sushy-tools running on https://${NET_GATEWAY}:${BMC_PORT}"

# --- Step 3: Firewall ---
ZONE=$(firewall-cmd --get-zone-of-interface="${NET_BRIDGE}" 2>/dev/null || echo "libvirt")
firewall-cmd --zone="${ZONE}" --add-port="${BMC_PORT}/tcp" 2>/dev/null || true
firewall-cmd --zone="${ZONE}" --add-port="${BMC_PORT}/tcp" --permanent 2>/dev/null || true
info "Firewall port ${BMC_PORT}/tcp opened in zone '${ZONE}'"

# Allow all traffic on the bridge (libvirt nftables defaults block non-DHCP/DNS)
nft insert rule ip filter LIBVIRT_INP iifname "${NET_BRIDGE}" accept 2>/dev/null || true
nft insert rule ip filter LIBVIRT_OUT oifname "${NET_BRIDGE}" accept 2>/dev/null || true
nft insert rule ip filter LIBVIRT_FWI oifname "${NET_BRIDGE}" accept 2>/dev/null || true
nft insert rule ip filter LIBVIRT_FWO iifname "${NET_BRIDGE}" accept 2>/dev/null || true
info "nftables rules added for ${NET_BRIDGE}"

# --- Step 4: Create VMs ---
info "Creating ${NUM_MASTERS} UEFI VMs..."

for i in $(seq 0 $((NUM_MASTERS - 1))); do
  VM_NAME="${VM_PREFIX}-${i}"
  VM_MAC="${VM_MAC_PREFIX}:$(printf '%02x' "$i")"

  if virsh dominfo "${VM_NAME}" &>/dev/null; then
    virsh destroy "${VM_NAME}" 2>/dev/null || true
    virsh undefine "${VM_NAME}" --remove-all-storage --nvram 2>/dev/null || true
  fi

  virt-install -n "${VM_NAME}" \
    --os-variant=rhel8.0 \
    --ram=${VM_RAM} --vcpus=${VM_CPUS} \
    --network network="${NET_NAME}",mac="${VM_MAC}" \
    --disk size=${VM_DISK_GB},bus=scsi,sparse=yes \
    --boot uefi,hd,cdrom,network \
    --check disk_size=off --noautoconsole 2>&1 | tail -1

  # Stop VM — Ironic will boot it via Redfish
  sleep 1
  virsh destroy "${VM_NAME}" 2>/dev/null || true
  info "  ${VM_NAME} created (MAC: ${VM_MAC})"
done

# --- Step 5: Output VM UUIDs ---
echo ""
info "=== VM System IDs (for wizard config) ==="
for i in $(seq 0 $((NUM_MASTERS - 1))); do
  VM_NAME="${VM_PREFIX}-${i}"
  SYSTEMS=$(curl -sk "https://${NET_GATEWAY}:${BMC_PORT}/redfish/v1/Systems")
  UUID=$(echo "${SYSTEMS}" | python3 -c "
import sys, json, urllib.request, ssl
ctx = ssl.create_default_context()
ctx.check_hostname = False
ctx.verify_mode = ssl.CERT_NONE
for m in json.load(sys.stdin)['Members']:
    sid = m['@odata.id'].split('/')[-1]
    r = urllib.request.urlopen('https://${NET_GATEWAY}:${BMC_PORT}/redfish/v1/Systems/' + sid, context=ctx)
    info = json.loads(r.read())
    if info['Name'] == '${VM_NAME}':
        print(sid)
        break
")
  echo "  ${VM_NAME}: ${UUID}"
done

echo ""
info "Setup complete. VMs are shut off — Ironic will boot them during deployment."
info "Sushy-tools: https://${NET_GATEWAY}:${BMC_PORT}"
info "DNS: api.${FQDN} → ${API_VIP}, *.apps.${FQDN} → ${INGRESS_VIP}"
