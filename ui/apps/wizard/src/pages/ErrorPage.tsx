import type React from "react";
import { useRouteError } from "react-router-dom";

export const ErrorPage: React.FC = () => {
  const error = useRouteError() as Error;
  return (
    <div style={{ padding: "2rem", fontFamily: "monospace" }}>
      <h1>Application Error</h1>
      <pre style={{ color: "red", whiteSpace: "pre-wrap" }}>
        {error?.message ?? String(error)}
      </pre>
      <pre style={{ color: "#666", whiteSpace: "pre-wrap", marginTop: "1rem" }}>
        {error?.stack}
      </pre>
    </div>
  );
};
