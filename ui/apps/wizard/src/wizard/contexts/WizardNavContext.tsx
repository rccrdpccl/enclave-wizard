import type React from "react";
import { createContext, useContext, useReducer } from "react";
import type { StepValidationError } from "../../schema/schemaUtils.ts";

export interface WizardNavState {
  currentStep: number;
  activeSubStep: number;
  validationErrors: StepValidationError[];
  showValidation: boolean;
}

export type WizardNavAction =
  | { type: "SET_STEP"; step: number }
  | { type: "SET_SUB_STEP"; subStep: number }
  | { type: "SET_VALIDATION_ERRORS"; errors: StepValidationError[] }
  | { type: "SET_SHOW_VALIDATION"; show: boolean };

export const initialWizardNavState: WizardNavState = {
  currentStep: 0,
  activeSubStep: 0,
  validationErrors: [],
  showValidation: false,
};

export function wizardNavReducer(
  state: WizardNavState,
  action: WizardNavAction,
): WizardNavState {
  switch (action.type) {
    case "SET_STEP":
      return { ...state, currentStep: action.step };
    case "SET_SUB_STEP":
      return { ...state, activeSubStep: action.subStep };
    case "SET_VALIDATION_ERRORS":
      return { ...state, validationErrors: action.errors };
    case "SET_SHOW_VALIDATION":
      return { ...state, showValidation: action.show };
    default:
      return state;
  }
}

interface WizardNavContextValue {
  state: WizardNavState;
  dispatch: React.Dispatch<WizardNavAction>;
}

const WizardNavContext = createContext<WizardNavContextValue | null>(null);

export const WizardNavProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [state, dispatch] = useReducer(wizardNavReducer, initialWizardNavState);
  return (
    <WizardNavContext.Provider value={{ state, dispatch }}>
      {children}
    </WizardNavContext.Provider>
  );
};

export function useWizardNav(): WizardNavContextValue {
  const context = useContext(WizardNavContext);
  if (context === null) {
    throw new Error("useWizardNav must be used within a WizardNavProvider.");
  }
  return context;
}
