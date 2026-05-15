import {
  Checkbox,
  Content,
  Form,
  Title,
} from "@patternfly/react-core";
import type React from "react";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles } from "./stepStyles.ts";

const GPU_AI_PLUGINS = [
  {
    id: "nvidia-gpu",
    label: "NVIDIA GPU Operator",
    description:
      "Manages NVIDIA GPU drivers, device plugins, and monitoring on OpenShift nodes.",
    required: true,
  },
  {
    id: "openshift-ai",
    label: "OpenShift AI (RHOAI)",
    description:
      "AI/ML platform for model training, serving, and inference workflows.",
    required: false,
  },
];

export const GpuAiStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const configData = state.configData as Record<string, unknown>;
  const globalData = (configData.global ?? {}) as Record<string, unknown>;
  const enabledPlugins: string[] = Array.isArray(globalData.enabled_plugins)
    ? (globalData.enabled_plugins as string[])
    : ["lvms"];

  const enabledSet = new Set(enabledPlugins);

  const togglePlugin = (pluginId: string) => {
    const next = new Set(enabledSet);
    if (next.has(pluginId)) {
      next.delete(pluginId);
      if (pluginId === "nvidia-gpu") {
        next.delete("openshift-ai");
      }
    } else {
      next.add(pluginId);
      if (pluginId === "openshift-ai") {
        next.add("nvidia-gpu");
      }
    }
    dispatch({
      type: "SET_FIELD",
      path: "global.enabled_plugins",
      value: [...next],
    });
  };

  return (
    <Form>
      <Title headingLevel="h2" size="xl">
        GPU & AI Configuration
      </Title>
      <Content component="p" className={stepStyles.subtitle}>
        Select which GPU and AI components to deploy on the hub cluster.
      </Content>

      <div style={{ marginTop: "1rem" }}>
        {GPU_AI_PLUGINS.map((plugin) => {
          const checked = enabledSet.has(plugin.id);
          const autoSelected =
            plugin.id === "nvidia-gpu" && enabledSet.has("openshift-ai");
          return (
            <Checkbox
              key={plugin.id}
              id={`gpu-ai-${plugin.id}`}
              label={plugin.label}
              description={
                autoSelected
                  ? `${plugin.description} (required by OpenShift AI)`
                  : plugin.description
              }
              isChecked={checked || autoSelected}
              isDisabled={autoSelected}
              onChange={() => togglePlugin(plugin.id)}
              style={{ marginBottom: "0.75rem" }}
            />
          );
        })}
      </div>
    </Form>
  );
};
