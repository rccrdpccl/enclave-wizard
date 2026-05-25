package defaults

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, baseDir, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(baseDir, relPath)
	os.MkdirAll(filepath.Dir(fullPath), 0750)
	if err := os.WriteFile(fullPath, []byte(content), 0640); err != nil {
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
	if d.StoragePlugin != "lvms" {
		t.Errorf("StoragePlugin: want lvms, got %s", d.StoragePlugin)
	}
}

func TestReadPluginDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "defaults/deployment.yaml", `---
storage_plugin: lvms
`)

	writeFile(t, dir, "plugins/lvms/plugin.yaml", `---
name: lvms
type: foundation
defaults:
  lvmsDefaults:
    deviceClassName: vg1
    defaultStorageClass: true
    thinPoolConfig:
      name: vg1-pool-1
      sizePercent: 90
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

	if d.PluginDefaults == nil {
		t.Fatal("PluginDefaults is nil")
	}

	if _, ok := d.PluginDefaults["lvmsDefaults"]; !ok {
		t.Error("expected lvmsDefaults in PluginDefaults")
	}

	if _, ok := d.PluginDefaults["odfDefaults"]; !ok {
		t.Error("expected odfDefaults in PluginDefaults")
	}

	lvms, ok := d.PluginDefaults["lvmsDefaults"].(map[string]any)
	if !ok {
		t.Fatal("lvmsDefaults is not a map")
	}
	if lvms["deviceClassName"] != "vg1" {
		t.Errorf("lvmsDefaults.deviceClassName: want vg1, got %v", lvms["deviceClassName"])
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
	if d.PluginDefaults != nil {
		t.Errorf("PluginDefaults: want nil, got %v", d.PluginDefaults)
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

	if d.PluginDefaults != nil {
		t.Errorf("PluginDefaults: want nil for empty defaults, got %v", d.PluginDefaults)
	}
}
