---
name: add-wizard-step
description: Add a new wizard step or configuration sub-step — component creation, sub-step registration, routing, and validation wiring.
---

# Add a Wizard Step

Most new steps are **configuration sub-steps** (shown inside the Configure step). Top-level steps are rare.

## Adding a Config Sub-Step

### 1. Create the step component

File: `ui/apps/wizard/src/wizard/steps/MyStep.tsx`

Follow the existing pattern (e.g., `OsacStep.tsx`, `HubClusterStep.tsx`):

```typescript
import { SchemaFormRenderer } from "../../schema/SchemaFormRenderer.tsx";
import { useConfig } from "../contexts/ConfigContext.tsx";

const MY_FIELDS = [
  "global.myField1",
  "global.myField2",
];

export const MyStep = () => {
  const { state, dispatch } = useConfig();
  const configData = state.configData as Record<string, unknown>;
  const onChange = (path: string, value: unknown) =>
    dispatch({ type: "SET_FIELD", path, value });

  return (
    <SchemaFormRenderer
      schema={state.schema}
      fields={MY_FIELDS}
      values={configData}
      onChange={onChange}
      showValidation={state.showValidation}
    />
  );
};
```

### 2. Register the sub-step

File: `ui/apps/wizard/src/wizard/hooks/useSubSteps.ts`

Add to `buildConfigSubSteps()`:

```typescript
if (selectedFlavors.has("myexp")) {
  subs.push({ id: "my-step", label: "My Step" });
}
```

Or add to `BASE_CONFIG_SUBSTEPS` if it should always appear.

### 3. Route to the component

File: `ui/apps/wizard/src/wizard/steps/ConfigureStep.tsx`

Add to `SubStepContent` switch:

```typescript
case "my-step":
  return <MyStep />;
```

Also add in `ui/apps/wizard/src/wizard/WizardPage.tsx` if it has its own routing.

### 4. Add required fields

File: `ui/apps/wizard/src/wizard/stepFields.ts`

```typescript
"my-step": [
  "global.myField1",
  "global.myField2",
],
```

### 5. Add validation (if needed)

File: `ui/apps/wizard/src/wizard/hooks/useStepValidation.ts`

Add a block for custom validation logic.

### 6. Verify

```bash
make demo
# Navigate through wizard — new sub-step should appear in Configure
```

## Adding a Top-Level Step (rare)

Top-level steps (Welcome, Select Flavor, Configure, Review, Deploy) are defined in `WizardPage.tsx`. Adding one requires modifying the step array and the step content rendering. This is uncommon — prefer sub-steps.
