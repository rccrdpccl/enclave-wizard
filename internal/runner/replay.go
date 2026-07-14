package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

// ReplayRunner replays recorded task runs from JSON recording files.
// It writes events to the artifacts directory with timing delays proportional
// to the configured speed multiplier.
type ReplayRunner struct {
	recordingsDir string
	artifactsDir  string
	speed         float64

	mu        sync.Mutex
	activeRun *models.TaskRun
	done      chan struct{}
	cancelCh  chan struct{} // closed to request cancellation
}

// NewReplayRunner creates a runner that replays recordings from recordingsDir.
// speed controls replay speed: 0 = instant, 1 = real-time, 10 = 10x faster.
func NewReplayRunner(enclaveDir, recordingsDir string, speed float64) (*ReplayRunner, error) {
	if _, err := os.Stat(enclaveDir); err != nil {
		return nil, fmt.Errorf("enclave directory: %w", err)
	}
	if _, err := os.Stat(recordingsDir); err != nil {
		return nil, fmt.Errorf("recordings directory: %w", err)
	}

	artifactsDir := filepath.Join(enclaveDir, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0750); err != nil {
		return nil, fmt.Errorf("creating artifacts directory: %w", err)
	}

	return &ReplayRunner{
		recordingsDir: recordingsDir,
		artifactsDir:  artifactsDir,
		speed:         speed,
	}, nil
}

func (r *ReplayRunner) Start(req StartRequest) (*models.TaskRun, error) {
	r.mu.Lock()
	if r.activeRun != nil {
		r.mu.Unlock()
		return nil, ErrBusy
	}

	key := ScenarioKey(req.Playbook, req.Tags)
	recordingFile := filepath.Join(r.recordingsDir, key+".json")
	if _, err := os.Stat(recordingFile); err != nil {
		r.mu.Unlock()
		return nil, fmt.Errorf("%w: %s", ErrNoRecording, key)
	}

	runID := generateRunID()
	runDir := filepath.Join(r.artifactsDir, runID)
	if err := os.MkdirAll(runDir, 0750); err != nil {
		r.mu.Unlock()
		return nil, fmt.Errorf("creating run directory: %w", err)
	}

	now := time.Now()
	run := &models.TaskRun{
		ID:        runID,
		Type:      req.Type,
		Status:    models.TaskStatusRunning,
		Playbook:  req.Playbook,
		ExtraVars: req.ExtraVars,
		StartedAt: now,
	}
	writeRunJSON(runDir, run)

	slog.Info("replay task started", "run_id", runID, "playbook", req.Playbook, "speed", r.speed)

	runCopy := *run

	done := make(chan struct{})
	cancelCh := make(chan struct{})
	r.activeRun = run
	r.done = done
	r.cancelCh = cancelCh
	r.mu.Unlock()

	go r.runReplay(recordingFile, run, runDir, done, cancelCh)
	return &runCopy, nil
}

func (r *ReplayRunner) RunSync(ctx context.Context, req StartRequest) (*models.TaskRun, []byte, error) {
	run, err := r.Start(req)
	if err != nil {
		return nil, nil, err
	}

	r.mu.Lock()
	done := r.done
	r.mu.Unlock()

	select {
	case <-ctx.Done():
		return run, nil, ctx.Err()
	case <-done:
		updated, _ := r.Get(run.ID)
		if updated != nil {
			run = updated
		}
		logs, _ := r.Logs(run.ID)
		return run, logs, nil
	}
}

func (r *ReplayRunner) runReplay(recordingFile string, run *models.TaskRun, runDir string, done, cancelCh chan struct{}) {
	defer func() {
		r.mu.Lock()
		r.activeRun = nil
		r.done = nil
		r.cancelCh = nil
		r.mu.Unlock()
		close(done)
	}()

	data, err := os.ReadFile(recordingFile)
	if err != nil {
		slog.Error("failed to read recording", "error", err)
		return
	}

	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		slog.Error("failed to parse recording", "error", err)
		return
	}

	eventsDir := filepath.Join(runDir, "job_events")
	os.MkdirAll(eventsDir, 0750)

	var totalDuration time.Duration
	if rec.Run.EndedAt != nil {
		totalDuration = rec.Run.EndedAt.Sub(rec.Run.StartedAt)
	}

	var stdoutBuilder strings.Builder
	canceled := false
	for _, event := range rec.Events {
		if r.speed > 0 && totalDuration > 0 && len(rec.Events) > 1 {
			delay := time.Duration(float64(totalDuration) / float64(len(rec.Events)) / r.speed)
			select {
			case <-cancelCh:
				canceled = true
			case <-time.After(delay):
			}
		}
		if canceled {
			break
		}

		select {
		case <-cancelCh:
			canceled = true
		default:
		}
		if canceled {
			break
		}

		var meta struct {
			UUID    string `json:"uuid"`
			Counter int    `json:"counter"`
			Stdout  string `json:"stdout"`
		}
		json.Unmarshal(event, &meta)

		filename := fmt.Sprintf("%03d-%s.json", meta.Counter, meta.UUID)
		os.WriteFile(filepath.Join(eventsDir, filename), event, 0640)

		if meta.Stdout != "" {
			stdoutBuilder.WriteString(meta.Stdout)
			stdoutBuilder.WriteString("\n")
			os.WriteFile(filepath.Join(runDir, "stdout"), []byte(stdoutBuilder.String()), 0640)
		}
	}

	if canceled {
		now := time.Now()
		run.EndedAt = &now
		run.Status = models.TaskStatusCanceled
		writeRunJSON(runDir, run)
		slog.Info("replay task canceled", "run_id", run.ID)
		return
	}

	os.WriteFile(filepath.Join(runDir, "stdout"), []byte(rec.Stdout), 0640)
	os.WriteFile(filepath.Join(runDir, "stderr"), []byte(rec.Stderr), 0640)
	os.WriteFile(filepath.Join(runDir, "status"), []byte(rec.Status), 0640)
	os.WriteFile(filepath.Join(runDir, "rc"), []byte(fmt.Sprintf("%d", rec.RC)), 0640)

	now := time.Now()
	run.EndedAt = &now
	switch rec.Status {
	case "successful":
		run.Status = models.TaskStatusSuccessful
	default:
		run.Status = models.TaskStatusFailed
	}
	run.ExitCode = &rec.RC
	writeRunJSON(runDir, run)

	slog.Info("replay task completed", "run_id", run.ID, "status", run.Status, "events", len(rec.Events))
}

// Cancel stops a running replay task.
func (r *ReplayRunner) Cancel(id string) error {
	r.mu.Lock()
	if r.activeRun == nil || r.activeRun.ID != id {
		r.mu.Unlock()
		return ErrNotFound
	}
	cancelCh := r.cancelCh
	done := r.done
	r.mu.Unlock()

	close(cancelCh)
	<-done
	return nil
}

func (r *ReplayRunner) Get(id string) (*models.TaskRun, error) {
	return artifactGet(r.artifactsDir, id)
}

func (r *ReplayRunner) List() ([]models.TaskRun, error) {
	return artifactList(r.artifactsDir)
}

func (r *ReplayRunner) Logs(id string) ([]byte, error) {
	return artifactLogs(r.artifactsDir, id)
}

func (r *ReplayRunner) Events(id string) ([]json.RawMessage, error) {
	return artifactEvents(r.artifactsDir, id)
}

// Stream returns a channel that emits events for a task run.
func (r *ReplayRunner) Stream(id string) (<-chan Event, error) {
	runDir := filepath.Join(r.artifactsDir, id)
	if _, err := readRunJSON(runDir); err != nil {
		return nil, ErrNotFound
	}

	ch := make(chan Event, 64)

	r.mu.Lock()
	isActive := r.activeRun != nil && r.activeRun.ID == id
	var done <-chan struct{}
	if isActive {
		done = r.done
	}
	r.mu.Unlock()

	go streamEventsFromDir(filepath.Join(runDir, "job_events"), ch, done)

	return ch, nil
}

func (r *ReplayRunner) Delete(id string) error {
	r.mu.Lock()
	active := r.activeRun
	r.mu.Unlock()

	if active != nil && active.ID == id {
		return ErrRunning
	}

	runDir := filepath.Join(r.artifactsDir, id)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return ErrNotFound
	}

	if err := os.RemoveAll(runDir); err != nil {
		return err
	}
	slog.Info("task deleted", "run_id", id)
	return nil
}

func (r *ReplayRunner) Shutdown(_ context.Context) error {
	r.mu.Lock()
	cancelCh := r.cancelCh
	done := r.done
	r.mu.Unlock()

	if cancelCh == nil || done == nil {
		return nil
	}

	select {
	case <-cancelCh:
		// Already canceled.
	default:
		close(cancelCh)
	}
	<-done
	return nil
}

func (r *ReplayRunner) Recover() error {
	// Replay runner doesn't spawn real processes, so recovery just marks
	// any interrupted runs as failed.
	entries, err := os.ReadDir(r.artifactsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading artifacts dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runDir := filepath.Join(r.artifactsDir, entry.Name())
		run, err := readRunJSON(runDir)
		if err != nil {
			continue
		}
		if run.Status != models.TaskStatusRunning {
			continue
		}
		now := time.Now()
		run.EndedAt = &now
		run.Status = models.TaskStatusFailed
		run.Error = "recovered after server restart"
		writeRunJSON(runDir, run)
	}
	return nil
}

// streamEventsFromDir is a shared helper used by both ReplayRunner and
// potentially other runners for directory-based event streaming.
func streamEventsFromDir(eventsDir string, ch chan<- Event, done <-chan struct{}) {
	defer close(ch)

	emitted := 0
	for {
		entries, _ := os.ReadDir(eventsDir)
		sortEventEntries(entries)

		var jsonEntries []os.DirEntry
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".json") {
				jsonEntries = append(jsonEntries, e)
			}
		}

		for i := emitted; i < len(jsonEntries); i++ {
			data, err := os.ReadFile(filepath.Join(eventsDir, jsonEntries[i].Name()))
			if err != nil {
				continue
			}
			ch <- Event{Type: "log", Data: json.RawMessage(data)}
			emitted++
		}

		if done == nil {
			return
		}

		select {
		case <-done:
			remaining, _ := os.ReadDir(eventsDir)
			sortEventEntries(remaining)
			var jsonRemaining []os.DirEntry
			for _, e := range remaining {
				if strings.HasSuffix(e.Name(), ".json") {
					jsonRemaining = append(jsonRemaining, e)
				}
			}
			for i := emitted; i < len(jsonRemaining); i++ {
				data, err := os.ReadFile(filepath.Join(eventsDir, jsonRemaining[i].Name()))
				if err != nil {
					continue
				}
				ch <- Event{Type: "log", Data: json.RawMessage(data)}
			}
			return
		case <-time.After(250 * time.Millisecond):
		}
	}
}
