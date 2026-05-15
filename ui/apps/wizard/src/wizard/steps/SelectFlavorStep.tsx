import { Alert, Content, Flex, FlexItem, Title } from "@patternfly/react-core";
import { ClusterIcon, CpuIcon, BrainIcon } from "@patternfly/react-icons";
import type React from "react";
import type { ReactNode } from "react";
import { FlavorCard } from "../components/FlavorCard.tsx";
import { FLAVORS, type FlavorId } from "../flavors.ts";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles } from "./stepStyles.ts";

const FLAVOR_ICONS: Record<FlavorId, ReactNode> = {
  cluster: <ClusterIcon />,
  "nvidia-gpu": <CpuIcon />,
  "openshift-ai": <BrainIcon />,
};

export const SelectFlavorStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const autoSelected = state.selectedFlavors.has("openshift-ai") && !state.selectedFlavors.has("nvidia-gpu");

  return (
    <Flex direction={{ default: "column" }} gap={{ default: "gapLg" }}>
      <FlexItem>
        <Title headingLevel="h2" size="xl">
          Select your sovereign cloud setup
        </Title>
        <Content component="p" className={stepStyles.subtitle}>
          Choose additional services to deploy, or skip this step to set up
          only the landing zone and hub cluster.
        </Content>
      </FlexItem>
      <FlexItem>
        <Flex gap={{ default: "gapMd" }} flexWrap={{ default: "wrap" }}>
          {FLAVORS.map((flavor) => {
            const isSelected = state.selectedFlavors.has(flavor.id);
            const isAutoSelected =
              flavor.id === "nvidia-gpu" &&
              !isSelected &&
              state.selectedFlavors.has("openshift-ai");
            return (
              <FlexItem key={flavor.id} style={{ minWidth: 280, flex: 1 }}>
                <FlavorCard
                  title={flavor.title}
                  description={
                    isAutoSelected
                      ? `${flavor.description} (auto-selected by OpenShift AI)`
                      : flavor.description
                  }
                  icon={FLAVOR_ICONS[flavor.id]}
                  isSelected={isSelected || isAutoSelected}
                  onSelect={() =>
                    dispatch({ type: "TOGGLE_FLAVOR", flavor: flavor.id })
                  }
                />
              </FlexItem>
            );
          })}
        </Flex>
      </FlexItem>
      {state.selectedFlavors.has("openshift-ai") && (
        <FlexItem>
          <Alert variant="info" title="NVIDIA GPU Support auto-selected" isInline>
            OpenShift AI requires the NVIDIA GPU Operator.
          </Alert>
        </FlexItem>
      )}
    </Flex>
  );
};
