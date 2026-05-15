import type { EnclaveConfig } from "@enclave-wizard-ui/api-client";
import type { WizardState } from "./WizardContext.ts";
import { FLAVORS } from "./flavors.ts";

export function buildFinalConfig(state: WizardState): EnclaveConfig {
  const globalData = (state.configData.global ?? {}) as Record<string, unknown>;
  const configPlugins = Array.isArray(globalData.enabledPlugins)
    ? (globalData.enabledPlugins as string[])
    : ["lvms"];
  const flavorPlugins = FLAVORS
    .filter((f) => state.selectedFlavors.has(f.id))
    .flatMap((f) => f.plugins);
  const enabledPlugins = [...new Set([...configPlugins, ...flavorPlugins])];

  return {
    ...state.configData,
    global: {
      ...globalData,
      workingDir: "/home/enclave",
      disconnected: true,
      enabledPlugins,
    },
    certificates: state.configData.certificates ?? {},
    cloudInfra: state.configData.cloudInfra ?? { discovery_hosts: [] },
  } as unknown as EnclaveConfig;
}
