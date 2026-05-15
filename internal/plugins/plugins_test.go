package plugins

import (
	"testing"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func TestAllPluginsPresent(t *testing.T) {
	want := []string{"lvms", "odf", "vast-csi", "nvidia-gpu", "openshift-ai"}
	if len(All) != len(want) {
		t.Fatalf("expected %d plugins, got %d", len(want), len(All))
	}
	for i, name := range want {
		if All[i].Name != name {
			t.Errorf("All[%d]: want %s, got %s", i, name, All[i].Name)
		}
	}
}

func TestAllPluginTypes(t *testing.T) {
	foundations := 0
	addons := 0
	for _, p := range All {
		switch p.Type {
		case models.PluginTypeFoundation:
			foundations++
		case models.PluginTypeAddon:
			addons++
		default:
			t.Errorf("plugin %s has unknown type %s", p.Name, p.Type)
		}
	}
	if foundations != 3 {
		t.Errorf("expected 3 foundation plugins, got %d", foundations)
	}
	if addons != 2 {
		t.Errorf("expected 2 addon plugins, got %d", addons)
	}
}

func TestGet(t *testing.T) {
	p, ok := Get("lvms")
	if !ok {
		t.Fatal("Get(lvms) returned false")
	}
	if p.Name != "lvms" {
		t.Errorf("Get(lvms).Name = %s", p.Name)
	}
	if p.Type != models.PluginTypeFoundation {
		t.Errorf("Get(lvms).Type = %s, want foundation", p.Type)
	}

	_, ok = Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) returned true")
	}
}

func TestValidateCombinationValid(t *testing.T) {
	errs := ValidateCombination([]string{"lvms", "nvidia-gpu"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateCombinationUnknown(t *testing.T) {
	errs := ValidateCombination([]string{"lvms", "bogus", "also-bogus"})
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
	if errs[0].Message != "unknown plugin: bogus" {
		t.Errorf("errs[0].Message = %s", errs[0].Message)
	}
	if errs[1].Message != "unknown plugin: also-bogus" {
		t.Errorf("errs[1].Message = %s", errs[1].Message)
	}
}
