import { FormGroup, FormHelperText, TextArea } from "@patternfly/react-core";
import type React from "react";

interface CertificateFieldProps {
  label: string;
  description?: string;
  value: string;
  onChange: (value: string) => void;
}

export const CertificateField: React.FC<CertificateFieldProps> = ({
  label,
  description,
  value,
  onChange,
}) => {
  const fieldId = `cert-${label.toLowerCase().replace(/\s+/g, "-")}`;

  return (
    <FormGroup label={label} fieldId={fieldId}>
      <TextArea
        id={fieldId}
        value={value}
        onChange={(_e, v) => onChange(v)}
        placeholder="-----BEGIN CERTIFICATE-----"
        rows={6}
        resizeOrientation="vertical"
        aria-label={label}
      />
      {description && <FormHelperText>{description}</FormHelperText>}
    </FormGroup>
  );
};
