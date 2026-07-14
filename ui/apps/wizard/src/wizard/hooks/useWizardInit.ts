import { EnclaveConfigToJSON } from "@enclave-wizard-ui/api-client";
import { useEffect, useState } from "react";
import { useEnclaveApi } from "../../api/useEnclaveApi.ts";
import { useOpenApiSchema } from "../../schema/useOpenApiSchema.ts";
import { useCatalog } from "../contexts/CatalogContext.tsx";
import { useConfig } from "../contexts/ConfigContext.tsx";
import { EXPERIENCES } from "../experiences.ts";

export function useWizardInit(): { loading: boolean } {
  const { dispatch: configDispatch } = useConfig();
  const { setState: setCatalog } = useCatalog();
  const { schema, loading: schemaLoading } = useOpenApiSchema();
  const api = useEnclaveApi();
  const [initDone, setInitDone] = useState(false);

  useEffect(() => {
    if (schema) {
      setCatalog((prev) => ({ ...prev, schema }));
    }
  }, [schema, setCatalog]);

  useEffect(() => {
    if (initDone) return;
    const init = async () => {
      try {
        const [defaults, pluginsResult, existingConfig] =
          await Promise.allSettled([
            api.getDefaults(),
            api.getPlugins(),
            api.getConfig(),
          ]);

        if (defaults.status === "fulfilled") {
          const d = defaults.value;
          configDispatch({
            type: "SET_FIELD",
            path: "global.storage_plugin",
            value: d.storagePlugin,
          });
          configDispatch({
            type: "SET_FIELD",
            path: "global.defaultPrefix",
            value: 24,
          });
          configDispatch({
            type: "SET_FIELD",
            path: "global.quayBackend",
            value: "LocalStorage",
          });
          configDispatch({
            type: "SET_FIELD",
            path: "global.enabled_plugins",
            value: ["lvms"],
          });
        }

        if (pluginsResult.status === "fulfilled") {
          setCatalog((prev) => ({
            ...prev,
            plugins: (pluginsResult.value.plugins ?? []) as Array<{
              name: string;
              [key: string]: unknown;
            }>,
          }));
        }

        // Set experiences (hardcoded fallback for Phase 1)
        setCatalog((prev) => ({ ...prev, experiences: EXPERIENCES }));

        if (existingConfig.status === "fulfilled") {
          configDispatch({
            type: "LOAD_CONFIG",
            config: EnclaveConfigToJSON(existingConfig.value),
          });
        }
      } catch (err) {
        console.warn("Failed to load initial data:", err);
      }
      setCatalog((prev) => ({ ...prev, loading: false }));
      setInitDone(true);
    };
    init();
  }, [api, configDispatch, setCatalog, initDone]);

  return { loading: schemaLoading };
}
