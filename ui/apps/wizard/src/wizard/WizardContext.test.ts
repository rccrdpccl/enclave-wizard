import { describe, expect, it } from "vitest";
import { initialWizardState, wizardReducer } from "./WizardContext.tsx";

describe("wizardReducer", () => {
  it("sets the current step", () => {
    const state = wizardReducer(initialWizardState, {
      type: "SET_STEP",
      step: 3,
    });
    expect(state.currentStep).toBe(3);
  });

  it("toggles a flavor on", () => {
    const state = wizardReducer(initialWizardState, {
      type: "TOGGLE_FLAVOR",
      flavor: "cluster",
    });
    expect(state.selectedFlavors.has("cluster")).toBe(true);
  });

  it("toggles a flavor off", () => {
    let state = wizardReducer(initialWizardState, {
      type: "TOGGLE_FLAVOR",
      flavor: "cluster",
    });
    state = wizardReducer(state, {
      type: "TOGGLE_FLAVOR",
      flavor: "cluster",
    });
    expect(state.selectedFlavors.has("cluster")).toBe(false);
  });

  it("auto-selects nvidia-gpu when openshift-ai is selected", () => {
    const state = wizardReducer(initialWizardState, {
      type: "TOGGLE_FLAVOR",
      flavor: "openshift-ai",
    });
    expect(state.selectedFlavors.has("openshift-ai")).toBe(true);
    expect(state.selectedFlavors.has("nvidia-gpu")).toBe(true);
  });

  it("deselects openshift-ai when nvidia-gpu is deselected", () => {
    let state = wizardReducer(initialWizardState, {
      type: "TOGGLE_FLAVOR",
      flavor: "openshift-ai",
    });
    state = wizardReducer(state, {
      type: "TOGGLE_FLAVOR",
      flavor: "nvidia-gpu",
    });
    expect(state.selectedFlavors.has("nvidia-gpu")).toBe(false);
    expect(state.selectedFlavors.has("openshift-ai")).toBe(false);
  });

  it("sets a top-level config field via dot path", () => {
    const state = wizardReducer(initialWizardState, {
      type: "SET_FIELD",
      path: "global.baseDomain",
      value: "enclave.example.com",
    });
    expect(state.configData.global?.baseDomain).toBe("enclave.example.com");
  });

  it("sets a nested config field via dot path", () => {
    const state = wizardReducer(initialWizardState, {
      type: "SET_FIELD",
      path: "global.quayBackendRGWConfiguration.hostname",
      value: "rgw.example.com",
    });
    expect(state.configData.global?.quayBackendRGWConfiguration?.hostname).toBe(
      "rgw.example.com",
    );
  });

  it("preserves existing fields when setting a new one", () => {
    let state = wizardReducer(initialWizardState, {
      type: "SET_FIELD",
      path: "global.baseDomain",
      value: "example.com",
    });
    state = wizardReducer(state, {
      type: "SET_FIELD",
      path: "global.clusterName",
      value: "mgmt",
    });
    expect(state.configData.global?.baseDomain).toBe("example.com");
    expect(state.configData.global?.clusterName).toBe("mgmt");
  });

  it("loads a full config", () => {
    const config = {
      global: { baseDomain: "test.com", clusterName: "test" },
      certificates: {},
      cloudInfra: { discovery_hosts: [] },
    };
    const state = wizardReducer(initialWizardState, {
      type: "LOAD_CONFIG",
      config: config as never,
    });
    expect(state.configData.global?.baseDomain).toBe("test.com");
  });

  it("sets validation errors", () => {
    const errors = [{ field: "global.baseDomain", message: "Required" }];
    const state = wizardReducer(initialWizardState, {
      type: "SET_VALIDATION_ERRORS",
      errors,
    });
    expect(state.validationErrors).toEqual(errors);
  });
});
