package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

// fakeRunner is a minimal Runner for testing RecordingRunner without needing
// ansible-runner.
type fakeRunner struct {
	mu        sync.Mutex
	runs      map[string]*models.TaskRun
	artifacts string
	started   chan string
}

func newFakeRunner(t *testing.T) *fakeRunner {
	t.Helper()
	dir := t.TempDir()
	return &fakeRunner{
		runs:      make(map[string]*models.TaskRun),
		artifacts: dir,
		started:   make(chan string, 10),
	}
}

func (f *fakeRunner) Start(req StartRequest) (*models.TaskRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	id := generateRunID()
	now := time.Now()
	run := &models.TaskRun{
		ID:        id,
		Type:      req.Type,
		Status:    models.TaskStatusRunning,
		Playbook:  req.Playbook,
		ExtraVars: req.ExtraVars,
		StartedAt: now,
	}
	f.runs[id] = run

	runDir := filepath.Join(f.artifacts, id)
	os.MkdirAll(runDir, 0750)
	writeRunJSON(runDir, run)

	// Write some events for the recording to capture.
	eventsDir := filepath.Join(runDir, "job_events")
	os.MkdirAll(eventsDir, 0750)
	eventData, _ := json.Marshal(map[string]any{"uuid": "test-uuid", "counter": 1, "event": "runner_on_ok"})
	os.WriteFile(filepath.Join(eventsDir, "001-test-uuid.json"), eventData, 0640)
	os.WriteFile(filepath.Join(runDir, "stdout"), []byte("test output"), 0640)
	os.WriteFile(filepath.Join(runDir, "status"), []byte("successful"), 0640)
	os.WriteFile(filepath.Join(runDir, "rc"), []byte("0"), 0640)

	runCopy := *run

	// Complete immediately.
	go func() {
		time.Sleep(100 * time.Millisecond)
		f.mu.Lock()
		ended := time.Now()
		run.Status = models.TaskStatusSuccessful
		run.EndedAt = &ended
		rc := 0
		run.ExitCode = &rc
		writeRunJSON(runDir, run)
		f.mu.Unlock()
		f.started <- id
	}()

	return &runCopy, nil
}

func (f *fakeRunner) RunSync(_ context.Context, req StartRequest) (*models.TaskRun, []byte, error) {
	run, err := f.Start(req)
	if err != nil {
		return nil, nil, err
	}
	time.Sleep(200 * time.Millisecond)
	f.mu.Lock()
	r := f.runs[run.ID]
	f.mu.Unlock()
	return r, []byte("test output"), nil
}

func (f *fakeRunner) Cancel(_ string) error   { return nil }
func (f *fakeRunner) Get(id string) (*models.TaskRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.runs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}
func (f *fakeRunner) List() ([]models.TaskRun, error)                 { return nil, nil }
func (f *fakeRunner) Logs(_ string) ([]byte, error)                   { return nil, nil }
func (f *fakeRunner) Events(_ string) ([]json.RawMessage, error)      { return nil, nil }
func (f *fakeRunner) Stream(_ string) (<-chan Event, error)            { return nil, nil }
func (f *fakeRunner) Delete(_ string) error                           { return nil }
func (f *fakeRunner) Shutdown(_ context.Context) error                { return nil }
func (f *fakeRunner) Recover() error                                  { return nil }

func TestRecordingRunner_DelegatesToInner(t *testing.T) {
	inner := newFakeRunner(t)
	recDir := t.TempDir()

	rec, err := NewRecordingRunner(inner, inner.artifacts, recDir)
	if err != nil {
		t.Fatalf("NewRecordingRunner: %v", err)
	}

	run, err := rec.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if run.Status != models.TaskStatusRunning {
		t.Errorf("initial status: want running, got %s", run.Status)
	}

	// Wait for the fake runner to complete.
	<-inner.started
}

func TestRecordingRunner_SavesRecording(t *testing.T) {
	inner := newFakeRunner(t)
	recDir := t.TempDir()

	rec, err := NewRecordingRunner(inner, inner.artifacts, recDir)
	if err != nil {
		t.Fatalf("NewRecordingRunner: %v", err)
	}

	_, err = rec.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for the fake to complete and recording to be saved.
	<-inner.started
	time.Sleep(2 * time.Second)

	key := ScenarioKey("success.yaml", nil)
	path := filepath.Join(recDir, key+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("recording file not created at %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading recording: %v", err)
	}
	var recording Recording
	if err := json.Unmarshal(data, &recording); err != nil {
		t.Fatalf("parsing recording: %v", err)
	}
	if recording.Status != "successful" {
		t.Errorf("recording status: want successful, got %s", recording.Status)
	}
}

func TestRecordingRunner_RunSync_SavesRecording(t *testing.T) {
	inner := newFakeRunner(t)
	recDir := t.TempDir()

	rec, err := NewRecordingRunner(inner, inner.artifacts, recDir)
	if err != nil {
		t.Fatalf("NewRecordingRunner: %v", err)
	}

	ctx := context.Background()
	run, logs, err := rec.RunSync(ctx, StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "sync.yaml",
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	if run == nil {
		t.Fatal("RunSync returned nil run")
	}
	if len(logs) == 0 {
		t.Error("RunSync returned empty logs")
	}

	key := ScenarioKey("sync.yaml", nil)
	path := filepath.Join(recDir, key+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("recording file not created at %s", path)
	}
}

func TestRecordingRunner_ScenarioKeyWithTags(t *testing.T) {
	inner := newFakeRunner(t)
	recDir := t.TempDir()

	rec, err := NewRecordingRunner(inner, inner.artifacts, recDir)
	if err != nil {
		t.Fatalf("NewRecordingRunner: %v", err)
	}

	_, err = rec.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "validate.yaml",
		Tags:     []string{"z-tag", "a-tag"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	<-inner.started
	time.Sleep(2 * time.Second)

	// Tags should be sorted in the key.
	key := ScenarioKey("validate.yaml", []string{"z-tag", "a-tag"})
	path := filepath.Join(recDir, key+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("recording not saved with correct scenario key at %s", path)
	}
}
