import type { EnclaveConfig, TaskRun } from "@enclave-wizard-ui/api-client";
import { useCallback, useEffect, useMemo, useState } from "react";
import { usePolling } from "../tasks/hooks/usePolling.ts";
import { useEnclaveApi } from "./useEnclaveApi.ts";
import { useTasksApi } from "./useTasksApi.ts";

export type DeployPhase =
  | "idle"
  | "writing"
  | "deploying"
  | "complete"
  | "failed"
  | "error";

export interface DeploymentError {
  message: string;
  details: string[];
}

export interface DeploymentState {
  phase: DeployPhase;
  task: TaskRun | null;
  logs: string;
  error: DeploymentError | null;
  taskId: string | null;
}

export interface UseDeploymentReturn {
  state: DeploymentState;
  start: (config: EnclaveConfig) => Promise<void>;
}

export function useDeployment(): UseDeploymentReturn {
  const api = useEnclaveApi();
  const tasksApi = useTasksApi();
  const [phase, setPhase] = useState<DeployPhase>("idle");
  const [errorState, setErrorState] = useState<DeploymentError | null>(null);
  const [taskId, setTaskId] = useState<string | null>(null);
  const [taskDone, setTaskDone] = useState(false);

  const start = useCallback(
    async (config: EnclaveConfig) => {
      setPhase("writing");
      setErrorState(null);
      try {
        await api.writeConfig(config);
        setPhase("deploying");
        const task = await tasksApi.startDeploy();
        setTaskId(task.id);
        setTaskDone(false);
      } catch (err: unknown) {
        setPhase("error");
        const details: string[] = [];
        if (err && typeof err === "object" && "response" in err) {
          try {
            const body = await (err as { response: Response }).response.json();
            if (body.errors && Array.isArray(body.errors)) {
              for (const e of body.errors) {
                details.push(e.message ?? String(e));
              }
            }
            if (!details.length && body.detail) {
              details.push(body.detail);
            }
          } catch {
            // response not JSON
          }
        }
        setErrorState({
          message:
            details.length > 0
              ? "Configuration validation failed"
              : err instanceof Error
                ? err.message
                : "Failed to start deployment",
          details,
        });
      }
    },
    [api, tasksApi],
  );

  // Poll task status
  const fetchTask = useCallback(
    () => (taskId ? tasksApi.getTask(taskId) : Promise.resolve(null)),
    [tasksApi, taskId],
  );
  const { data: task } = usePolling(fetchTask, 3000, !!taskId && !taskDone);

  const isRunning = task?.status === "running";

  useEffect(() => {
    if (task == null) return;
    if (isRunning) return;
    setTaskDone(true);
    setPhase(task.status === "failed" ? "failed" : "complete");
  }, [task, isRunning]);

  // Poll logs
  const fetchLogs = useCallback(
    () => (taskId ? tasksApi.getTaskLogs(taskId) : Promise.resolve("")),
    [tasksApi, taskId],
  );
  const { data: logs } = usePolling(fetchLogs, 2000, !!taskId && !taskDone);

  const state: DeploymentState = useMemo(
    () => ({
      phase,
      task: task ?? null,
      logs: logs ?? "",
      error: errorState,
      taskId,
    }),
    [phase, task, logs, errorState, taskId],
  );

  return { state, start };
}
