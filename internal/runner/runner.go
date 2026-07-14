package runner

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

var (
	ErrBusy        = errors.New("a task is already running")
	ErrNotFound    = errors.New("run not found")
	ErrRunnerBin   = errors.New("ansible-runner binary not found in PATH")
	ErrRunning     = errors.New("task is still running")
	ErrNoRecording = errors.New("no recording found for this playbook/tags combination")
)

// StartRequest describes a task to execute.
type StartRequest struct {
	Type      models.TaskType
	Playbook  string
	ExtraVars map[string]string
	Tags      []string
	Env       map[string]string
}

// Event represents a structured event emitted during a task run.
type Event struct {
	Type string          // "status", "progress", "log"
	Data json.RawMessage
}

//go:generate go run go.uber.org/mock/mockgen -source=runner.go -destination=mock_runner.go -package=runner
type Runner interface {
	Start(req StartRequest) (*models.TaskRun, error)
	RunSync(ctx context.Context, req StartRequest) (*models.TaskRun, []byte, error)
	Cancel(id string) error
	Get(id string) (*models.TaskRun, error)
	List() ([]models.TaskRun, error)
	Logs(id string) ([]byte, error)
	Events(id string) ([]json.RawMessage, error)
	Stream(id string) (<-chan Event, error)
	Delete(id string) error
	Shutdown(ctx context.Context) error
	Recover() error
}
