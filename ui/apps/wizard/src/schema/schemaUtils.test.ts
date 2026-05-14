import { describe, expect, it } from "vitest";
import {
  extractFieldMeta,
  type FieldMeta,
  getNestedSchema,
  listSchemaFields,
} from "./schemaUtils.ts";

const SAMPLE_SCHEMA = {
  type: "object",
  properties: {
    global: {
      type: "object",
      required: ["baseDomain", "clusterName"],
      properties: {
        baseDomain: {
          type: "string",
          minLength: 1,
          doc: "Base DNS domain for the cluster",
        },
        clusterName: {
          type: "string",
          minLength: 1,
          doc: "OpenShift cluster name",
        },
        apiVIP: {
          type: "string",
          pattern:
            "^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$",
          doc: "Virtual IP for Kubernetes API server",
        },
        disconnected: {
          type: "boolean",
          doc: "Air-gapped deployment mode",
        },
        defaultPrefix: {
          type: "integer",
          minimum: 1,
          maximum: 32,
          doc: "Subnet prefix length",
        },
        quayBackend: {
          type: "string",
          enum: ["RadosGWStorage", "LocalStorage"],
          doc: "Quay image storage backend",
        },
        defaultNtpServers: {
          type: "array",
          items: { type: "string" },
          doc: "Additional NTP server addresses",
        },
      },
    },
  },
};

describe("getNestedSchema", () => {
  it("retrieves a nested property schema by dot path", () => {
    const result = getNestedSchema(SAMPLE_SCHEMA, "global.baseDomain");
    expect(result).toEqual({
      type: "string",
      minLength: 1,
      doc: "Base DNS domain for the cluster",
    });
  });

  it("returns undefined for missing paths", () => {
    const result = getNestedSchema(SAMPLE_SCHEMA, "global.nonExistent");
    expect(result).toBeUndefined();
  });
});

describe("extractFieldMeta", () => {
  it("extracts string field metadata", () => {
    const meta = extractFieldMeta(SAMPLE_SCHEMA, "global.baseDomain", [
      "baseDomain",
    ]);
    expect(meta).toEqual<FieldMeta>({
      path: "global.baseDomain",
      label: "Base Domain",
      description: "Base DNS domain for the cluster",
      type: "string",
      required: true,
      enum: undefined,
      pattern: undefined,
      minimum: undefined,
      maximum: undefined,
      items: undefined,
    });
  });

  it("extracts enum field metadata", () => {
    const meta = extractFieldMeta(SAMPLE_SCHEMA, "global.quayBackend", []);
    expect(meta?.type).toBe("string");
    expect(meta?.enum).toEqual(["RadosGWStorage", "LocalStorage"]);
  });

  it("extracts boolean field metadata", () => {
    const meta = extractFieldMeta(SAMPLE_SCHEMA, "global.disconnected", []);
    expect(meta?.type).toBe("boolean");
  });

  it("extracts integer field metadata with min/max", () => {
    const meta = extractFieldMeta(SAMPLE_SCHEMA, "global.defaultPrefix", []);
    expect(meta?.type).toBe("integer");
    expect(meta?.minimum).toBe(1);
    expect(meta?.maximum).toBe(32);
  });

  it("extracts array field metadata", () => {
    const meta = extractFieldMeta(
      SAMPLE_SCHEMA,
      "global.defaultNtpServers",
      [],
    );
    expect(meta?.type).toBe("array");
    expect(meta?.items).toEqual({ type: "string" });
  });

  it("extracts pattern for IP fields", () => {
    const meta = extractFieldMeta(SAMPLE_SCHEMA, "global.apiVIP", []);
    expect(meta?.pattern).toBeDefined();
  });
});

describe("listSchemaFields", () => {
  it("lists all top-level field paths under a parent", () => {
    const fields = listSchemaFields(SAMPLE_SCHEMA, "global");
    expect(fields).toContain("global.baseDomain");
    expect(fields).toContain("global.clusterName");
    expect(fields).toContain("global.apiVIP");
    expect(fields).toContain("global.disconnected");
    expect(fields).toContain("global.defaultPrefix");
    expect(fields).toContain("global.quayBackend");
    expect(fields).toContain("global.defaultNtpServers");
    expect(fields).toHaveLength(7);
  });
});
