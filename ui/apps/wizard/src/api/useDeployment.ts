import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { EnclaveConfig, TaskRun } from "@enclave-wizard-ui/api-client";
import { useEnclaveApi } from "./useEnclaveApi.ts";
import { useTasksApi } from "./useTasksApi.ts";

export interface DeploymentError {
  message: string;
  details: string[];
}

export type DeployPhase =
  | "idle"
  | "writing"
  | "deploying"
  | "complete"
  | "failed"
  | "error";

export interface DeploymentState {
  phase: DeployPhase;
  deploymentId: string | null;
  taskId: string | null;
  task: TaskRun | null;
  logs: string;
  error: DeploymentError | null;
}

export interface UseDeploymentReturn {
  state: DeploymentState;
  start: (config: EnclaveConfig) => Promise<void>;
  cancel: () => Promise<void>;
}

const INITIAL_STATE: DeploymentState = {
  phase: "idle",
  deploymentId: null,
  taskId: null,
  task: null,
  logs: "",
  error: null,
};

/** Try a fetch; return the Response on success, null on 404. Throws on other errors. */
async function fetchWithFallback(
  url: string,
  init?: RequestInit,
): Promise<Response | null> {
  const resp = await fetch(url, init);
  if (resp.status === 404) return null;
  if (!resp.ok) {
    const text = await resp.text().catch(() => "");
    throw new Error(`${resp.status}: ${text}`);
  }
  return resp;
}

async function extractErrorDetails(err: unknown): Promise<DeploymentError> {
  const details: string[] = [];
  if (err && typeof err === "object" && "response" in err) {
    try {
      const body = await (err as { response: Response }).response.json();
      if (body.errors && Array.isArray(body.errors)) {
        for (const e of body.errors) {
          details.push(e.message ?? String(e));
        }
      } else if (body.detail) {
        details.push(body.detail);
      }
    } catch {
      // response not JSON
    }
  }
  return {
    message:
      details.length > 0
        ? "Configuration validation failed"
        : err instanceof Error
          ? err.message
          : "Failed to start deployment",
    details,
  };
}

export function useDeployment(): UseDeploymentReturn {
  const api = useEnclaveApi();
  const tasksApi = useTasksApi();
  const [state, setState] = useState<DeploymentState>(INITIAL_STATE);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const mountedRef = useRef(true);

  // Cleanup on unmount
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, []);

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, []);

  const pollDeployment = useCallback(
    (deploymentId: string | null, taskId: string) => {
      stopPolling();

      const tick = async () => {
        if (!mountedRef.current) return;

        try {
          // Try new progress endpoint first, fall back to task status
          let task: TaskRun | null = null;
          if (deploymentId) {
            const resp = await fetchWithFallback(
              `/api/v1/deployments/${deploymentId}/progress`,
            );
            if (resp) {
              task = await resp.json();
            }
          }

          if (!task) {
            task = await tasksApi.getTask(taskId);
          }

          // Fetch logs
          let logs = "";
          try {
            logs = await tasksApi.getTaskLogs(taskId);
          } catch {
            // logs may not be available yet
          }

          if (!mountedRef.current) return;

          const isRunning = task.status === "running";
          const isFailed = task.status === "failed" || task.status === "error";
          const phase: DeployPhase = isRunning
            ? "deploying"
            : isFailed
              ? "failed"
              : "complete";

          setState((prev) => ({
            ...prev,
            phase,
            task,
            logs,
            error: task?.error
              ? { message: task.error, details: [] }
              : prev.error,
          }));

          if (!isRunning) {
            stopPolling();
          }
        } catch {
          // Polling errors are transient; keep trying
        }
      };

      tick();
      pollRef.current = setInterval(tick, 3000);
    },
    [tasksApi, stopPolling],
  );

  const start = useCallback(
    async (config: EnclaveConfig) => {
      setState({
        ...INITIAL_STATE,
        phase: "writing",
      });

      try {
        // Write config
        await api.writeConfig(config);

        if (!mountedRef.current) return;
        setState((prev) => ({ ...prev, phase: "deploying" }));

        // Try new deployment endpoint first
        let deploymentId: string | null = null;
        let taskId: string | null = null;

        const resp = await fetchWithFallback("/api/v1/deployments", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({}),
        });

        if (resp) {
          const body = await resp.json();
          deploymentId = body.id ?? null;
          taskId = body.taskId ?? body.id ?? null;
        }

        // Fall back to old endpoint
        if (!taskId) {
          const task = await tasksApi.startDeploy();
          taskId = task.id;
        }

        if (!mountedRef.current) return;
        setState((prev) => ({
          ...prev,
          deploymentId,
          taskId,
        }));

        pollDeployment(deploymentId, taskId);
      } catch (err: unknown) {
        if (!mountedRef.current) return;
        const error = await extractErrorDetails(err);
        setState((prev) => ({
          ...prev,
          phase: "error",
          error,
        }));
      }
    },
    [api, tasksApi, pollDeployment],
  );

  const cancel = useCallback(async () => {
    const { deploymentId, taskId } = state;
    stopPolling();

    try {
      if (deploymentId) {
        const resp = await fetchWithFallback(
          `/api/v1/deployments/${deploymentId}`,
          { method: "DELETE" },
        );
        if (resp) {
          setState(INITIAL_STATE);
          return;
        }
      }

      // Fall back: delete the task via old API
      if (taskId) {
        await tasksApi.deleteTask(taskId);
      }

      setState(INITIAL_STATE);
    } catch (err: unknown) {
      setState((prev) => ({
        ...prev,
        phase: "error",
        error: {
          message:
            err instanceof Error ? err.message : "Failed to cancel deployment",
          details: [],
        },
      }));
    }
  }, [state, tasksApi, stopPolling]);

  // Reconnect to an in-progress deployment on mount
  useEffect(() => {
    const reconnect = async () => {
      try {
        // Try new endpoint first
        const resp = await fetchWithFallback("/api/v1/deployments/current");
        if (resp) {
          const body = await resp.json();
          const deploymentId = body.id ?? null;
          const taskId = body.taskId ?? body.id ?? null;
          if (taskId) {
            setState((prev) => ({
              ...prev,
              phase: "deploying",
              deploymentId,
              taskId,
            }));
            pollDeployment(deploymentId, taskId);
          }
          return;
        }
      } catch {
        // New endpoint not available
      }

      // Fall back: check old deployment endpoint
      try {
        const resp = await fetch("/api/v1/deployment");
        if (resp.ok) {
          const body = await resp.json();
          if (body.taskId && body.status === "running") {
            setState((prev) => ({
              ...prev,
              phase: "deploying",
              taskId: body.taskId,
            }));
            pollDeployment(null, body.taskId);
          }
        }
      } catch {
        // No active deployment
      }
    };

    reconnect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return useMemo(
    () => ({ state, start, cancel }),
    [state, start, cancel],
  );
}
