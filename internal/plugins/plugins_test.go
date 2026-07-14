package plugins

import (
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func testRegistry() *Registry {
	return NewRegistry([]models.Plugin{
		{Name: "lvms", Type: models.PluginTypeFoundation, Order: 10},
		{Name: "odf", Type: models.PluginTypeFoundation, Order: 10},
		{Name: "vast-csi", Type: models.PluginTypeAddon, Order: 10},
		{Name: "nvidia-gpu", Type: models.PluginTypeAddon, Order: 110},
		{Name: "openshift-ai", Type: models.PluginTypeAddon, Order: 100},
		{Name: "aap", Type: models.PluginTypeAddon, Order: 103},
		{Name: "authorino", Type: models.PluginTypeAddon, Order: 102},
		{Name: "cnv", Type: models.PluginTypeAddon, Order: 104},
		{Name: "trust-manager", Type: models.PluginTypeAddon, Order: 100},
	})
}

func TestRegistryAll(t *testing.T) {
	r := testRegistry()
	if len(r.All()) != 9 {
		t.Fatalf("expected 9 plugins, got %d", len(r.All()))
	}
}

func TestRegistryGet(t *testing.T) {
	r := testRegistry()
	p, ok := r.Get("lvms")
	if !ok {
		t.Fatal("Get(lvms) returned false")
	}
	if p.Name != "lvms" {
		t.Errorf("Get(lvms).Name = %s", p.Name)
	}
	if p.Type != models.PluginTypeFoundation {
		t.Errorf("Get(lvms).Type = %s, want foundation", p.Type)
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) returned true")
	}
}

func TestValidateCombinationValid(t *testing.T) {
	r := testRegistry()
	errs := r.ValidateCombination([]string{"lvms", "nvidia-gpu"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateCombinationUnknown(t *testing.T) {
	r := testRegistry()
	errs := r.ValidateCombination([]string{"lvms", "bogus", "also-bogus"})
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
	if errs[0].Message != "unknown plugin: bogus" {
		t.Errorf("errs[0].Message = %s", errs[0].Message)
	}
}

func TestValidateCombinationCNV(t *testing.T) {
	r := testRegistry()
	errs := r.ValidateCombination([]string{"lvms", "cnv"})
	if len(errs) != 0 {
		t.Errorf("expected no errors for cnv, got %v", errs)
	}
}

func TestValidateCombinationAuthorino(t *testing.T) {
	r := testRegistry()
	errs := r.ValidateCombination([]string{"lvms", "authorino"})
	if len(errs) != 0 {
		t.Errorf("expected no errors for authorino, got %v", errs)
	}
}

func TestRegistrySetAndGetSchema(t *testing.T) {
	r := testRegistry()
	schema := []byte(`{"type":"object"}`)
	r.SetSchema("lvms", schema)

	got, err := r.GetSchema("lvms")
	if err != nil {
		t.Fatalf("GetSchema: %v", err)
	}
	if string(got) != string(schema) {
		t.Errorf("GetSchema = %s, want %s", got, schema)
	}
}

func TestRegistryGetSchema_NotFound(t *testing.T) {
	r := testRegistry()

	_, err := r.GetSchema("lvms")
	if err == nil {
		t.Fatal("expected error for missing schema")
	}
	if !errors.Is(err, ErrSchemaNotFound) {
		t.Errorf("expected ErrSchemaNotFound, got %v", err)
	}
}

func TestRegistryGetSchema_UnknownPlugin(t *testing.T) {
	r := testRegistry()

	_, err := r.GetSchema("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown plugin schema")
	}
	if !errors.Is(err, ErrSchemaNotFound) {
		t.Errorf("expected ErrSchemaNotFound, got %v", err)
	}
}
