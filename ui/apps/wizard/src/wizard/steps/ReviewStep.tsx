import {
  Alert,
  Button,
  Content,
  DescriptionList,
  DescriptionListDescription,
  DescriptionListGroup,
  DescriptionListTerm,
  Spinner,
  Title,
} from "@patternfly/react-core";
import type React from "react";
import { useCallback, useState } from "react";
import type { EnclaveConfig } from "@enclave-wizard-ui/api-client";
import { useEnclaveApi } from "../../api/useEnclaveApi.ts";
import { useWizard } from "../WizardContext.tsx";

function flattenConfig(
  obj: unknown,
  prefix = "",
): { key: string; value: string }[] {
  const entries: { key: string; value: string }[] = [];
  if (obj == null || typeof obj !== "object") return entries;

  for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
    const path = prefix ? `${prefix}.${k}` : k;
    if (Array.isArray(v)) {
      if (v.length === 0) {
        entries.push({ key: path, value: "(none)" });
      } else if (v.every((item) => typeof item === "string")) {
        entries.push({ key: path, value: v.join(", ") });
      } else {
        for (let i = 0; i < v.length; i++) {
          entries.push(...flattenConfig(v[i], `${path}[${i}]`));
        }
      }
    } else if (v && typeof v === "object") {
      entries.push(...flattenConfig(v, path));
    } else {
      const display =
        typeof v === "string" && v.includes("BEGIN")
          ? "(PEM certificate)"
          : String(v ?? "");
      entries.push({ key: path, value: display });
    }
  }
  return entries;
}

export const ReviewStep: React.FC = () => {
  const { state, dispatch } = useWizard();
  const api = useEnclaveApi();
  const [validating, setValidating] = useState(false);
  const [validationDone, setValidationDone] = useState(false);

  const handleValidate = useCallback(async () => {
    setValidating(true);
    setValidationDone(false);
    try {
      const result = await api.validateConfig(
        state.configData as unknown as EnclaveConfig,
      );
      dispatch({
        type: "SET_VALIDATION_ERRORS",
        errors: result.errors ?? [],
      });
      setValidationDone(true);
    } catch (err) {
      dispatch({
        type: "SET_VALIDATION_ERRORS",
        errors: [
          {
            field: "",
            message:
              err instanceof Error
                ? err.message
                : "Validation request failed",
          },
        ],
      });
      setValidationDone(true);
    } finally {
      setValidating(false);
    }
  }, [api, state.configData, dispatch]);

  const entries = flattenConfig(state.configData);

  return (
    <div>
      <Title headingLevel="h2" size="xl">
        Review your deployment
      </Title>
      <Content component="p" style={{ marginTop: "0.5rem" }}>
        Confirm your selections below. Click Validate to check the
        configuration against the server before generating.
      </Content>

      <div style={{ margin: "1rem 0" }}>
        <Button
          variant="secondary"
          onClick={handleValidate}
          isLoading={validating}
          isDisabled={validating}
        >
          {validating ? "Validating..." : "Validate configuration"}
        </Button>
      </div>

      {!validating && validationDone && state.validationErrors.length > 0 && (
        <Alert
          variant="danger"
          title="Validation errors"
          isInline
          style={{ marginBottom: "1rem" }}
        >
          <ul>
            {state.validationErrors.map((err, i) => (
              <li key={`${err.field}-${i}`}>
                {err.field ? <strong>{err.field}:</strong> : null} {err.message}
              </li>
            ))}
          </ul>
        </Alert>
      )}

      {!validating && validationDone && state.validationErrors.length === 0 && (
        <Alert
          variant="success"
          title="Configuration is valid"
          isInline
          style={{ marginBottom: "1rem" }}
        />
      )}

      <DescriptionList>
        {entries.map(({ key, value }) => (
          <DescriptionListGroup key={key}>
            <DescriptionListTerm>{key}</DescriptionListTerm>
            <DescriptionListDescription>
              {value || "—"}
            </DescriptionListDescription>
          </DescriptionListGroup>
        ))}
      </DescriptionList>
    </div>
  );
};
