package runner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func TestReplayRunner_Start_Success(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	createTestRecording(t, recDir, "success.yaml", "successful", 3)

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	run, err := r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if run.Status != models.TaskStatusRunning {
		t.Errorf("initial status: want running, got %s", run.Status)
	}

	completed := pollRun(t, r, run.ID, 30*time.Second)
	if completed.Status != models.TaskStatusSuccessful {
		t.Errorf("final status: want successful, got %s", completed.Status)
	}
	if completed.ExitCode == nil || *completed.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %v", completed.ExitCode)
	}
}

func TestReplayRunner_Start_Failure(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	createTestRecording(t, recDir, "fail.yaml", "failed", 2)

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	run, err := r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "fail.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	completed := pollRun(t, r, run.ID, 30*time.Second)
	if completed.Status != models.TaskStatusFailed {
		t.Errorf("final status: want failed, got %s", completed.Status)
	}
}

func TestReplayRunner_Start_NoRecording(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}

	_, err = r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "nonexistent.yaml",
	})
	if err == nil {
		t.Fatal("expected error for missing recording")
	}
}

func TestReplayRunner_Start_Busy(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	createTestRecording(t, recDir, "success.yaml", "successful", 3)

	// Use slow speed to keep the replay running.
	r, err := NewReplayRunner(enclaveDir, recDir, 0.01)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	_, err = r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("first Start: %v", err)
	}

	// Give the goroutine time to start.
	time.Sleep(50 * time.Millisecond)

	_, err = r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != ErrBusy {
		t.Errorf("expected ErrBusy, got %v", err)
	}
}

func TestReplayRunner_Cancel(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	createTestRecording(t, recDir, "success.yaml", "successful", 100)

	// Use slow speed so we can cancel during replay.
	r, err := NewReplayRunner(enclaveDir, recDir, 0.1)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	run, err := r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Give it time to begin replaying.
	time.Sleep(100 * time.Millisecond)

	if err := r.Cancel(run.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	stored, err := r.Get(run.ID)
	if err != nil {
		t.Fatalf("Get after Cancel: %v", err)
	}
	if stored.Status != models.TaskStatusCanceled {
		t.Errorf("expected canceled, got %s", stored.Status)
	}
}

func TestReplayRunner_WritesEvents(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	createTestRecording(t, recDir, "success.yaml", "successful", 3)

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	run, err := r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	pollRun(t, r, run.ID, 30*time.Second)

	events, err := r.Events(run.ID)
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestReplayRunner_WritesStdout(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	createTestRecording(t, recDir, "success.yaml", "successful", 2)

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	run, err := r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	pollRun(t, r, run.ID, 30*time.Second)

	logs, err := r.Logs(run.ID)
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if len(logs) == 0 {
		t.Error("expected non-empty stdout after replay")
	}
}

func TestReplayRunner_Stream(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	createTestRecording(t, recDir, "success.yaml", "successful", 3)

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	run, err := r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	pollRun(t, r, run.ID, 30*time.Second)

	// Stream a completed run.
	ch, err := r.Stream(run.ID)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	var events []Event
	for ev := range ch {
		events = append(events, ev)
	}
	if len(events) < 3 {
		t.Errorf("expected at least 3 events from stream, got %d", len(events))
	}
}

func TestReplayRunner_Recover(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}

	// Seed a run that looks like it was interrupted.
	artifactsDir := filepath.Join(enclaveDir, "artifacts")
	seedRun(t, artifactsDir, &models.TaskRun{
		ID:        "interrupted",
		Status:    models.TaskStatusRunning,
		StartedAt: time.Now(),
	})

	if err := r.Recover(); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	recovered, err := r.Get("interrupted")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if recovered.Status != models.TaskStatusFailed {
		t.Errorf("expected failed, got %s", recovered.Status)
	}
}

func TestReplayRunner_SpeedMultiplier(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	// Create a recording with events spanning 10 seconds.
	createTestRecording(t, recDir, "success.yaml", "successful", 5)

	// Speed 0 = instant.
	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}
	t.Cleanup(func() { r.Shutdown(nil) })

	start := time.Now()
	run, err := r.Start(StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "success.yaml",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	pollRun(t, r, run.ID, 30*time.Second)
	duration := time.Since(start)

	// With speed=0, should complete nearly instantly.
	if duration > 5*time.Second {
		t.Errorf("speed=0 should be instant, took %v", duration)
	}
}

func TestReplayRunner_Delete(t *testing.T) {
	enclaveDir := t.TempDir()
	recDir := t.TempDir()

	r, err := NewReplayRunner(enclaveDir, recDir, 0)
	if err != nil {
		t.Fatalf("NewReplayRunner: %v", err)
	}

	artifactsDir := filepath.Join(enclaveDir, "artifacts")
	runDir := seedRun(t, artifactsDir, &models.TaskRun{
		ID:     "to-delete",
		Status: models.TaskStatusSuccessful,
	})

	if err := r.Delete("to-delete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Error("expected run directory to be removed")
	}
}
