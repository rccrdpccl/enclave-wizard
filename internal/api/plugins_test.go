package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
)

func setupPluginsAPI() *httptest.Server {
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("test", "0.0.0"))
	NewPluginsHandler().Register(api)
	return httptest.NewServer(mux)
}

// --- List plugins ---

func TestListPlugins_ReturnsOK(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	assertEqual(t, "status", http.StatusOK, resp.StatusCode)
}

func TestListPlugins_ReturnsAllPlugins(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	var out struct {
		Plugins []models.Plugin `json:"plugins"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(out.Plugins) != len(plugins.All) {
		t.Errorf("expected %d plugins, got %d", len(plugins.All), len(out.Plugins))
	}
}

func TestListPlugins_EachPluginHasNameTypeDescription(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	var out struct {
		Plugins []models.Plugin `json:"plugins"`
	}
	json.NewDecoder(resp.Body).Decode(&out)

	for _, p := range out.Plugins {
		if p.Name == "" {
			t.Errorf("plugin has empty name: %+v", p)
		}
		if p.Type == "" {
			t.Errorf("plugin %q has empty type", p.Name)
		}
		if p.Description == "" {
			t.Errorf("plugin %q has empty description", p.Name)
		}
	}
}

func TestListPlugins_ContainsExpectedPlugins(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	var out struct {
		Plugins []models.Plugin `json:"plugins"`
	}
	json.NewDecoder(resp.Body).Decode(&out)

	byName := make(map[string]models.Plugin, len(out.Plugins))
	for _, p := range out.Plugins {
		byName[p.Name] = p
	}

	for _, want := range []struct {
		name    string
		typ     models.PluginType
	}{
		{"lvms", models.PluginTypeFoundation},
		{"odf", models.PluginTypeFoundation},
		{"vast-csi", models.PluginTypeFoundation},
		{"nvidia-gpu", models.PluginTypeAddon},
		{"openshift-ai", models.PluginTypeAddon},
	} {
		got, ok := byName[want.name]
		if !ok {
			t.Errorf("plugin %q missing from response", want.name)
			continue
		}
		if got.Type != want.typ {
			t.Errorf("plugin %q: want type %q, got %q", want.name, want.typ, got.Type)
		}
	}
}

// --- Validate plugin combination ---

func postValidate(t *testing.T, srv *httptest.Server, pluginNames []string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"plugins": pluginNames})
	resp, err := http.Post(srv.URL+"/api/v1/plugins/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST validate: %v", err)
	}
	return resp
}

func TestValidateCombination_ValidSinglePlugin(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp := postValidate(t, srv, []string{"lvms"})
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var out struct {
		Valid  bool                     `json:"valid"`
		Errors []models.ValidationError `json:"errors"`
	}
	json.NewDecoder(resp.Body).Decode(&out)

	if !out.Valid {
		t.Errorf("expected valid=true, got errors: %v", out.Errors)
	}
	if len(out.Errors) != 0 {
		t.Errorf("expected no errors, got %v", out.Errors)
	}
}

func TestValidateCombination_ValidMultiplePlugins(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp := postValidate(t, srv, []string{"lvms", "nvidia-gpu", "openshift-ai"})
	defer resp.Body.Close()

	var out struct {
		Valid bool `json:"valid"`
	}
	json.NewDecoder(resp.Body).Decode(&out)

	if !out.Valid {
		t.Error("expected valid=true for known plugin combination")
	}
}

func TestValidateCombination_UnknownPlugin(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp := postValidate(t, srv, []string{"bogus-plugin"})
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var out struct {
		Valid  bool                     `json:"valid"`
		Errors []models.ValidationError `json:"errors"`
	}
	json.NewDecoder(resp.Body).Decode(&out)

	if out.Valid {
		t.Error("expected valid=false for unknown plugin")
	}
	if len(out.Errors) == 0 {
		t.Error("expected at least one error for unknown plugin")
	}
}

func TestValidateCombination_MixedKnownAndUnknown(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp := postValidate(t, srv, []string{"lvms", "not-a-plugin"})
	defer resp.Body.Close()

	var out struct {
		Valid  bool                     `json:"valid"`
		Errors []models.ValidationError `json:"errors"`
	}
	json.NewDecoder(resp.Body).Decode(&out)

	if out.Valid {
		t.Error("expected valid=false when any plugin is unknown")
	}
	if len(out.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(out.Errors))
	}
}

func TestValidateCombination_ErrorIncludesPluginName(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	resp := postValidate(t, srv, []string{"mystery-plugin"})
	defer resp.Body.Close()

	var out struct {
		Errors []models.ValidationError `json:"errors"`
	}
	json.NewDecoder(resp.Body).Decode(&out)

	if len(out.Errors) == 0 {
		t.Fatal("expected at least one error")
	}
	msg := out.Errors[0].Message
	if msg == "" {
		t.Error("error message is empty")
	}
	// The error should name the offending plugin.
	found := false
	for i := range msg {
		if i+len("mystery-plugin") <= len(msg) && msg[i:i+len("mystery-plugin")] == "mystery-plugin" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("error message %q does not mention the unknown plugin name", msg)
	}
}

func TestValidateCombination_EmptyBodyReturnsBadRequest(t *testing.T) {
	srv := setupPluginsAPI()
	defer srv.Close()

	// huma validates minItems:1 on the plugins field
	body, _ := json.Marshal(map[string]any{"plugins": []string{}})
	resp, err := http.Post(srv.URL+"/api/v1/plugins/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("expected non-200 for empty plugins list (minItems:1 constraint)")
	}
}
