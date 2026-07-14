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
import type React from "react";
import { useEffect, useRef, useState } from "react";
import { useDeployment } from "../../api/useDeployment.ts";
import { tasksStyles as styles } from "../../tasks/tasksStyles.ts";
import { buildFinalConfig } from "../buildFinalConfig.ts";
import { useWizard } from "../WizardContext.tsx";

function formatDuration(start?: Date | null): string {
  if (!start) return "—";
  const elapsed = Date.now() - start.getTime();
  const seconds = Math.floor(elapsed / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

export const DeployStep: React.FC = () => {
  const { state: wizardState } = useWizard();
  const { state: deployment, start } = useDeployment();

  const handleDeploy = async () => {
    await start(buildFinalConfig(wizardState));
  };

  const isRunning = deployment.phase === "deploying";
  const isComplete =
    deployment.phase === "complete" || deployment.phase === "failed";
  const isFailed = deployment.phase === "failed";

  const logsEndRef = useRef<HTMLDivElement>(null);
  const [follow, setFollow] = useState(true);
  const [logsOpen, setLogsOpen] = useState(false);

  // biome-ignore lint/correctness/useExhaustiveDependencies: deployment.logs triggers scroll-to-bottom
  useEffect(() => {
    if (follow && isRunning && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [deployment.logs, follow, isRunning]);

  // Duration ticker
  const [, setTick] = useState(0);
  useEffect(() => {
    if (isComplete) return;
    if (!deployment.taskId) return;
    const id = setInterval(() => setTick((t) => t + 1), 1000);
    return () => clearInterval(id);
  }, [isComplete, deployment.taskId]);

  if (deployment.phase === "writing") {
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

  if (deployment.phase === "idle" || deployment.phase === "error") {
    return (
      <div>
        <Title headingLevel="h2" size="xl">
          Deploy
        </Title>
        {deployment.phase === "error" && deployment.error && (
          <Alert
            variant="danger"
            title={deployment.error.message}
            isInline
            style={{ margin: "1rem 0" }}
          >
            {deployment.error.details.length > 0 && (
              <ul style={{ margin: 0, paddingLeft: "1.25rem" }}>
                {deployment.error.details.map((d) => (
                  <li key={d}>{d}</li>
                ))}
              </ul>
            )}
          </Alert>
        )}
        <p style={{ margin: "1rem 0" }}>
          Write the configuration and start a full deployment.
        </p>
        <Button variant="primary" size="lg" onClick={handleDeploy}>
          Deploy
        </Button>
      </div>
    );
  }

  // Deploying / complete / failed
  const progressVariant = isFailed
    ? ProgressVariant.danger
    : isComplete
      ? ProgressVariant.success
      : undefined;

  return (
    <Stack hasGutter>
      <StackItem>
        <Split hasGutter>
          <SplitItem isFilled>
            <Title headingLevel="h2" size="xl">
              Deployment
            </Title>
          </SplitItem>
          <SplitItem>
            <span
              style={{
                color: "var(--pf-t--global--text--color--subtle)",
                fontSize: "0.875rem",
              }}
            >
              {formatDuration(deployment.startTime)}
              {isRunning && (
                <Spinner size="sm" style={{ marginLeft: "0.5rem" }} />
              )}
            </span>
          </SplitItem>
        </Split>
      </StackItem>

      <StackItem>
        <Progress
          value={isComplete ? 100 : (deployment.progress?.percentage ?? 0)}
          title={
            deployment.progress?.currentTask || (isRunning ? "Starting..." : "")
          }
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
                  <Button
                    variant="link"
                    isInline
                    onClick={() => setFollow((f) => !f)}
                  >
                    {follow ? "Unfollow" : "Follow"}
                  </Button>
                </div>
              )}
              <div className={styles.logsContainer}>
                {deployment.logs ? (
                  <pre
                    style={{
                      margin: 0,
                      whiteSpace: "pre-wrap",
                      wordBreak: "break-word",
                    }}
                  >
                    {deployment.logs}
                  </pre>
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
          <Alert
            variant="success"
            title="Deployment completed successfully"
            isInline
          />
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
