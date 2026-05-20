package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

type AnsibleRunner struct {
	enclaveDir   string
	artifactsDir string

	mu       sync.Mutex
	lockFile *os.File

	activeRun *models.TaskRun
	activeCmd *exec.Cmd
	// Closed by waitForCompletion when the process exits.
	done chan struct{}
}

func NewAnsibleRunner(enclaveDir string) (*AnsibleRunner, error) {
	if _, err := os.Stat(enclaveDir); err != nil {
		return nil, fmt.Errorf("enclave directory: %w", err)
	}
	if _, err := exec.LookPath("ansible-runner"); err != nil {
		return nil, ErrRunnerBin
	}

	artifactsDir := filepath.Join(enclaveDir, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0750); err != nil {
		return nil, fmt.Errorf("creating artifacts directory: %w", err)
	}

	return &AnsibleRunner{
		enclaveDir:   enclaveDir,
		artifactsDir: artifactsDir,
	}, nil
}

func (r *AnsibleRunner) Start(req StartRequest) (*models.TaskRun, error) {
	if err := r.acquireLock(); err != nil {
		return nil, err
	}

	runID := generateRunID()
	runDir := filepath.Join(r.artifactsDir, runID)
	if err := os.MkdirAll(runDir, 0750); err != nil {
		r.releaseLock()
		return nil, fmt.Errorf("creating run directory: %w", err)
	}

	args := []string{"run", r.enclaveDir, "-p", req.Playbook, "--ident", runID}
	if len(req.ExtraVars) > 0 {
		// ansible-runner does not accept --extra-vars directly; extra vars must
		// be forwarded to ansible-playbook via --cmdline.
		keys := make([]string, 0, len(req.ExtraVars))
		for k := range req.ExtraVars {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var cmdParts []string
		for _, k := range keys {
			cmdParts = append(cmdParts, "--extra-vars", k+"="+req.ExtraVars[k])
		}
		args = append(args, "--cmdline", strings.Join(cmdParts, " "))
	}

	cmd := exec.Command("ansible-runner", args...)
	cmd.Dir = r.enclaveDir
	cmd.Env = append(os.Environ(),
		"ANSIBLE_CONFIG="+filepath.Join(r.enclaveDir, "ansible.cfg"),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	now := time.Now()
	run := &models.TaskRun{
		ID:        runID,
		Type:      req.Type,
		Status:    models.TaskStatusRunning,
		Playbook:  req.Playbook,
		ExtraVars: req.ExtraVars,
		CreatedAt: now,
		StartedAt: &now,
	}

	if err := cmd.Start(); err != nil {
		r.releaseLock()
		slog.Error("failed to start ansible-runner", "run_id", runID, "playbook", req.Playbook, "error", err)
		return nil, fmt.Errorf("starting ansible-runner: %w", err)
	}

	run.PID = cmd.Process.Pid
	if err := writeRunJSON(runDir, run); err != nil {
		cmd.Process.Kill()
		r.releaseLock()
		slog.Error("failed to write run metadata", "run_id", runID, "error", err)
		return nil, fmt.Errorf("writing run metadata: %w", err)
	}

	slog.Info("task started", "run_id", runID, "type", req.Type, "playbook", req.Playbook, "pid", run.PID)

	done := make(chan struct{})
	r.mu.Lock()
	r.activeRun = run
	r.activeCmd = cmd
	r.done = done

	r.mu.Unlock()

	go r.waitForCompletion(cmd, run, runDir, done)

	return run, nil
}

func (r *AnsibleRunner) waitForCompletion(cmd *exec.Cmd, run *models.TaskRun, runDir string, done chan struct{}) {
	_ = cmd.Wait()
	close(done)

	now := time.Now()
	run.EndedAt = &now
	duration := now.Sub(*run.StartedAt)

	arStatus := readAnsibleRunnerStatus(runDir)
	switch arStatus {
	case "successful":
		run.Status = models.TaskStatusSuccessful
	case "failed":
		run.Status = models.TaskStatusFailed
	default:
		if run.Status != models.TaskStatusCanceled {
			run.Status = models.TaskStatusFailed
		}
	}

	if rc, err := readAnsibleRunnerRC(runDir); err == nil {
		run.ExitCode = &rc
	}

	switch run.Status {
	case models.TaskStatusSuccessful:
		slog.Info("task completed", "run_id", run.ID, "playbook", run.Playbook, "duration", duration)
	default:
		slog.Warn("task did not complete successfully", "run_id", run.ID, "playbook", run.Playbook, "status", run.Status, "duration", duration)
	}

	writeRunJSON(runDir, run)

	r.releaseLock()
}

func (r *AnsibleRunner) Get(id string) (*models.TaskRun, error) {
	runDir := filepath.Join(r.artifactsDir, id)
	run, err := readRunJSON(runDir)
	if err != nil {
		return nil, ErrNotFound
	}
	return run, nil
}

func (r *AnsibleRunner) List() ([]models.TaskRun, error) {
	entries, err := os.ReadDir(r.artifactsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.TaskRun{}, nil
		}
		return nil, err
	}

	var runs []models.TaskRun
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		run, err := r.Get(entry.Name())
		if err != nil {
			continue
		}
		runs = append(runs, *run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].CreatedAt.After(runs[j].CreatedAt)
	})

	return runs, nil
}

func (r *AnsibleRunner) Logs(id string) ([]byte, error) {
	runDir := filepath.Join(r.artifactsDir, id)
	if _, err := readRunJSON(runDir); err != nil {
		return nil, ErrNotFound
	}

	data, err := os.ReadFile(filepath.Join(runDir, "stdout"))
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, nil
		}
		return nil, err
	}
	return data, nil
}

func (r *AnsibleRunner) Events(id string) ([]json.RawMessage, error) {
	runDir := filepath.Join(r.artifactsDir, id)
	if _, err := readRunJSON(runDir); err != nil {
		return nil, ErrNotFound
	}

	eventsDir := filepath.Join(runDir, "job_events")
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []json.RawMessage{}, nil
		}
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.SplitN(entries[i].Name(), "-", 2)[0])
		nj, _ := strconv.Atoi(strings.SplitN(entries[j].Name(), "-", 2)[0])
		return ni < nj
	})

	var events []json.RawMessage
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(eventsDir, entry.Name()))
		if err != nil {
			continue
		}
		events = append(events, json.RawMessage(data))
	}

	return events, nil
}

func (r *AnsibleRunner) Delete(id string) error {
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

func (r *AnsibleRunner) Recover() error {
	entries, err := os.ReadDir(r.artifactsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading artifacts dir: %w", err)
	}

	recovered := 0
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

		slog.Warn("recovering interrupted task", "run_id", run.ID, "playbook", run.Playbook, "pid", run.PID)

		if run.PID > 0 && processAlive(run.PID) {
			syscall.Kill(run.PID, syscall.SIGTERM)
			time.Sleep(5 * time.Second)
			syscall.Kill(run.PID, syscall.SIGKILL)
		}

		now := time.Now()
		arStatus := readAnsibleRunnerStatus(runDir)
		if arStatus == "successful" || arStatus == "failed" {
			run.Status = models.TaskStatus(arStatus)
			run.EndedAt = &now
			if rc, err := readAnsibleRunnerRC(runDir); err == nil {
				run.ExitCode = &rc
			}
		} else {
			run.Status = models.TaskStatusFailed
			run.EndedAt = &now
			run.Error = "recovered after server restart: process no longer running"
		}

		writeRunJSON(runDir, run)
		recovered++
	}

	if recovered > 0 {
		slog.Info("task recovery complete", "recovered", recovered)
	}

	return nil
}

func (r *AnsibleRunner) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	cmd := r.activeCmd
	run := r.activeRun
	done := r.done
	runDir := ""
	if run != nil {
		runDir = filepath.Join(r.artifactsDir, run.ID)
	}
	r.mu.Unlock()

	if cmd == nil || run == nil {
		return nil
	}

	slog.Info("shutting down active task", "run_id", run.ID, "playbook", run.Playbook)

	if cmd.Process != nil {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}

	select {
	case <-done:
	case <-ctx.Done():
		slog.Warn("task shutdown timed out, force-killing", "run_id", run.ID)
		if cmd.Process != nil {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-done
	}

	// waitForCompletion may have already updated the run; override with canceled.
	run.Status = models.TaskStatusCanceled
	now := time.Now()
	run.EndedAt = &now
	run.Error = "server shutdown"
	if runDir != "" {
		writeRunJSON(runDir, run)
	}

	return nil
}

func (r *AnsibleRunner) acquireLock() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeRun != nil {
		return ErrBusy
	}

	f, err := os.OpenFile(
		filepath.Join(r.artifactsDir, ".runner.lock"),
		os.O_CREATE|os.O_RDWR, 0640,
	)
	if err != nil {
		return fmt.Errorf("opening lock file: %w", err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return ErrBusy
	}

	r.lockFile = f
	return nil
}

func (r *AnsibleRunner) releaseLock() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lockFile != nil {
		syscall.Flock(int(r.lockFile.Fd()), syscall.LOCK_UN)
		r.lockFile.Close()
		r.lockFile = nil
	}
	r.activeRun = nil
	r.activeCmd = nil
	r.done = nil
}
