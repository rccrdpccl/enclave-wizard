import type React from "react";
import { useCallback } from "react";
import { useDeployment } from "../../api/useDeployment.ts";
import { buildFinalConfig } from "../buildFinalConfig.ts";
import { useWizard } from "../WizardContext.tsx";
import { DeployIdle } from "./deploy/DeployIdle.tsx";
import { DeployProgress } from "./deploy/DeployProgress.tsx";
import { DeployWriting } from "./deploy/DeployWriting.tsx";

export const DeployStep: React.FC = () => {
  const { state: wizardState } = useWizard();
  const { state: deployment, start } = useDeployment();

  const handleDeploy = useCallback(
    () => start(buildFinalConfig(wizardState)),
    [start, wizardState],
  );

  if (deployment.phase === "writing") {
    return <DeployWriting />;
  }

  if (
    deployment.phase === "idle" ||
    (deployment.phase === "error" && !deployment.taskId)
  ) {
    return <DeployIdle error={deployment.error} onDeploy={handleDeploy} />;
  }

  return <DeployProgress state={deployment} />;
};
