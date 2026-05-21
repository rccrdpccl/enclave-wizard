package defaults

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0640); err != nil {
		t.Fatal(err)
	}
}

func TestReadDeploymentDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "defaults/deployment.yaml", `---
disconnected: true
masterMaxPods: 500
diskEncryption: false
ocMirrorLogLevel: info
storage_plugin: lvms
enabled_plugins:
  - "{{ storage_plugin }}"
pullSecretPath: "{{ workingDir }}/config/pull-secret.json"
`)

	d, err := NewReader(dir).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if d.Disconnected != true {
		t.Errorf("Disconnected: want true, got %v", d.Disconnected)
	}
	if d.MasterMaxPods != 500 {
		t.Errorf("MasterMaxPods: want 500, got %d", d.MasterMaxPods)
	}
	if d.DiskEncryption != false {
		t.Errorf("DiskEncryption: want false, got %v", d.DiskEncryption)
	}
	if d.OCMirrorLogLevel != "info" {
		t.Errorf("OCMirrorLogLevel: want info, got %s", d.OCMirrorLogLevel)
	}
	if d.StoragePlugin != "lvms" {
		t.Errorf("StoragePlugin: want lvms, got %s", d.StoragePlugin)
	}
}

func TestReadPluginDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "defaults/deployment.yaml", `---
disconnected: true
masterMaxPods: 500
diskEncryption: false
ocMirrorLogLevel: info
storage_plugin: lvms
`)

	writeFile(t, dir, "plugins/lvms/plugin.yaml", `---
name: lvms
type: foundation
defaults:
  lvmsConfigDefaults:
    deviceSelector:
      forceWipeDevicesAndDestroyAllData: true
  lvmsDefaults:
    deviceClassName: vg1
    defaultStorageClass: true
    thinPoolConfig:
      name: vg1-pool-1
      sizePercent: 90
      overprovisionRatio: 10
`)

	writeFile(t, dir, "plugins/odf/plugin.yaml", `---
name: odf
type: foundation
defaults:
  odfDefaults:
    defaultStorageClass: true
`)

	d, err := NewReader(dir).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if d.LVMSDefaults == nil {
		t.Fatal("LVMSDefaults is nil")
	}
	if d.LVMSDefaults.DeviceClassName != "vg1" {
		t.Errorf("LVMS deviceClassName: want vg1, got %s", d.LVMSDefaults.DeviceClassName)
	}
	if d.LVMSDefaults.DefaultStorageClass != true {
		t.Errorf("LVMS defaultStorageClass: want true, got %v", d.LVMSDefaults.DefaultStorageClass)
	}
	if d.LVMSDefaults.ThinPoolConfig.Name != "vg1-pool-1" {
		t.Errorf("LVMS thinPoolConfig.name: want vg1-pool-1, got %s", d.LVMSDefaults.ThinPoolConfig.Name)
	}
	if d.LVMSDefaults.ThinPoolConfig.SizePercent != 90 {
		t.Errorf("LVMS thinPoolConfig.sizePercent: want 90, got %d", d.LVMSDefaults.ThinPoolConfig.SizePercent)
	}
	if d.LVMSDefaults.ThinPoolConfig.OverprovisionRatio != 10 {
		t.Errorf("LVMS thinPoolConfig.overprovisionRatio: want 10, got %d", d.LVMSDefaults.ThinPoolConfig.OverprovisionRatio)
	}

	if d.ODFDefaults == nil {
		t.Fatal("ODFDefaults is nil")
	}
	if d.ODFDefaults.DefaultStorageClass != true {
		t.Errorf("ODF defaultStorageClass: want true, got %v", d.ODFDefaults.DefaultStorageClass)
	}
}

func TestReadVastCSIPluginDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "defaults/deployment.yaml", `---
disconnected: true
masterMaxPods: 500
diskEncryption: false
ocMirrorLogLevel: info
storage_plugin: vast-csi
`)

	writeFile(t, dir, "plugins/vast-csi/plugin.yaml", `---
name: vast-csi
type: foundation
defaults:
  vastDefaults:
    infraTenant: infra
    storagePath: /osac
    viewPolicyId: 1
    quayPvcSize: 1000Gi
    tiers:
      - name: quay
        protocol: nfs
`)

	d, err := NewReader(dir).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if d.VASTDefaults == nil {
		t.Fatal("VASTDefaults is nil")
	}
	if d.VASTDefaults.InfraTenant == nil || *d.VASTDefaults.InfraTenant != "infra" {
		t.Errorf("vastDefaults.infraTenant: want infra, got %v", d.VASTDefaults.InfraTenant)
	}
	if d.VASTDefaults.StoragePath == nil || *d.VASTDefaults.StoragePath != "/osac" {
		t.Errorf("vastDefaults.storagePath: want /osac, got %v", d.VASTDefaults.StoragePath)
	}
	if d.VASTDefaults.ViewPolicyID == nil || *d.VASTDefaults.ViewPolicyID != 1 {
		t.Errorf("vastDefaults.viewPolicyId: want 1, got %v", d.VASTDefaults.ViewPolicyID)
	}
	if d.VASTDefaults.QuayPvcSize == nil || *d.VASTDefaults.QuayPvcSize != "1000Gi" {
		t.Errorf("vastDefaults.quayPvcSize: want 1000Gi, got %v", d.VASTDefaults.QuayPvcSize)
	}
	if len(d.VASTDefaults.Tiers) != 1 {
		t.Fatalf("vastDefaults.tiers: want 1 tier, got %d", len(d.VASTDefaults.Tiers))
	}
	if d.VASTDefaults.Tiers[0].Name != "quay" {
		t.Errorf("vastDefaults.tiers[0].name: want quay, got %s", d.VASTDefaults.Tiers[0].Name)
	}
	if d.VASTDefaults.Tiers[0].Protocol != "nfs" {
		t.Errorf("vastDefaults.tiers[0].protocol: want nfs, got %s", d.VASTDefaults.Tiers[0].Protocol)
	}
}

func TestMissingFilesReturnZeros(t *testing.T) {
	dir := t.TempDir()

	d, err := NewReader(dir).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if d.Disconnected != false {
		t.Errorf("Disconnected: want false, got %v", d.Disconnected)
	}
	if d.MasterMaxPods != 0 {
		t.Errorf("MasterMaxPods: want 0, got %d", d.MasterMaxPods)
	}
	if d.LVMSDefaults != nil {
		t.Errorf("LVMSDefaults: want nil, got %+v", d.LVMSDefaults)
	}
	if d.ODFDefaults != nil {
		t.Errorf("ODFDefaults: want nil, got %+v", d.ODFDefaults)
	}
	if d.VASTDefaults != nil {
		t.Errorf("VASTDefaults: want nil, got %+v", d.VASTDefaults)
	}
}

func TestPluginWithEmptyDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "plugins/nvidia-gpu/plugin.yaml", `---
name: nvidia-gpu
type: addon
defaults: {}
`)

	d, err := NewReader(dir).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if d.LVMSDefaults != nil {
		t.Errorf("LVMSDefaults: want nil, got %+v", d.LVMSDefaults)
	}
}
