package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromDir(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "lvms", `---
name: lvms
type: foundation
order: 10
`)
	writePlugin(t, dir, "odf", `---
name: odf
type: foundation
order: 10
`)
	writePlugin(t, dir, "nvidia-gpu", `---
name: nvidia-gpu
type: addon
order: 110
`)

	plugins, err := LoadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(plugins))
	}
	if plugins[0].Name != "lvms" && plugins[0].Name != "odf" {
		t.Errorf("expected foundation plugin first, got %s", plugins[0].Name)
	}
	if plugins[2].Name != "nvidia-gpu" {
		t.Errorf("expected nvidia-gpu last (order 110), got %s", plugins[2].Name)
	}
}

func TestLoadFromDir_SkipsExample(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "lvms", `---
name: lvms
type: foundation
order: 10
`)
	writePlugin(t, dir, "example", `---
name: example
type: addon
order: 999
`)

	plugins, err := LoadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin (example skipped), got %d", len(plugins))
	}
}

func TestLoadFromDir_SkipsInvalid(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "good", `---
name: good
type: foundation
`)
	writePlugin(t, dir, "bad", `not: valid: yaml: [`)

	plugins, err := LoadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin (bad skipped), got %d", len(plugins))
	}
}

func TestLoadFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "plugins"), 0750)

	plugins, err := LoadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func writePlugin(t *testing.T, baseDir, name, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, "plugins", name)
	os.MkdirAll(dir, 0750)
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(content), 0640)
}
