package deploy

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"
)

// mockRunner is a minimal Runner for deployer tests.
type mockRunner struct {
	mu   sync.Mutex
	runs map[string]*models.TaskRun
}

func newMockRunner() *mockRunner {
	return &mockRunner{runs: make(map[string]*models.TaskRun)}
}

func (m *mockRunner) Start(req runner.StartRequest) (*models.TaskRun, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := "mock-run-001"
	now := time.Now()
	run := &models.TaskRun{
		ID:        id,
		Type:      req.Type,
		Status:    models.TaskStatusRunning,
		Playbook:  req.Playbook,
		ExtraVars: req.ExtraVars,
		StartedAt: now,
	}
	m.runs[id] = run

	// Complete after a short delay.
	go func() {
		time.Sleep(200 * time.Millisecond)
		m.mu.Lock()
		defer m.mu.Unlock()
		ended := time.Now()
		run.Status = models.TaskStatusSuccessful
		run.EndedAt = &ended
	}()

	return run, nil
}

func (m *mockRunner) RunSync(_ context.Context, req runner.StartRequest) (*models.TaskRun, []byte, error) {
	run, err := m.Start(req)
	if err != nil {
		return nil, nil, err
	}
	time.Sleep(300 * time.Millisecond)
	return run, nil, nil
}

func (m *mockRunner) Cancel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	run, ok := m.runs[id]
	if !ok {
		return runner.ErrNotFound
	}
	run.Status = models.TaskStatusCanceled
	now := time.Now()
	run.EndedAt = &now
	return nil
}

func (m *mockRunner) Get(id string) (*models.TaskRun, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	run, ok := m.runs[id]
	if !ok {
		return nil, runner.ErrNotFound
	}
	return run, nil
}

func (m *mockRunner) List() ([]models.TaskRun, error)                 { return nil, nil }
func (m *mockRunner) Logs(_ string) ([]byte, error)                   { return nil, nil }
func (m *mockRunner) Events(_ string) ([]json.RawMessage, error)      { return nil, nil }
func (m *mockRunner) Stream(_ string) (<-chan runner.Event, error)     { return nil, nil }
func (m *mockRunner) Delete(_ string) error                           { return nil }
func (m *mockRunner) Shutdown(_ context.Context) error                { return nil }
func (m *mockRunner) Recover() error                                  { return nil }

func newTestDeployer(t *testing.T) (*Deployer, *mockRunner) {
	t.Helper()

	enclaveDir := t.TempDir()
	stateDir := t.TempDir()

	// Create a minimal config file so ReadAll works.
	cfgDir := enclaveDir
	writeTestConfig(t, cfgDir)

	r := newMockRunner()
	state, err := NewState(stateDir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	reader := config.NewReader(enclaveDir)
	writer := config.NewWriter(enclaveDir)
	registry := plugins.NewRegistry(nil)

	d := NewDeployer(r, registry, reader, writer, enclaveDir, state)
	t.Cleanup(func() { d.Stop() })
	return d, r
}

func writeTestConfig(t *testing.T, dir string) {
	t.Helper()

	// Create minimal config files that the reader expects.
	globalYAML := `---
baseDomain: test.example.com
clusterName: test-cluster
`
	certsYAML := `---
caCert: ""
`
	infraYAML := `---
cloudType: baremetal
`
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(writeFileHelper(dir+"/global.yaml", []byte(globalYAML)))
	must(writeFileHelper(dir+"/certificates.yaml", []byte(certsYAML)))
	must(writeFileHelper(dir+"/cloud_infra.yaml", []byte(infraYAML)))
}

func writeFileHelper(path string, data []byte) error {
	return writeFile(path, data)
}

func TestDeployer_Start(t *testing.T) {
	d, _ := newTestDeployer(t)

	dep, run, err := d.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if dep == nil {
		t.Fatal("Start returned nil deployment")
	}
	if run == nil {
		t.Fatal("Start returned nil run")
	}
	if dep.Status != models.TaskStatusRunning {
		t.Errorf("deployment status: want running, got %s", dep.Status)
	}
	if len(dep.Phases) == 0 {
		t.Fatal("deployment has no phases")
	}
	if dep.Phases[0].Name != "main" {
		t.Errorf("first phase: want main, got %s", dep.Phases[0].Name)
	}
	if dep.Phases[0].Status != models.TaskStatusRunning {
		t.Errorf("first phase status: want running, got %s", dep.Phases[0].Status)
	}
	if dep.TotalTasks == 0 {
		t.Error("TotalTasks should be > 0")
	}
	if dep.StartedAt.IsZero() {
		t.Error("StartedAt not set")
	}
}

func TestDeployer_Start_PersistsState(t *testing.T) {
	d, _ := newTestDeployer(t)

	dep, _, err := d.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	loaded, err := d.Get(dep.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if loaded.ID != dep.ID {
		t.Errorf("ID: want %s, got %s", dep.ID, loaded.ID)
	}
}

func TestDeployer_Get_NotFound(t *testing.T) {
	d, _ := newTestDeployer(t)

	_, err := d.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing deployment")
	}
}

func TestDeployer_GetLatest(t *testing.T) {
	d, _ := newTestDeployer(t)

	dep, _, err := d.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	latest, err := d.GetLatest()
	if err != nil {
		t.Fatalf("GetLatest: %v", err)
	}
	if latest == nil {
		t.Fatal("GetLatest returned nil")
	}
	if latest.ID != dep.ID {
		t.Errorf("GetLatest.ID: want %s, got %s", dep.ID, latest.ID)
	}
}

func TestDeployer_Cancel(t *testing.T) {
	d, _ := newTestDeployer(t)

	dep, _, err := d.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := d.Cancel(dep.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	loaded, err := d.Get(dep.ID)
	if err != nil {
		t.Fatalf("Get after Cancel: %v", err)
	}
	if loaded.Status != models.TaskStatusCanceled {
		t.Errorf("status after cancel: want canceled, got %s", loaded.Status)
	}
}

func TestDeployer_WatchUpdatesState(t *testing.T) {
	d, _ := newTestDeployer(t)

	dep, _, err := d.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// watchDeploy polls every 2 seconds; mock completes in 200ms.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		loaded, err := d.Get(dep.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if loaded.Status != models.TaskStatusRunning {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Error("deployment status should have been updated from running within 10 seconds")
}
