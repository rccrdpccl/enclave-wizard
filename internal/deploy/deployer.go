package deploy

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"
)

// Deployer manages the deployment lifecycle: starting deployments, tracking
// phases, watching progress, and persisting state.
type Deployer struct {
	runner       runner.Runner
	registry     *plugins.Registry
	configReader *config.Reader
	configWriter *config.Writer
	enclaveDir   string
	state        *State
	stopCh       chan struct{}
}

// NewDeployer creates a Deployer that coordinates deployments using the given
// runner and persists state via the given State.
func NewDeployer(r runner.Runner, registry *plugins.Registry, configReader *config.Reader, configWriter *config.Writer, enclaveDir string, state *State) *Deployer {
	return &Deployer{
		runner:       r,
		registry:     registry,
		configReader: configReader,
		configWriter: configWriter,
		enclaveDir:   enclaveDir,
		state:        state,
		stopCh:       make(chan struct{}),
	}
}

// Stop signals all background goroutines to exit.
func (d *Deployer) Stop() {
	select {
	case <-d.stopCh:
	default:
		close(d.stopCh)
	}
}

// Start initiates a new deployment. It reads config, writes it to disk,
// generates the deployment playbook, starts the runner task, and watches
// for completion in the background.
func (d *Deployer) Start() (*models.Deployment, *models.TaskRun, error) {
	cfg, err := d.configReader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("reading config: %w", err)
	}

	// WORKAROUND: force connected mode until disconnected support is ready
	f := false
	cfg.Global.Disconnected = &f

	if err := d.configWriter.WriteAll(cfg); err != nil {
		return nil, nil, fmt.Errorf("writing config: %w", err)
	}

	addonPlugins := d.addonPluginsFromConfig()

	playbook := "playbooks/main.yaml"
	if len(addonPlugins) > 0 {
		generated, err := d.generateDeployPlaybook(addonPlugins)
		if err != nil {
			return nil, nil, fmt.Errorf("generating deploy playbook: %w", err)
		}
		playbook = generated
		slog.Info("generated deploy playbook with addon plugins", "playbook", playbook, "plugins", addonPlugins)
	}

	run, err := d.runner.Start(runner.StartRequest{
		Type:     models.TaskTypeDeploy,
		Playbook: playbook,
		ExtraVars: map[string]string{
			"workingDir": d.enclaveDir,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	phases := []models.DeploymentPhase{
		{Name: "main", TaskID: run.ID, Status: models.TaskStatusRunning},
	}
	for _, p := range addonPlugins {
		phases = append(phases, models.DeploymentPhase{Name: p, TaskID: run.ID, Status: models.TaskStatusPending})
	}

	dep := &models.Deployment{
		ID:         run.ID,
		Status:     models.TaskStatusRunning,
		Phases:     phases,
		TotalTasks: 350 + len(addonPlugins)*30,
		StartedAt:  time.Now(),
	}

	if err := d.state.Save(dep); err != nil {
		slog.Error("failed to save deployment state", "error", err)
	}

	go d.watchDeploy(dep.ID)

	depCopy := *dep
	return &depCopy, run, nil
}

// Cancel cancels a deployment by stopping the running phase's task.
func (d *Deployer) Cancel(id string) error {
	dep, err := d.state.Load(id)
	if err != nil {
		return err
	}

	for _, p := range dep.Phases {
		if p.Status == models.TaskStatusRunning {
			if err := d.runner.Cancel(p.TaskID); err != nil {
				return fmt.Errorf("canceling task %s: %w", p.TaskID, err)
			}
			break
		}
	}

	dep.Status = models.TaskStatusCanceled
	return d.state.Save(dep)
}

// Get returns a deployment by ID.
func (d *Deployer) Get(id string) (*models.Deployment, error) {
	return d.state.Load(id)
}

// GetLatest returns the most recent deployment.
func (d *Deployer) GetLatest() (*models.Deployment, error) {
	return d.state.LoadLatest()
}

// GetProgress calculates live deployment progress from runner events.
func (d *Deployer) GetProgress(id string) (*models.DeploymentProgress, error) {
	dep, err := d.state.Load(id)
	if err != nil {
		return nil, err
	}

	completedTasks := 0
	currentTask := ""

	events, err := d.runner.Events(dep.ID)
	if err == nil {
		for _, raw := range events {
			var ev struct {
				Event     string `json:"event"`
				EventData struct {
					Task string `json:"task"`
					Play string `json:"play"`
				} `json:"event_data"`
			}
			if json.Unmarshal(raw, &ev) != nil {
				continue
			}
			switch {
			case strings.HasPrefix(ev.Event, "runner_on_"):
				completedTasks++
				currentTask = ev.EventData.Task
			case ev.Event == "playbook_on_task_start" && ev.EventData.Task != "":
				currentTask = ev.EventData.Task
			}
		}
	}

	total := dep.TotalTasks
	if total == 0 {
		total = 350
	}
	pct := completedTasks * 100 / total
	if pct > 99 && dep.Status == models.TaskStatusRunning {
		pct = 99
	}
	if dep.Status == models.TaskStatusSuccessful {
		pct = 100
	}

	return &models.DeploymentProgress{
		Completed:    completedTasks,
		Total:        total,
		Percentage:   pct,
		CurrentPhase: "",
		CurrentTask:  currentTask,
	}, nil
}

func (d *Deployer) watchDeploy(depID string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
		}

		run, err := d.runner.Get(depID)
		if err != nil {
			continue
		}
		if run.Status == models.TaskStatusRunning {
			continue
		}

		dep, err := d.state.Load(depID)
		if err != nil {
			slog.Error("failed to load deployment state for update", "error", err)
			return
		}
		dep.Status = run.Status
		if err := d.state.Save(dep); err != nil {
			slog.Error("failed to save deployment state after completion", "error", err)
		}
		return
	}
}

func (d *Deployer) addonPluginsFromConfig() []string {
	if d.configReader == nil {
		return nil
	}
	cfg, err := d.configReader.ReadAll()
	if err != nil {
		return nil
	}
	type addonInfo struct {
		name  string
		order int
	}
	var addons []addonInfo
	for _, name := range cfg.Global.EnabledPlugins {
		p, ok := d.registry.Get(name)
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

func (d *Deployer) generateDeployPlaybook(addonPlugins []string) (string, error) {
	var buf strings.Builder
	buf.WriteString("---\n# Auto-generated: main deploy + addon plugins\n")
	buf.WriteString("- ansible.builtin.import_playbook: main.yaml\n")
	for _, name := range addonPlugins {
		fmt.Fprintf(&buf, "- ansible.builtin.import_playbook: deploy-plugin.yaml\n")
		fmt.Fprintf(&buf, "  vars:\n")
		fmt.Fprintf(&buf, "    plugin_name: %s\n", name)
		fmt.Fprintf(&buf, "    plugin_mirror: true\n")
	}

	dir := fmt.Sprintf("%s/playbooks", d.enclaveDir)
	if err := makeDir(dir); err != nil {
		return "", err
	}
	path := fmt.Sprintf("%s/deploy-all.yaml", dir)
	if err := writeFile(path, []byte(buf.String())); err != nil {
		return "", fmt.Errorf("writing deploy-all.yaml: %w", err)
	}
	return "playbooks/deploy-all.yaml", nil
}

// Thin wrappers to aid testing.
var makeDir = func(path string) error {
	return os.MkdirAll(path, 0750)
}

var writeFile = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0640)
}
