package plugins

import (
	"testing"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func testRegistry() *Registry {
	return NewRegistry([]models.Plugin{
		{Name: "lvms", Type: models.PluginTypeFoundation, Order: 10},
		{Name: "odf", Type: models.PluginTypeFoundation, Order: 10},
		{Name: "vast-csi", Type: models.PluginTypeFoundation, Order: 10},
		{Name: "nvidia-gpu", Type: models.PluginTypeAddon, Order: 110},
		{Name: "openshift-ai", Type: models.PluginTypeAddon, Order: 100},
	})
}

func TestRegistryAll(t *testing.T) {
	r := testRegistry()
	if len(r.All()) != 5 {
		t.Fatalf("expected 5 plugins, got %d", len(r.All()))
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
