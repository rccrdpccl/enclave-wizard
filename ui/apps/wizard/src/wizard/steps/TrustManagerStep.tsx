import {
  Form,
  FormGroup,
  FormHelperText,
  TextInput,
  Title,
} from "@patternfly/react-core";
import type React from "react";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles } from "./stepStyles.ts";

function getValueByPath(obj: Record<string, unknown>, path: string): unknown {
  const keys = path.split(".");
  let current: unknown = obj;
  for (const key of keys) {
    if (current == null || typeof current !== "object") return undefined;
    current = (current as Record<string, unknown>)[key];
  }
  return current;
}

export const TrustManagerStep: React.FC = () => {
  const { state, dispatch } = useWizard();

  const onChange = (path: string, value: unknown) =>
    dispatch({ type: "SET_FIELD", path, value });

  const configData = state.configData as Record<string, unknown>;
  const tmDefaults = (getValueByPath(
    configData,
    "global.trustManagerDefaults",
  ) ?? {}) as Record<string, unknown>;

  const setTM = (field: string, value: unknown) =>
    onChange(`global.trustManagerDefaults.${field}`, value);

  return (
    <Form>
      <Title headingLevel="h2" size="xl">
        Trust Manager
      </Title>

      <Title
        headingLevel="h3"
        size="lg"
        className={stepStyles.firstSectionTitle}
      >
        CA Issuer
      </Title>

      <FormGroup label="CA certificate lifetime" fieldId="tm-ca-duration">
        <TextInput
          id="tm-ca-duration"
          aria-label="CA certificate lifetime"
          value={(tmDefaults.trust_manager_ca_issuer_duration as string) ?? ""}
          onChange={(_e, v) =>
            setTM("trust_manager_ca_issuer_duration", v || undefined)
          }
          placeholder="87600h"
        />
        <FormHelperText>
          Lifetime of the CA certificate (e.g. 87600h for 10 years). Leave empty
          for default.
        </FormHelperText>
      </FormGroup>

      <FormGroup label="CA renewal period" fieldId="tm-ca-renew">
        <TextInput
          id="tm-ca-renew"
          aria-label="CA renewal period"
          value={
            (tmDefaults.trust_manager_ca_issuer_renew_before as string) ?? ""
          }
          onChange={(_e, v) =>
            setTM("trust_manager_ca_issuer_renew_before", v || undefined)
          }
          placeholder="8760h"
        />
        <FormHelperText>
          How long before expiry to renew the CA (e.g. 8760h for 1 year). Leave
          empty for default.
        </FormHelperText>
      </FormGroup>
    </Form>
  );
};
