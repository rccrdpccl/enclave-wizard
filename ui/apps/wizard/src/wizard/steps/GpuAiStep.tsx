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

const GPU_PLUGINS = getExperiencePlugins("gpu");

export const GpuAiStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const globalData = ((state.configData as Record<string, unknown>).global ??
    {}) as Record<string, unknown>;
  const enabledPlugins = Array.isArray(globalData.enabled_plugins)
    ? (globalData.enabled_plugins as string[])
    : [];

  const isEnabled = GPU_PLUGINS.every((p) => enabledPlugins.includes(p));

  const toggle = (checked: boolean) => {
    const current = new Set(enabledPlugins);
    for (const p of GPU_PLUGINS) {
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
          GPU Compute
        </Title>
        <Content component="p" className={stepStyles.subtitle}>
          Enable GPU-accelerated compute on your virtual machine
          infrastructure.
        </Content>
      </FlexItem>

      <FlexItem>
        <Checkbox
          id="enable-gpu"
          label="Enable NVIDIA GPU support"
          description="Installs the NVIDIA GPU Operator for automatic driver and runtime provisioning on GPU-equipped nodes."
          isChecked={isEnabled}
          onChange={(_e, checked) => toggle(checked)}
        />
      </FlexItem>
    </Flex>
  );
};
