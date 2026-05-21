package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)

func TestParseFailedEvents_StructuredErrors(t *testing.T) {
	event := map[string]any{
		"event": "runner_on_failed",
		"event_data": map[string]any{
			"task": "Validate merged config",
			"res": map[string]any{
				"msg": "Validation errors",
				"errors": []any{
					map[string]any{
						"message":     "'odfExternalConfig' is a required property",
						"data_path":   "",
						"schema_path": "allOf.1.if.then.required",
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(event)
	errs := parseFailedEvents([]json.RawMessage{raw})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Message != "'odfExternalConfig' is a required property" {
		t.Errorf("got %q", errs[0].Message)
	}
	if errs[0].Field != "allOf.1.if.then.required" {
		t.Errorf("expected schema_path as field, got %q", errs[0].Field)
	}
}

func TestParseFailedEvents_MsgFallback(t *testing.T) {
	event := map[string]any{
		"event": "runner_on_failed",
		"event_data": map[string]any{
			"task": "Validate config",
			"res":  map[string]any{"msg": "simple error"},
		},
	}
	raw, _ := json.Marshal(event)
	errs := parseFailedEvents([]json.RawMessage{raw})
	if len(errs) != 1 || errs[0].Message != "simple error" {
		t.Errorf("expected 'simple error', got %v", errs)
	}
}

func TestParseFailedEvents_CensoredFallsBackToTaskName(t *testing.T) {
	event := map[string]any{
		"event": "runner_on_failed",
		"event_data": map[string]any{
			"task": "Validate merged config against variables schema",
			"res": map[string]any{
				"censored": "the output has been hidden",
				"changed":  false,
			},
		},
	}
	raw, _ := json.Marshal(event)
	errs := parseFailedEvents([]json.RawMessage{raw})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Message != "task failed: Validate merged config against variables schema" {
		t.Errorf("expected task name in message, got %q", errs[0].Message)
	}
}

func TestParseFailedEvents_IgnoresNonFailure(t *testing.T) {
	event := map[string]any{
		"event":      "runner_on_ok",
		"event_data": map[string]any{"res": map[string]any{"msg": "ignored"}},
	}
	raw, _ := json.Marshal(event)
	errs := parseFailedEvents([]json.RawMessage{raw})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}
}

func TestNewValidator_Unavailable(t *testing.T) {
	v := NewValidator("/nonexistent/path", nil)
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

func TestStructToMap_RemovesNulls(t *testing.T) {
	type inner struct {
		A string  `json:"a"`
		B *string `json:"b"`
	}
	m := structToMap(inner{A: "val", B: nil})
	if _, ok := m["b"]; ok {
		t.Error("expected null field removed")
	}
}

func TestWriteConfigToDir_StoragePlugin(t *testing.T) {
	dir := t.TempDir()
	cfg := &models.EnclaveConfig{}
	cfg.Global.StoragePlugin = "odf"
	cfg.Global.BaseDomain = "test.local"

	if err := writeConfigToDir(dir, cfg); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "global.yaml"))
	var m map[string]any
	yaml.Unmarshal(data, &m)

	if m["storage_plugin"] != "odf" {
		t.Errorf("expected storage_plugin=odf, got %v", m["storage_plugin"])
	}
}

func TestWriteConfigToDir_WritesAllFiles(t *testing.T) {
	dir := t.TempDir()
	if err := writeConfigToDir(dir, &models.EnclaveConfig{}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"global.yaml", "certificates.yaml", "cloud_infra.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to exist", name)
		}
	}
}

func TestWriteConfigToDir_CertificatesPreserved(t *testing.T) {
	dir := t.TempDir()
	cert := "-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----"
	cfg := &models.EnclaveConfig{}
	cfg.Certificates.SSLCACertificate = &cert

	writeConfigToDir(dir, cfg)

	data, _ := os.ReadFile(filepath.Join(dir, "certificates.yaml"))
	if !strings.Contains(string(data), "TEST") {
		t.Error("expected certificate content")
	}
}
