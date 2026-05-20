package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func skipIfNoAnsibleRunner(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ansible-runner"); err != nil {
		t.Skip("ansible-runner not in PATH")
	}
}

// newTestProject creates a minimal ansible-runner private data directory in a
// temp dir. The project/ subdirectory contains several test playbooks.
func newTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	must(os.MkdirAll(filepath.Join(dir, "project"), 0755))
	must(os.MkdirAll(filepath.Join(dir, "inventory"), 0755))

	must(os.WriteFile(filepath.Join(dir, "inventory", "hosts"),
		[]byte("[all]\nlocalhost ansible_connection=local\n"), 0644))

	must(os.WriteFile(filepath.Join(dir, "ansible.cfg"),
		[]byte("[defaults]\nhost_key_checking = False\n"), 0644))

	must(os.WriteFile(filepath.Join(dir, "project", "success.yaml"), []byte(`---
- name: Success
  hosts: localhost
  gather_facts: false
  tasks:
    - name: Print message
      ansible.builtin.debug:
        msg: "hello from test"
`), 0644))

	must(os.WriteFile(filepath.Join(dir, "project", "fail.yaml"), []byte(`---
- name: Failure
  hosts: localhost
  gather_facts: false
  tasks:
    - name: Fail task
      ansible.builtin.fail:
        msg: "intentional failure"
`), 0644))

	must(os.WriteFile(filepath.Join(dir, "project", "echo_var.yaml"), []byte(`---
- name: Echo variable
  hosts: localhost
  gather_facts: false
  tasks:
    - name: Fail if var is unset
      ansible.builtin.fail:
        msg: "my_var was not passed"
      when: my_var is not defined
    - name: Print var
      ansible.builtin.debug:
        msg: "value={{ my_var }}"
`), 0644))

	// slow.yaml sleeps long enough for busy/shutdown tests; Shutdown will
	// SIGTERM the process group so the test does not actually wait 30 s.
	must(os.WriteFile(filepath.Join(dir, "project", "slow.yaml"), []byte(`---
- name: Slow
  hosts: localhost
  gather_facts: false
  tasks:
    - name: Wait
      ansible.builtin.command: sleep 30
`), 0644))

	return dir
}

// newRunner creates an AnsibleRunner and registers Shutdown as test cleanup so
// long-running processes are terminated even if the test fails.
func newRunner(t *testing.T, dir string) *AnsibleRunner {
	t.Helper()
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(dir)
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		runner.Shutdown(ctx) //nolint:errcheck
	})
	return runner
}

// pollRun polls Get until the run leaves "running" status.
func pollRun(t *testing.T, runner *AnsibleRunner, id string, timeout time.Duration) *models.TaskRun {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		run, err := runner.Get(id)
		if err != nil {
			t.Fatalf("Get(%q): %v", id, err)
		}
		if run.Status != models.TaskStatusRunning {
			return run
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("run %q did not finish within %v", id, timeout)
	return nil
}

// seedRun writes a fake run.json into the runner's artifacts directory so
// tests can exercise Get/List/Logs/Events/Recover without actually executing
// ansible-runner.
func seedRun(t *testing.T, artifactsDir string, run *models.TaskRun) string {
	t.Helper()
	runDir := filepath.Join(artifactsDir, run.ID)
	if err := os.MkdirAll(runDir, 0750); err != nil {
		t.Fatalf("mkdir %s: %v", runDir, err)
	}
	if err := writeRunJSON(runDir, run); err != nil {
		t.Fatalf("writeRunJSON: %v", err)
	}
	return runDir
}

// --- NewAnsibleRunner ---

func TestNewAnsibleRunner_DirNotFound(t *testing.T) {
	_, err := NewAnsibleRunner("/nonexistent/does-not-exist-xyz")
	if err == nil {
		t.Fatal("expected error for missing directory, got nil")
	}
}

func TestNewAnsibleRunner_CreatesArtifactsDir(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	dir := t.TempDir()
	if _, err := NewAnsibleRunner(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "artifacts")); err != nil {
		t.Fatalf("artifacts dir not created: %v", err)
	}
}

// --- Start ---

func TestAnsibleRunner_Start_Success(t *testing.T) {
	runner := newRunner(t, newTestProject(t))

	run, err := runner.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if run.Status != models.TaskStatusRunning {
		t.Errorf("initial status: want running, got %s", run.Status)
	}
	if run.PID == 0 {
		t.Error("PID not set")
	}
	if run.StartedAt == nil {
		t.Error("StartedAt not set")
	}

	completed := pollRun(t, runner, run.ID, 60*time.Second)
	if completed.Status != models.TaskStatusSuccessful {
		t.Errorf("final status: want successful, got %s", completed.Status)
	}
	if completed.ExitCode == nil || *completed.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %v", completed.ExitCode)
	}
	if completed.EndedAt == nil {
		t.Error("EndedAt not set after completion")
	}
}

func TestAnsibleRunner_Start_Failure(t *testing.T) {
	runner := newRunner(t, newTestProject(t))

	run, err := runner.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "fail.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	completed := pollRun(t, runner, run.ID, 60*time.Second)
	if completed.Status != models.TaskStatusFailed {
		t.Errorf("final status: want failed, got %s", completed.Status)
	}
	if completed.ExitCode == nil {
		t.Error("ExitCode not set after failure")
	} else if *completed.ExitCode == 0 {
		t.Error("expected non-zero exit code, got 0")
	}
	if completed.EndedAt == nil {
		t.Error("EndedAt not set after failure")
	}
}

func TestAnsibleRunner_Start_Busy(t *testing.T) {
	runner := newRunner(t, newTestProject(t))

	// Start a slow playbook; the runner holds its lock while the process runs.
	if _, err := runner.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "slow.yaml",
	}); err != nil {
		t.Fatalf("first Start: %v", err)
	}

	_, err := runner.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != ErrBusy {
		t.Errorf("expected ErrBusy, got %v", err)
	}
}

func TestAnsibleRunner_Start_ExtraVars(t *testing.T) {
	runner := newRunner(t, newTestProject(t))

	run, err := runner.Start(StartRequest{
		Type:      models.TaskTypeDeploy,
		Playbook:  "echo_var.yaml",
		ExtraVars: map[string]string{"my_var": "test_value"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	completed := pollRun(t, runner, run.ID, 60*time.Second)
	if completed.Status != models.TaskStatusSuccessful {
		t.Errorf("final status: want successful, got %s", completed.Status)
	}
	if completed.ExtraVars["my_var"] != "test_value" {
		t.Errorf("ExtraVars not preserved in stored run: %v", completed.ExtraVars)
	}
}

// --- Get ---

func TestAnsibleRunner_Get_NotFound(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	if _, err := runner.Get("does-not-exist"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAnsibleRunner_Get_ReturnsStoredFields(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	want := &models.TaskRun{
		ID:        "stored-run",
		Type:      models.TaskTypeDeploy,
		Status:    models.TaskStatusSuccessful,
		Playbook:  "success.yaml",
		ExtraVars: map[string]string{"k": "v"},
		CreatedAt: now,
	}
	seedRun(t, runner.artifactsDir, want)

	got, err := runner.Get("stored-run")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID: want %s, got %s", want.ID, got.ID)
	}
	if got.Status != want.Status {
		t.Errorf("Status: want %s, got %s", want.Status, got.Status)
	}
	if got.ExtraVars["k"] != "v" {
		t.Errorf("ExtraVars not round-tripped: %v", got.ExtraVars)
	}
}

// --- List ---

func TestAnsibleRunner_List_Empty(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	runs, err := runner.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestAnsibleRunner_List_SortedNewestFirst(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	now := time.Now()
	older := &models.TaskRun{ID: "older", Status: models.TaskStatusSuccessful, CreatedAt: now.Add(-1 * time.Hour)}
	middle := &models.TaskRun{ID: "middle", Status: models.TaskStatusSuccessful, CreatedAt: now.Add(-30 * time.Minute)}
	newer := &models.TaskRun{ID: "newer", Status: models.TaskStatusSuccessful, CreatedAt: now}

	for _, r := range []*models.TaskRun{older, newer, middle} {
		seedRun(t, runner.artifactsDir, r)
	}

	runs, err := runner.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
	if runs[0].ID != "newer" {
		t.Errorf("[0] want newer, got %s", runs[0].ID)
	}
	if runs[1].ID != "middle" {
		t.Errorf("[1] want middle, got %s", runs[1].ID)
	}
	if runs[2].ID != "older" {
		t.Errorf("[2] want older, got %s", runs[2].ID)
	}
}

func TestAnsibleRunner_List_IgnoresDirsWithoutRunJSON(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID: "valid-run", Status: models.TaskStatusSuccessful, CreatedAt: time.Now(),
	})
	// A directory with no run.json should be silently skipped.
	if err := os.MkdirAll(filepath.Join(runner.artifactsDir, "orphan-dir"), 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	runs, err := runner.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run (orphan-dir ignored), got %d", len(runs))
	}
}

// --- Logs ---

func TestAnsibleRunner_Logs_NotFound(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	if _, err := runner.Logs("does-not-exist"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAnsibleRunner_Logs_EmptyWhenStdoutMissing(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID: "no-stdout", Status: models.TaskStatusRunning, CreatedAt: time.Now(),
	})

	logs, err := runner.Logs("no-stdout")
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected empty logs, got %d bytes", len(logs))
	}
}

func TestAnsibleRunner_Logs_Integration(t *testing.T) {
	runner := newRunner(t, newTestProject(t))

	run, err := runner.Start(StartRequest{Type: models.TaskTypeDeploy, Playbook: "success.yaml"})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	pollRun(t, runner, run.ID, 60*time.Second)

	logs, err := runner.Logs(run.ID)
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if len(logs) == 0 {
		t.Error("expected non-empty logs after successful run")
	}
}

// --- Events ---

func TestAnsibleRunner_Events_NotFound(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	if _, err := runner.Events("does-not-exist"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAnsibleRunner_Events_EmptyWhenEventsDirMissing(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID: "no-events", Status: models.TaskStatusSuccessful, CreatedAt: time.Now(),
	})

	events, err := runner.Events("no-events")
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestAnsibleRunner_Events_OrderedByNumericPrefix(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	runDir := seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID: "ordered-run", Status: models.TaskStatusSuccessful, CreatedAt: time.Now(),
	})
	eventsDir := filepath.Join(runDir, "job_events")
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		t.Fatalf("mkdir events: %v", err)
	}

	// Write three event files with numeric prefixes out of order.
	eventFiles := []struct {
		name    string
		counter int
	}{
		{"3-abc-runner_on_ok.json", 3},
		{"1-def-playbook_on_start.json", 1},
		{"2-ghi-runner_on_task.json", 2},
	}
	for _, f := range eventFiles {
		data := fmt.Appendf(nil, `{"counter":%d}`, f.counter)
		if err := os.WriteFile(filepath.Join(eventsDir, f.name), data, 0644); err != nil {
			t.Fatalf("write event file: %v", err)
		}
	}

	got, err := runner.Events("ordered-run")
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}
	for i, ev := range got {
		var m map[string]any
		if err := json.Unmarshal(ev, &m); err != nil {
			t.Fatalf("unmarshal event[%d]: %v", i, err)
		}
		if m["counter"] != float64(i+1) {
			t.Errorf("events[%d]: want counter %d, got %v", i, i+1, m["counter"])
		}
	}
}

func TestAnsibleRunner_Events_SkipsNonJSONFiles(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	runDir := seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID: "skip-non-json", Status: models.TaskStatusSuccessful, CreatedAt: time.Now(),
	})
	eventsDir := filepath.Join(runDir, "job_events")
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		t.Fatalf("mkdir events: %v", err)
	}

	os.WriteFile(filepath.Join(eventsDir, "1-abc-event.json"), []byte(`{"counter":1}`), 0644)
	os.WriteFile(filepath.Join(eventsDir, "2-abc-event.txt"), []byte("not json"), 0644)
	os.WriteFile(filepath.Join(eventsDir, "metadata"), []byte("{}"), 0644)

	got, err := runner.Events("skip-non-json")
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 event (only .json files), got %d", len(got))
	}
}

func TestAnsibleRunner_Events_Integration(t *testing.T) {
	runner := newRunner(t, newTestProject(t))

	run, err := runner.Start(StartRequest{Type: models.TaskTypeDeploy, Playbook: "success.yaml"})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	pollRun(t, runner, run.ID, 60*time.Second)

	events, err := runner.Events(run.ID)
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(events) == 0 {
		t.Error("expected at least one event after a completed run")
	}
}

// --- Delete ---

func TestAnsibleRunner_Delete_NotFound(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	if err := runner.Delete("does-not-exist"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAnsibleRunner_Delete_RemovesDirectory(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	runDir := seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID:     "to-delete",
		Status: models.TaskStatusSuccessful,
		CreatedAt: time.Now(),
	})

	if err := runner.Delete("to-delete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Error("expected run directory to be removed")
	}
}

func TestAnsibleRunner_Delete_ActiveRunReturnsErrRunning(t *testing.T) {
	runner := newRunner(t, newTestProject(t))

	run, err := runner.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "slow.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := runner.Delete(run.ID); err != ErrRunning {
		t.Errorf("expected ErrRunning, got %v", err)
	}
}

// --- Recover ---

func TestAnsibleRunner_Recover_DeadProcessNoStatus(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	now := time.Now()
	// Use a PID that is astronomically unlikely to be alive.
	seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID:        "dead-no-status",
		Status:    models.TaskStatusRunning,
		Playbook:  "success.yaml",
		PID:       999999999,
		CreatedAt: now,
		StartedAt: &now,
	})

	if err := runner.Recover(); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	recovered, err := runner.Get("dead-no-status")
	if err != nil {
		t.Fatalf("Get after Recover: %v", err)
	}
	if recovered.Status != models.TaskStatusFailed {
		t.Errorf("expected failed, got %s", recovered.Status)
	}
	if recovered.Error == "" {
		t.Error("expected error message to be set")
	}
	if recovered.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

func TestAnsibleRunner_Recover_DeadProcessWithSuccessStatus(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	now := time.Now()
	runDir := seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID:        "dead-successful",
		Status:    models.TaskStatusRunning,
		Playbook:  "success.yaml",
		PID:       999999999,
		CreatedAt: now,
		StartedAt: &now,
	})
	// Pre-populate ansible-runner status files to simulate a completed run.
	os.WriteFile(filepath.Join(runDir, "status"), []byte("successful"), 0640)
	os.WriteFile(filepath.Join(runDir, "rc"), []byte("0"), 0640)

	if err := runner.Recover(); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	recovered, err := runner.Get("dead-successful")
	if err != nil {
		t.Fatalf("Get after Recover: %v", err)
	}
	if recovered.Status != models.TaskStatusSuccessful {
		t.Errorf("expected successful, got %s", recovered.Status)
	}
	if recovered.ExitCode == nil || *recovered.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %v", recovered.ExitCode)
	}
	if recovered.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

func TestAnsibleRunner_Recover_SkipsAlreadyCompletedRuns(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	now := time.Now()
	seedRun(t, runner.artifactsDir, &models.TaskRun{
		ID:        "already-done",
		Status:    models.TaskStatusSuccessful,
		Playbook:  "success.yaml",
		CreatedAt: now,
		StartedAt: &now,
	})

	if err := runner.Recover(); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	// Status must remain unchanged.
	recovered, err := runner.Get("already-done")
	if err != nil {
		t.Fatalf("Get after Recover: %v", err)
	}
	if recovered.Status != models.TaskStatusSuccessful {
		t.Errorf("expected successful (unchanged), got %s", recovered.Status)
	}
}

// --- Shutdown ---

func TestAnsibleRunner_Shutdown_NoActiveRun(t *testing.T) {
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(t.TempDir())
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}
	if err := runner.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown with no active run: %v", err)
	}
}

func TestAnsibleRunner_Shutdown_CancelsActiveRun(t *testing.T) {
	// Don't use newRunner — we manage Shutdown ourselves in this test.
	skipIfNoAnsibleRunner(t)
	runner, err := NewAnsibleRunner(newTestProject(t))
	if err != nil {
		t.Fatalf("NewAnsibleRunner: %v", err)
	}

	run, err := runner.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "slow.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := runner.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	stored, err := runner.Get(run.ID)
	if err != nil {
		t.Fatalf("Get after Shutdown: %v", err)
	}
	if stored.Status != models.TaskStatusCanceled {
		t.Errorf("expected canceled, got %s", stored.Status)
	}
	if stored.Error != "server shutdown" {
		t.Errorf("expected error %q, got %q", "server shutdown", stored.Error)
	}
}
