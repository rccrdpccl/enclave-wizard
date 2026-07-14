import type { EnclaveConfig } from "@enclave-wizard-ui/api-client";
import { EnclaveConfigFromJSON } from "@enclave-wizard-ui/api-client";
import type { ConfigData } from "./contexts/ConfigContext.tsx";

export function buildFinalConfig(state: {
  configData: ConfigData;
}): EnclaveConfig {
  const globalData = (state.configData.global ?? {}) as Record<string, unknown>;

  const raw = {
    ...state.configData,
    global: {
      ...globalData,
      workingDir: globalData.workingDir || "/home/enclave",
    },
    certificates: state.configData.certificates ?? {},
    cloudInfra: state.configData.cloudInfra ?? { discovery_hosts: [] },
  };

  return EnclaveConfigFromJSON(raw);
}
