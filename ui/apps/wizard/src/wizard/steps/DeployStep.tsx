import {
  Alert,
  Button,
  EmptyState,
  EmptyStateBody,
  ExpandableSection,
  Progress,
  ProgressMeasureLocation,
  ProgressVariant,
  Spinner,
  Split,
  SplitItem,
  Stack,
  StackItem,
  Title,
} from "@patternfly/react-core";
import { AnsiUp } from "ansi_up";
import type React from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useEnclaveApi } from "../../api/useEnclaveApi.ts";
import { useTasksApi } from "../../api/useTasksApi.ts";
import { useWizard } from "../WizardContext.tsx";
import { buildFinalConfig } from "../buildFinalConfig.ts";
import { usePolling } from "../../tasks/hooks/usePolling.ts";
import { tasksStyles as styles } from "../../tasks/tasksStyles.ts";

type DeployStatus = "idle" | "writing" | "deploying" | "error";

function formatDuration(start?: Date | null): string {
  if (!start) return "—";
  const elapsed = (new Date()).getTime() - start.getTime();
  const seconds = Math.floor(elapsed / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

export const DeployStep: React.FC = () => {
  const { state } = useWizard();
  const api = useEnclaveApi();
  const tasksApi = useTasksApi();
  const [status, setStatus] = useState<DeployStatus>("idle");
  const [errorMessage, setErrorMessage] = useState("");
  const [errorDetails, setErrorDetails] = useState<string[]>([]);
  const [taskId, setTaskId] = useState<string | null>(null);
  const [taskDone, setTaskDone] = useState(false);
  const [startTime, setStartTime] = useState<Date | null>(null);

  // On mount, reconnect to a running deployment
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
        setTaskId(dep.id);
        setStartTime(new Date());
        setStatus("deploying");
      } catch { /* no deployment */ }
    })();
    return () => { cancelled = true; };
  }, []);

  const handleDeploy = useCallback(async () => {
    setStatus("writing");
    setErrorMessage("");
    setErrorDetails([]);
    try {
      await api.writeConfig(buildFinalConfig(state));
      setStatus("deploying");
      setStartTime(new Date());
      setTaskDone(false);
      const run = await tasksApi.startDeploy();
      setTaskId(run.id);
    } catch (err: unknown) {
      setStatus("error");
      const details: string[] = [];
      const resp = err && typeof err === "object" && "response" in err
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
        } catch { /* response not JSON */ }
      }
      setErrorDetails(details);
      setErrorMessage(
        details.length > 0
          ? "Configuration validation failed"
          : err instanceof Error ? err.message : "Failed to start deployment",
      );
    }
  }, [api, tasksApi, state]);

  // Poll task status (same pattern as TaskDetail in TasksPage)
  const fetchTask = useCallback(
    () => taskId ? tasksApi.getTask(taskId) : Promise.resolve(null),
    [tasksApi, taskId],
  );
  const { data: task } = usePolling(fetchTask, 3000, !!taskId && !taskDone);

  const isRunning = task?.status === "running";

  useEffect(() => {
    if (task && !isRunning) {
      setTaskDone(true);
    }
  }, [task, isRunning]);

  // Poll progress
  const fetchProgress = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/deployment/progress");
      if (!res.ok) return null;
      return res.json();
    } catch { return null; }
  }, []);
  const { data: progress } = usePolling(fetchProgress, 3000, !!taskId && !taskDone);

  // Poll logs
  const fetchLogs = useCallback(
    () => taskId ? tasksApi.getTaskLogs(taskId) : Promise.resolve(""),
    [tasksApi, taskId],
  );
  const { data: logs } = usePolling(fetchLogs, 2000, !!taskId && !taskDone);

  const ansiUp = useMemo(() => {
    const instance = new AnsiUp();
    instance.use_classes = false;
    return instance;
  }, []);

  const logsHtml = useMemo(() => logs ? ansiUp.ansi_to_html(logs) : "", [logs, ansiUp]);

  const logsEndRef = useRef<HTMLDivElement>(null);
  const [follow, setFollow] = useState(true);
  const [logsOpen, setLogsOpen] = useState(false);

  useEffect(() => {
    if (follow && isRunning && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [logsHtml, follow, isRunning]);

  // Duration ticker
  const [, setTick] = useState(0);
  useEffect(() => {
    if (!isRunning && taskDone) return;
    if (!taskId) return;
    const id = setInterval(() => setTick(t => t + 1), 1000);
    return () => clearInterval(id);
  }, [isRunning, taskDone, taskId]);

  if (status === "writing") {
    return (
      <EmptyState variant="lg" titleText="Writing configuration..." headingLevel="h2" icon={Spinner}>
        <EmptyStateBody>Writing config files before starting deployment.</EmptyStateBody>
      </EmptyState>
    );
  }

  if (status === "idle" || status === "error") {
    return (
      <div>
        <Title headingLevel="h2" size="xl">Deploy</Title>
        {status === "error" && (
          <Alert variant="danger" title={errorMessage} isInline style={{ margin: "1rem 0" }}>
            {errorDetails.length > 0 && (
              <ul style={{ margin: 0, paddingLeft: "1.25rem" }}>
                {errorDetails.map((d, i) => <li key={i}>{d}</li>)}
              </ul>
            )}
          </Alert>
        )}
        <p style={{ margin: "1rem 0" }}>
          Write the configuration and start a full deployment.
        </p>
        <Button variant="primary" size="lg" onClick={handleDeploy}>Deploy</Button>
      </div>
    );
  }

  // Deploying — show status + logs (same pattern as TaskDetail)
  const isComplete = task && !isRunning;
  const isFailed = task?.status === "failed";
  const progressVariant = isFailed ? ProgressVariant.danger : isComplete ? ProgressVariant.success : undefined;

  return (
    <Stack hasGutter>
      <StackItem>
        <Split hasGutter>
          <SplitItem isFilled>
            <Title headingLevel="h2" size="xl">Deployment</Title>
          </SplitItem>
          <SplitItem>
            <span style={{ color: "var(--pf-t--global--text--color--subtle)", fontSize: "0.875rem" }}>
              {formatDuration(startTime)}
              {isRunning && <Spinner size="sm" style={{ marginLeft: "0.5rem" }} />}
            </span>
          </SplitItem>
        </Split>
      </StackItem>

      <StackItem>
        <Progress
          value={isComplete ? 100 : (progress?.percentage ?? 0)}
          title={progress?.currentTask || (isRunning ? "Starting..." : "")}
          measureLocation={ProgressMeasureLocation.top}
          variant={progressVariant}
        />
      </StackItem>

      {/* Logs */}
      <StackItem>
        <ExpandableSection
          toggleText={logsOpen ? "Hide output" : "Show output"}
          isExpanded={logsOpen}
          onToggle={(_e, expanded) => setLogsOpen(expanded)}
        >
          <Stack hasGutter>
            <StackItem>
              {isRunning && (
                <div style={{ textAlign: "right", marginBottom: "0.5rem" }}>
                  <Button variant="link" isInline onClick={() => setFollow(f => !f)}>
                    {follow ? "Unfollow" : "Follow"}
                  </Button>
                </div>
              )}
              <div className={styles.logsContainer}>
                {logsHtml ? (
                  <div dangerouslySetInnerHTML={{ __html: logsHtml }} />
                ) : isRunning ? (
                  "Waiting for output..."
                ) : (
                  "No output available."
                )}
                <div ref={logsEndRef} />
              </div>
            </StackItem>
          </Stack>
        </ExpandableSection>
      </StackItem>

      {isComplete && !isFailed && (
        <StackItem>
          <Alert variant="success" title="Deployment completed successfully" isInline />
        </StackItem>
      )}

      {isFailed && (
        <StackItem>
          <Alert variant="danger" title="Deployment failed" isInline>
            Check the output above for details.
          </Alert>
        </StackItem>
      )}
    </Stack>
  );
};
