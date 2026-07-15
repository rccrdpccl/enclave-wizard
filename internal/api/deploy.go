package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/deploy"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"
)

// DeployHandler serves the /api/v1/deployments endpoints.
// These are the separated deployment endpoints (per the spec in section 7)
// that coexist with the legacy task-based deployment endpoint.
type DeployHandler struct {
	runner       runner.Runner
	registry     *plugins.Registry
	configReader *config.Reader
	configWriter *config.Writer
	enclaveDir   string

	mu          sync.Mutex
	deployments map[string]*models.Deployment
	latestID    string
}

func NewDeployHandler(r runner.Runner, registry *plugins.Registry, configReader *config.Reader, configWriter *config.Writer, enclaveDir string) *DeployHandler {
	return &DeployHandler{
		runner:       r,
		registry:     registry,
		configReader: configReader,
		configWriter: configWriter,
		enclaveDir:   enclaveDir,
		deployments:  make(map[string]*models.Deployment),
	}
}

// --- Request / Response types ---

type StartDeploymentOutput struct {
	Body models.Deployment
}

type GetDeploymentByIDInput struct {
	ID string `path:"id" doc:"Deployment identifier" minLength:"1"`
}

type GetDeploymentByIDOutput struct {
	Body models.Deployment
}

type GetDeploymentProgressByIDInput struct {
	ID string `path:"id" doc:"Deployment identifier" minLength:"1"`
}

type GetDeploymentProgressByIDOutput struct {
	Body models.DeploymentProgress
}

type CancelDeploymentInput struct {
	ID string `path:"id" doc:"Deployment identifier" minLength:"1"`
}

type GetCurrentDeploymentOutput struct {
	Body models.Deployment
}

// --- Registration ---

func (h *DeployHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "start-deployment",
		Method:      http.MethodPost,
		Path:        "/api/v1/deployments",
		Summary:     "Start deployment",
		Description: "Starts the full deployment chain (main playbook + addon plugins).",
		Tags:        []string{"Deployments"},
	}, h.startDeployment)

	huma.Register(api, huma.Operation{
		OperationID: "get-current-deployment",
		Method:      http.MethodGet,
		Path:        "/api/v1/deployments/current",
		Summary:     "Get current deployment",
		Description: "Returns the most recent deployment. Use this for reconnection after page reload.",
		Tags:        []string{"Deployments"},
	}, h.getCurrentDeployment)

	huma.Register(api, huma.Operation{
		OperationID: "get-deployment-by-id",
		Method:      http.MethodGet,
		Path:        "/api/v1/deployments/{id}",
		Summary:     "Get deployment by ID",
		Description: "Returns the deployment state with phases for the given deployment ID.",
		Tags:        []string{"Deployments"},
	}, h.getDeploymentByID)

	huma.Register(api, huma.Operation{
		OperationID: "get-deployment-progress-by-id",
		Method:      http.MethodGet,
		Path:        "/api/v1/deployments/{id}/progress",
		Summary:     "Get deployment progress",
		Description: "Returns live progress with completed task count, percentage, and current phase/task.",
		Tags:        []string{"Deployments"},
	}, h.getDeploymentProgressByID)

	huma.Register(api, huma.Operation{
		OperationID: "cancel-deployment",
		Method:      http.MethodDelete,
		Path:        "/api/v1/deployments/{id}",
		Summary:     "Cancel deployment",
		Description: "Cancels a running deployment.",
		Tags:        []string{"Deployments"},
	}, h.cancelDeployment)
}

// --- Handlers ---

func (h *DeployHandler) startDeployment(_ context.Context, _ *struct{}) (*StartDeploymentOutput, error) {
	run, err := h.runner.Start(runner.StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "playbooks/main.yaml",
		ExtraVars: map[string]string{
			"workingDir": h.enclaveDir,
		},
	})
	if err != nil {
		return nil, mapTaskError(err)
	}

	dep := &models.Deployment{
		ID:         run.ID,
		Status:     models.TaskStatusRunning,
		Phases:     []models.DeploymentPhase{{Name: "main", TaskID: run.ID, Status: models.TaskStatusRunning}},
		TotalTasks: 350,
		StartedAt:  time.Now(),
	}

	h.mu.Lock()
	h.deployments[dep.ID] = dep
	h.latestID = dep.ID
	h.mu.Unlock()

	go h.watchDeployment(dep.ID)

	return &StartDeploymentOutput{Body: *dep}, nil
}

func (h *DeployHandler) getCurrentDeployment(_ context.Context, _ *struct{}) (*GetCurrentDeploymentOutput, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.latestID == "" {
		return nil, huma.Error404NotFound("no deployment found")
	}
	dep, ok := h.deployments[h.latestID]
	if !ok {
		return nil, huma.Error404NotFound("no deployment found")
	}
	return &GetCurrentDeploymentOutput{Body: *dep}, nil
}

func (h *DeployHandler) getDeploymentByID(_ context.Context, input *GetDeploymentByIDInput) (*GetDeploymentByIDOutput, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	dep, ok := h.deployments[input.ID]
	if !ok {
		return nil, huma.Error404NotFound("deployment not found: " + input.ID)
	}
	return &GetDeploymentByIDOutput{Body: *dep}, nil
}

func (h *DeployHandler) getDeploymentProgressByID(_ context.Context, input *GetDeploymentProgressByIDInput) (*GetDeploymentProgressByIDOutput, error) {
	h.mu.Lock()
	dep, ok := h.deployments[input.ID]
	h.mu.Unlock()

	if !ok {
		return nil, huma.Error404NotFound("deployment not found: " + input.ID)
	}

	events, _ := h.runner.Events(dep.ID)
	progress := deploy.CalculateProgress(dep, events)
	return &GetDeploymentProgressByIDOutput{Body: progress}, nil
}

func (h *DeployHandler) cancelDeployment(_ context.Context, input *CancelDeploymentInput) (*struct{}, error) {
	h.mu.Lock()
	dep, ok := h.deployments[input.ID]
	h.mu.Unlock()

	if !ok {
		return nil, huma.Error404NotFound("deployment not found: " + input.ID)
	}

	if err := h.runner.Cancel(dep.ID); err != nil {
		return nil, mapTaskError(err)
	}

	h.mu.Lock()
	dep.Status = models.TaskStatusCanceled
	h.mu.Unlock()

	return nil, nil
}

// --- Helpers ---

func (h *DeployHandler) watchDeployment(deployID string) {
	for {
		time.Sleep(10 * time.Second)
		run, err := h.runner.Get(deployID)
		if err != nil {
			continue
		}
		if run.Status == models.TaskStatusRunning {
			continue
		}

		h.mu.Lock()
		if dep, ok := h.deployments[deployID]; ok {
			dep.Status = run.Status
		}
		h.mu.Unlock()
		return
	}
}

