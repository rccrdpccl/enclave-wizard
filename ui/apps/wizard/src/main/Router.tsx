import { createBrowserRouter, Link, Navigate } from "react-router-dom";
import { ErrorPage } from "../pages/ErrorPage.tsx";
import { AuthGuard } from "./AuthGuard.tsx";

export const router = createBrowserRouter([
  {
    path: "/",
    index: true,
    element: <Navigate to="/login" />,
  },
  {
    path: "/login",
    errorElement: <ErrorPage />,
    lazy: async () => {
      const { LoginPage } = await import("../auth/LoginPage.tsx");
      return { Component: LoginPage };
    },
  },
  {
    path: "/wizard",
    errorElement: <ErrorPage />,
    element: <AuthGuard />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { WizardPage } = await import("../wizard/WizardPage.tsx");
          return { Component: WizardPage };
        },
      },
    ],
  },
  {
    path: "*",
    element: (
      <div style={{ padding: "2rem", textAlign: "center" }}>
        <h1>404 — Page not found</h1>
        <p><Link to="/login">Go to login</Link></p>
      </div>
    ),
  },
]);
