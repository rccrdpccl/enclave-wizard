import { css } from "@emotion/css";
import {
  Divider,
  ProgressStep,
  ProgressStepper,
  Split,
  SplitItem,
} from "@patternfly/react-core";
import { ListIcon } from "@patternfly/react-icons";
import type React from "react";
import { Link } from "react-router-dom";
import { RedHatLogo } from "../../common/components/RedHatLogo.tsx";
import { wizardStyles as styles } from "../wizardStyles.ts";

const taskNavButton = css`
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 1rem;
  border: 1px solid var(--pf-t--global--border--color--default);
  border-radius: var(--pf-t--global--border--radius--small);
  color: var(--pf-t--global--text--color--regular);
  text-decoration: none;
  font-size: 0.875rem;
  &:hover {
    background-color: var(--pf-t--global--background--color--secondary--hover);
  }
`;

interface StepDef {
  id: string;
  label: string;
}

const TOP_STEPS: StepDef[] = [
  { id: "welcome", label: "Welcome" },
  { id: "flavor", label: "Select" },
  { id: "configure", label: "Configure" },
  { id: "review", label: "Review" },
  { id: "deploy", label: "Deploy" },
];

export const WizardHeader: React.FC<{
  currentStep: number;
}> = ({ currentStep }) => {
  const isWelcome = currentStep === 0;

  return (
    <header className={styles.header}>
      <div className={styles.headerInner}>
        <Split hasGutter>
          <SplitItem isFilled>
            <RedHatLogo width={240} />
          </SplitItem>
          <SplitItem>
            <Link to="/tasks" className={taskNavButton}>
              <ListIcon /> Tasks
            </Link>
          </SplitItem>
        </Split>
      </div>
      <Divider />
      {!isWelcome && (
        <div className={styles.headerInner}>
          <ProgressStepper aria-label="Wizard progress">
            {TOP_STEPS.map((step, i) => (
              <ProgressStep
                key={step.id}
                id={step.id}
                titleId={`step-title-${step.id}`}
                variant={
                  i < currentStep
                    ? "success"
                    : i === currentStep
                      ? "info"
                      : "pending"
                }
                isCurrent={i === currentStep}
                aria-label={step.label}
              >
                {step.label}
              </ProgressStep>
            ))}
          </ProgressStepper>
        </div>
      )}
    </header>
  );
};
