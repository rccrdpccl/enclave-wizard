import type React from "react";
import { SchemaFormRenderer } from "../../schema/SchemaFormRenderer.tsx";

interface ToggleFieldGroupProps {
  schema: unknown;
  toggleField: string;
  dependentFields: string[];
  values: Record<string, unknown>;
  onChange: (path: string, value: unknown) => void;
  showValidation?: boolean;
}

function getValueByPath(obj: Record<string, unknown>, path: string): unknown {
  let current: unknown = obj;
  for (const key of path.split(".")) {
    if (current == null || typeof current !== "object") return undefined;
    current = (current as Record<string, unknown>)[key];
  }
  return current;
}

export const ToggleFieldGroup: React.FC<ToggleFieldGroupProps> = ({
  schema,
  toggleField,
  dependentFields,
  values,
  onChange,
  showValidation = false,
}) => {
  const enabled = Boolean(getValueByPath(values, toggleField));

  return (
    <>
      <SchemaFormRenderer
        schema={schema}
        fields={[toggleField]}
        values={values}
        onChange={onChange}
        showValidation={showValidation}
      />
      {enabled && (
        <SchemaFormRenderer
          schema={schema}
          fields={dependentFields}
          values={values}
          onChange={onChange}
          showValidation={showValidation}
        />
      )}
    </>
  );
};
