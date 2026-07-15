package deploy

import (
	"encoding/json"
	"strings"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

// CalculateProgress computes deployment progress from runner job events.
func CalculateProgress(dep *models.Deployment, events []json.RawMessage) models.DeploymentProgress {
	completedTasks := 0
	currentTask := ""

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
		switch {
		case strings.HasPrefix(ev.Event, "runner_on_"):
			completedTasks++
			currentTask = ev.EventData.Task
		case ev.Event == "playbook_on_task_start" && ev.EventData.Task != "":
			currentTask = ev.EventData.Task
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

	return models.DeploymentProgress{
		Completed:    completedTasks,
		Total:        total,
		Percentage:   pct,
		CurrentPhase: "",
		CurrentTask:  currentTask,
	}
}
