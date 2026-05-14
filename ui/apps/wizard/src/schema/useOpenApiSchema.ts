import { useCallback, useEffect, useState } from "react";

interface OpenApiSchema {
  components?: {
    schemas?: Record<string, unknown>;
  };
  [key: string]: unknown;
}

interface UseOpenApiSchemaResult {
  schema: unknown | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

function resolveRef(
  root: OpenApiSchema,
  ref: string,
): Record<string, unknown> | undefined {
  const path = ref.replace("#/", "").split("/");
  let current: unknown = root;
  for (const segment of path) {
    if (current == null || typeof current !== "object") return undefined;
    current = (current as Record<string, unknown>)[segment];
  }
  return current as Record<string, unknown> | undefined;
}

function resolveSchemaRefs(
  root: OpenApiSchema,
  node: Record<string, unknown>,
): Record<string, unknown> {
  if (node.$ref && typeof node.$ref === "string") {
    const resolved = resolveRef(root, node.$ref);
    if (resolved) return resolveSchemaRefs(root, { ...resolved });
    return node;
  }

  const result: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(node)) {
    if (value && typeof value === "object" && !Array.isArray(value)) {
      result[key] = resolveSchemaRefs(root, value as Record<string, unknown>);
    } else {
      result[key] = value;
    }
  }
  return result;
}

export function useOpenApiSchema(): UseOpenApiSchemaResult {
  const [schema, setSchema] = useState<unknown | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchSchema = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch("/openapi.json");
      if (!response.ok) {
        throw new Error(`Failed to fetch schema: ${response.status}`);
      }
      const raw: OpenApiSchema = await response.json();

      const enclaveConfigSchema = raw.components?.schemas?.EnclaveConfig as
        | Record<string, unknown>
        | undefined;

      if (!enclaveConfigSchema) {
        throw new Error("EnclaveConfig schema not found in OpenAPI spec");
      }

      const resolved = resolveSchemaRefs(raw, enclaveConfigSchema);
      setSchema(resolved);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Unknown error fetching schema";
      setError(message);
      console.error("Failed to load OpenAPI schema:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSchema();
  }, [fetchSchema]);

  return { schema, loading, error, refetch: fetchSchema };
}
