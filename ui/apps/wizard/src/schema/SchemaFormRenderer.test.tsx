import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { SchemaFormRenderer } from "./SchemaFormRenderer.tsx";

const MOCK_SCHEMA = {
  type: "object",
  properties: {
    global: {
      type: "object",
      required: ["baseDomain"],
      properties: {
        baseDomain: {
          type: "string",
          minLength: 1,
          doc: "Base DNS domain",
        },
        disconnected: {
          type: "boolean",
          doc: "Air-gapped mode",
        },
        defaultPrefix: {
          type: "integer",
          minimum: 1,
          maximum: 32,
          doc: "Subnet prefix",
        },
        quayBackend: {
          type: "string",
          enum: ["RadosGWStorage", "LocalStorage"],
          doc: "Quay backend",
        },
        apiVIP: {
          type: "string",
          pattern: "^[0-9.]+$",
          doc: "API VIP address",
        },
      },
    },
  },
};

describe("SchemaFormRenderer", () => {
  it("renders a text input for a string field", () => {
    render(
      <SchemaFormRenderer
        schema={MOCK_SCHEMA}
        fields={["global.baseDomain"]}
        values={{ global: { baseDomain: "test.com" } }}
        onChange={vi.fn()}
      />,
    );
    const input = screen.getByLabelText("Base Domain");
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue("test.com");
  });

  it("renders a checkbox for a boolean field", () => {
    render(
      <SchemaFormRenderer
        schema={MOCK_SCHEMA}
        fields={["global.disconnected"]}
        values={{ global: { disconnected: true } }}
        onChange={vi.fn()}
      />,
    );
    const checkbox = screen.getByLabelText("Disconnected (Air-Gapped) Mode");
    expect(checkbox).toBeChecked();
  });

  it("renders a select for an enum field", () => {
    render(
      <SchemaFormRenderer
        schema={MOCK_SCHEMA}
        fields={["global.quayBackend"]}
        values={{ global: { quayBackend: "LocalStorage" } }}
        onChange={vi.fn()}
      />,
    );
    const select = screen.getByLabelText("Quay Storage Backend");
    expect(select).toHaveValue("LocalStorage");
  });

  it("renders helper text from the doc annotation", () => {
    render(
      <SchemaFormRenderer
        schema={MOCK_SCHEMA}
        fields={["global.baseDomain"]}
        values={{}}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByText("Base DNS domain")).toBeInTheDocument();
  });

  it("calls onChange with the correct path and value", async () => {
    const onChange = vi.fn();
    render(
      <SchemaFormRenderer
        schema={MOCK_SCHEMA}
        fields={["global.baseDomain"]}
        values={{ global: { baseDomain: "" } }}
        onChange={onChange}
      />,
    );
    const input = screen.getByLabelText("Base Domain");
    await userEvent.type(input, "x");
    expect(onChange).toHaveBeenCalledWith("global.baseDomain", "x");
  });
});
