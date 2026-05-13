# Enclave Configuration Reference

All configuration is loaded by `playbooks/common/load-vars.yaml` in this order:

1. `defaults/deployment.yaml` — deployment defaults
2. `config/global.yaml` — user-supplied cluster config (required)
3. `config/certificates.yaml` — user-supplied TLS certificates (required)
4. `config/cloud_infra.yaml` — user-supplied discovery hosts (required)
5. `defaults/control_binaries.yaml`, `defaults/operators.yaml`, `defaults/platforms.yaml`, `defaults/catalogs.yaml`, `defaults/mirror_registry.yaml`, `defaults/quay_operator.yaml`, `defaults/k8s.yaml`, `defaults/oc_mirror.yaml` — internal defaults

Schemas live in `schemas/` and are validated by `playbooks/validation/`. Fields marked **overrideable in global.yaml** can be set in `config/global.yaml` to replace the default value.

---

## `config/global.yaml`

The primary user configuration file. Validated against `schemas/variables.yaml`.

### Required Fields

| Field | Type | Description |
|---|---|---|
| `workingDir` | string | Absolute path to the root working directory (e.g. `/home/enclave`). Used as the base for all generated files, binaries, and data. |
| `baseDomain` | string | Base DNS domain for the cluster (e.g. `enclave-test.nodns.in`). Combined with `clusterName` to form the cluster FQDN. |
| `clusterName` | string | Name of the OpenShift cluster (e.g. `mgmt`). Combined with `baseDomain`. |
| `machineNetwork` | CIDR | Network CIDR containing all cluster node IPs (e.g. `192.168.2.0/24`). |
| `apiVIP` | IPv4 | Virtual IP for the Kubernetes API server. Must be within `machineNetwork`. |
| `ingressVIP` | IPv4 | Virtual IP for the ingress wildcard. Must be within `machineNetwork`. |
| `rendezvousIP` | IPv4 | IP of the first control-plane node. Must match the `ipAddress` of the first entry in `agent_hosts`. |
| `defaultDNS` | IPv4 | DNS server IP for cluster nodes. |
| `defaultGateway` | IPv4 | Default gateway IP for cluster nodes. |
| `defaultPrefix` | integer (1–32) | Subnet prefix length for the machine network (e.g. `24` for /24). |
| `lzBmcIP` | IPv4 | IP of the landing zone host that serves the boot ISO over HTTP to BMC/Redfish nodes. |
| `quayUser` | string | Admin username for the Quay registry. |
| `quayPassword` | string | Admin password for the Quay registry. |
| `quayBackend` | `RadosGWStorage` \| `LocalStorage` | Quay image storage backend. `RadosGWStorage` uses an external S3/Ceph RGW endpoint. `LocalStorage` mounts a local PVC (not recommended for production). |
| `blockStorageBackend` | `lvms` \| `odf` | Block storage backend for the Quay database and the Assisted Installer. Selects which storage plugin is deployed. Use `storage_plugin` in new configurations (see optional fields). |
| `pullSecret` | object | OpenShift pull secret object (`{"auths":{…}}`). Obtain from [console.redhat.com](https://console.redhat.com/openshift/install/pull-secret). |
| `sshPubPath` | string | Absolute path to the SSH public key file used for cluster node access. |
| `agent_hosts` | array (exactly 3) | Control-plane node definitions. See [Host Object Fields](#host-object-fields). |

### Conditional Requirements

| Condition | Additional required fields |
|---|---|
| `quayBackend: RadosGWStorage` | `quayBackendRGWConfiguration` |
| `blockStorageBackend: odf` | `odfExternalConfig` |
| `sslAPICertificateFullChain` is set | `sslAPICertificateKey`, `sslIngressCertificateFullChain`, `sslIngressCertificateKey` |
| `ironicHTTPSCertificate` is set | `ironicHTTPSKey` |

### Optional Fields

| Field | Default | Description |
|---|---|---|
| `storage_plugin` | `lvms` | Modern replacement for `blockStorageBackend`. Selects which storage plugin (`plugins/lvms/` or `plugins/odf/`) to deploy. `blockStorageBackend` is derived from this value and will be removed after migration completes. |
| `enabled_plugins` | `[storage_plugin]` | List of plugin names to deploy during the pipeline. By default only the selected storage plugin is enabled. Extend this list to deploy additional plugins (e.g. `[lvms, nvidia-gpu, openshift-ai]`). |
| `disconnected` | `true` | Whether the deployment is air-gapped. Set to `false` for connected deployments that pull images from upstream registries. When `false`, Phase 2 (mirroring) is skipped. |
| `masterMaxPods` | `500` | Maximum number of pods per node. Configures the High Density Pod Limits for Agent-Based Installer. |
| `diskEncryption` | `false` | Enable TPM v2 disk encryption on all cluster nodes. |
| `ocMirrorLogLevel` | `info` | Log verbosity for `oc-mirror`. Options: `trace`, `debug`, `info`, `error`. |
| `defaultNtpServers` | (none) | Array of additional NTP server addresses for cluster nodes. |
| `lzBmcHostname` | (none) | DNS hostname for the landing zone BMC interface. When set, Ironic HTTPS vmedia URLs use this hostname instead of `lzBmcIP`. Required when using publicly trusted certificates (e.g. Let's Encrypt) that only support DNS SANs. Must resolve to `lzBmcIP` from the BMC network. |
| `odfExternalConfig` | (none) | JSON string from the `ceph-external-cluster-details-exporter.py` script. Required when `blockStorageBackend: odf`. |
| `lvmsConfig` | (none) | LVMS device selector configuration. When omitted, LVMS manages all available disks. |
| `lvmsConfig.deviceSelector.optionalPaths` | (none) | Array of disk paths (by-path) to restrict LVMS to specific devices (e.g. `/dev/disk/by-path/pci-0000:00:11.4-ata-1.0`). |
| `discovery_hosts` | (none) | Additional worker nodes for hardware discovery (CaaS). See [Host Object Fields](#host-object-fields). Set to `[]` if no discovery hosts are needed. Can also be placed in `config/cloud_infra.yaml`. |
| `sslAPICertificateFullChain` | (none) | PEM-encoded full certificate chain for the API server (`api.<clusterName>.<baseDomain>`). |
| `sslAPICertificateKey` | (none) | PEM-encoded private key for the API server certificate. Required if `sslAPICertificateFullChain` is set. |
| `sslIngressCertificateFullChain` | (none) | PEM-encoded full certificate chain for the ingress wildcard (`*.apps.<clusterName>.<baseDomain>`). Required if `sslAPICertificateFullChain` is set. |
| `sslIngressCertificateKey` | (none) | PEM-encoded private key for the ingress certificate. Required if `sslAPICertificateFullChain` is set. |
| `sslCACertificate` | (none) | PEM-encoded root CA certificate. |
| `ironicHTTPSCertificate` | (none) | PEM-encoded TLS certificate for the Ironic vmedia HTTPS server. The SAN must cover `lzBmcHostname` (DNS SAN) or `lzBmcIP` (IP SAN). |
| `ironicHTTPSKey` | (none) | PEM-encoded private key for the Ironic HTTPS certificate. Required if `ironicHTTPSCertificate` is set. |

### `quayBackendRGWConfiguration`

Required when `quayBackend: RadosGWStorage`. Configures the external S3/RadosGW storage backend for Quay.

| Field | Required | Type | Default | Description |
|---|---|---|---|---|
| `access_key` | yes | string | — | S3/RGW access key ID. |
| `secret_key` | yes | string | — | S3/RGW secret access key. |
| `bucket_name` | yes | string | — | Name of the S3 bucket for Quay image storage. |
| `hostname` | yes | string | — | Hostname of the RadosGW/S3 endpoint. |
| `is_secure` | no | boolean | `true` | Use HTTPS for the backend storage connection. |
| `port` | no | integer (1–65535) | `443` | Port for the backend storage endpoint. |
| `minimum_chunk_size_mb` | no | integer ≥1 | `100` | Minimum multipart upload chunk size in MB. |
| `maximum_chunk_size_mb` | no | integer ≥1 | `500` | Maximum multipart upload chunk size in MB. |
| `server_side_assembly` | no | boolean | `true` | Enable server-side assembly for multipart uploads. |
| `storage_path` | no | string (must start with `/`) | `/datastorage/registry` | Path prefix for stored objects within the bucket. |

### Host Object Fields

Used by both `agent_hosts` (exactly 3 entries required) and `discovery_hosts` (any number).

| Field | Required | Type | Description |
|---|---|---|---|
| `name` | yes | string | Node name label. |
| `macAddress` | yes | MAC address | MAC address in `xx:xx:xx:xx:xx:xx` format. Used for network boot and agent identification. |
| `ipAddress` | yes | IPv4 | Node IP address. Must be within `machineNetwork`. For the first `agent_hosts` entry, must match `rendezvousIP`. |
| `redfish` | yes | IPv4 | Redfish/IPMI management IP for remote power control and virtual media. |
| `redfishUser` | yes | string | Redfish username. |
| `redfishPassword` | yes | string | Redfish password. |
| `rootDisk` | yes | string | Path to the root disk for OS installation (e.g. `/dev/disk/by-path/pci-0000:00:11.4-ata-1.0`). By-path addressing is recommended for stability. |
| `bmcSystemId` | no | string | Redfish system ID path component. Defaults to `"1"`. Override when the BMC exposes the system under a different ID (e.g. `"2"`). |
| `mapInterfaces` | no | object | Interface mapping configuration for complex network topologies. |
| `networkConfig` | no | object | NMState network configuration object. Use instead of `macAddress`/`ipAddress` for multi-NIC or VLAN configurations. |

---

## `config/certificates.yaml`

Certificate fields can alternatively be placed directly in `global.yaml`. This file exists as a separate secrets file for organizational purposes.

| Field | Description |
|---|---|
| `sslAPICertificateKey` | PEM-encoded private key for the API server certificate. |
| `sslAPICertificateFullChain` | PEM-encoded full certificate chain for the API server. When this field is non-empty, `sslAPICertificateKey`, `sslIngressCertificateFullChain`, and `sslIngressCertificateKey` are all required. |
| `sslIngressCertificateKey` | PEM-encoded private key for the ingress wildcard certificate. |
| `sslIngressCertificateFullChain` | PEM-encoded full certificate chain for the ingress wildcard certificate. |
| `sslCACertificate` | PEM-encoded root CA certificate. Trusted by the cluster and installed as a custom CA. |
| `ironicHTTPSCertificate` | PEM-encoded TLS certificate for the Ironic vmedia HTTPS server. Optional; required only when serving boot ISOs over HTTPS. |
| `ironicHTTPSKey` | PEM-encoded private key for the Ironic HTTPS certificate. Required when `ironicHTTPSCertificate` is set. |

---

## `config/cloud_infra.yaml`

| Field | Type | Description |
|---|---|---|
| `discovery_hosts` | array of [Host Objects](#host-object-fields) | Worker nodes to be discovered and added to the cluster via hardware discovery (CaaS). Set to `[]` if no discovery hosts are needed. |

---

## Internal Defaults

These files in `defaults/` set values that are used when the corresponding field is not present in `global.yaml`. All overrideable fields can be set directly in `global.yaml`.

---

### `defaults/deployment.yaml`

Schema: `schemas/deployment.yaml`

| Field | Default | Overrideable | Description |
|---|---|---|---|
| `disconnected` | `true` | yes | Air-gapped deployment mode. |
| `masterMaxPods` | `500` | yes | Maximum pods per node. |
| `diskEncryption` | `false` | yes | TPM v2 disk encryption. |
| `ocMirrorLogLevel` | `info` | yes | oc-mirror log level (`trace`, `debug`, `info`, `error`). |
| `storage_plugin` | `lvms` | yes | Active storage plugin (`lvms` or `odf`). |
| `blockStorageBackend` | `"{{ storage_plugin }}"` | yes | Backward-compatible alias for `storage_plugin`. Derived automatically; set `storage_plugin` instead. |
| `enabled_plugins` | `["{{ storage_plugin }}"]` | yes | Plugins to deploy. Always includes `storage_plugin` even if overridden. |
| `pullSecretPath` | `"{{ workingDir }}/config/pull-secret.json"` | yes | Path where the pull secret JSON file is written before deployment. |

---

### `defaults/platforms.yaml`

Schema: `schemas/platforms.yaml`

| Field | Default | Overrideable | Description |
|---|---|---|---|
| `mgmt_openshift_version` | `4.20.8` | yes | OpenShift version for the management cluster. Must appear in the `openshift_versions` list. |
| `openshift_versions` | `[{version: "4.20.8"}]` | yes | Array of `{version: string}` objects listing all OpenShift versions to make available for spoke cluster deployment. `mgmt_openshift_version` must be present in this list. |

---

### `defaults/catalogs.yaml`

Schema: `schemas/catalogs.yaml`

| Field | Default | Overrideable | Description |
|---|---|---|---|
| `rh_operator_catalog` | `registry.redhat.io/redhat/redhat-operator-index` | yes | Source address of the Red Hat Operator catalog. |
| `rh_operator_catalog_version` | `v4.20` | yes | Version tag for the Red Hat Operator catalog image. |
| `certified_operator_catalog` | `registry.redhat.io/redhat/certified-operator-index` | yes | Source address of the Certified Operator catalog. |
| `certified_operator_catalog_version` | `v4.20` | yes | Version tag for the Certified Operator catalog image. |
| `mirror_rh_operator_catalog` | `mirror-redhat-operators` | yes | Name of the mirrored Red Hat Operator CatalogSource created in the cluster. |
| `mirror_certified_rh_operator_catalog` | `mirror-certified-operators` | yes | Name of the mirrored Certified Operator CatalogSource created in the cluster. |

---

### `defaults/mirror_registry.yaml`

Schema: `schemas/mirror_registry.yaml`

| Field | Default | Overrideable | Description |
|---|---|---|---|
| `quayHostname` | `"mirror.{{ baseDomain }}"` | yes | Hostname for the Quay mirror registry. Defaults to `mirror.<baseDomain>`. |
| `quayCAPath` | `"{{ workingDir }}/data/quay-rootCA/rootCA.pem"` | yes | Path to the Quay CA certificate file on the landing zone host. |

---

### `defaults/quay_operator.yaml`

Schema: `schemas/quay_operator.yaml`

| Field | Default | Overrideable | Description |
|---|---|---|---|
| `quayFeatureProxyStorage` | `true` | yes | Enable the Quay proxy storage feature. |
| `quayFeatureQuotaManagement` | `false` | yes | Enable the Quay quota management feature. |
| `quayMaximumLayerSize` | `100G` | yes | Maximum image layer size Quay will accept. Format: `[0-9]+(G\|M)` (e.g. `100G`, `500M`). |
| `quayBackendDefaults.storage_path` | `/datastorage/registry` | yes | Default object storage path applied to all Quay backends. Must begin with `/`. |
| `quayBackendRGWDefaults.is_secure` | `true` | yes | Default: use HTTPS for RadosGW backend connections. Applied when `quayBackend: RadosGWStorage` and the field is not set in `quayBackendRGWConfiguration`. |
| `quayBackendRGWDefaults.port` | `443` | yes | Default port for the RadosGW backend endpoint. |
| `quayBackendRGWDefaults.minimum_chunk_size_mb` | `100` | yes | Default minimum multipart upload chunk size in MB. |
| `quayBackendRGWDefaults.maximum_chunk_size_mb` | `500` | yes | Default maximum multipart upload chunk size in MB. |
| `quayBackendRGWDefaults.server_side_assembly` | `true` | yes | Default: enable server-side assembly for multipart uploads. |
| `quayOAuthApp.enabled` | `true` | yes | Automatically create a Quay OAuth application during initial setup. |
| `quayOAuthApp.name` | `default-application` | yes | Name of the OAuth application. Required when `enabled: true`. |
| `quayOAuthApp.organization` | `default-org` | yes | Quay organization for the OAuth application. Pattern: `^[a-z0-9]+(?:[._-][a-z0-9]+)*$`. Required when `enabled: true`. |
| `quayOAuthApp.redirect_uri` | `http://localhost:8080/callback` | yes | OAuth redirect URI. Pattern: `^https?://.+`. Required when `enabled: true`. |
| `quayOAuthApp.scopes` | `repo:read repo:write repo:admin repo:create user:read user:admin org:admin` | yes | Space-separated OAuth scope string. Required when `enabled: true`. |

---

### `defaults/oc_mirror.yaml`

Schema: `schemas/oc_mirror.yaml`. These control parallelism and retry behaviour of the mirroring phase. Commented as internal defaults; override only when tuning mirror performance.

| Field | Default | Overrideable | Description |
|---|---|---|---|
| `ocMirrorParallelImages` | `10` | yes | Number of images downloaded in parallel by oc-mirror. |
| `ocMirrorParallelLayers` | `10` | yes | Number of image layers downloaded in parallel by oc-mirror. |
| `ocMirrorParallelLayersLocalStorage` | `1` | yes | Parallel layer downloads when `quayBackend: LocalStorage`. Lower default prevents I/O contention on local storage. |
| `ocMirrorRetryTimes` | `10` | yes | Number of internal oc-mirror retries per image on failure. |
| `ocMirrorRetryDelay` | `10s` | yes | Delay between oc-mirror internal retries. Format: `[0-9]+[smh]?` (e.g. `10s`, `1m`). |
| `ocMirrorImageTimeout` | `40m0s` | yes | Timeout for mirroring a single image. Format: `[0-9]+[smh][0-9]*[smh]?` (e.g. `40m0s`, `1h`). |
| `ocMirrorAnsibleRetries` | `10` | yes | Number of Ansible-level retries wrapping the oc-mirror invocation. |
| `ocMirrorAnsibleDelay` | `10` | yes | Seconds between Ansible-level oc-mirror retries. |
| `ocMirrorCacheAnsibleRetries` | `5` | yes | Ansible-level retries for oc-mirror cache tasks specifically. |

---

### `defaults/k8s.yaml`

Schema: `schemas/k8s.yaml`. Internal retry settings for `kubernetes.core.k8s` module calls. Not intended for user override.

| Field | Default | Description |
|---|---|---|
| `k8s_retries` | `12` | Number of retries on `k8s` state: present/absent/patched calls. |
| `k8s_delay` | `10` | Seconds between `k8s` module retries. |

---

### `defaults/control_binaries.yaml`

Schema: `schemas/control_binaries.yaml`. Defines download URLs and SHA256 checksums for required binaries. Override to pin to different versions or use internal mirrors.

Each entry under `control_binaries` has:

| Sub-field | Description |
|---|---|
| `url` | Download URL for the binary or tarball. |
| `checksum` | SHA256 checksum in the format `sha256:<64 hex chars>`. |

| Binary key | Default version | Description |
|---|---|---|
| `control_binaries.openshift_client` | `4.20.8` | OpenShift CLI (`oc`, `kubectl`) tarball. |
| `control_binaries.helm` | `3.17.1` | Helm binary. |
| `control_binaries.mirror_registry` | `1.3.11` | `mirror-registry` binary (installs Quay). |
| `control_binaries.oc_mirror` | `4.20.15` | `oc-mirror` binary for disconnected image mirroring. |
| `control_binaries.clairctl` | `v4.8.0` | `clairctl` binary for Clair vulnerability scanning. |

---

### `defaults/operators.yaml`

Schema: `schemas/operators.yaml`. Defines the core OLM operators installed on the management cluster. Each entry in the `operators` array has:

| Field | Required | Description |
|---|---|---|
| `name` | yes | OLM package name as it appears in the catalog. |
| `version` | yes | Operator version to install. |
| `channel` | yes | OLM update channel. |
| `init_version` | yes | Initial/bootstrap version used during the seed deployment phase. |
| `namespace` | no | Kubernetes namespace where the operator is installed. |
| `source` | no | CatalogSource name. Defaults to the cluster-wide catalog. |
| `csvNames` | no | Array of ClusterServiceVersion names. Required when `csvMirror: true`. |
| `csvMirror` | no | Mirror the operator packages listed under `csvNames` instead of the package name. Requires `csvNames`. |
| `extraMirrorPackages` | no | Additional operator packages to mirror as bare entries without version constraints. |
| `global` | no | Configure the operator to watch all namespaces (`AllNamespaces` install mode). |
| `seed` | no | Deploy this operator before all others during the seed phase. |

Default operators installed (versions as of the current defaults):

| Operator | Version | Namespace |
|---|---|---|
| `quay-operator` | 3.15.3 | `quay-enterprise` |
| `multicluster-engine` | 2.10.2 | `multicluster-engine` (seed) |
| `advanced-cluster-management` | 2.15.1 | `open-cluster-management` (seed) |
| `cincinnati-operator` | 5.0.3 | `openshift-update-service` |
| `openshift-gitops-operator` | 1.19.2 | `openshift-gitops-operator` (global) |
| `openshift-pipelines-operator-rh` | 1.20.3 | `openshift-pipelines` (global) |
| `netobserv-operator` | 1.11.0 | `openshift-netobserv-operator` (global) |
| `cluster-logging` | 6.4.2 | `openshift-logging` |
| `loki-operator` | 6.4.2 | `openshift-operators-redhat` (global) |
| `redhat-oadp-operator` | 1.5.5 | `openshift-oadp` |
| `openshift-cert-manager-operator` | 1.18.1 | `cert-manager-operator` |
| `cluster-observability-operator` | 1.3.1 | `openshift-cluster-observability-operator` (global) |
| `openshift-external-secrets-operator` | 1.0.0 | `external-secrets-operator` (global) |
| `compliance-operator` | 1.8.2 | `openshift-compliance` |
| `metallb-operator` | 4.20.0-202602261925 | `metallb-system` (global) |

---

## Plugin System

Plugins live in `plugins/<name>/plugin.yaml`. Validated against `schemas/plugin.yaml`.

### `plugin.yaml` Fields

| Field | Required | Type | Description |
|---|---|---|---|
| `name` | yes | string | Plugin identifier. Must match the directory name. Pattern: `^[A-Za-z0-9][A-Za-z0-9._-]*$`. |
| `type` | yes | `foundation` \| `addon` | `foundation` plugins deploy before core OLM operators. `addon` plugins deploy after. |
| `order` | no | integer | Deployment order among plugins of the same type. Lower values deploy first. |
| `mirror` | no | `core` \| `plugin` \| `none` | Image mirroring strategy. `core`: included in the main oc-mirror run. `plugin`: the plugin runs its own oc-mirror invocation during deploy. `none`: no mirroring. |
| `catalog` | no | `redhat` \| `certified` | Which operator catalog to include in the plugin's image set. Defaults to `redhat`. |
| `operators` | no | array | OLM operators to install. Same schema as `defaults/operators.yaml` entries. |
| `installOperators` | no | boolean | Whether to install OLM operators during the deploy phase. Defaults to `true`. Set to `false` when the plugin installs operators via ACM policies instead. |
| `helm` | no | array | Helm charts to install, in order. Runs after OLM operators and before `tasks/deploy.yaml`. |
| `defaults` | no | object | Ansible variables loaded into scope before plugin tasks run. These values are overrideable in `global.yaml`. |
| `registries` | no | array | Registry mirror entries for MCE custom-registries patching. Each entry has `location` (source registry) and `mirror` (path under the internal Quay registry). |
| `additionalImages` | no | array of strings | Additional image tags or digests to include in the image set. |
| `blockedImages` | no | array of strings | Images to block from mirroring (full tags, digests, or patterns). |
| `requires.vars` | no | array | Ansible variables that must be defined before the plugin runs. Each entry: `{name, description?, when?}`. `when` is a Jinja2 condition; the check is skipped when it evaluates to false. |
| `requires.files` | no | array | Files that must exist within the plugin directory. Each entry: `{path, description?, when?}`. |

### `helm` Entry Fields

| Field | Required | Description |
|---|---|---|
| `release` | yes | Helm release name. |
| `namespace` | yes | Kubernetes namespace for the release. |
| `chart` | conditional | Chart name when `repo` is set, or local path relative to the plugin directory. Defaults to `charts/<release>` for local charts. Required when `repo` is present. |
| `repo` | no | Remote Helm repository URL. When set, the chart is fetched from this repo rather than the plugin directory. |
| `version` | no | Chart version constraint (e.g. `1.2.3`). |
| `valuesTemplate` | no | Path to a Jinja2 values template, relative to the plugin directory. Mutually exclusive with `valuesFile`. |
| `valuesFile` | no | Path to a static values file, relative to the plugin directory. Mutually exclusive with `valuesTemplate`. |
| `createNamespace` | no | Pass `--create-namespace` to Helm. Default: `true`. |
| `timeout` | no | Helm operation timeout (e.g. `15m`). Default: `15m`. |
| `wait` | no | Pass `--wait` to Helm. Default: `true`. |

---

## Built-in Plugins

### `plugins/lvms/` — LVMS (Logical Volume Manager Storage)

- **Type:** `foundation` — deploys before core operators
- **Order:** `10`
- **Mirror:** `core` — images included in the main oc-mirror run

Installs `lvms-operator` into `openshift-storage`. The following defaults are set by `plugin.yaml` and can be overridden in `global.yaml`:

#### `lvmsDefaults`

Controls the `LVMCluster` custom resource spec.

| Field | Default | Description |
|---|---|---|
| `lvmsDefaults.deviceClassName` | `vg1` | Name of the LVM device class. Also used as the StorageClass name suffix (`lvms-vg1`) and the Quay PVC StorageClass (`lvms-vg1`). |
| `lvmsDefaults.defaultStorageClass` | `true` | Whether this device class is the cluster's default StorageClass. |
| `lvmsDefaults.thinPoolConfig.name` | `vg1-pool-1` | Name of the LVM thin pool. |
| `lvmsDefaults.thinPoolConfig.sizePercent` | `90` | Percentage of the volume group used for the thin pool. |
| `lvmsDefaults.thinPoolConfig.overprovisionRatio` | `10` | Thin pool overprovisioning ratio. |

#### `lvmsConfigDefaults`

Merged with the user-supplied `lvmsConfig` only when `lvmsConfig.deviceSelector.optionalPaths` is defined.

| Field | Default | Description |
|---|---|---|
| `lvmsConfigDefaults.deviceSelector.forceWipeDevicesAndDestroyAllData` | `true` | Whether LVMS wipes devices before use. Set to `false` to preserve existing data on specified disks. |

---

### `plugins/odf/` — ODF (OpenShift Data Foundation)

- **Type:** `foundation`
- **Order:** `10`
- **Mirror:** `core`

Installs `odf-operator` and creates an external `StorageCluster` connected to a pre-existing Ceph cluster. Requires `odfExternalConfig` in `global.yaml`.

#### `odfDefaults`

| Field | Default | Description |
|---|---|---|
| `odfDefaults.defaultStorageClass` | `true` | Whether the ODF Ceph block pool StorageClass is the cluster default. |

The `StorageCluster` spec (encryption settings, external storage mode, `cephObjectStores` reconcile strategy) is hardcoded in `tasks/deploy.yaml` and is not user-configurable without modifying that file.

---

### `plugins/openshift-ai/` — OpenShift AI (RHOAI)

- **Type:** `addon`
- **Order:** `100`
- **Mirror:** `plugin` — runs its own oc-mirror invocation
- **`installOperators: false`** — operators are installed via ACM policies, not directly by OLM

No plugin-level overrideable defaults (`defaults: {}`).

Deploys via ACM policies applied to managed clusters:

1. `model-catalogsources` — Creates plugin-specific CatalogSources for the mirrored catalog. Also creates a `redhat-operators` alias CatalogSource as a workaround for OCPBUGS-78330 (temporary, until OCP 4.22).
2. `model-prerequisites` — Installs NFD, cert-manager, Service Mesh, RHCL, Leader Worker Set, and RHOAI operators via Subscriptions.
3. `model-operators-crd-ready` — Gates on the `LeaderWorkerSet` CRD being registered before proceeding.
4. `model-configuration` — Creates the `DataScienceCluster` CR with KServe `Managed` and all other components set to `Removed`.

The `DataScienceCluster` component configuration and all Subscription channels are hardcoded in the template files (`files/*.yaml.j2`). There are no user-overrideable variables beyond the global catalog/cluster variables consumed from defaults.

---

### `plugins/nvidia-gpu/` — NVIDIA GPU Operator

- **Type:** `addon`
- **Order:** `110`
- **Mirror:** `plugin`
- **Catalog:** `certified` — uses the Certified Operator catalog
- **`installOperators: false`** — installed via ACM policies

No plugin-level overrideable defaults (`defaults: {}`).

Deploys via ACM policies:

1. `nvidia-gpu-catalogsource` — Creates a plugin-specific CatalogSource from the mirrored certified catalog.
2. `nvidia-gpu-configuration` — Installs the `gpu-operator-certified` Subscription and creates the `ClusterPolicy` CR.

The `ClusterPolicy` spec (driver settings, DCGM exporter, MIG strategy, CDI, vGPU, GDS, GDRCopy, etc.) is fully hardcoded in `files/10-policy.yaml.j2`. There are no user-overrideable variables.

---

### `plugins/example/` — Example Plugin

- **Type:** `addon`
- **Order:** `999`

A minimal template plugin demonstrating the plugin structure.

#### `defaults`

| Field | Default | Description |
|---|---|---|
| `example_message` | `Hello from example plugin` | Message printed by the deploy task. Override in `global.yaml` to customize. |
