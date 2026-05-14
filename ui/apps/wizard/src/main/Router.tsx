import { createBrowserRouter, Navigate, useRouteError } from "react-router-dom";

function ErrorPage() {
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
}

export const router = createBrowserRouter([
  {
    path: "/",
    index: true,
    element: <Navigate to="/wizard" />,
  },
  {
    path: "/wizard",
    errorElement: <ErrorPage />,
    lazy: async () => {
      const { WizardPage } = await import("../wizard/WizardPage.tsx");
      return { Component: WizardPage };
    },
  },
]);
