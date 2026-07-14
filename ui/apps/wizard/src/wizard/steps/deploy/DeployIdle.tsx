import { Alert, Button, Title } from "@patternfly/react-core";
import type React from "react";
import type { DeploymentError } from "../../../api/useDeployment.ts";

export const DeployIdle: React.FC<{
  error: DeploymentError | null;
  onDeploy: () => void;
}> = ({ error, onDeploy }) => {
  return (
    <div>
      <Title headingLevel="h2" size="xl">
        Deploy
      </Title>

      {error && (
        <Alert
          variant="danger"
          title={error.message}
          isInline
          style={{ margin: "1rem 0" }}
        >
          {error.details.length > 0 && (
            <ul style={{ margin: 0, paddingLeft: "1.25rem" }}>
              {error.details.map((d, i) => (
                // biome-ignore lint/suspicious/noArrayIndexKey: error details are a static list
                <li key={i}>{d}</li>
              ))}
            </ul>
          )}
        </Alert>
      )}

      <p style={{ margin: "1rem 0" }}>
        Write the configuration and start a full deployment (all 7 phases).
      </p>

      <Button variant="primary" size="lg" onClick={onDeploy}>
        Deploy
      </Button>
    </div>
  );
};
