import { ExpandableSection, Form, Title } from "@patternfly/react-core";
import type React from "react";
import { useState } from "react";
import { SchemaFormRenderer } from "../../schema/SchemaFormRenderer.tsx";
import { useWizard } from "../WizardContext.tsx";
import { CertificateField } from "../components/CertificateField.tsx";
import { stepStyles } from "./stepStyles.ts";

const LZ_FIELDS = ["global.lzBmcIP", "global.lzBmcHostname"];

const QUAY_FIELDS = [
  "global.quayUser",
  "global.quayPassword",
  "global.quayBackend",
];

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
  "global.quayUser",
  "global.quayPassword",
  "global.quayBackend",
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
        Quay Registry
      </Title>
      <SchemaFormRenderer
        schema={state.schema}
        fields={QUAY_FIELDS}
        values={configData}
        onChange={onChange}
        showValidation={state.showValidation}
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
