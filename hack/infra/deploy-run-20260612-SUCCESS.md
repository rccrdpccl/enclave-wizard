# Successful Deployment Run — 2026-06-12

## Result: SUCCESS (rc=0)

**ok=591, changed=160, failed=0, skipped=177**
**Total time: 1h23m05s**

## Environment
- **Host**: rdu-infra-edge-03.infra-edge.lab.eng.rdu2.redhat.com
- **Wizard VM**: 192.168.122.106 (CentOS Stream 9, 4GB/2CPU)
- **BM VMs**: 3x enclave-cp (32GB RAM, 16 CPU, 2x120GB disks, PXE boot)
- **Sushy-tools**: SUSHY_EMULATOR_IGNORE_BOOT_DEVICE=True
- **OCP version**: 4.20.21
- **Branch**: feat/osac-services-and-deployment-fixes
- **Wizard flags**: --no-auth --record

## Timeline
| Time (EDT) | Duration | Event |
|---|---|---|
| 12:55 | 0:00 | Deploy triggered |
| 12:57 | 0:02 | Ironic pod created |
| 13:06 | 0:11 | VMs booted, 3-min pause |
| 13:09 | 0:14 | Bootstrap Kube API Initialized |
| 13:16 | 0:21 | All nodes rebooted to disk (boot order OK) |
| 13:27 | 0:32 | Bootstrap complete |
| 13:44 | 0:49 | **OCP Install complete** |
| 13:49 | 0:54 | MCH reached Running (~5 min) |
| 14:02 | 1:07 | Quay user created (DNS fix OK) |
| 14:10 | 1:15 | Remaining operators installing |
| 14:18 | 1:23 | **FULL SUCCESS** |

## All Fixes Applied
1. Boot order: --pxe VMs + SUSHY_EMULATOR_IGNORE_BOOT_DEVICE
2. Python env: uv-managed Python 3.12 with ansible-runner
3. RPM post-install: XDG env vars + absolute path for uv
4. VM resources: 32GB RAM, 16 CPU, 2x120GB disks
5. DNS priority: enclave-bmc DNS first (-10 priority)
6. OSAC config: fields split to config/plugins/osac.yaml

## Recordings
Saved to `fixtures/recordings/`:
- `playbooks-main.yaml.json` (12MB) — full deploy recording
- `playbooks-validate-plugins.yaml.json` (331KB)
- `playbooks-validation-validate-schema.yaml--validate-config.json` (184KB)

## Credentials
- **Wizard**: https://rdu-infra-edge-03:3443/wizard (no auth)
- **OCP Console**: kubeadmin / KEHZ9-MITfo-ZH9cR-XZDzT
- **VM SSH**: ssh -J root@rdu-infra-edge-03 wizard@192.168.122.106

## Top Tasks by Duration
| Task | Duration |
|---|---|
| Wait for bootstrap | 1229s |
| Wait for installation | 1062s |
| MCH ready | 349s |
| QuayRegistry available | 245s |
| ClusterImageSet cleanup | 229s |
| Baremetal reboot pause | 180s |
| TektonConfig ready | 161s |
| Quay rolling update | 100s |
| Hosts ready | 93s |
