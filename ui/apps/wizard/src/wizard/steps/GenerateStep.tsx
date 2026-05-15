import {
  Alert,
  Button,
  EmptyState,
  EmptyStateBody,
  Spinner,
  Title,
} from "@patternfly/react-core";
import { CheckCircleIcon } from "@patternfly/react-icons";
import type React from "react";
import { useCallback, useState } from "react";
import type { EnclaveConfig } from "@enclave-wizard-ui/api-client";
import { useEnclaveApi } from "../../api/useEnclaveApi.ts";
import { useWizard } from "../WizardContext.tsx";

type GenerateStatus = "idle" | "writing" | "success" | "error";

export const GenerateStep: React.FC = () => {
  const { state } = useWizard();
  const api = useEnclaveApi();
  const [status, setStatus] = useState<GenerateStatus>("idle");
  const [errorMessage, setErrorMessage] = useState<string>("");

  const handleGenerate = useCallback(async () => {
    setStatus("writing");
    setErrorMessage("");
    try {
      const enabledPlugins = ["lvms"];
      const configToWrite = {
        ...state.configData,
        global: {
          ...(state.configData.global as Record<string, unknown>),
          workingDir: "/home/enclave",
          disconnected: true,
          enabled_plugins: enabledPlugins,
        },
        certificates: state.configData.certificates ?? {},
        cloudInfra: state.configData.cloudInfra ?? { discovery_hosts: [] },
      };
      await api.writeConfig(configToWrite as unknown as EnclaveConfig);
      setStatus("success");
    } catch (err) {
      setStatus("error");
      setErrorMessage(
        err instanceof Error ? err.message : "Failed to write configuration",
      );
    }
  }, [api, state.configData]);

  if (status === "writing") {
    return (
      <EmptyState
        variant="lg"
        titleText="Writing configuration..."
        headingLevel="h2"
        icon={Spinner}
      >
        <EmptyStateBody>
          Writing config files to the landing zone. This may take a moment.
        </EmptyStateBody>
      </EmptyState>
    );
  }

  if (status === "success") {
    return (
      <EmptyState
        variant="lg"
        titleText="Configuration written successfully"
        headingLevel="h2"
        icon={CheckCircleIcon}
      >
        <EmptyStateBody>
          The configuration files (global.yaml, certificates.yaml,
          cloud_infra.yaml) have been written to the enclave config directory.
          You can now run the installer.
        </EmptyStateBody>
      </EmptyState>
    );
  }

  return (
    <div>
      <Title headingLevel="h2" size="xl">
        Generate configuration
      </Title>

      {status === "error" && (
        <Alert
          variant="danger"
          title="Failed to write configuration"
          isInline
          style={{ margin: "1rem 0" }}
        >
          {errorMessage}
        </Alert>
      )}

      <p style={{ margin: "1rem 0" }}>
        Click the button below to write the deployment configuration to disk.
      </p>

      <Button variant="primary" size="lg" onClick={handleGenerate}>
        Write configuration
      </Button>
    </div>
  );
};
