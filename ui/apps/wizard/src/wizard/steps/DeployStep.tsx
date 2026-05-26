import {
  Alert,
  Button,
  DescriptionList,
  DescriptionListDescription,
  DescriptionListGroup,
  DescriptionListTerm,
  EmptyState,
  EmptyStateBody,
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
import type { TaskRun } from "@enclave-wizard-ui/api-client";
import { useEnclaveApi } from "../../api/useEnclaveApi.ts";
import { useTasksApi } from "../../api/useTasksApi.ts";
import { useWizard } from "../WizardContext.tsx";
import { buildFinalConfig } from "../buildFinalConfig.ts";
import { usePolling } from "../../tasks/hooks/usePolling.ts";
import { TaskStatusLabel } from "../../tasks/components/TaskStatusLabel.tsx";
import { tasksStyles as styles } from "../../tasks/tasksStyles.ts";

type DeployStatus = "idle" | "writing" | "deploying" | "error";

function formatDuration(start?: Date | null, end?: Date | null): string {
  if (!start) return "—";
  const elapsed = (end ?? new Date()).getTime() - start.getTime();
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
  const [taskId, setTaskId] = useState<string | null>(null);
  const [taskDone, setTaskDone] = useState(false);

  const handleDeploy = useCallback(async () => {
    setStatus("writing");
    setErrorMessage("");
    try {
      await api.writeConfig(buildFinalConfig(state));
      setStatus("deploying");
      const task = await tasksApi.startDeploy();
      setTaskId(task.id);
    } catch (err) {
      setStatus("error");
      setErrorMessage(
        err instanceof Error ? err.message : "Failed to start deployment",
      );
    }
  }, [api, tasksApi, state]);

  // Poll task status
  const fetchTask = useCallback(
    () => (taskId ? tasksApi.getTask(taskId) : Promise.resolve(null)),
    [tasksApi, taskId],
  );
  const { data: task } = usePolling(fetchTask, 3000, !!taskId && !taskDone);

  const isRunning = task?.status === "running";

  useEffect(() => {
    if (task != null && !isRunning) {
      setTaskDone(true);
    }
  }, [task, isRunning]);

  // Poll logs
  const fetchLogs = useCallback(
    () => (taskId ? tasksApi.getTaskLogs(taskId) : Promise.resolve("")),
    [tasksApi, taskId],
  );
  const { data: logs } = usePolling(fetchLogs, 2000, !!taskId && !taskDone);

  const ansiUp = useMemo(() => {
    const instance = new AnsiUp();
    instance.use_classes = false;
    return instance;
  }, []);

  const logsHtml = useMemo(() => {
    if (!logs) return "";
    return ansiUp.ansi_to_html(logs);
  }, [logs, ansiUp]);

  const logsEndRef = useRef<HTMLDivElement>(null);
  const [follow, setFollow] = useState(true);

  useEffect(() => {
    if (follow && isRunning && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [logsHtml, follow, isRunning]);

  if (status === "writing") {
    return (
      <EmptyState
        variant="lg"
        titleText="Writing configuration..."
        headingLevel="h2"
        icon={Spinner}
      >
        <EmptyStateBody>
          Writing config files before starting deployment.
        </EmptyStateBody>
      </EmptyState>
    );
  }

  if (status === "idle" || (status === "error" && !taskId)) {
    return (
      <div>
        <Title headingLevel="h2" size="xl">
          Deploy
        </Title>

        {status === "error" && (
          <Alert
            variant="danger"
            title="Deployment failed to start"
            isInline
            style={{ margin: "1rem 0" }}
          >
            {errorMessage}
          </Alert>
        )}

        <p style={{ margin: "1rem 0" }}>
          Write the configuration and start a full deployment (all 7 phases).
        </p>

        <Button variant="primary" size="lg" onClick={handleDeploy}>
          Deploy
        </Button>
      </div>
    );
  }

  // Deploying — show task output
  return (
    <Stack hasGutter>
      <StackItem>
        <Title headingLevel="h2" size="xl">
          Deployment
        </Title>
      </StackItem>

      {task && (
        <StackItem>
          <DescriptionList isHorizontal isCompact>
            <DescriptionListGroup>
              <DescriptionListTerm>Status</DescriptionListTerm>
              <DescriptionListDescription>
                <TaskStatusLabel status={task.status} />
              </DescriptionListDescription>
            </DescriptionListGroup>
            <DescriptionListGroup>
              <DescriptionListTerm>Duration</DescriptionListTerm>
              <DescriptionListDescription>
                {formatDuration(task.startedAt, task.endedAt)}
                {isRunning && (
                  <Spinner size="sm" style={{ marginLeft: "0.5rem" }} />
                )}
              </DescriptionListDescription>
            </DescriptionListGroup>
            {task.exitCode != null && (
              <DescriptionListGroup>
                <DescriptionListTerm>Exit Code</DescriptionListTerm>
                <DescriptionListDescription>
                  {task.exitCode}
                </DescriptionListDescription>
              </DescriptionListGroup>
            )}
          </DescriptionList>
        </StackItem>
      )}

      {task?.error && (
        <StackItem>
          <Alert variant="danger" title="Error" isInline>
            {task.error}
          </Alert>
        </StackItem>
      )}

      <StackItem>
        <Split hasGutter>
          <SplitItem isFilled>
            <Title headingLevel="h3" size="md">
              Output
            </Title>
          </SplitItem>
          {isRunning && (
            <SplitItem>
              <Button
                variant="link"
                isInline
                onClick={() => setFollow((f) => !f)}
              >
                {follow ? "Unfollow" : "Follow"}
              </Button>
            </SplitItem>
          )}
        </Split>
      </StackItem>

      <StackItem>
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
  );
};
