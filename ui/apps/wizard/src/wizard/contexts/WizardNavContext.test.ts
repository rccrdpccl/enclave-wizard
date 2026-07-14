import { describe, expect, it } from "vitest";
import {
  initialWizardNavState,
  wizardNavReducer,
} from "./WizardNavContext.tsx";

describe("wizardNavReducer", () => {
  it("sets the current step", () => {
    const state = wizardNavReducer(initialWizardNavState, {
      type: "SET_STEP",
      step: 3,
    });
    expect(state.currentStep).toBe(3);
  });

  it("sets the active sub-step", () => {
    const state = wizardNavReducer(initialWizardNavState, {
      type: "SET_SUB_STEP",
      subStep: 2,
    });
    expect(state.activeSubStep).toBe(2);
  });

  it("sets validation errors", () => {
    const errors = [
      { path: "global.baseDomain", label: "Base Domain", message: "Required" },
    ];
    const state = wizardNavReducer(initialWizardNavState, {
      type: "SET_VALIDATION_ERRORS",
      errors,
    });
    expect(state.validationErrors).toEqual(errors);
  });

  it("clears validation errors", () => {
    const withErrors = wizardNavReducer(initialWizardNavState, {
      type: "SET_VALIDATION_ERRORS",
      errors: [{ path: "x", label: "X", message: "err" }],
    });
    const cleared = wizardNavReducer(withErrors, {
      type: "SET_VALIDATION_ERRORS",
      errors: [],
    });
    expect(cleared.validationErrors).toEqual([]);
  });

  it("sets showValidation to true", () => {
    const state = wizardNavReducer(initialWizardNavState, {
      type: "SET_SHOW_VALIDATION",
      show: true,
    });
    expect(state.showValidation).toBe(true);
  });

  it("sets showValidation to false", () => {
    const withShow = wizardNavReducer(initialWizardNavState, {
      type: "SET_SHOW_VALIDATION",
      show: true,
    });
    const hidden = wizardNavReducer(withShow, {
      type: "SET_SHOW_VALIDATION",
      show: false,
    });
    expect(hidden.showValidation).toBe(false);
  });

  it("preserves other state when setting step", () => {
    const withErrors = wizardNavReducer(initialWizardNavState, {
      type: "SET_VALIDATION_ERRORS",
      errors: [{ path: "a", label: "A", message: "err" }],
    });
    const stepped = wizardNavReducer(withErrors, {
      type: "SET_STEP",
      step: 2,
    });
    expect(stepped.currentStep).toBe(2);
    expect(stepped.validationErrors).toHaveLength(1);
  });

  it("returns unchanged state for unknown action", () => {
    const state = wizardNavReducer(initialWizardNavState, {
      type: "UNKNOWN",
    } as never);
    expect(state).toBe(initialWizardNavState);
  });
});
