package tasks

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)


var (
	ErrBusy      = errors.New("a task is already running")
	ErrNotFound  = errors.New("run not found")
	ErrRunnerBin = errors.New("ansible-runner binary not found in PATH")
)

type StartRequest struct {
	Type      models.TaskType
	Playbook  string
	ExtraVars map[string]string
}

//go:generate go run go.uber.org/mock/mockgen -source=runner.go -destination=mock_runner.go -package=tasks
type Runner interface {
	Start(req StartRequest) (*models.TaskRun, error)
	Get(id string) (*models.TaskRun, error)
	List() ([]models.TaskRun, error)
	Logs(id string) ([]byte, error)
	Events(id string) ([]json.RawMessage, error)
	Shutdown(ctx context.Context) error
	Recover() error
}

