import {
  Button,
  Content,
  Flex,
  FlexItem,
  Form,
  Title,
} from "@patternfly/react-core";
import { MinusCircleIcon, PlusCircleIcon } from "@patternfly/react-icons";
import type React from "react";
import { useWizard } from "../WizardContext.tsx";
import { HostEntryCard, type HostEntry } from "../components/HostEntryCard.tsx";
import { stepStyles } from "./stepStyles.ts";

const EMPTY_HOST: HostEntry = {
  name: "",
  macAddress: "",
  ipAddress: "",
  redfish: "",
  redfishUser: "",
  redfishPassword: "",
  rootDisk: "",
};

export const CaasStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const configData = state.configData as Record<string, unknown>;
  const discoveryHosts: HostEntry[] = Array.isArray(
    (configData.cloudInfra as Record<string, unknown>)?.discovery_hosts,
  )
    ? ((configData.cloudInfra as Record<string, unknown>)
        .discovery_hosts as HostEntry[])
    : [];

  const setDiscoveryHosts = (hosts: HostEntry[]) =>
    dispatch({
      type: "SET_FIELD",
      path: "cloudInfra.discovery_hosts",
      value: hosts,
    });

  return (
    <Form>
      <Title headingLevel="h2" size="xl">
        Cluster as a Service
      </Title>
      <Content component="p">
        Add nodes to the discovery pool. These nodes will be available for
        provisioning managed clusters. You can add nodes later through the
        management interface.
      </Content>

      <Flex
        justifyContent={{ default: "justifyContentSpaceBetween" }}
        alignItems={{ default: "alignItemsCenter" }}
        className={stepStyles.sectionTitle}
      >
        <FlexItem>
          <Title headingLevel="h3" size="lg">
            Discovery Pool ({discoveryHosts.length} node{discoveryHosts.length !== 1 ? "s" : ""})
          </Title>
        </FlexItem>
        <FlexItem>
          <Button
            variant="link"
            icon={<PlusCircleIcon />}
            onClick={() => setDiscoveryHosts([...discoveryHosts, { ...EMPTY_HOST }])}
          >
            Add node
          </Button>
        </FlexItem>
      </Flex>

      {discoveryHosts.length === 0 && (
        <p className={stepStyles.emptyHint}>
          No nodes in the discovery pool. Click &quot;Add node&quot; to add
          nodes, or skip this step to add them later.
        </p>
      )}

      <Flex direction={{ default: "column" }} gap={{ default: "gapMd" }} className={stepStyles.hostSection}>
        {discoveryHosts.map((host, i) => (
          <FlexItem key={`discovery-${i}`}>
            <Flex alignItems={{ default: "alignItemsFlexStart" }} gap={{ default: "gapSm" }}>
              <FlexItem grow={{ default: "grow" }}>
                <HostEntryCard
                  index={i}
                  host={host}
                  onChange={(h) => {
                    const updated = [...discoveryHosts];
                    updated[i] = h;
                    setDiscoveryHosts(updated);
                  }}
                  label="Discovery node"
                />
              </FlexItem>
              <FlexItem>
                <Button
                  variant="plain"
                  aria-label={`Remove discovery node ${i + 1}`}
                  onClick={() =>
                    setDiscoveryHosts(discoveryHosts.filter((_, idx) => idx !== i))
                  }
                  className={stepStyles.removeButton}
                >
                  <MinusCircleIcon />
                </Button>
              </FlexItem>
            </Flex>
          </FlexItem>
        ))}
      </Flex>
    </Form>
  );
};
