import {
  Alert,
  Button,
  Content,
  Flex,
  FlexItem,
  Form,
  FormGroup,
  FormSelect,
  FormSelectOption,
  HelperText,
  HelperTextItem,
  TextInput,
  Title,
} from "@patternfly/react-core";
import { MinusCircleIcon, PlusCircleIcon } from "@patternfly/react-icons";
import type React from "react";
import { useCallback } from "react";
import { type HostEntry, HostEntryCard } from "../components/HostEntryCard.tsx";
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

const AWS_KEY_PATTERN = /^AKIA[A-Z0-9]{16}$/;
const AWS_SECRET_MIN_LENGTH = 40;

export const CaasStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const configData = state.configData as Record<string, unknown>;
  const globalData = (configData.global ?? {}) as Record<string, unknown>;
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

  const clusterFulfillmentConfig = (globalData.clusterFulfillmentConfig ??
    {}) as Record<string, string>;
  const awsAccessKeyId = clusterFulfillmentConfig.AWS_ACCESS_KEY_ID ?? "";
  const awsSecretAccessKey =
    clusterFulfillmentConfig.AWS_SECRET_ACCESS_KEY ?? "";
  const dnsZone = clusterFulfillmentConfig.DNS_ZONE ?? "";

  const setClusterFulfillmentField = useCallback(
    (key: string, value: string) => {
      const updated = { ...clusterFulfillmentConfig, [key]: value };
      if (!value) {
        delete updated[key];
      }
      dispatch({
        type: "SET_FIELD",
        path: "global.clusterFulfillmentConfig",
        value: updated,
      });
    },
    [clusterFulfillmentConfig, dispatch],
  );

  return (
    <Form>
      {/* DNS Configuration */}
      <Title headingLevel="h2" size="xl">
        DNS Configuration
      </Title>
      <Content component="p">
        Configure DNS backend for automated cluster DNS record management during
        provisioning. The platform creates DNS records for cluster API endpoints
        and ingress routes.
      </Content>

      <Alert
        variant="info"
        isInline
        title="AWS Route 53 Permissions Required"
        className={stepStyles.infoAlert}
      >
        <p>
          The AWS IAM user must have permissions to create, modify, and delete
          records in the target Route 53 hosted zone.
        </p>
      </Alert>

      <FormGroup label="DNS Backend" isRequired fieldId="dns-backend">
        <FormSelect
          id="dns-backend"
          value="dns.route53.dns"
          isDisabled
          aria-label="DNS Backend"
        >
          <FormSelectOption value="dns.route53.dns" label="Route 53" />
        </FormSelect>
        <HelperText>
          <HelperTextItem>
            DNS backend for automated cluster DNS record management
          </HelperTextItem>
        </HelperText>
      </FormGroup>

      <FormGroup
        label="AWS Access Key ID"
        isRequired
        fieldId="aws-access-key-id"
      >
        <TextInput
          id="aws-access-key-id"
          type="password"
          value={awsAccessKeyId}
          onChange={(_event, value) =>
            setClusterFulfillmentField("AWS_ACCESS_KEY_ID", value)
          }
          placeholder="AKIA..."
          aria-label="AWS Access Key ID"
          validated={
            awsAccessKeyId && !AWS_KEY_PATTERN.test(awsAccessKeyId)
              ? "error"
              : "default"
          }
        />
        <HelperText>
          <HelperTextItem
            variant={
              awsAccessKeyId && !AWS_KEY_PATTERN.test(awsAccessKeyId)
                ? "error"
                : "default"
            }
          >
            {awsAccessKeyId && !AWS_KEY_PATTERN.test(awsAccessKeyId)
              ? "Invalid AWS access key format (must start with AKIA and be 20 characters)"
              : "AWS IAM access key with Route 53 permissions"}
          </HelperTextItem>
        </HelperText>
      </FormGroup>

      <FormGroup
        label="AWS Secret Access Key"
        isRequired
        fieldId="aws-secret-access-key"
      >
        <TextInput
          id="aws-secret-access-key"
          type="password"
          value={awsSecretAccessKey}
          onChange={(_event, value) =>
            setClusterFulfillmentField("AWS_SECRET_ACCESS_KEY", value)
          }
          aria-label="AWS Secret Access Key"
          validated={
            awsSecretAccessKey &&
            awsSecretAccessKey.length < AWS_SECRET_MIN_LENGTH
              ? "error"
              : "default"
          }
        />
        <HelperText>
          <HelperTextItem
            variant={
              awsSecretAccessKey &&
              awsSecretAccessKey.length < AWS_SECRET_MIN_LENGTH
                ? "error"
                : "default"
            }
          >
            {awsSecretAccessKey &&
            awsSecretAccessKey.length < AWS_SECRET_MIN_LENGTH
              ? "AWS secret access key must be at least 40 characters"
              : "AWS IAM secret access key"}
          </HelperTextItem>
        </HelperText>
      </FormGroup>

      <FormGroup label="DNS Zone" fieldId="dns-zone">
        <TextInput
          id="dns-zone"
          value={dnsZone}
          onChange={(_event, value) =>
            setClusterFulfillmentField("DNS_ZONE", value)
          }
          placeholder="example.com"
          aria-label="DNS Zone"
        />
        <HelperText>
          <HelperTextItem>
            Route 53 hosted zone name. Leave empty to use the base domain.
          </HelperTextItem>
        </HelperText>
      </FormGroup>

      {/* Bare Metal Hosts */}
      <Title headingLevel="h2" size="xl">
        Bare Metal Hosts
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
    </Form>
  );
};
