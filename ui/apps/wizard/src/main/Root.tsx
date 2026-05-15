import "@patternfly/react-core/dist/styles/base.css";

import {
  Configuration,
  ConfigApi,
  DefaultsApi,
  PluginsApi,
} from "@enclave-wizard-ui/api-client";
import {
  Container,
  Provider as DependencyInjectionProvider,
} from "@enclave-wizard-ui/ioc";
import { Spinner } from "@patternfly/react-core";
import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { router } from "./Router.tsx";
import { Symbols } from "./Symbols.ts";

function getApiBasePath(): string {
  return window.location.origin;
}

function getConfiguredContainer(): Container {
  const apiConfig = new Configuration({
    basePath: getApiBasePath(),
    fetchApi: (url, init) => fetch(url, { ...init, cache: "no-store" }),
  });

  const container = new Container();
  container.register(Symbols.ConfigApi, new ConfigApi(apiConfig));
  container.register(Symbols.DefaultsApi, new DefaultsApi(apiConfig));
  container.register(Symbols.PluginsApi, new PluginsApi(apiConfig));
  return container;
}

function main(): void {
  const root = document.getElementById("root");
  if (!root) {
    throw new Error("Root element not found.");
  }

  root.style.height = "inherit";
  const container = getConfiguredContainer();
  ReactDOM.createRoot(root).render(
    <React.StrictMode>
      <DependencyInjectionProvider container={container}>
        <React.Suspense fallback={<Spinner />}>
          <RouterProvider router={router} />
        </React.Suspense>
      </DependencyInjectionProvider>
    </React.StrictMode>,
  );
}

main();
