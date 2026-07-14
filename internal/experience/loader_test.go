package experience

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromDir_ValidExperiences(t *testing.T) {
	experiences, err := LoadFromDir("testdata")
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if len(experiences) != 2 {
		t.Fatalf("expected 2 experiences, got %d", len(experiences))
	}

	// Sorted alphabetically by ID
	assertEqual(t, "first.id", "caas", experiences[0].ID)
	assertEqual(t, "first.name", "Containers as a Service", experiences[0].Name)
	if len(experiences[0].Plugins) != 5 {
		t.Errorf("caas plugins: expected 5, got %d", len(experiences[0].Plugins))
	}

	assertEqual(t, "second.id", "vmaas", experiences[1].ID)
	assertEqual(t, "second.name", "VMs as a Service", experiences[1].Name)
	if len(experiences[1].Plugins) != 6 {
		t.Errorf("vmaas plugins: expected 6, got %d", len(experiences[1].Plugins))
	}
}

func TestLoadFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "experiences"), 0755)

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if len(experiences) != 0 {
		t.Errorf("expected 0 experiences, got %d", len(experiences))
	}
}

func TestLoadFromDir_NoExperiencesDir(t *testing.T) {
	dir := t.TempDir()

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir should not error on missing dir: %v", err)
	}
	if experiences != nil {
		t.Errorf("expected nil, got %v", experiences)
	}
}

func TestLoadFromDir_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	expDir := filepath.Join(dir, "experiences", "broken")
	os.MkdirAll(expDir, 0755)
	os.WriteFile(filepath.Join(expDir, "experience.yaml"), []byte(":\n  :\n  bad: [yaml"), 0644)

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	// Malformed YAML is skipped, not an error
	if len(experiences) != 0 {
		t.Errorf("expected 0 experiences (malformed skipped), got %d", len(experiences))
	}
}

func TestLoadFromDir_IDFromDirName(t *testing.T) {
	dir := t.TempDir()
	expDir := filepath.Join(dir, "experiences", "my-exp")
	os.MkdirAll(expDir, 0755)
	os.WriteFile(filepath.Join(expDir, "experience.yaml"), []byte("name: My Experience\ndescription: test\n"), 0644)

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if len(experiences) != 1 {
		t.Fatalf("expected 1 experience, got %d", len(experiences))
	}
	assertEqual(t, "id", "my-exp", experiences[0].ID)
	assertEqual(t, "name", "My Experience", experiences[0].Name)
}

func TestLoadFromDir_SkipsMissingYAML(t *testing.T) {
	dir := t.TempDir()
	// A subdirectory without experience.yaml
	os.MkdirAll(filepath.Join(dir, "experiences", "no-yaml"), 0755)
	// A regular file (not a directory) in experiences/
	os.WriteFile(filepath.Join(dir, "experiences", "README.md"), []byte("# hello"), 0644)

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if len(experiences) != 0 {
		t.Errorf("expected 0 experiences, got %d", len(experiences))
	}
}

func TestLoadFromDir_PluginOrderPreserved(t *testing.T) {
	dir := t.TempDir()
	expDir := filepath.Join(dir, "experiences", "ordered")
	os.MkdirAll(expDir, 0755)
	os.WriteFile(filepath.Join(expDir, "experience.yaml"), []byte(`
id: ordered
name: Ordered Experience
description: test plugin ordering
plugins:
  - name: zeta
    order: 300
  - name: alpha
    order: 100
  - name: beta
    order: 200
`), 0644)

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if len(experiences) != 1 {
		t.Fatalf("expected 1 experience, got %d", len(experiences))
	}
	plugins := experiences[0].Plugins
	if len(plugins) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(plugins))
	}
	assertEqual(t, "plugin[0]", "zeta", plugins[0].Name)
	assertEqual(t, "plugin[0].order", 300, plugins[0].Order)
	assertEqual(t, "plugin[1]", "alpha", plugins[1].Name)
	assertEqual(t, "plugin[1].order", 100, plugins[1].Order)
	assertEqual(t, "plugin[2]", "beta", plugins[2].Name)
	assertEqual(t, "plugin[2].order", 200, plugins[2].Order)
}

func TestLoadFromDir_ExtraFieldsIgnored(t *testing.T) {
	dir := t.TempDir()
	expDir := filepath.Join(dir, "experiences", "extra")
	os.MkdirAll(expDir, 0755)
	os.WriteFile(filepath.Join(expDir, "experience.yaml"), []byte(`
id: extra
name: Extra Fields
description: has unknown fields
version: "2.0"
author: someone
plugins:
  - name: foo
    order: 1
    custom_field: bar
`), 0644)

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if len(experiences) != 1 {
		t.Fatalf("expected 1 experience, got %d", len(experiences))
	}
	assertEqual(t, "id", "extra", experiences[0].ID)
	assertEqual(t, "name", "Extra Fields", experiences[0].Name)
	if len(experiences[0].Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(experiences[0].Plugins))
	}
}

func TestLoadFromDir_DuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	// Two directories with experience.yaml files specifying the same ID.
	for _, sub := range []string{"aaa-first", "zzz-second"} {
		expDir := filepath.Join(dir, "experiences", sub)
		os.MkdirAll(expDir, 0755)
		os.WriteFile(filepath.Join(expDir, "experience.yaml"), []byte("id: same-id\nname: "+sub+"\ndescription: dup test\n"), 0644)
	}

	experiences, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	// Both are loaded; the loader does not deduplicate.
	if len(experiences) != 2 {
		t.Fatalf("expected 2 experiences (no dedup), got %d", len(experiences))
	}
	assertEqual(t, "both.id", "same-id", experiences[0].ID)
	assertEqual(t, "both.id", "same-id", experiences[1].ID)
}

func assertEqual[T comparable](t *testing.T, field string, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %v, got %v", field, want, got)
	}
}
