import {
  Checkbox,
  Content,
  Flex,
  FlexItem,
  Title,
} from "@patternfly/react-core";
import {
  ClusterIcon,
  CubesIcon,
  ServerIcon,
} from "@patternfly/react-icons";
import type React from "react";
import type { ReactNode } from "react";
import { FlavorCard } from "../components/FlavorCard.tsx";
import { FLAVORS, type FlavorId } from "../flavors.ts";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles } from "./stepStyles.ts";
import { css } from "@emotion/css";

const FLAVOR_ICONS: Record<FlavorId, ReactNode> = {
  caas: <ClusterIcon />,
  vmaas: <CubesIcon />,
  bmaas: <ServerIcon />,
};

const addonArea = css`
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--pf-t--global--border--color--default);
`;

export const SelectFlavorStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const enabledPlugins = (() => {
    const g = (state.configData as Record<string, unknown>).global as
      | Record<string, unknown>
      | undefined;
    return Array.isArray(g?.enabled_plugins)
      ? (g.enabled_plugins as string[])
      : [];
  })();

  const toggleAddonPlugin = (plugins: string[], checked: boolean) => {
    const current = new Set(enabledPlugins);
    for (const p of plugins) {
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
        <Title headingLevel="h2" size="xl">
          Select your services
        </Title>
        <Content component="p" className={stepStyles.subtitle}>
          Choose which services to deploy on your sovereign cloud. You can
          combine multiple services. The landing zone and hub cluster are always
          included.
        </Content>
      </FlexItem>
      <FlexItem>
        <Flex gap={{ default: "gapMd" }} flexWrap={{ default: "wrap" }}>
          {FLAVORS.map((flavor) => {
            const isSelected = state.selectedFlavors.has(flavor.id);
            return (
              <FlexItem key={flavor.id} style={{ minWidth: 280, flex: 1 }}>
                <FlavorCard
                  title={flavor.title}
                  subtitle={flavor.subtitle}
                  description={flavor.description}
                  icon={FLAVOR_ICONS[flavor.id]}
                  isSelected={isSelected}
                  onSelect={() =>
                    dispatch({ type: "TOGGLE_FLAVOR", flavor: flavor.id })
                  }
                />
                {isSelected && flavor.addons && flavor.addons.length > 0 && (
                  <div className={addonArea}>
                    {flavor.addons.map((addon) => {
                      const isAddonEnabled = addon.plugins.every((p) =>
                        enabledPlugins.includes(p),
                      );
                      return (
                        <Checkbox
                          key={addon.id}
                          id={`addon-${flavor.id}-${addon.id}`}
                          label={addon.label}
                          description={addon.description}
                          isChecked={isAddonEnabled}
                          onChange={(_e, checked) =>
                            toggleAddonPlugin(addon.plugins, checked)
                          }
                        />
                      );
                    })}
                  </div>
                )}
              </FlexItem>
            );
          })}
        </Flex>
      </FlexItem>
    </Flex>
  );
};
