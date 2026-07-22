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
import { useCallback } from "react";
import { usePluginSchema } from "../../api/usePluginSchema.ts";
import { type HostEntry, HostEntryCard } from "../components/HostEntryCard.tsx";
import { ToggleFieldGroup } from "../components/ToggleFieldGroup.tsx";
import { useWizard } from "../WizardContext.tsx";
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

const BCM_TOGGLE = "global.osacBcmEnabled";

const BCM_DETAIL_FIELDS = [
  "global.osacBcmApiUrl",
  "global.osacBcmClientCert",
  "global.osacBcmClientKey",
  "global.osacBcmValidateCerts",
  "global.osacBcmDisableBmcCertVerification",
];

function buildMergedSchema(
  pluginSchema: Record<string, unknown>,
  fieldPaths: string[],
): unknown {
  const root: Record<string, unknown> = {
    type: "object",
    properties: {},
  };

  for (const path of fieldPaths) {
    const segments = path.split(".");
    let current = root;

    for (let i = 0; i < segments.length - 1; i++) {
      const seg = segments[i];
      const props = current.properties as Record<
        string,
        Record<string, unknown>
      >;
      if (!props[seg]) {
        props[seg] = { type: "object", properties: {} };
      }
      current = props[seg];
    }

    const leafName = segments[segments.length - 1];
    const props = current.properties as Record<string, unknown>;
    const pluginProps = pluginSchema.properties as
      | Record<string, unknown>
      | undefined;
    if (pluginProps?.[leafName]) {
      props[leafName] = pluginProps[leafName];
    }
  }

  return root;
}

export const CaasStep: React.FC = () => {
  const { state, dispatch } = useWizard();
  const { schema: pluginSchema } = usePluginSchema("osac");

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

  const onChange = useCallback(
    (path: string, value: unknown) =>
      dispatch({ type: "SET_FIELD", path, value }),
    [dispatch],
  );

  const allBcmFields = [BCM_TOGGLE, ...BCM_DETAIL_FIELDS];
  const bcmSchema = pluginSchema
    ? buildMergedSchema(pluginSchema as Record<string, unknown>, allBcmFields)
    : null;

  return (
    <Form>
      <Title headingLevel="h2" size="xl">
        CaaS
      </Title>
      <Content component="p">
        Register bare metal machines that form the resource pool for CaaS. The
        platform draws from this pool when provisioning managed clusters. Each
        machine is enrolled via its BMC (Redfish/IPMI) interface. You can add
        machines later through the management interface.
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
            onClick={() =>
              setDiscoveryHosts([...discoveryHosts, { ...EMPTY_HOST }])
            }
          >
            Add host
          </Button>
        </FlexItem>
      </Flex>

      {discoveryHosts.length === 0 && (
        <p className={stepStyles.emptyHint}>
          No hosts registered yet. Click &quot;Add host&quot; to register bare
          metal hosts, or skip this step to add them later.
        </p>
      )}

      <Flex
        direction={{ default: "column" }}
        gap={{ default: "gapMd" }}
        className={stepStyles.hostSection}
      >
        {discoveryHosts.map((host, i) => (
          <FlexItem key={`discovery-${i}`}>
            <Flex
              alignItems={{ default: "alignItemsFlexStart" }}
              gap={{ default: "gapSm" }}
            >
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
                    setDiscoveryHosts(
                      discoveryHosts.filter((_, idx) => idx !== i),
                    )
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

      <Title headingLevel="h2" size="xl" style={{ marginTop: 32 }}>
        BCM Inventory
      </Title>
      <Content component="p">
        Optionally enable NVIDIA Base Command Manager as an additional bare
        metal inventory source. BCM-discovered servers are imported alongside
        any hosts defined above.
      </Content>

      {bcmSchema && (
        <ToggleFieldGroup
          schema={bcmSchema}
          toggleField={BCM_TOGGLE}
          dependentFields={BCM_DETAIL_FIELDS}
          values={configData}
          onChange={onChange}
          showValidation={state.showValidation}
        />
      )}
    </Form>
  );
};
