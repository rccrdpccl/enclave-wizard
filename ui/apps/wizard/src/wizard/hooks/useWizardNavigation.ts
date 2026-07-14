import { useCallback } from "react";
import { useWizardNav } from "../contexts/WizardNavContext.tsx";
import { useStepValidation } from "./useStepValidation.ts";
import type { ConfigSubStep } from "./useSubSteps.ts";

const TOP_STEPS_COUNT = 5;
const CONFIGURE_INDEX = 2;

export function useWizardNavigation(
  configSubSteps: ConfigSubStep[],
  currentSubStepId: string | undefined,
) {
  const { state: nav, dispatch: navDispatch } = useWizardNav();
  const validateCurrentSubStep = useStepValidation(currentSubStepId);

  const isConfigure = nav.currentStep === CONFIGURE_INDEX;
  const isFirstSubStep = nav.activeSubStep === 0;
  const isLastSubStep = nav.activeSubStep === configSubSteps.length - 1;

  const goBack = useCallback(() => {
    navDispatch({ type: "SET_VALIDATION_ERRORS", errors: [] });
    navDispatch({ type: "SET_SHOW_VALIDATION", show: false });

    if (isConfigure && !isFirstSubStep) {
      navDispatch({ type: "SET_SUB_STEP", subStep: nav.activeSubStep - 1 });
      return;
    }

    if (isConfigure && isFirstSubStep) {
      navDispatch({ type: "SET_SUB_STEP", subStep: 0 });
    }

    navDispatch({
      type: "SET_STEP",
      step: Math.max(0, nav.currentStep - 1),
    });
  }, [
    isConfigure,
    isFirstSubStep,
    nav.currentStep,
    nav.activeSubStep,
    navDispatch,
  ]);

  const goNext = useCallback(() => {
    if (isConfigure) {
      const errors = validateCurrentSubStep();
      if (errors.length > 0) {
        navDispatch({ type: "SET_VALIDATION_ERRORS", errors });
        navDispatch({ type: "SET_SHOW_VALIDATION", show: true });
        return;
      }
      navDispatch({ type: "SET_VALIDATION_ERRORS", errors: [] });
      navDispatch({ type: "SET_SHOW_VALIDATION", show: false });

      if (!isLastSubStep) {
        navDispatch({ type: "SET_SUB_STEP", subStep: nav.activeSubStep + 1 });
        return;
      }
    }

    navDispatch({ type: "SET_VALIDATION_ERRORS", errors: [] });
    navDispatch({ type: "SET_SHOW_VALIDATION", show: false });
    navDispatch({
      type: "SET_STEP",
      step: Math.min(TOP_STEPS_COUNT - 1, nav.currentStep + 1),
    });

    if (nav.currentStep + 1 === CONFIGURE_INDEX) {
      navDispatch({ type: "SET_SUB_STEP", subStep: 0 });
    }
  }, [
    isConfigure,
    isLastSubStep,
    nav.currentStep,
    nav.activeSubStep,
    navDispatch,
    validateCurrentSubStep,
  ]);

  return { goBack, goNext };
}
