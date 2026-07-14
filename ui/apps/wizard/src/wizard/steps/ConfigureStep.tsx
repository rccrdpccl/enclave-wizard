import { css } from "@emotion/css";
import { Title } from "@patternfly/react-core";
import type React from "react";
import type { ConfigSubStep } from "../hooks/useSubSteps.ts";
import { AAPStep } from "./AAPStep.tsx";
import { CaasStep } from "./CaasStep.tsx";
import { GpuAiStep } from "./GpuAiStep.tsx";
import { HubClusterStep } from "./HubClusterStep.tsx";
import { LandingZoneStep } from "./LandingZoneStep.tsx";
import { OsacStep } from "./OsacStep.tsx";
import { StorageStep } from "./StorageStep.tsx";
import { TrustManagerStep } from "./TrustManagerStep.tsx";

const configLayout = css`
  display: flex;
  gap: 2rem;
`;

const configNav = css`
  min-width: 200px;
  padding-top: 0.5rem;
`;

const configNavItem = css`
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.5rem 0;
  cursor: pointer;
  color: var(--pf-t--global--text--color--subtle);
  &:hover {
    color: var(--pf-t--global--text--color--regular);
  }
`;

const configNavItemActive = css`
  color: var(--pf-t--global--text--color--regular);
  font-weight: 600;
`;

const configNavCircle = css`
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid var(--pf-t--global--border--color--default);
  flex-shrink: 0;
  margin-top: 2px;
`;

const configNavCircleActive = css`
  border-color: var(--pf-t--global--icon--color--brand--default);
  background-color: var(--pf-t--global--icon--color--brand--default);
`;

const configNavLine = css`
  width: 2px;
  height: 24px;
  background-color: var(--pf-t--global--border--color--default);
  margin-left: 9px;
`;

const configContent = css`
  flex: 1;
  min-width: 0;
`;

function SubStepContent({
  subStepId,
}: {
  subStepId: string;
}): React.ReactElement {
  switch (subStepId) {
    case "landing-zone":
      return <LandingZoneStep />;
    case "storage":
      return <StorageStep />;
    case "hub-cluster":
      return <HubClusterStep />;
    case "osac":
      return <OsacStep />;
    case "gpu-ai":
      return <GpuAiStep />;
    case "aap":
      return <AAPStep />;
    case "trust-manager":
      return <TrustManagerStep />;
    case "caas":
      return <CaasStep />;
    default:
      return <div>Unknown section</div>;
  }
}

export const ConfigureStep: React.FC<{
  subSteps: ConfigSubStep[];
  activeSubStep: number;
  onSubStepChange: (index: number) => void;
}> = ({ subSteps, activeSubStep, onSubStepChange }) => {
  const currentSub = subSteps[activeSubStep];

  return (
    <div>
      <Title headingLevel="h2" size="xl" style={{ marginBottom: "0.25rem" }}>
        Configure your deployment
      </Title>
      <p
        style={{
          color: "var(--pf-t--global--text--color--subtle)",
          marginBottom: "1.5rem",
        }}
      >
        Answer a few questions to set up your chosen services.
      </p>

      <div className={configLayout}>
        <nav className={configNav}>
          {subSteps.map((sub, i) => (
            <div key={sub.id}>
              {/* biome-ignore lint/a11y/useSemanticElements: custom nav styling requires div with role=button */}
              <div
                className={`${configNavItem} ${i === activeSubStep ? configNavItemActive : ""}`}
                onClick={() => onSubStepChange(i)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => e.key === "Enter" && onSubStepChange(i)}
              >
                <div
                  className={`${configNavCircle} ${i === activeSubStep ? configNavCircleActive : ""}`}
                />
                <span>{sub.label}</span>
              </div>
              {i < subSteps.length - 1 && <div className={configNavLine} />}
            </div>
          ))}
        </nav>

        <div className={configContent}>
          <p
            style={{
              color: "var(--pf-t--global--text--color--subtle)",
              marginBottom: "1rem",
              fontSize: "0.875rem",
            }}
          >
            Step {activeSubStep + 1} of {subSteps.length} &middot;{" "}
            {currentSub?.label}
          </p>
          <SubStepContent subStepId={currentSub?.id ?? ""} />
        </div>
      </div>
    </div>
  );
};
