package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"
)

var phasePlaybooks = map[int]string{
	1: "playbooks/01-prepare.yaml",
	2: "playbooks/02-mirror.yaml",
	3: "playbooks/03-deploy.yaml",
	4: "playbooks/04-post-install.yaml",
	5: "playbooks/05-operators.yaml",
	6: "playbooks/06-day2.yaml",
	7: "playbooks/07-configure-discovery.yaml",
}

type TasksHandler struct {
	runner       tasks.Runner
	registry     *plugins.Registry
	configReader *config.Reader
	configWriter *config.Writer
	enclaveDir   string

	deployMu   sync.Mutex
	deployment *models.Deployment
}

func NewTasksHandler(runner tasks.Runner, registry *plugins.Registry, configReader *config.Reader, configWriter *config.Writer, enclaveDir string) *TasksHandler {
	return &TasksHandler{runner: runner, registry: registry, configReader: configReader, configWriter: configWriter, enclaveDir: enclaveDir}
}

// --- Request / Response types ---

type StartDeployInput struct{}

type StartDeployPhaseInput struct {
	Phase int `path:"phase" doc:"Deployment phase number (1-7)" minimum:"1" maximum:"7"`
}

type StartDeployPluginInput struct {
	Name string `path:"name" doc:"Plugin name" minLength:"1"`
}

type StartTaskOutput struct {
	Body models.TaskRun
}

type ListTasksOutput struct {
	Body struct {
		Runs []models.TaskRun `json:"runs" doc:"All known task runs"`
	}
}

type GetTaskInput struct {
	ID string `path:"id" doc:"Run identifier" minLength:"1"`
}

type GetTaskOutput struct {
	Body models.TaskRun
}

type GetTaskLogsInput struct {
	ID string `path:"id" doc:"Run identifier" minLength:"1"`
}

type GetTaskLogsOutput struct {
	Body string
}

type GetTaskEventsInput struct {
	ID string `path:"id" doc:"Run identifier" minLength:"1"`
}

type GetTaskEventsOutput struct {
	Body struct {
		Events []json.RawMessage `json:"events" doc:"Ansible Runner job events"`
	}
}

type DeleteTaskInput struct {
	ID string `path:"id" doc:"Run identifier" minLength:"1"`
}

// --- Registration ---

func (h *TasksHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "start-deploy",
		Method:      http.MethodPost,
		Path:        "/api/v1/tasks/deploy",
		Summary:     "Start full deployment",
		Description: "Runs the main.yaml playbook (all 7 phases).",
		Tags:        []string{"Tasks"},
	}, h.startDeploy)

	huma.Register(api, huma.Operation{
		OperationID: "start-deploy-phase",
		Method:      http.MethodPost,
		Path:        "/api/v1/tasks/deploy/{phase}",
		Summary:     "Start a specific deployment phase",
		Description: "Runs a single deployment phase (1-7).",
		Tags:        []string{"Tasks"},
	}, h.startDeployPhase)

	huma.Register(api, huma.Operation{
		OperationID: "start-deploy-plugin",
		Method:      http.MethodPost,
		Path:        "/api/v1/tasks/plugins/{name}",
		Summary:     "Deploy a plugin",
		Description: "Runs the deploy-plugin.yaml playbook for the named plugin.",
		Tags:        []string{"Tasks"},
	}, h.startDeployPlugin)

	huma.Register(api, huma.Operation{
		OperationID: "list-tasks",
		Method:      http.MethodGet,
		Path:        "/api/v1/tasks",
		Summary:     "List all task runs",
		Description: "Returns all known task runs, most recent first.",
		Tags:        []string{"Tasks"},
	}, h.listTasks)

	huma.Register(api, huma.Operation{
		OperationID: "get-task",
		Method:      http.MethodGet,
		Path:        "/api/v1/tasks/{id}",
		Summary:     "Get task run details",
		Description: "Returns status and metadata for a specific run.",
		Tags:        []string{"Tasks"},
	}, h.getTask)

	huma.Register(api, huma.Operation{
		OperationID: "get-task-logs",
		Method:      http.MethodGet,
		Path:        "/api/v1/tasks/{id}/logs",
		Summary:     "Get task output logs",
		Description: "Returns ansible-runner stdout as text/plain. Use the offset query parameter for incremental reads.",
		Tags:        []string{"Tasks"},
	}, h.getTaskLogs)

	huma.Register(api, huma.Operation{
		OperationID: "get-task-events",
		Method:      http.MethodGet,
		Path:        "/api/v1/tasks/{id}/events",
		Summary:     "Get task job events",
		Description: "Returns ansible-runner job events as a JSON array.",
		Tags:        []string{"Tasks"},
	}, h.getTaskEvents)

	huma.Register(api, huma.Operation{
		OperationID: "delete-task",
		Method:      http.MethodDelete,
		Path:        "/api/v1/tasks/{id}",
		Summary:     "Delete a task run",
		Description: "Removes the ansible-runner directory for the given run. Returns 409 if the task is still running.",
		Tags:        []string{"Tasks"},
	}, h.deleteTask)

	huma.Register(api, huma.Operation{
		OperationID: "start-validate",
		Method:      http.MethodPost,
		Path:        "/api/v1/tasks/validate",
		Summary:     "Run operational validation (validations.sh)",
		Description: "Runs the enclave operational validation script that checks DNS, Redfish, certificates, and registry connectivity.",
		Tags:        []string{"Tasks"},
	}, h.startValidate)

	huma.Register(api, huma.Operation{
		OperationID: "get-deployment",
		Method:      http.MethodGet,
		Path:        "/api/v1/deployment",
		Summary:     "Get current deployment state",
		Description: "Returns the full deployment chain state including main playbook and all addon plugin phases.",
		Tags:        []string{"Deployment"},
	}, h.getDeployment)

	huma.Register(api, huma.Operation{
		OperationID: "get-deployment-progress",
		Method:      http.MethodGet,
		Path:        "/api/v1/deployment/progress",
		Summary:     "Get deployment progress",
		Description: "Returns live progress with completed task count, percentage, and current phase/task.",
		Tags:        []string{"Deployment"},
	}, h.getDeploymentProgress)
}

// --- Handlers ---

func (h *TasksHandler) startDeploy(ctx context.Context, _ *StartDeployInput) (*StartTaskOutput, error) {
	cfg, err := h.configReader.ReadAll()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read config", err)
	}

	// WORKAROUND: force connected mode until disconnected support is ready
	f := false
	cfg.Global.Disconnected = &f

	if err := h.configWriter.WriteAll(cfg); err != nil {
		return nil, huma.Error500InternalServerError("failed to write config before deploy", err)
	}

	addonPlugins := h.addonPluginsFromConfig()

	run, err := h.runner.Start(tasks.StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: "playbooks/main.yaml",
		ExtraVars: map[string]string{
			"workingDir": h.enclaveDir,
		},
	})
	if err != nil {
		return nil, mapTaskError(err)
	}

	// Create deployment tracking
	phases := []models.DeploymentPhase{
		{Name: "main", TaskID: run.ID, Status: models.TaskStatusRunning},
	}
	for _, p := range addonPlugins {
		phases = append(phases, models.DeploymentPhase{Name: p, Status: models.TaskStatusPending})
	}

	h.deployMu.Lock()
	h.deployment = &models.Deployment{
		ID:     run.ID,
		Status: models.TaskStatusRunning,
		Phases: phases,
	}
	h.deployMu.Unlock()

	if len(addonPlugins) > 0 {
		slog.Info("deploy with addon plugins queued", "plugins", addonPlugins)
		go h.chainAddonPlugins(run.ID, addonPlugins)
	} else {
		go h.watchMainDeploy(run.ID)
	}

	return &StartTaskOutput{Body: *run}, nil
}

func (h *TasksHandler) addonPluginsFromConfig() []string {
	if h.configReader == nil {
		return nil
	}
	cfg, err := h.configReader.ReadAll()
	if err != nil {
		return nil
	}
	type addonInfo struct {
		name  string
		order int
	}
	var addons []addonInfo
	for _, name := range cfg.Global.EnabledPlugins {
		p, ok := h.registry.Get(name)
		if !ok {
			continue
		}
		if p.Type == models.PluginTypeAddon {
			addons = append(addons, addonInfo{name: p.Name, order: p.Order})
		}
	}
	sort.Slice(addons, func(i, j int) bool { return addons[i].order < addons[j].order })
	result := make([]string, len(addons))
	for i, a := range addons {
		result[i] = a.name
	}
	return result
}

func (h *TasksHandler) setPhaseStatus(name string, status models.TaskStatus, taskID string) {
	h.deployMu.Lock()
	defer h.deployMu.Unlock()
	if h.deployment == nil {
		return
	}
	for i := range h.deployment.Phases {
		if h.deployment.Phases[i].Name == name {
			h.deployment.Phases[i].Status = status
			if taskID != "" {
				h.deployment.Phases[i].TaskID = taskID
			}
			break
		}
	}
}

func (h *TasksHandler) setDeploymentStatus(status models.TaskStatus) {
	h.deployMu.Lock()
	defer h.deployMu.Unlock()
	if h.deployment != nil {
		h.deployment.Status = status
	}
}

func (h *TasksHandler) watchMainDeploy(mainRunID string) {
	for {
		time.Sleep(10 * time.Second)
		mainRun, err := h.runner.Get(mainRunID)
		if err != nil {
			continue
		}
		if mainRun.Status == models.TaskStatusRunning {
			continue
		}
		h.setPhaseStatus("main", mainRun.Status, "")
		h.setDeploymentStatus(mainRun.Status)
		return
	}
}

func (h *TasksHandler) chainAddonPlugins(mainRunID string, pluginNames []string) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("addon plugin chain panicked", "error", r)
		}
	}()

	slog.Info("waiting for main deploy to complete before addon plugins", "plugins", pluginNames)

	for {
		time.Sleep(10 * time.Second)
		mainRun, err := h.runner.Get(mainRunID)
		if err != nil {
			continue
		}
		if mainRun.Status == models.TaskStatusRunning {
			continue
		}
		if mainRun.Status != models.TaskStatusSuccessful {
			slog.Warn("skipping addon plugins — main deploy did not succeed",
				"run_id", mainRunID, "status", mainRun.Status)
			h.setPhaseStatus("main", mainRun.Status, "")
			h.setDeploymentStatus(mainRun.Status)
			return
		}
		break
	}

	h.setPhaseStatus("main", models.TaskStatusSuccessful, "")
	slog.Info("main deploy completed, deploying addon plugins", "plugins", pluginNames)

	for _, name := range pluginNames {
		slog.Info("deploying addon plugin", "plugin", name)
		h.setPhaseStatus(name, models.TaskStatusRunning, "")

		run, _, err := h.runner.RunSync(context.Background(), tasks.StartRequest{
			Type:     models.TaskTypeDeployPlugin,
			Playbook: "playbooks/deploy-plugin.yaml",
			ExtraVars: map[string]string{
				"plugin_name": name,
				"workingDir":  h.enclaveDir,
			},
		})
		if err != nil {
			slog.Error("addon plugin deploy failed to start", "plugin", name, "error", err)
			h.setPhaseStatus(name, models.TaskStatusFailed, "")
			h.setDeploymentStatus(models.TaskStatusFailed)
			return
		}

		h.setPhaseStatus(name, run.Status, run.ID)

		if run.Status != models.TaskStatusSuccessful {
			slog.Error("addon plugin deploy failed", "plugin", name, "status", run.Status)
			h.setDeploymentStatus(models.TaskStatusFailed)
			return
		}
		slog.Info("addon plugin deployed successfully", "plugin", name)
	}

	slog.Info("all addon plugins deployed successfully")
	h.setDeploymentStatus(models.TaskStatusSuccessful)
}

func (h *TasksHandler) startDeployPhase(ctx context.Context, input *StartDeployPhaseInput) (*StartTaskOutput, error) {
	playbook, ok := phasePlaybooks[input.Phase]
	if !ok {
		return nil, huma.Error400BadRequest(fmt.Sprintf("invalid phase: %d", input.Phase))
	}
	run, err := h.runner.Start(tasks.StartRequest{
		Type:     models.TaskTypeDeployPhase,
		Playbook: playbook,
		ExtraVars: map[string]string{
			"workingDir": h.enclaveDir,
		},
	})
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &StartTaskOutput{Body: *run}, nil
}

func (h *TasksHandler) startDeployPlugin(ctx context.Context, input *StartDeployPluginInput) (*StartTaskOutput, error) {
	if _, ok := h.registry.Get(input.Name); !ok {
		return nil, huma.Error404NotFound("unknown plugin: " + input.Name)
	}
	run, err := h.runner.Start(tasks.StartRequest{
		Type:     models.TaskTypeDeployPlugin,
		Playbook: "playbooks/deploy-plugin.yaml",
		ExtraVars: map[string]string{
			"plugin_name": input.Name,
			"workingDir":  h.enclaveDir,
		},
	})
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &StartTaskOutput{Body: *run}, nil
}

func (h *TasksHandler) listTasks(_ context.Context, _ *struct{}) (*ListTasksOutput, error) {
	runs, err := h.runner.List()
	if err != nil {
		return nil, mapTaskError(err)
	}
	out := &ListTasksOutput{}
	if runs == nil {
		runs = []models.TaskRun{}
	}
	out.Body.Runs = runs
	return out, nil
}

func (h *TasksHandler) getTask(_ context.Context, input *GetTaskInput) (*GetTaskOutput, error) {
	run, err := h.runner.Get(input.ID)
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &GetTaskOutput{Body: *run}, nil
}

func (h *TasksHandler) getTaskLogs(_ context.Context, input *GetTaskLogsInput) (*GetTaskLogsOutput, error) {
	data, err := h.runner.Logs(input.ID)
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &GetTaskLogsOutput{Body: string(data)}, nil
}

func (h *TasksHandler) getTaskEvents(_ context.Context, input *GetTaskEventsInput) (*GetTaskEventsOutput, error) {
	events, err := h.runner.Events(input.ID)
	if err != nil {
		return nil, mapTaskError(err)
	}
	out := &GetTaskEventsOutput{}
	out.Body.Events = events
	return out, nil
}

func (h *TasksHandler) deleteTask(_ context.Context, input *DeleteTaskInput) (*struct{}, error) {
	if err := h.runner.Delete(input.ID); err != nil {
		return nil, mapTaskError(err)
	}
	return nil, nil
}

type StartValidateInput struct{}

func (h *TasksHandler) startValidate(ctx context.Context, _ *StartValidateInput) (*StartTaskOutput, error) {
	run, err := h.runner.Start(tasks.StartRequest{
		Type:     models.TaskTypeValidate,
		Playbook: "validations.sh",
	})
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &StartTaskOutput{Body: *run}, nil
}

// --- Deployment endpoints ---

type GetDeploymentOutput struct {
	Body models.Deployment
}

func (h *TasksHandler) getDeployment(_ context.Context, _ *struct{}) (*GetDeploymentOutput, error) {
	h.deployMu.Lock()
	defer h.deployMu.Unlock()

	if h.deployment == nil {
		return nil, huma.Error404NotFound("no active deployment")
	}
	return &GetDeploymentOutput{Body: *h.deployment}, nil
}

type GetDeploymentProgressOutput struct {
	Body models.DeploymentProgress
}

func (h *TasksHandler) getDeploymentProgress(_ context.Context, _ *struct{}) (*GetDeploymentProgressOutput, error) {
	h.deployMu.Lock()
	dep := h.deployment
	h.deployMu.Unlock()

	if dep == nil {
		return nil, huma.Error404NotFound("no active deployment")
	}

	completedPhases := 0
	totalPhases := len(dep.Phases)
	currentPhase := ""
	currentTask := ""

	for _, phase := range dep.Phases {
		switch phase.Status {
		case models.TaskStatusSuccessful:
			completedPhases++
		case models.TaskStatusRunning:
			currentPhase = phase.Name
		}

		if phase.Status != models.TaskStatusRunning || phase.TaskID == "" {
			continue
		}
		events, err := h.runner.Events(phase.TaskID)
		if err != nil {
			continue
		}
		for _, raw := range events {
			var ev struct {
				Event     string `json:"event"`
				EventData struct {
					Task string `json:"task"`
				} `json:"event_data"`
			}
			if json.Unmarshal(raw, &ev) != nil {
				continue
			}
			if strings.HasPrefix(ev.Event, "runner_on_") {
				currentTask = ev.EventData.Task
			}
		}
	}

	pct := 0
	if dep.Status == models.TaskStatusSuccessful {
		pct = 100
	} else if dep.Status == models.TaskStatusFailed {
		pct = completedPhases * 100 / totalPhases
	} else if totalPhases > 0 {
		pct = completedPhases * 100 / totalPhases
	}

	return &GetDeploymentProgressOutput{
		Body: models.DeploymentProgress{
			Completed:    completedPhases,
			Total:        totalPhases,
			Percentage:   pct,
			CurrentPhase: currentPhase,
			CurrentTask:  currentTask,
		},
	}, nil
}

func mapTaskError(err error) error {
	switch {
	case errors.Is(err, tasks.ErrBusy):
		return huma.Error409Conflict("a task is already running")
	case errors.Is(err, tasks.ErrRunning):
		return huma.Error409Conflict("task is still running")
	case errors.Is(err, tasks.ErrNotFound):
		return huma.Error404NotFound("run not found")
	default:
		return huma.Error500InternalServerError("task operation failed", err)
	}
}
