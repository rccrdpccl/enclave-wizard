# Enclave Wizard Deployment

## Overview

The enclave wizard consists of two components deployed inside a VM on a target host:

- **wizard-api** — Go binary running as a native systemd service, with full access to enclave configs and host tools (ansible, oc, make)
- **wizard-ui** — nginx container (podman quadlet) serving the React UI and proxying API requests to the wizard-api

The deployment scripts handle the full lifecycle: build artifacts locally, boot a Fedora VM on the target host via libvirt, install everything via RPM, and set up port forwarding so the wizard is accessible externally.

## Prerequisites

**Local machine:**
- podman
- SSH access to the target host

**Target host:**
- libvirt + virt-install + qemu-img
- A `default` network in libvirt
- SSH access as root
- Internet access (to download Fedora Cloud image on first run)

## Quick Start

```bash
# Build everything (Go binary + UI container + RPM)
make rpm

# Deploy to a target host (boots VM, installs RPM, verifies)
make deploy TARGET=root@myserver.example.com

# Tear down (destroys VM, cleans up port forwarding)
make teardown TARGET=root@myserver.example.com
```

## What Happens During Deployment

1. **Transfer** — The RPM is copied to the target host
2. **VM setup** — A Fedora 42 Cloud VM is booted via libvirt (4GB RAM, 2 vCPUs, 20GB disk) with cloud-init for SSH access and podman
3. **RPM install** — The RPM is copied into the VM and installed. The post-install scriptlet:
   - Installs ORAS
   - Pulls the enclave distribution from `quay.io/edge-infrastructure/enclave`
   - Extracts enclave to `/opt/enclave` and creates config files from examples
   - Loads the wizard-ui container image
   - Opens firewall ports and starts all services
4. **Port forwarding** — iptables rules on the target host forward port 3001 from the target to the VM
5. **Verification** — The script checks that the API, UI, and external access all work

## Accessing the Wizard

After deployment, the wizard is available at:

```
http://<target-host>:3001/wizard
```

## SSH into the VM

```bash
ssh -J root@<target-host> wizard@<vm-ip>
```

The VM IP is shown in the deploy output.

## Architecture

```
Your laptop                    Target host                      VM (enclave-wizard-lz)
─────────────                  ──────────────                   ──────────────────────
                               iptables forwards
browser ──── :3001 ──────────► :3001 ─────────────────────────► nginx (:3001)
                                                                  │
                                                                  ├─► /api/v1/* → wizard-api (:8080)
                                                                  └─► /*        → static UI files
                                                                
                                                                wizard-api (native systemd)
                                                                  └─► reads/writes /opt/enclave/config/
```

## Building the RPM

```bash
make rpm
```

This runs inside containers (no local Go or Node.js needed):
1. Cross-compiles the Go binary for linux/amd64
2. Builds the wizard-ui container image (nginx + static files)
3. Packages everything into an RPM with `rpmbuild`
4. Generates a SHA256 checksum

Output: `out/enclave-wizard-0.0.1-1.fc44.x86_64.rpm`

Verify checksum:
```bash
cd out && sha256sum -c enclave-wizard-*.sha256
```

## Manual RPM Installation

If you want to install the RPM directly on a machine (without the VM):

```bash
scp out/enclave-wizard-*.rpm root@host:/tmp/
ssh root@host "rpm -Uvh /tmp/enclave-wizard-*.rpm"
```

The RPM handles everything: ORAS, enclave extraction, container image loading, firewall, and service startup.

To uninstall:
```bash
rpm -e enclave-wizard
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build Go binary (requires local Go) |
| `make build-linux` | Cross-compile Go binary for linux/amd64 (uses container) |
| `make build-ui` | Build wizard-ui container image |
| `make rpm` | Build RPM package (runs build-linux + build-ui first) |
| `make deploy TARGET=user@host` | Deploy to target host in a VM |
| `make teardown TARGET=user@host` | Tear down deployment |
| `make test` | Run Go tests |
| `make lint` | Run Go linter |

## Troubleshooting

**Deployment fails at "Waiting for VM":**
- Check that libvirt's `default` network is active: `virsh net-list` on the target
- Check the VM console: `virsh console enclave-wizard-lz` on the target

**UI not accessible externally:**
- Check iptables forwarding: `iptables -t nat -L PREROUTING -n` on the target
- Check firewall: `firewall-cmd --list-ports` on the target
- Check VM is running: `virsh list` on the target

**Services not starting in the VM:**
```bash
ssh -J root@target wizard@<vm-ip> "sudo journalctl -u enclave-wizard-api -u enclave-wizard-ui --no-pager -n 30"
```
