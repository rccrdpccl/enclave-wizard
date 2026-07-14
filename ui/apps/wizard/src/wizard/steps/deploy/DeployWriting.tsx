import { EmptyState, EmptyStateBody, Spinner } from "@patternfly/react-core";
import type React from "react";

export const DeployWriting: React.FC = () => {
  return (
    <EmptyState
      variant="lg"
      titleText="Writing configuration..."
      headingLevel="h2"
      icon={Spinner}
    >
      <EmptyStateBody>
        Writing config files before starting deployment.
      </EmptyStateBody>
    </EmptyState>
  );
};
