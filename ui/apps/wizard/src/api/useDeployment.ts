import { useCallback, useEffect, useRef, useState } from "react";
import { usePolling } from "../tasks/hooks/usePolling.ts";
import { useEnclaveApi } from "./useEnclaveApi.ts";
import { type SSEEvent, useSSE } from "./useSSE.ts";
import { useTasksApi } from "./useTasksApi.ts";

export interface DeploymentProgress {
  percentage: number;
  currentTask: string;
}

export interface DeploymentError {
  message: string;
  details: string[];
}

export type DeploymentPhase =
  | "idle"
  | "writing"
  | "deploying"
  | "complete"
  | "failed"
  | "error";

export interface DeploymentState {
  phase: DeploymentPhase;
  progress: DeploymentProgress | null;
  logs: string;
  startTime: Date | null;
  error: DeploymentError | null;
  taskId: string | null;
}

export interface UseDeploymentReturn {
  state: DeploymentState;
  start: (config: unknown) => Promise<void>;
}

const INITIAL_STATE: DeploymentState = {
  phase: "idle",
  progress: null,
  logs: "",
  startTime: null,
  error: null,
  taskId: null,
};

function stripAnsi(text: string): string {
  // biome-ignore lint/suspicious/noControlCharactersInRegex: ANSI escape codes are control characters by definition
  return text.replace(/\x1b\[[0-9;]*m/g, "").replace(/\r/g, "");
}

export function useDeployment(): UseDeploymentReturn {
  const api = useEnclaveApi();
  const tasksApi = useTasksApi();
  const [state, setState] = useState<DeploymentState>(INITIAL_STATE);
  const [useSSETransport, setUseSSETransport] = useState(true);
  const [sseFailed, setSseFailed] = useState(false);

  // --- SSE transport ---
  const sseUrl =
    state.taskId && useSSETransport && !sseFailed
      ? `/api/v1/tasks/${state.taskId}/stream`
      : null;

  const handleSSEEvent = useCallback((event: SSEEvent) => {
    try {
      const data = JSON.parse(event.data);
      switch (event.type) {
        case "status":
          setState((prev) => ({
            ...prev,
            phase: data.phase ?? data.status ?? prev.phase,
          }));
          break;
        case "progress":
          setState((prev) => ({
            ...prev,
            progress: {
              percentage: data.percentage ?? 0,
              currentTask: data.currentTask ?? "",
            },
          }));
          break;
        case "log":
          setState((prev) => ({
            ...prev,
            logs: prev.logs + stripAnsi(data.line ?? data.message ?? ""),
          }));
          break;
        case "done":
          setState((prev) => ({
            ...prev,
            phase: data.status === "failed" ? "failed" : "complete",
            progress: prev.progress
              ? { ...prev.progress, percentage: 100 }
              : { percentage: 100, currentTask: "" },
          }));
          break;
      }
    } catch {
      // Ignore JSON parse errors
    }
  }, []);

  const handleSSEError = useCallback(() => {
    // If SSE fails, fall back to polling
    setSseFailed(true);
    setUseSSETransport(false);
  }, []);

  useSSE(sseUrl, {
    onEvent: handleSSEEvent,
    onError: handleSSEError,
  });

  // --- Polling transport (fallback) ---
  const pollingEnabled =
    !!state.taskId && sseFailed && state.phase === "deploying";

  const fetchTask = useCallback(
    () =>
      state.taskId ? tasksApi.getTask(state.taskId) : Promise.resolve(null),
    [tasksApi, state.taskId],
  );

  const fetchProgress = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/deployment/progress");
      if (!res.ok) return null;
      return res.json();
    } catch {
      return null;
    }
  }, []);

  const fetchLogs = useCallback(
    () =>
      state.taskId ? tasksApi.getTaskLogs(state.taskId) : Promise.resolve(""),
    [tasksApi, state.taskId],
  );

  const { data: polledTask } = usePolling(fetchTask, 3000, pollingEnabled);
  const { data: polledProgress } = usePolling(
    fetchProgress,
    3000,
    pollingEnabled,
  );
  const { data: polledLogs } = usePolling(fetchLogs, 2000, pollingEnabled);

  // Apply polled data to state
  const prevPolledTaskRef = useRef(polledTask);
  useEffect(() => {
    if (!pollingEnabled || polledTask === prevPolledTaskRef.current) return;
    prevPolledTaskRef.current = polledTask;
    if (!polledTask) return;

    const isRunning = polledTask.status === "running";
    if (!isRunning) {
      setState((prev) => ({
        ...prev,
        phase: polledTask.status === "failed" ? "failed" : "complete",
      }));
    }
  }, [polledTask, pollingEnabled]);

  useEffect(() => {
    if (!pollingEnabled || !polledProgress) return;
    setState((prev) => ({
      ...prev,
      progress: {
        percentage: polledProgress.percentage ?? 0,
        currentTask: polledProgress.currentTask ?? "",
      },
    }));
  }, [polledProgress, pollingEnabled]);

  useEffect(() => {
    if (!pollingEnabled || polledLogs === null || polledLogs === undefined)
      return;
    setState((prev) => ({
      ...prev,
      logs: stripAnsi(polledLogs),
    }));
  }, [polledLogs, pollingEnabled]);

  // --- Mount reconnection ---
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await fetch("/api/v1/deployment");
        if (!res.ok || cancelled) return;
        const dep = await res.json();
        if (cancelled) return;
        if (dep.status !== "running" && dep.status !== "pending") return;
        if (!dep.id) return;
        setState((prev) => ({
          ...prev,
          taskId: dep.id,
          startTime: new Date(),
          phase: "deploying",
        }));
      } catch {
        /* no deployment */
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  // --- Start deployment ---
  const start = useCallback(
    async (config: unknown) => {
      setState({
        ...INITIAL_STATE,
        phase: "writing",
      });
      setSseFailed(false);
      setUseSSETransport(true);

      try {
        await api.writeConfig(config as Parameters<typeof api.writeConfig>[0]);
        setState((prev) => ({
          ...prev,
          phase: "deploying",
          startTime: new Date(),
        }));

        const run = await tasksApi.startDeploy();
        setState((prev) => ({
          ...prev,
          taskId: run.id,
        }));
      } catch (err: unknown) {
        const details: string[] = [];
        const resp =
          err && typeof err === "object" && "response" in err
            ? (err as { response: Response }).response
            : null;
        if (resp) {
          try {
            const body = await resp.json();
            if (body.errors && Array.isArray(body.errors)) {
              for (const e of body.errors) {
                details.push(e.message ?? String(e));
              }
            } else if (body.detail) {
              details.push(body.detail);
            }
          } catch {
            /* response not JSON */
          }
        }
        const message =
          details.length > 0
            ? "Configuration validation failed"
            : err instanceof Error
              ? err.message
              : "Failed to start deployment";
        setState({
          ...INITIAL_STATE,
          phase: "error",
          error: { message, details },
        });
      }
    },
    [api, tasksApi],
  );

  return { state, start };
}
