package plugins

import (
	"encoding/json"
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

func TestLoadFromDirWithSchemas_YAMLSchema(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "myplugin", `---
name: myplugin
type: addon
order: 100
`)
	writeSchema(t, dir, "myplugin", "config.yaml", `type: object
properties:
  replicas:
    type: integer
    minimum: 1
`)

	result, err := LoadFromDirWithSchemas(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result.Plugins))
	}
	schema, ok := result.Schemas["myplugin"]
	if !ok {
		t.Fatal("expected schema for myplugin")
	}
	if !json.Valid(schema) {
		t.Fatalf("schema is not valid JSON: %s", schema)
	}
	// Verify the schema contains expected fields
	var parsed map[string]any
	json.Unmarshal(schema, &parsed)
	if parsed["type"] != "object" {
		t.Errorf("expected type=object, got %v", parsed["type"])
	}
}

func TestLoadFromDirWithSchemas_JSONSchema(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "jsonplugin", `---
name: jsonplugin
type: addon
order: 50
`)
	writeSchema(t, dir, "jsonplugin", "config.json", `{"type":"object","properties":{"name":{"type":"string"}}}`)

	result, err := LoadFromDirWithSchemas(dir)
	if err != nil {
		t.Fatal(err)
	}
	schema, ok := result.Schemas["jsonplugin"]
	if !ok {
		t.Fatal("expected schema for jsonplugin")
	}
	if !json.Valid(schema) {
		t.Fatalf("schema is not valid JSON: %s", schema)
	}
}

func TestLoadFromDirWithSchemas_NoSchema(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "noplugin", `---
name: noplugin
type: foundation
order: 10
`)

	result, err := LoadFromDirWithSchemas(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result.Plugins))
	}
	if _, ok := result.Schemas["noplugin"]; ok {
		t.Error("did not expect schema for noplugin")
	}
}

func TestLoadFromDirWithSchemas_FromTestdata(t *testing.T) {
	result, err := LoadFromDirWithSchemas("testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Plugins) < 2 {
		t.Fatalf("expected at least 2 plugins from testdata, got %d", len(result.Plugins))
	}

	// with-schema should have a schema
	if _, ok := result.Schemas["with-schema"]; !ok {
		t.Error("expected schema for with-schema plugin")
	}
	// json-schema should have a schema
	if _, ok := result.Schemas["json-schema"]; !ok {
		t.Error("expected schema for json-schema plugin")
	}
	// no-schema should not have a schema
	if _, ok := result.Schemas["no-schema"]; ok {
		t.Error("did not expect schema for no-schema plugin")
	}
}

func TestLoadFromDirWithSchemas_MatchesDisk(t *testing.T) {
	result, err := LoadFromDirWithSchemas("testdata")
	if err != nil {
		t.Fatal(err)
	}

	// Read the JSON fixture directly and compare
	diskJSON, err := os.ReadFile("testdata/plugins/json-schema/schemas/config.json")
	if err != nil {
		t.Fatal(err)
	}

	schema, ok := result.Schemas["json-schema"]
	if !ok {
		t.Fatal("expected schema for json-schema plugin")
	}

	// Both should parse to the same JSON structure
	var diskParsed, schemaParsed any
	json.Unmarshal(diskJSON, &diskParsed)
	json.Unmarshal(schema, &schemaParsed)

	diskBytes, _ := json.Marshal(diskParsed)
	schemaBytes, _ := json.Marshal(schemaParsed)
	if string(diskBytes) != string(schemaBytes) {
		t.Errorf("schema mismatch:\ndisk:   %s\nloaded: %s", diskBytes, schemaBytes)
	}
}

func writePlugin(t *testing.T, baseDir, name, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, "plugins", name)
	os.MkdirAll(dir, 0750)
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(content), 0640)
}

func writeSchema(t *testing.T, baseDir, pluginName, filename, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, "plugins", pluginName, "schemas")
	os.MkdirAll(dir, 0750)
	os.WriteFile(filepath.Join(dir, filename), []byte(content), 0640)
}
