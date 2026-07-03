package models

import "time"

type TaskStatus string

const (
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusSuccessful TaskStatus = "successful"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCanceled   TaskStatus = "canceled"
)

type TaskType string

const (
	TaskTypeDeploy       TaskType = "deploy"
	TaskTypeDeployPhase  TaskType = "deploy-phase"
	TaskTypeDeployPlugin TaskType = "deploy-plugin"
	TaskTypeValidate     TaskType = "validate"
)

const TaskStatusPending TaskStatus = "pending"

type DeploymentPhase struct {
	Name   string     `json:"name" doc:"Phase name (main, trust-manager, osac, etc.)"`
	TaskID string     `json:"taskId,omitempty" doc:"Task run ID for this phase"`
	Status TaskStatus `json:"status" doc:"Phase status" enum:"pending,running,successful,failed"`
}

type Deployment struct {
	ID         string            `json:"id" doc:"Deployment identifier"`
	Status     TaskStatus        `json:"status" doc:"Overall deployment status"`
	Phases     []DeploymentPhase `json:"phases" doc:"Ordered list of deployment phases"`
	TotalTasks int               `json:"totalTasks" doc:"Estimated total ansible tasks across all phases"`
}

type DeploymentProgress struct {
	Completed    int    `json:"completed" doc:"Number of completed ansible tasks"`
	Total        int    `json:"total" doc:"Estimated total tasks"`
	Percentage   int    `json:"percentage" doc:"Completion percentage (0-100)"`
	CurrentPhase string `json:"currentPhase" doc:"Name of the currently running phase"`
	CurrentTask  string `json:"currentTask" doc:"Name of the currently running ansible task"`
}

type TaskRun struct {
	ID        string            `json:"id" doc:"Unique run identifier"`
	Type      TaskType          `json:"type" doc:"Type of task" enum:"deploy,deploy-phase,deploy-plugin,validate"`
	Status    TaskStatus        `json:"status" doc:"Current execution status" enum:"running,successful,failed,canceled"`
	Playbook  string            `json:"playbook" doc:"Playbook path relative to enclave directory"`
	ExtraVars map[string]string `json:"extraVars,omitempty" doc:"Extra variables passed to ansible-runner"`
	PID       int               `json:"pid,omitempty" doc:"OS process ID of ansible-runner"`
	ExitCode  *int              `json:"exitCode,omitempty" doc:"Process exit code"`
	StartedAt time.Time          `json:"startedAt" doc:"When ansible-runner started"`
	EndedAt   *time.Time        `json:"endedAt,omitempty" doc:"When the run completed"`
	Error     string            `json:"error,omitempty" doc:"Error message if failed"`
}
