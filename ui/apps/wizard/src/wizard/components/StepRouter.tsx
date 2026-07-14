import type React from "react";
import type { ConfigSubStep } from "../hooks/useSubSteps.ts";
import { ConfigureStep } from "../steps/ConfigureStep.tsx";
import { DeployStep } from "../steps/DeployStep.tsx";
import { ReviewStep } from "../steps/ReviewStep.tsx";
import { SelectFlavorStep } from "../steps/SelectFlavorStep.tsx";
import { WelcomeStep } from "../steps/WelcomeStep.tsx";

const TOP_STEPS = ["welcome", "flavor", "configure", "review", "deploy"];

export const StepRouter: React.FC<{
  currentStep: number;
  configSubSteps: ConfigSubStep[];
  activeSubStep: number;
  onSubStepChange: (index: number) => void;
}> = ({ currentStep, configSubSteps, activeSubStep, onSubStepChange }) => {
  const stepId = TOP_STEPS[currentStep] ?? "welcome";

  switch (stepId) {
    case "welcome":
      return <WelcomeStep />;
    case "flavor":
      return <SelectFlavorStep />;
    case "configure":
      return (
        <ConfigureStep
          subSteps={configSubSteps}
          activeSubStep={activeSubStep}
          onSubStepChange={onSubStepChange}
        />
      );
    case "review":
      return <ReviewStep />;
    case "deploy":
      return <DeployStep />;
    default:
      return <div>Unknown step</div>;
  }
};
