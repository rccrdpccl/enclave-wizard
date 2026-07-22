---
name: add-wizard-field
description: Add a new config field to the wizard — Go struct tag, UI label, schema rendering, validation, and step wiring. Use when adding a field, config option, or parameter.
---

# Add a Wizard Config Field

## 1. Add the Go struct field

File: `internal/models/config.go`

Add to the appropriate sub-struct with JSON, YAML, and `doc` tags:

```go
MyField *string `json:"myField,omitempty" yaml:"myField,omitempty" doc:"Human-readable description" minLength:"1"`
```

Use `*Type` + `omitempty` for optional, value types for required. Tags: `doc`, `pattern`, `enum`, `minimum`, `maximum`, `minLength`, `maxLength`, `minItems`, `maxItems`.

## 2. Rebuild and regenerate API client

```bash
make build
```

Then regenerate the TypeScript API client so the new field is included in
the `EnclaveConfigFromJSON`/`EnclaveConfigToJSON` serialization. Without this,
the Review step and config writer will silently drop the field.

```bash
# Start backend temporarily to get updated OpenAPI spec
./enclave-wizard --no-auth --enclave-dir ../enclave --tls-cert hack/tls/server.crt --tls-key hack/tls/server.key &
curl -sk https://localhost:3443/openapi.yaml -o ui/packages/api-client/api/openapi.yaml
kill %1
# Regenerate TypeScript client
cd ui && make api-client
```

## 3. Add label via schema `title` (preferred) or label override

Preferred: add `title: "Human Label"` to the field in the enclave plugin schema
(`plugins/osac/schemas/config.yaml`). The renderer uses `title` automatically.

Fallback: `ui/apps/wizard/src/schema/schemaUtils.ts` — `LABEL_OVERRIDES` map.

## 4. Add field to step component

Add the dot-path (e.g., `"global.myField"`) to the fields array in the step:

- `ui/apps/wizard/src/wizard/steps/LandingZoneStep.tsx`
- `ui/apps/wizard/src/wizard/steps/HubClusterStep.tsx`
- `ui/apps/wizard/src/wizard/steps/OsacStep.tsx`
- `ui/apps/wizard/src/wizard/steps/AAPStep.tsx`
- `ui/apps/wizard/src/wizard/steps/CaasStep.tsx`

## 5. Schema-driven rendering hints

In the enclave plugin schema (`plugins/osac/schemas/config.yaml`):

- `title: "Label"` — display label (avoids LABEL_OVERRIDES)
- `format: textarea` — renders as multi-line TextArea (use for PEM certs, YAML blocks)
- `enumLabels:` — map of value→display label for enum dropdowns (e.g., `bcm: "BCM"`)

For boolean fields that gate dependent fields, use `ToggleFieldGroup` component
(`ui/apps/wizard/src/wizard/components/ToggleFieldGroup.tsx`).

## 6. Add placeholder hint (optional)

File: `ui/apps/wizard/src/schema/SchemaFormRenderer.tsx` — `PLACEHOLDER_HINTS` map.

## 6. Mark required (if applicable)

File: `ui/apps/wizard/src/wizard/stepFields.ts` — add to `STEP_REQUIRED_FIELDS`.

## 7. Custom validation (if needed)

File: `ui/apps/wizard/src/wizard/hooks/useStepValidation.ts` — add logic in the step's `if` block.

## 8. Verify

Run `make demo`, navigate to the step, confirm field renders correctly. Run `make test`.

See `docs/howto-add-field.md` for the full guide with examples.
