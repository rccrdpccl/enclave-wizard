import {
  Checkbox,
  Content,
  Flex,
  FlexItem,
  Title,
} from "@patternfly/react-core";
import type React from "react";
import { getExperiencePlugins } from "../experiences.ts";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles } from "./stepStyles.ts";

const AI_PLUGINS = getExperiencePlugins("aiaas");

export const GpuAiStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const globalData = ((state.configData as Record<string, unknown>).global ??
    {}) as Record<string, unknown>;
  const enabledPlugins = Array.isArray(globalData.enabled_plugins)
    ? (globalData.enabled_plugins as string[])
    : [];

  const isEnabled = AI_PLUGINS.every((p) => enabledPlugins.includes(p));

  const toggle = (checked: boolean) => {
    const current = new Set(enabledPlugins);
    for (const p of AI_PLUGINS) {
      if (checked) current.add(p);
      else current.delete(p);
    }
    dispatch({
      type: "SET_FIELD",
      path: "global.enabled_plugins",
      value: [...current],
    });
  };

  return (
    <Flex direction={{ default: "column" }} gap={{ default: "gapLg" }}>
      <FlexItem>
        <Title headingLevel="h3" size="lg">
          GPU & AI Workloads
        </Title>
        <Content component="p" className={stepStyles.subtitle}>
          Enable GPU-accelerated compute and AI/ML platform capabilities on
          your virtual machine infrastructure.
        </Content>
      </FlexItem>

      <FlexItem>
        <Checkbox
          id="enable-gpu-ai"
          label="Enable GPU & AI support"
          description="Installs the NVIDIA GPU Operator and OpenShift AI (RHOAI) for running ML/AI workloads on GPU-equipped nodes."
          isChecked={isEnabled}
          onChange={(_e, checked) => toggle(checked)}
        />
      </FlexItem>

      {isEnabled && (
        <FlexItem>
          <Content component="p" className={stepStyles.subtitle}>
            The following will be deployed:
          </Content>
          <Content component="ul" style={{ marginTop: "0.5rem" }}>
            <Content component="li">
              <strong>NVIDIA GPU Operator</strong> — automatic driver and
              runtime provisioning for GPU nodes
            </Content>
            <Content component="li">
              <strong>OpenShift AI</strong> — model serving, notebooks, and
              data science pipelines
            </Content>
          </Content>
        </FlexItem>
      )}
    </Flex>
  );
};
