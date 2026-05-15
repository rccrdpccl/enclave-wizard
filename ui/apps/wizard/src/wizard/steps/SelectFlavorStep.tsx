import { Content, Flex, FlexItem, Title } from "@patternfly/react-core";
import { ClusterIcon, CpuIcon } from "@patternfly/react-icons";
import type React from "react";
import type { ReactNode } from "react";
import { FlavorCard } from "../components/FlavorCard.tsx";
import { FLAVORS, type FlavorId } from "../flavors.ts";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles } from "./stepStyles.ts";

const FLAVOR_ICONS: Record<FlavorId, ReactNode> = {
  cluster: <ClusterIcon />,
  "gpu-ai": <CpuIcon />,
};

export const SelectFlavorStep: React.FC = () => {
  const { state, dispatch } = useWizard();

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
          {FLAVORS.map((flavor) => (
            <FlexItem key={flavor.id} style={{ minWidth: 280, flex: 1 }}>
              <FlavorCard
                title={flavor.title}
                description={flavor.description}
                icon={FLAVOR_ICONS[flavor.id]}
                isSelected={state.selectedFlavors.has(flavor.id)}
                onSelect={() =>
                  dispatch({ type: "TOGGLE_FLAVOR", flavor: flavor.id })
                }
              />
            </FlexItem>
          ))}
        </Flex>
      </FlexItem>
    </Flex>
  );
};
