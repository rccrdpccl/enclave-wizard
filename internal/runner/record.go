package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

// RecordingRunner wraps any Runner and saves completed runs as recording
// files for later replay by ReplayRunner.
type RecordingRunner struct {
	inner         Runner
	artifactsDir  string
	recordingsDir string
}

// NewRecordingRunner creates a decorator that saves recordings after each run.
func NewRecordingRunner(inner Runner, artifactsDir, recordingsDir string) (*RecordingRunner, error) {
	if err := os.MkdirAll(recordingsDir, 0750); err != nil {
		return nil, fmt.Errorf("creating recordings directory: %w", err)
	}
	return &RecordingRunner{
		inner:         inner,
		artifactsDir:  artifactsDir,
		recordingsDir: recordingsDir,
	}, nil
}

func (r *RecordingRunner) Start(req StartRequest) (*models.TaskRun, error) {
	run, err := r.inner.Start(req)
	if err != nil {
		return nil, err
	}
	go r.waitAndSave(req, run.ID)
	return run, nil
}

func (r *RecordingRunner) RunSync(ctx context.Context, req StartRequest) (*models.TaskRun, []byte, error) {
	run, logs, err := r.inner.RunSync(ctx, req)
	if err != nil {
		return run, logs, err
	}
	r.saveRecording(req, run.ID)
	return run, logs, nil
}

func (r *RecordingRunner) waitAndSave(req StartRequest, runID string) {
	for {
		time.Sleep(500 * time.Millisecond)
		run, err := r.inner.Get(runID)
		if err != nil {
			return
		}
		if run.Status == models.TaskStatusRunning {
			continue
		}
		r.saveRecording(req, runID)
		return
	}
}

func (r *RecordingRunner) saveRecording(req StartRequest, runID string) {
	key := ScenarioKey(req.Playbook, req.Tags)
	src := filepath.Join(r.artifactsDir, runID)
	dst := filepath.Join(r.recordingsDir, key+".json")

	os.Remove(dst)
	if err := SaveRecording(src, dst); err != nil {
		slog.Error("failed to save recording", "key", key, "error", err)
		return
	}
	slog.Info("recording saved", "key", key, "source", runID)
}

func (r *RecordingRunner) Cancel(id string) error {
	return r.inner.Cancel(id)
}

func (r *RecordingRunner) Get(id string) (*models.TaskRun, error) {
	return r.inner.Get(id)
}

func (r *RecordingRunner) List() ([]models.TaskRun, error) {
	return r.inner.List()
}

func (r *RecordingRunner) Logs(id string) ([]byte, error) {
	return r.inner.Logs(id)
}

func (r *RecordingRunner) Events(id string) ([]json.RawMessage, error) {
	return r.inner.Events(id)
}

func (r *RecordingRunner) Stream(id string) (<-chan Event, error) {
	return r.inner.Stream(id)
}

func (r *RecordingRunner) Delete(id string) error {
	return r.inner.Delete(id)
}

func (r *RecordingRunner) Shutdown(ctx context.Context) error {
	return r.inner.Shutdown(ctx)
}

func (r *RecordingRunner) Recover() error {
	return r.inner.Recover()
}
