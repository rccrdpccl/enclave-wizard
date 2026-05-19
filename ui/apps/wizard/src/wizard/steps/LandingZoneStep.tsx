import {
  Checkbox,
  ExpandableSection,
  Form,
  Title,
} from "@patternfly/react-core";
import type React from "react";
import { useState } from "react";
import { SchemaFormRenderer } from "../../schema/SchemaFormRenderer.tsx";
import { useWizard } from "../WizardContext.tsx";
import { CertificateField } from "../components/CertificateField.tsx";
import { stepStyles } from "./stepStyles.ts";

const LZ_FIELDS = ["global.lzBmcIP", "global.lzBmcHostname"];

const IRONIC_CERTS = [
  { path: "certificates.ironicHTTPSCertificate", label: "Ironic HTTPS Certificate" },
  { path: "certificates.ironicHTTPSKey", label: "Ironic HTTPS Key" },
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

export const LZ_REQUIRED_FIELDS = [
  "global.lzBmcIP",
];

export const LandingZoneStep: React.FC = () => {
  const { state, dispatch } = useWizard();
  const [certsOpen, setCertsOpen] = useState(false);

  const onChange = (path: string, value: unknown) =>
    dispatch({ type: "SET_FIELD", path, value });

  if (!state.schema) {
    return <div>Loading schema...</div>;
  }

  const configData = state.configData as Record<string, unknown>;
  const globalData = (configData.global ?? {}) as Record<string, unknown>;
  const disconnected = globalData.disconnected !== false;

  const toggleDisconnected = (checked: boolean) => {
    onChange("global.disconnected", checked);
    if (!checked) {
      onChange("global.quayUser", "");
      onChange("global.quayPassword", "");
      onChange("global.quayBackend", "");
      onChange("global.quayBackendRGWConfiguration", undefined);
    } else if (!globalData.quayBackend) {
      onChange("global.quayBackend", "LocalStorage");
    }
  };

  return (
    <Form>
      <Title headingLevel="h2" size="xl">
        Landing Zone
      </Title>

      <Title headingLevel="h3" size="lg" className={stepStyles.firstSectionTitle}>
        Landing Zone Network
      </Title>
      <SchemaFormRenderer
        schema={state.schema}
        fields={LZ_FIELDS}
        values={configData}
        onChange={onChange}
        showValidation={state.showValidation}
      />

      <Title headingLevel="h3" size="lg" className={stepStyles.sectionTitle}>
        Deployment Mode
      </Title>
      <Checkbox
        id="disconnected-toggle"
        label="Disconnected (air-gapped) deployment"
        isChecked={disconnected}
        onChange={(_e, checked) => toggleDisconnected(checked)}
        description="When enabled, a local Quay mirror registry is configured for image distribution."
      />

      <ExpandableSection
        toggleText={certsOpen ? "Hide TLS certificates" : "TLS certificates (optional)"}
        isExpanded={certsOpen}
        onToggle={(_e, expanded) => setCertsOpen(expanded)}
        className={stepStyles.sectionTitle}
      >
        {IRONIC_CERTS.map(({ path, label }) => (
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
