package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func writePlugin(t *testing.T, dir, name, content string) {
	t.Helper()
	pluginDir := filepath.Join(dir, "plugins", name)
	if err := os.MkdirAll(pluginDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(content), 0640); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAllDiscoversPlugins(t *testing.T) {
	dir := t.TempDir()

	writePlugin(t, dir, "lvms", `---
name: lvms
type: foundation
order: 10
mirror: core
operators:
  - name: lvms-operator
    version: 4.20.0
    channel: stable-4.20
    init_version: 4.20.0
    namespace: openshift-storage
    source: cs-mirror-redhat-operators
defaults:
  lvmsDefaults:
    deviceClassName: vg1
registries:
  - location: "registry.redhat.io/lvms4"
    mirror: "lvms4"
requires:
  files:
    - path: "tasks/deploy.yaml"
      description: "LVMS deployment tasks"
`)

	writePlugin(t, dir, "nvidia-gpu", `---
name: nvidia-gpu
type: addon
order: 110
mirror: plugin
catalog: certified
operators:
  - name: gpu-operator-certified
    version: 25.10.1
    channel: v25.10
    init_version: 25.10.1
installOperators: false
defaults: {}
registries:
  - location: "nvcr.io/nvidia"
    mirror: "nvidia"
  - location: "registry.connect.redhat.com/nvidia"
    mirror: "nvidia"
`)

	writePlugin(t, dir, "odf", `---
name: odf
type: foundation
order: 10
mirror: core
`)

	loader := NewLoader(dir)
	plugins, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	if len(plugins) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(plugins))
	}

	// Foundation plugins should come before addon, sorted by order
	if plugins[0].Type != "foundation" || plugins[1].Type != "foundation" {
		t.Errorf("expected first two plugins to be foundation, got %s, %s", plugins[0].Type, plugins[1].Type)
	}
	if plugins[2].Type != "addon" {
		t.Errorf("expected last plugin to be addon, got %s", plugins[2].Type)
	}
	if plugins[2].Name != "nvidia-gpu" {
		t.Errorf("expected addon plugin to be nvidia-gpu, got %s", plugins[2].Name)
	}

	// Verify lvms fields
	var lvms *struct{ idx int }
	for i, p := range plugins {
		if p.Name == "lvms" {
			lvms = &struct{ idx int }{i}
			break
		}
	}
	if lvms == nil {
		t.Fatal("lvms plugin not found")
	}
	lp := plugins[lvms.idx]

	if lp.Order == nil || *lp.Order != 10 {
		t.Errorf("lvms order: want 10, got %v", lp.Order)
	}
	if lp.Mirror == nil || *lp.Mirror != "core" {
		t.Errorf("lvms mirror: want core, got %v", lp.Mirror)
	}
	if len(lp.Operators) != 1 {
		t.Fatalf("lvms operators: want 1, got %d", len(lp.Operators))
	}
	op := lp.Operators[0]
	if op.Name != "lvms-operator" {
		t.Errorf("operator name: want lvms-operator, got %s", op.Name)
	}
	if op.Version != "4.20.0" {
		t.Errorf("operator version: want 4.20.0, got %s", op.Version)
	}
	if op.Channel != "stable-4.20" {
		t.Errorf("operator channel: want stable-4.20, got %s", op.Channel)
	}
	if op.InitVersion != "4.20.0" {
		t.Errorf("operator init_version: want 4.20.0, got %s", op.InitVersion)
	}
	if op.Namespace != "openshift-storage" {
		t.Errorf("operator namespace: want openshift-storage, got %s", op.Namespace)
	}
	if op.Source != "cs-mirror-redhat-operators" {
		t.Errorf("operator source: want cs-mirror-redhat-operators, got %s", op.Source)
	}
	if len(lp.Registries) != 1 || lp.Registries[0].Location != "registry.redhat.io/lvms4" {
		t.Errorf("lvms registries: unexpected %+v", lp.Registries)
	}
	if lp.Requires == nil || len(lp.Requires.Files) != 1 {
		t.Fatalf("lvms requires.files: want 1, got %v", lp.Requires)
	}
	if lp.Requires.Files[0].Path != "tasks/deploy.yaml" {
		t.Errorf("requires file path: want tasks/deploy.yaml, got %s", lp.Requires.Files[0].Path)
	}

	// Verify nvidia-gpu fields
	gpu := plugins[2]
	if gpu.Catalog == nil || *gpu.Catalog != "certified" {
		t.Errorf("nvidia-gpu catalog: want certified, got %v", gpu.Catalog)
	}
	if gpu.InstallOperators == nil || *gpu.InstallOperators != false {
		t.Errorf("nvidia-gpu installOperators: want false, got %v", gpu.InstallOperators)
	}
	if len(gpu.Registries) != 2 {
		t.Errorf("nvidia-gpu registries: want 2, got %d", len(gpu.Registries))
	}
}

func TestLoadAllMissingPluginsDir(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir)
	plugins, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("expected nil error for missing plugins dir, got %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestLoadAllSkipsDirsWithoutPluginYaml(t *testing.T) {
	dir := t.TempDir()

	// Directory with plugin.yaml
	writePlugin(t, dir, "lvms", `---
name: lvms
type: foundation
order: 10
mirror: core
`)

	// Directory without plugin.yaml
	if err := os.MkdirAll(filepath.Join(dir, "plugins", "scripts"), 0750); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(dir)
	plugins, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name != "lvms" {
		t.Errorf("expected lvms, got %s", plugins[0].Name)
	}
}

func TestLoadAllRejectsInvalidYaml(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "bad", `{{invalid yaml`)

	loader := NewLoader(dir)
	_, err := loader.LoadAll()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}
