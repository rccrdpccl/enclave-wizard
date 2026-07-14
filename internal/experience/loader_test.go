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

func assertEqual[T comparable](t *testing.T, field string, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %v, got %v", field, want, got)
	}
}
