import { useMemo } from "react";
import { useConfig } from "../contexts/ConfigContext.tsx";

export interface ConfigSubStep {
  id: string;
  label: string;
}

const BASE_CONFIG_SUBSTEPS: ConfigSubStep[] = [
  { id: "landing-zone", label: "Landing Zone" },
  { id: "storage", label: "Storage" },
  { id: "hub-cluster", label: "Hub Cluster" },
];

export function buildConfigSubSteps(
  selectedFlavors: Set<string>,
  enabledPlugins: string[],
): ConfigSubStep[] {
  const subs = [...BASE_CONFIG_SUBSTEPS];
  const hasOsac =
    selectedFlavors.has("caas") ||
    selectedFlavors.has("vmaas") ||
    selectedFlavors.has("bmaas");

  if (hasOsac) {
    subs.push({ id: "osac", label: "OSAC Platform" });
  }
  if (enabledPlugins.includes("aap") && !hasOsac) {
    subs.push({ id: "aap", label: "AAP Config" });
  }
  if (enabledPlugins.includes("trust-manager") && !hasOsac) {
    subs.push({ id: "trust-manager", label: "Trust Manager" });
  }
  if (selectedFlavors.has("vmaas")) {
    subs.push({ id: "gpu-ai", label: "Virtual Machines" });
  }
  if (selectedFlavors.has("caas")) {
    subs.push({ id: "caas", label: "Bare Metal Hosts" });
  }
  return subs;
}

export function useSubSteps(): ConfigSubStep[] {
  const { state } = useConfig();
  const globalData = (state.configData as Record<string, unknown>).global as
    | Record<string, unknown>
    | undefined;
  const enabledPlugins = Array.isArray(globalData?.enabled_plugins)
    ? (globalData.enabled_plugins as string[])
    : [];

  return useMemo(
    () => buildConfigSubSteps(state.selectedFlavors, enabledPlugins),
    [state.selectedFlavors, enabledPlugins],
  );
}
