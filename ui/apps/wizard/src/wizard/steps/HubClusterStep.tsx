import {
  Button,
  Checkbox,
  ExpandableSection,
  Flex,
  FlexItem,
  Form,
  FormGroup,
  FormHelperText,
  TextArea,
  Title,
} from "@patternfly/react-core";
import { MinusCircleIcon, PlusCircleIcon } from "@patternfly/react-icons";
import type React from "react";
import { useState } from "react";
import { SchemaFormRenderer } from "../../schema/SchemaFormRenderer.tsx";
import { useWizard } from "../WizardContext.tsx";
import { CertificateField } from "../components/CertificateField.tsx";
import { HostEntryCard, type HostEntry } from "../components/HostEntryCard.tsx";
import { stepStyles } from "./stepStyles.ts";

const CLUSTER_FIELDS = ["global.baseDomain", "global.clusterName"];

const NETWORK_FIELDS = [
  "global.machineNetwork",
  "global.apiVIP",
  "global.ingressVIP",
  "global.rendezvousIP",
  "global.defaultDNS",
  "global.defaultGateway",
  "global.defaultPrefix",
];


const EMPTY_HOST: HostEntry = {
  name: "",
  macAddress: "",
  ipAddress: "",
  redfish: "",
  redfishUser: "",
  redfishPassword: "",
  rootDisk: "",
};

const HUB_CERTS = [
  { path: "certificates.sslAPICertificateFullChain", label: "API Certificate (Full Chain)" },
  { path: "certificates.sslAPICertificateKey", label: "API Certificate Key" },
  { path: "certificates.sslIngressCertificateFullChain", label: "Ingress Certificate (Full Chain)" },
  { path: "certificates.sslIngressCertificateKey", label: "Ingress Certificate Key" },
  { path: "certificates.sslCACertificate", label: "Root CA Certificate" },
];

function getValueByPath(obj: Record<string, unknown>, path: string): unknown {
  const keys = path.split(".");
  let current: unknown = obj;
  for (const key of keys) {
    if (current == null || typeof current !== "object") return undefined;
    current = (current as Record<string, unknown>)[key];
  }
  return current;
}

function StorageSelection({
  configData,
  onChange,
}: {
  configData: Record<string, unknown>;
  onChange: (path: string, value: unknown) => void;
}) {
  const globalData = (configData.global ?? {}) as Record<string, unknown>;
  const enabledPlugins: string[] = Array.isArray(globalData.enabled_plugins)
    ? (globalData.enabled_plugins as string[])
    : ["lvms"];
  const odfEnabled = enabledPlugins.includes("odf");
  const odfExternalConfig = (globalData.odfExternalConfig as string) ?? "";

  const toggleOdf = (checked: boolean) => {
    const next = checked
      ? [...new Set([...enabledPlugins, "odf"])]
      : enabledPlugins.filter((p) => p !== "odf");
    onChange("global.enabled_plugins", next);
    onChange("global.storage_plugin", checked ? "odf" : "lvms");
    onChange("global.blockStorageBackend", checked ? "odf" : "lvms");
    if (!checked) {
      onChange("global.odfExternalConfig", "");
    }
  };

  return (
    <>
      <Checkbox
        id="storage-lvms"
        label="LVMS (Logical Volume Manager Storage)"
        isChecked={true}
        isDisabled={true}
        description="Always enabled. Default block storage for the hub cluster."
      />
      <Checkbox
        id="storage-odf"
        label="ODF (OpenShift Data Foundation)"
        isChecked={odfEnabled}
        onChange={(_e, checked) => toggleOdf(checked)}
        description="External Ceph storage. Requires connection details below."
        style={{ marginTop: "0.5rem" }}
      />
      {odfEnabled && (
        <FormGroup
          label="ODF External Ceph Connection Details"
          isRequired
          fieldId="odf-external-config"
          style={{ marginTop: "0.75rem", marginLeft: "1.5rem" }}
        >
          <TextArea
            id="odf-external-config"
            value={odfExternalConfig}
            onChange={(_e, v) => onChange("global.odfExternalConfig", v)}
            placeholder='[{"name": "external-cluster-user-command", "kind": "ConfigMap", "data": ...}]'
            rows={4}
            isRequired
            aria-label="ODF External Ceph Connection Details"
          />
          <FormHelperText>
            JSON output from ceph-external-cluster-details-exporter.py
          </FormHelperText>
        </FormGroup>
      )}
    </>
  );
}

export const HUB_REQUIRED_FIELDS = [
  "global.baseDomain",
  "global.clusterName",
  "global.machineNetwork",
  "global.apiVIP",
  "global.ingressVIP",
  "global.rendezvousIP",
  "global.defaultDNS",
  "global.defaultGateway",
  "global.defaultPrefix",
  "global.blockStorageBackend",
  "global.pullSecret",
  "global.sshPubPath",
];

export const HubClusterStep: React.FC = () => {
  const { state, dispatch } = useWizard();
  const [certsOpen, setCertsOpen] = useState(false);

  const onChange = (path: string, value: unknown) =>
    dispatch({ type: "SET_FIELD", path, value });

  const configData = state.configData as Record<string, unknown>;
  const globalData = (configData.global ?? {}) as Record<string, unknown>;

  const agentHosts: HostEntry[] = Array.isArray(globalData.agentHosts)
    ? (globalData.agentHosts as HostEntry[])
    : [];

  const setAgentHosts = (hosts: HostEntry[]) =>
    dispatch({ type: "SET_FIELD", path: "global.agentHosts", value: hosts });

  const hostCount = agentHosts.length;
  const canAddHost = hostCount < 3;

  if (!state.schema) {
    return <div>Loading schema...</div>;
  }

  return (
    <Form>
      <Title headingLevel="h2" size="xl">
        Hub Cluster
      </Title>

      <Title headingLevel="h3" size="lg" className={stepStyles.firstSectionTitle}>
        Cluster Identity
      </Title>
      <SchemaFormRenderer
        schema={state.schema}
        fields={CLUSTER_FIELDS}
        values={configData}
        onChange={onChange}
        showValidation={state.showValidation}
      />

      <Title headingLevel="h3" size="lg" className={stepStyles.sectionTitle}>
        Network
      </Title>
      <SchemaFormRenderer
        schema={state.schema}
        fields={NETWORK_FIELDS}
        values={configData}
        onChange={onChange}
        showValidation={state.showValidation}
      />

      <Title headingLevel="h3" size="lg" className={stepStyles.sectionTitle}>
        Storage
      </Title>
      <StorageSelection configData={configData} onChange={onChange} />

      <Title headingLevel="h3" size="lg" className={stepStyles.sectionTitle}>
        Authentication
      </Title>
      <FormGroup label="Pull Secret" isRequired fieldId="pull-secret">
        <TextArea
          id="pull-secret"
          value={(globalData.pullSecret as string) ?? ""}
          onChange={(_e, v) => onChange("global.pullSecret", v)}
          placeholder='{"auths":{}}'
          rows={4}
          isRequired
          aria-label="Pull Secret"
        />
      </FormGroup>
      <FormGroup label="SSH Public Key" isRequired fieldId="ssh-pub-key">
        <TextArea
          id="ssh-pub-key"
          value={(globalData.sshPubPath as string) ?? ""}
          onChange={(_e, v) => onChange("global.sshPubPath", v)}
          placeholder="ssh-rsa AAAA..."
          rows={3}
          isRequired
          aria-label="SSH Public Key"
        />
      </FormGroup>

      <Flex
        justifyContent={{ default: "justifyContentSpaceBetween" }}
        alignItems={{ default: "alignItemsCenter" }}
        className={stepStyles.sectionTitle}
      >
        <FlexItem>
          <Title headingLevel="h3" size="lg">
            Control Plane Nodes ({hostCount}/3)
          </Title>
        </FlexItem>
        <FlexItem>
          <Button
            variant="link"
            icon={<PlusCircleIcon />}
            onClick={() => setAgentHosts([...agentHosts, { ...EMPTY_HOST }])}
            isDisabled={!canAddHost}
          >
            Add node
          </Button>
        </FlexItem>
      </Flex>
      {hostCount === 0 && (
        <p className={stepStyles.emptyHint}>
          Add 3 control plane nodes to proceed. Click &quot;Add node&quot; to get started.
        </p>
      )}
      {hostCount > 0 && hostCount < 3 && (
        <p className={stepStyles.warningHint}>
          {3 - hostCount} more node{3 - hostCount > 1 ? "s" : ""} required.
        </p>
      )}
      <Flex direction={{ default: "column" }} gap={{ default: "gapMd" }} className={stepStyles.hostSection}>
        {agentHosts.map((host, i) => (
          <FlexItem key={`agent-${i}`}>
            <Flex alignItems={{ default: "alignItemsFlexStart" }} gap={{ default: "gapSm" }}>
              <FlexItem grow={{ default: "grow" }}>
                <HostEntryCard
                  index={i}
                  host={host}
                  onChange={(h) => {
                    const updated = [...agentHosts];
                    updated[i] = h;
                    setAgentHosts(updated);
                  }}
                  label="Node"
                />
              </FlexItem>
              <FlexItem>
                <Button
                  variant="plain"
                  aria-label={`Remove node ${i + 1}`}
                  onClick={() => setAgentHosts(agentHosts.filter((_, idx) => idx !== i))}
                  className={stepStyles.removeButton}
                >
                  <MinusCircleIcon />
                </Button>
              </FlexItem>
            </Flex>
          </FlexItem>
        ))}
      </Flex>

      <ExpandableSection
        toggleText={certsOpen ? "Hide TLS certificates" : "TLS certificates (optional)"}
        isExpanded={certsOpen}
        onToggle={(_e, expanded) => setCertsOpen(expanded)}
        className={stepStyles.sectionTitle}
      >
        {HUB_CERTS.map(({ path, label }) => (
          <CertificateField
            key={path}
            label={label}
            value={(getValueByPath(configData, path) as string) ?? ""}
            onChange={(v) => onChange(path, v)}
          />
        ))}
      </ExpandableSection>
    </Form>
  );
};
