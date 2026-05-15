import { createBrowserRouter, Link, Navigate } from "react-router-dom";
import { ErrorPage } from "../pages/ErrorPage.tsx";

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
  {
    path: "*",
    element: (
      <div style={{ padding: "2rem", textAlign: "center" }}>
        <h1>404 — Page not found</h1>
        <p><Link to="/wizard">Go to wizard</Link></p>
      </div>
    ),
  },
]);
