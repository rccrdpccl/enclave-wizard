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
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  InProgressIcon,
  PendingIcon,
} from "@patternfly/react-icons";
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

interface DeploymentPhase {
  name: string;
  taskId?: string;
  status: string;
}

interface Deployment {
  id: string;
  status: string;
  phases: DeploymentPhase[];
  totalTasks: number;
}

interface DeploymentProgress {
  completed: number;
  total: number;
  percentage: number;
  currentPhase: string;
  currentTask: string;
}

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

function phaseIcon(status: string) {
  switch (status) {
    case "successful": return <CheckCircleIcon color="var(--pf-t--global--color--status--success--default)" />;
    case "failed": return <ExclamationCircleIcon color="var(--pf-t--global--color--status--danger--default)" />;
    case "running": return <InProgressIcon color="var(--pf-t--global--color--status--info--default)" />;
    default: return <PendingIcon color="var(--pf-t--global--text--color--subtle)" />;
  }
}

function phaseLabel(name: string): string {
  if (name === "main") return "Base Platform";
  return name.split("-").map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(" ");
}

export const DeployStep: React.FC = () => {
  const { state } = useWizard();
  const api = useEnclaveApi();
  const tasksApi = useTasksApi();
  const [status, setStatus] = useState<DeployStatus>("idle");
  const [errorMessage, setErrorMessage] = useState("");
  const [errorDetails, setErrorDetails] = useState<string[]>([]);
  const [deploying, setDeploying] = useState(false);
  const [startTime, setStartTime] = useState<Date | null>(null);

  // Track which phase logs we've already accumulated
  const [accumulatedLogs, setAccumulatedLogs] = useState("");
  const [lastPhaseTaskId, setLastPhaseTaskId] = useState<string | null>(null);
  const logsReconstructedRef = useRef(false);

  // On mount, check for an existing deployment and reconnect
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await fetch("/api/v1/deployment");
        if (!res.ok || cancelled) return;
        const dep: Deployment = await res.json();
        if (cancelled) return;
        if (dep.status === "running" || dep.status === "pending") {
          setStatus("deploying");
          setDeploying(true);
          setStartTime(new Date());
        } else if (dep.status === "successful" || dep.status === "failed") {
          setStatus("deploying");
          setDeploying(false);
          setStartTime(new Date());
        }
      } catch { /* no deployment */ }
    })();
    return () => { cancelled = true; };
  }, []);

  const handleDeploy = useCallback(async () => {
    logsReconstructedRef.current = false;
    setStatus("writing");
    setErrorMessage("");
    setErrorDetails([]);
    setAccumulatedLogs("");
    setLastPhaseTaskId(null);
    try {
      await api.writeConfig(buildFinalConfig(state));
      setStatus("deploying");
      setDeploying(true);
      setStartTime(new Date());
      await tasksApi.startDeploy();
    } catch (err: unknown) {
      setStatus("error");
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

  // Poll deployment state (also fetch once for completed deployments on reconnect)
  const fetchDeployment = useCallback(async (): Promise<Deployment | null> => {
    try {
      const res = await fetch("/api/v1/deployment");
      if (!res.ok) return null;
      return res.json();
    } catch { return null; }
  }, []);
  const pollDeployment = status === "deploying";
  const { data: deployment } = usePolling(fetchDeployment, 3000, pollDeployment);

  // Poll progress
  const fetchProgress = useCallback(async (): Promise<DeploymentProgress | null> => {
    try {
      const res = await fetch("/api/v1/deployment/progress");
      if (!res.ok) return null;
      return res.json();
    } catch { return null; }
  }, []);
  const { data: progress } = usePolling(fetchProgress, 3000, pollDeployment);

  // Find the currently running phase's task ID for log polling
  const currentRunningTaskId = useMemo(() => {
    if (!deployment) return null;
    const running = deployment.phases.find(p => p.status === "running");
    return running?.taskId ?? null;
  }, [deployment]);

  // Poll logs for the current running task
  const fetchLogs = useCallback(
    () => currentRunningTaskId ? tasksApi.getTaskLogs(currentRunningTaskId) : Promise.resolve(""),
    [tasksApi, currentRunningTaskId],
  );
  const { data: currentLogs } = usePolling(fetchLogs, 2000, deploying && !!currentRunningTaskId);

  // Reconstruct logs from all phases on reconnect (page reload)
  useEffect(() => {
    if (!deployment || logsReconstructedRef.current) return;
    const hasTaskIds = deployment.phases.some(p => p.taskId);
    if (!hasTaskIds) return;
    logsReconstructedRef.current = true;

    (async () => {
      let logs = "";
      for (const phase of deployment.phases) {
        if (!phase.taskId) continue;
        try {
          const phaseLogs = await tasksApi.getTaskLogs(phase.taskId);
          if (phaseLogs) {
            logs += `\n=== ${phaseLabel(phase.name)} ===\n\n${phaseLogs}`;
            if (phase.status === "successful" || phase.status === "failed") {
              logs += `\n\n=== ${phaseLabel(phase.name)} complete ===\n\n`;
            }
          }
        } catch { /* skip */ }
      }
      setAccumulatedLogs(logs);
      const running = deployment.phases.find(p => p.status === "running");
      if (running?.taskId) {
        setLastPhaseTaskId(running.taskId);
      }
    })();
  }, [deployment, tasksApi]);

  // When phase changes, accumulate the previous phase's logs
  useEffect(() => {
    if (!currentRunningTaskId || currentRunningTaskId === lastPhaseTaskId) return;

    // A new phase started — fetch the completed phase's final logs
    if (lastPhaseTaskId) {
      tasksApi.getTaskLogs(lastPhaseTaskId).then(finalLogs => {
        const phaseName = deployment?.phases.find(p => p.taskId === lastPhaseTaskId)?.name ?? "";
        setAccumulatedLogs(prev =>
          prev + finalLogs + `\n\n=== ${phaseLabel(phaseName)} complete ===\n\n`
        );
      }).catch(() => {});
    }
    setLastPhaseTaskId(currentRunningTaskId);
  }, [currentRunningTaskId, lastPhaseTaskId, tasksApi, deployment]);

  // Detect deployment completion
  useEffect(() => {
    if (deployment && deployment.status !== "running" && deployment.status !== "pending") {
      // Fetch final phase logs
      if (lastPhaseTaskId) {
        tasksApi.getTaskLogs(lastPhaseTaskId).then(finalLogs => {
          setAccumulatedLogs(prev => prev + finalLogs);
        }).catch(() => {});
      }
      setDeploying(false);
    }
  }, [deployment, lastPhaseTaskId, tasksApi]);

  // Combine accumulated + current logs
  const fullLogs = useMemo(() => {
    if (!currentRunningTaskId) return accumulatedLogs;
    const phaseName = deployment?.phases.find(p => p.taskId === currentRunningTaskId)?.name ?? "";
    const separator = accumulatedLogs && !accumulatedLogs.endsWith(`=== ${phaseLabel(phaseName)} ===\n\n`)
      ? `\n=== ${phaseLabel(phaseName)} ===\n\n`
      : "";
    return accumulatedLogs + separator + (currentLogs ?? "");
  }, [accumulatedLogs, currentLogs, currentRunningTaskId, deployment]);

  const ansiUp = useMemo(() => {
    const instance = new AnsiUp();
    instance.use_classes = false;
    return instance;
  }, []);

  const logsHtml = useMemo(() => fullLogs ? ansiUp.ansi_to_html(fullLogs) : "", [fullLogs, ansiUp]);

  const logsEndRef = useRef<HTMLDivElement>(null);
  const [follow, setFollow] = useState(true);
  const [logsOpen, setLogsOpen] = useState(false);

  useEffect(() => {
    if (follow && deploying && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [logsHtml, follow, deploying]);

  // Duration ticker
  const [, setTick] = useState(0);
  useEffect(() => {
    if (!deploying) return;
    const id = setInterval(() => setTick(t => t + 1), 1000);
    return () => clearInterval(id);
  }, [deploying]);

  if (status === "writing") {
    return (
      <EmptyState variant="lg" titleText="Writing configuration..." headingLevel="h2" icon={Spinner}>
        <EmptyStateBody>Writing config files before starting deployment.</EmptyStateBody>
      </EmptyState>
    );
  }

  if (status === "idle" || (status === "error" && !deploying)) {
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

  // Deploying — show progress + phases + terminal
  const isComplete = deployment && deployment.status !== "running" && deployment.status !== "pending";
  const isFailed = deployment?.status === "failed";
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
              {deploying && <Spinner size="sm" style={{ marginLeft: "0.5rem" }} />}
            </span>
          </SplitItem>
        </Split>
      </StackItem>

      {/* Progress bar */}
      <StackItem>
        <Progress
          value={progress?.percentage ?? 0}
          title={progress?.currentTask ?? "Starting..."}
          measureLocation={ProgressMeasureLocation.top}
          variant={progressVariant}
          label={`${progress?.completed ?? 0} / ${progress?.total ?? "?"} phases`}
        />
      </StackItem>

      {/* Details: phases + logs */}
      <StackItem>
        <ExpandableSection
          toggleText={logsOpen ? "Hide details" : "Show details"}
          isExpanded={logsOpen}
          onToggle={(_e, expanded) => setLogsOpen(expanded)}
        >
          <Stack hasGutter>
            {/* Phase list */}
            {deployment && (
              <StackItem>
                {deployment.phases.map(phase => (
                  <div key={phase.name} style={{ display: "flex", alignItems: "center", gap: "0.5rem", padding: "0.25rem 0" }}>
                    {phaseIcon(phase.status)}
                    <span style={{ fontWeight: phase.status === "running" ? 600 : 400 }}>
                      {phaseLabel(phase.name)}
                    </span>
                  </div>
                ))}
              </StackItem>
            )}

            {/* Logs */}
            <StackItem>
              {deploying && (
                <div style={{ textAlign: "right", marginBottom: "0.5rem" }}>
                  <Button variant="link" isInline onClick={() => setFollow(f => !f)}>
                    {follow ? "Unfollow" : "Follow"}
                  </Button>
                </div>
              )}
              <div className={styles.logsContainer}>
                {logsHtml ? (
                  <div dangerouslySetInnerHTML={{ __html: logsHtml }} />
                ) : deploying ? (
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
