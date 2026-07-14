package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"
	"go.uber.org/mock/gomock"
)

const testDeployEnclaveDir = "/opt/enclave"

func setupDeployAPI(r runner.Runner) (*httptest.Server, *DeployHandler) {
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("test", "0.0.0"))
	registry := plugins.NewRegistry([]models.Plugin{
		{Name: "lvms", Type: models.PluginTypeFoundation},
	})
	h := NewDeployHandler(r, registry, nil, nil, testDeployEnclaveDir)
	h.Register(api)
	return httptest.NewServer(mux), h
}

// --- POST /api/v1/deployments ---

func TestStartDeployment_CreatesAndReturns(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)
	run := sampleRun()
	m.EXPECT().Start(runner.StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "playbooks/main.yaml",
		ExtraVars: map[string]string{
			"workingDir": testDeployEnclaveDir,
		},
	}).Return(run, nil)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/v1/deployments", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var got models.Deployment
	json.NewDecoder(resp.Body).Decode(&got)
	assertEqual(t, "id", "run-123", got.ID)
	assertEqual(t, "status", models.TaskStatusRunning, got.Status)
	if got.StartedAt.IsZero() {
		t.Error("StartedAt should be set")
	}
	if len(got.Phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(got.Phases))
	}
	assertEqual(t, "phase_name", "main", got.Phases[0].Name)
}

func TestStartDeployment_Busy(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)
	m.EXPECT().Start(gomock.Any()).Return(nil, runner.ErrBusy)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/v1/deployments", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()
	assertEqual(t, "status", http.StatusConflict, resp.StatusCode)
}

// --- GET /api/v1/deployments/current ---

func TestGetCurrentDeployment_ReturnsLatest(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)
	run := sampleRun()
	m.EXPECT().Start(gomock.Any()).Return(run, nil)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	// First create a deployment
	resp, err := http.Post(srv.URL+"/api/v1/deployments", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	// Now get current
	resp, err = http.Get(srv.URL + "/api/v1/deployments/current")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var got models.Deployment
	json.NewDecoder(resp.Body).Decode(&got)
	assertEqual(t, "id", "run-123", got.ID)
}

func TestGetCurrentDeployment_NoDeployment_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/deployments/current")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	assertEqual(t, "status", http.StatusNotFound, resp.StatusCode)
}

// --- GET /api/v1/deployments/{id} ---

func TestGetDeploymentByID_ReturnsDeployment(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)
	run := sampleRun()
	m.EXPECT().Start(gomock.Any()).Return(run, nil)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	// Create deployment
	resp, err := http.Post(srv.URL+"/api/v1/deployments", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	// Get by ID
	resp, err = http.Get(srv.URL + "/api/v1/deployments/run-123")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var got models.Deployment
	json.NewDecoder(resp.Body).Decode(&got)
	assertEqual(t, "id", "run-123", got.ID)
	assertEqual(t, "status", models.TaskStatusRunning, got.Status)
}

func TestGetDeploymentByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/deployments/nonexistent")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	assertEqual(t, "status", http.StatusNotFound, resp.StatusCode)
}

// --- GET /api/v1/deployments/{id}/progress ---

func TestGetDeploymentProgress_ReturnsProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)
	run := sampleRun()
	m.EXPECT().Start(gomock.Any()).Return(run, nil)

	events := []json.RawMessage{
		json.RawMessage(`{"event":"runner_on_ok","event_data":{"task":"Install packages"}}`),
		json.RawMessage(`{"event":"runner_on_ok","event_data":{"task":"Configure network"}}`),
	}
	m.EXPECT().Events("run-123").Return(events, nil)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	// Create deployment
	resp, err := http.Post(srv.URL+"/api/v1/deployments", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	// Get progress
	resp, err = http.Get(srv.URL + "/api/v1/deployments/run-123/progress")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	var got models.DeploymentProgress
	json.NewDecoder(resp.Body).Decode(&got)
	assertEqual(t, "completed", 2, got.Completed)
	assertEqual(t, "total", 350, got.Total)
	assertEqual(t, "currentTask", "Configure network", got.CurrentTask)
}

func TestGetDeploymentProgress_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/deployments/nonexistent/progress")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	assertEqual(t, "status", http.StatusNotFound, resp.StatusCode)
}

// --- DELETE /api/v1/deployments/{id} ---

func TestCancelDeployment_CancelsRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)
	run := sampleRun()
	m.EXPECT().Start(gomock.Any()).Return(run, nil)
	m.EXPECT().Cancel("run-123").Return(nil)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	// Create deployment
	resp, err := http.Post(srv.URL+"/api/v1/deployments", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	// Cancel
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/deployments/run-123", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	resp.Body.Close()
	assertEqual(t, "status", http.StatusNoContent, resp.StatusCode)
}

func TestCancelDeployment_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/deployments/nonexistent", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	resp.Body.Close()
	assertEqual(t, "status", http.StatusNotFound, resp.StatusCode)
}

func TestCancelDeployment_UpdatesStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := runner.NewMockRunner(ctrl)
	run := sampleRun()
	m.EXPECT().Start(gomock.Any()).Return(run, nil)
	m.EXPECT().Cancel("run-123").Return(nil)

	srv, _ := setupDeployAPI(m)
	defer srv.Close()

	// Create deployment
	resp, err := http.Post(srv.URL+"/api/v1/deployments", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	// Cancel
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/deployments/run-123", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	resp.Body.Close()

	// Verify status is canceled
	resp, err = http.Get(srv.URL + "/api/v1/deployments/run-123")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	var got models.Deployment
	json.NewDecoder(resp.Body).Decode(&got)
	assertEqual(t, "status_after_cancel", models.TaskStatusCanceled, got.Status)
}
