package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/experience"
)

func setupExperiencesAPI(experiences []experience.Experience) *httptest.Server {
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("test", "0.0.0"))
	NewExperiencesHandler(experiences).Register(api)
	return httptest.NewServer(mux)
}

func TestListExperiences_WithData(t *testing.T) {
	exps := []experience.Experience{
		{
			ID:          "caas",
			Name:        "Containers as a Service",
			Description: "CaaS experience",
			Plugins: []experience.ExperiencePlugin{
				{Name: "trust-manager", Order: 100},
				{Name: "osac", Order: 200},
			},
		},
		{
			ID:          "vmaas",
			Name:        "VMs as a Service",
			Description: "VMaaS experience",
			Plugins: []experience.ExperiencePlugin{
				{Name: "cnv", Order: 104},
			},
		},
	}

	srv := setupExperiencesAPI(exps)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/experiences")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var out struct {
		Experiences []experience.Experience `json:"experiences"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	if len(out.Experiences) != 2 {
		t.Fatalf("expected 2 experiences, got %d", len(out.Experiences))
	}
	assertEqual(t, "first.id", "caas", out.Experiences[0].ID)
	assertEqual(t, "first.plugins", 2, len(out.Experiences[0].Plugins))
	assertEqual(t, "second.id", "vmaas", out.Experiences[1].ID)
}

func TestListExperiences_Empty(t *testing.T) {
	srv := setupExperiencesAPI(nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/experiences")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var out struct {
		Experiences []experience.Experience `json:"experiences"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	if len(out.Experiences) != 0 {
		t.Errorf("expected 0 experiences, got %d", len(out.Experiences))
	}
}

func TestListExperiences_EncodesAsArray(t *testing.T) {
	srv := setupExperiencesAPI(nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/experiences")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	var raw map[string]json.RawMessage
	json.NewDecoder(resp.Body).Decode(&raw)
	if string(raw["experiences"]) == "null" {
		t.Error("experiences field must encode as [] not null")
	}
}
