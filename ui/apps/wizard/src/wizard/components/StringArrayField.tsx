import {
  Button,
  FormGroup,
  FormHelperText,
  TextInput,
} from "@patternfly/react-core";
import { MinusCircleIcon, PlusCircleIcon } from "@patternfly/react-icons";
import type React from "react";
import { stepStyles } from "../steps/stepStyles.ts";

interface StringArrayFieldProps {
  label: string;
  description?: string;
  values: string[];
  onChange: (values: string[]) => void;
  isRequired?: boolean;
}

export const StringArrayField: React.FC<StringArrayFieldProps> = ({
  label,
  description,
  values,
  onChange,
  isRequired = false,
}) => {
  const addItem = () => onChange([...values, ""]);

  const removeItem = (index: number) =>
    onChange(values.filter((_, i) => i !== index));

  const updateItem = (index: number, value: string) =>
    onChange(values.map((v, i) => (i === index ? value : v)));

  return (
    <FormGroup label={label} isRequired={isRequired} fieldId={`field-${label}`}>
      {values.map((value, index) => (
        <div
          key={`${label}-${index}`}
          className={stepStyles.arrayItemRow}
        >
          <TextInput
            id={`field-${label}-${index}`}
            value={value}
            onChange={(_e, v) => updateItem(index, v)}
            aria-label={`${label} item ${index + 1}`}
          />
          <Button
            variant="plain"
            aria-label={`Remove ${label} item ${index + 1}`}
            onClick={() => removeItem(index)}
          >
            <MinusCircleIcon />
          </Button>
        </div>
      ))}
      <Button variant="link" icon={<PlusCircleIcon />} onClick={addItem}>
        Add {label.toLowerCase()}
      </Button>
      {description && <FormHelperText>{description}</FormHelperText>}
    </FormGroup>
  );
};
