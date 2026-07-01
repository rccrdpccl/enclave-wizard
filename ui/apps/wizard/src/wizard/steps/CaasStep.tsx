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
        Bare Metal Hosts
      </Title>
      <Content component="p">
        Register bare metal machines that form the resource pool for CaaS.
        The platform draws from this pool when provisioning managed clusters.
        Each machine is enrolled via its BMC (Redfish/IPMI) interface.
        You can add machines later through the management interface.
      </Content>

      <Flex
        justifyContent={{ default: "justifyContentSpaceBetween" }}
        alignItems={{ default: "alignItemsCenter" }}
        className={stepStyles.sectionTitle}
      >
        <FlexItem>
          <Title headingLevel="h3" size="lg">
            Available Hosts ({discoveryHosts.length})
          </Title>
        </FlexItem>
        <FlexItem>
          <Button
            variant="link"
            icon={<PlusCircleIcon />}
            onClick={() => setDiscoveryHosts([...discoveryHosts, { ...EMPTY_HOST }])}
          >
            Add host
          </Button>
        </FlexItem>
      </Flex>

      {discoveryHosts.length === 0 && (
        <p className={stepStyles.emptyHint}>
          No hosts registered yet. Click &quot;Add host&quot; to register
          bare metal hosts, or skip this step to add them later.
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
                  label="Host"
                />
              </FlexItem>
              <FlexItem>
                <Button
                  variant="plain"
                  aria-label={`Remove host ${i + 1}`}
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
