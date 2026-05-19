package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)

func TestNewValidator_Unavailable(t *testing.T) {
	v := NewValidator("/nonexistent/path")
	if v.available {
		t.Error("expected validator to be unavailable for nonexistent path")
	}
}

func TestValidate_SkipsWhenUnavailable(t *testing.T) {
	v := &Validator{enclaveDir: "/nonexistent", available: false}
	errs := v.Validate(&models.EnclaveConfig{})
	if errs != nil {
		t.Errorf("expected nil errors when unavailable, got %v", errs)
	}
}

func TestParseAnsibleErrors_JSONCallback(t *testing.T) {
	output := `{"plays":[{"tasks":[{"hosts":{"localhost":{"failed":true,"msg":"Validation errors","errors":[{"message":"'odfExternalConfig' is a required property","data_path":"","schema_path":"allOf.1.if.then.required"}]}}}]}]}`
	errs := parseAnsibleErrors(output)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Message != "'odfExternalConfig' is a required property" {
		t.Errorf("got %q", errs[0].Message)
	}
	if errs[0].Field != "allOf.1.if.then.required" {
		t.Errorf("expected field 'allOf.1.if.then.required', got %q", errs[0].Field)
	}
}

func TestParseAnsibleErrors_JSONCallbackMultipleErrors(t *testing.T) {
	output := `{"plays":[{"tasks":[{"hosts":{"localhost":{"failed":true,"msg":"Validation errors","errors":[{"message":"error one","data_path":"field1","schema_path":""},{"message":"error two","data_path":"field2","schema_path":""}]}}}]}]}`
	errs := parseAnsibleErrors(output)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "field1" {
		t.Errorf("expected field1, got %q", errs[0].Field)
	}
}

func TestParseAnsibleErrors_FatalFallback(t *testing.T) {
	output := `fatal: [localhost]: FAILED! => {"changed": false, "msg": "simple error message"}`
	errs := parseAnsibleErrors(output)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Message != "simple error message" {
		t.Errorf("expected 'simple error message', got %q", errs[0].Message)
	}
}

func TestParseAnsibleErrors_NoFatalLine(t *testing.T) {
	output := "PLAY RECAP\nlocalhost: ok=0 changed=0 unreachable=0 failed=1"
	errs := parseAnsibleErrors(output)
	if len(errs) != 1 {
		t.Fatalf("expected 1 fallback error, got %d", len(errs))
	}
	if !contains(errs[0].Message, "Schema validation failed") {
		t.Errorf("expected fallback error message, got %q", errs[0].Message)
	}
}

func TestParseAnsibleErrors_EmptyOutput(t *testing.T) {
	errs := parseAnsibleErrors("")
	if len(errs) != 1 {
		t.Fatalf("expected 1 fallback error, got %d", len(errs))
	}
}

func TestParseAnsibleErrors_JSONCallbackMsgOnly(t *testing.T) {
	output := `{"plays":[{"tasks":[{"hosts":{"localhost":{"failed":true,"msg":"Validation errors were found.\nAt 'required' 'odfExternalConfig' is a required property."}}}]}]}`
	errs := parseAnsibleErrors(output)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !contains(errs[0].Message, "odfExternalConfig") {
		t.Errorf("expected error about odfExternalConfig, got %q", errs[0].Message)
	}
}

func TestStructToMap_RemovesNulls(t *testing.T) {
	type inner struct {
		A string  `json:"a"`
		B *string `json:"b"`
	}
	m := structToMap(inner{A: "val", B: nil})
	if _, ok := m["b"]; ok {
		t.Error("expected null field 'b' to be removed")
	}
	if m["a"] != "val" {
		t.Errorf("expected a=val, got %v", m["a"])
	}
}

func TestStructToMap_PreservesValues(t *testing.T) {
	s := "hello"
	type inner struct {
		X string  `json:"x"`
		Y *string `json:"y"`
	}
	m := structToMap(inner{X: "world", Y: &s})
	if m["x"] != "world" {
		t.Errorf("expected x=world, got %v", m["x"])
	}
	if m["y"] != "hello" {
		t.Errorf("expected y=hello, got %v", m["y"])
	}
}

func TestWriteConfigToDir_MapsAndStripsFields(t *testing.T) {
	dir := t.TempDir()

	cfg := &models.EnclaveConfig{}
	cfg.Global.BlockStorageBackend = "odf"
	cfg.Global.BaseDomain = "test.local"

	if err := writeConfigToDir(dir, cfg); err != nil {
		t.Fatalf("writeConfigToDir failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "global.yaml"))
	if err != nil {
		t.Fatalf("read global.yaml: %v", err)
	}

	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse global.yaml: %v", err)
	}

	for _, field := range wizardOnlyFields {
		if _, ok := m[field]; ok {
			t.Errorf("wizard-only field %q should have been stripped", field)
		}
	}
	if _, ok := m["blockStorageBackend"]; ok {
		t.Error("blockStorageBackend should have been mapped to storage_plugin")
	}
	if m["storage_plugin"] != "odf" {
		t.Errorf("expected storage_plugin=odf, got %v", m["storage_plugin"])
	}

	if m["baseDomain"] != "test.local" {
		t.Errorf("expected baseDomain=test.local, got %v", m["baseDomain"])
	}
}

func TestWriteConfigToDir_WritesAllFiles(t *testing.T) {
	dir := t.TempDir()
	cfg := &models.EnclaveConfig{}

	if err := writeConfigToDir(dir, cfg); err != nil {
		t.Fatalf("writeConfigToDir failed: %v", err)
	}

	for _, name := range []string{"global.yaml", "certificates.yaml", "cloud_infra.yaml"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist", name)
		}
	}
}

func TestWriteConfigToDir_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	cfg := &models.EnclaveConfig{}
	cfg.Global.BaseDomain = "example.com"
	cfg.Global.ClusterName = "mgmt"

	if err := writeConfigToDir(dir, cfg); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"global.yaml", "certificates.yaml", "cloud_infra.yaml"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := yaml.Unmarshal(data, &m); err != nil {
			t.Errorf("%s is not valid YAML: %v", name, err)
		}
	}
}

func TestWriteConfigToDir_CertificatesPreserved(t *testing.T) {
	dir := t.TempDir()
	cert := "-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----"
	cfg := &models.EnclaveConfig{}
	cfg.Certificates.SSLCACertificate = &cert

	if err := writeConfigToDir(dir, cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "certificates.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(data), "TEST") {
		t.Error("expected certificate content in certificates.yaml")
	}
}

func TestParseAnsibleErrors_MultipleFatalLines(t *testing.T) {
	output := `fatal: [localhost]: FAILED! => {"msg": "error one"}
TASK [another] ***
fatal: [localhost]: FAILED! => {"msg": "error two"}`
	errs := parseAnsibleErrors(output)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
