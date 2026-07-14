import { describe, expect, it } from "vitest";
import {
  configReducer,
  initialConfigState,
} from "./contexts/ConfigContext.tsx";
import { EXPERIENCES } from "./experiences.ts";
import { FLAVORS } from "./flavors.ts";

describe("configReducer", () => {
  it("toggles a flavor on", () => {
    const state = configReducer(initialConfigState, {
      type: "TOGGLE_FLAVOR",
      flavor: "caas",
      experiences: EXPERIENCES,
      flavors: FLAVORS,
    });
    expect(state.selectedFlavors.has("caas")).toBe(true);
  });

  it("toggles a flavor off", () => {
    let state = configReducer(initialConfigState, {
      type: "TOGGLE_FLAVOR",
      flavor: "caas",
      experiences: EXPERIENCES,
      flavors: FLAVORS,
    });
    state = configReducer(state, {
      type: "TOGGLE_FLAVOR",
      flavor: "caas",
      experiences: EXPERIENCES,
      flavors: FLAVORS,
    });
    expect(state.selectedFlavors.has("caas")).toBe(false);
  });

  it("sets enabled_plugins when toggling a flavor on", () => {
    const state = configReducer(initialConfigState, {
      type: "TOGGLE_FLAVOR",
      flavor: "caas",
      experiences: EXPERIENCES,
      flavors: FLAVORS,
    });
    const plugins = (state.configData.global as Record<string, unknown>)
      ?.enabled_plugins as string[];
    expect(plugins).toContain("osac");
    expect(plugins).toContain("aap");
    expect(plugins).toContain("trust-manager");
  });

  it("sets osacProfile to development when multiple flavors selected", () => {
    let state = configReducer(initialConfigState, {
      type: "TOGGLE_FLAVOR",
      flavor: "caas",
      experiences: EXPERIENCES,
      flavors: FLAVORS,
    });
    state = configReducer(state, {
      type: "TOGGLE_FLAVOR",
      flavor: "vmaas",
      experiences: EXPERIENCES,
      flavors: FLAVORS,
    });
    const profile = (state.configData.global as Record<string, unknown>)
      ?.osacProfile;
    expect(profile).toBe("development");
  });

  it("sets a top-level config field via dot path", () => {
    const state = configReducer(initialConfigState, {
      type: "SET_FIELD",
      path: "global.baseDomain",
      value: "enclave.example.com",
    });
    expect(state.configData.global?.baseDomain).toBe("enclave.example.com");
  });

  it("sets a nested config field via dot path", () => {
    const state = configReducer(initialConfigState, {
      type: "SET_FIELD",
      path: "global.quayBackendRGWConfiguration.hostname",
      value: "rgw.example.com",
    });
    expect(
      (state.configData.global as Record<string, unknown>)
        ?.quayBackendRGWConfiguration,
    ).toEqual({ hostname: "rgw.example.com" });
  });

  it("preserves existing fields when setting a new one", () => {
    let state = configReducer(initialConfigState, {
      type: "SET_FIELD",
      path: "global.baseDomain",
      value: "example.com",
    });
    state = configReducer(state, {
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
    const state = configReducer(initialConfigState, {
      type: "LOAD_CONFIG",
      config: config as never,
    });
    expect(state.configData.global?.baseDomain).toBe("test.com");
  });
});
