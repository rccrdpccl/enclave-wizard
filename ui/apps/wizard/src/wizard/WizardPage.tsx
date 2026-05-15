import {
  Alert,
  Button,
  Card,
  CardBody,
  Divider,
  Flex,
  FlexItem,
  ProgressStep,
  ProgressStepper,
  Spinner,
} from "@patternfly/react-core";
import type React from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useEnclaveApi } from "../api/useEnclaveApi.ts";
import { RedHatLogo } from "../common/components/RedHatLogo.tsx";
import {
  validateFields,
  validateHostEntries,
  type StepValidationError,
} from "../schema/schemaUtils.ts";
import { useOpenApiSchema } from "../schema/useOpenApiSchema.ts";
import { STEP_REQUIRED_FIELDS } from "./stepFields.ts";
import { CaasStep } from "./steps/CaasStep.tsx";
import { GenerateStep } from "./steps/GenerateStep.tsx";
import { GpuAiStep } from "./steps/GpuAiStep.tsx";
import { HubClusterStep } from "./steps/HubClusterStep.tsx";
import { LandingZoneStep } from "./steps/LandingZoneStep.tsx";
import { ReviewStep } from "./steps/ReviewStep.tsx";
import { SelectFlavorStep } from "./steps/SelectFlavorStep.tsx";
import { WelcomeStep } from "./steps/WelcomeStep.tsx";
import { useWizard, WizardProvider } from "./WizardContext.tsx";
import { wizardStyles as styles } from "./wizardStyles.ts";

interface StepDef {
  id: string;
  label: string;
}

const BASE_STEPS: StepDef[] = [
  { id: "welcome", label: "Welcome" },
  { id: "flavor", label: "Select" },
  { id: "landing-zone", label: "Landing Zone" },
  { id: "hub-cluster", label: "Hub Cluster" },
];

const CAAS_STEP: StepDef = { id: "caas", label: "Cluster as a Service" };
const GPU_AI_STEP: StepDef = { id: "gpu-ai", label: "GPU & AI" };

const TAIL_STEPS: StepDef[] = [
  { id: "review", label: "Review" },
  { id: "generate", label: "Generate" },
];

function buildSteps(selectedFlavors: Set<string>): StepDef[] {
  const steps = [...BASE_STEPS];
  if (selectedFlavors.has("cluster")) {
    steps.push(CAAS_STEP);
  }
  if (selectedFlavors.has("gpu-ai")) {
    steps.push(GPU_AI_STEP);
  }
  steps.push(...TAIL_STEPS);
  return steps;
}

function StepContent({ stepId }: { stepId: string }): React.ReactElement {
  switch (stepId) {
    case "welcome":
      return <WelcomeStep />;
    case "flavor":
      return <SelectFlavorStep />;
    case "landing-zone":
      return <LandingZoneStep />;
    case "hub-cluster":
      return <HubClusterStep />;
    case "caas":
      return <CaasStep />;
    case "gpu-ai":
      return <GpuAiStep />;
    case "review":
      return <ReviewStep />;
    case "generate":
      return <GenerateStep />;
    default:
      return <div>Unknown step</div>;
  }
}

function WizardContent(): React.ReactElement {
  const { state, dispatch } = useWizard();
  const { schema, loading: schemaLoading } = useOpenApiSchema();
  const api = useEnclaveApi();
  const [initDone, setInitDone] = useState(false);
  const [stepErrors, setStepErrors] = useState<StepValidationError[]>([]);

  const steps = useMemo(
    () => buildSteps(state.selectedFlavors),
    [state.selectedFlavors],
  );

  useEffect(() => {
    if (schema) {
      dispatch({ type: "SET_SCHEMA", schema });
    }
  }, [schema, dispatch]);

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
          dispatch({ type: "SET_FIELD", path: "global.disconnected", value: d.disconnected });
          dispatch({ type: "SET_FIELD", path: "global.storage_plugin", value: d.storagePlugin });
          dispatch({ type: "SET_FIELD", path: "global.blockStorageBackend", value: d.storagePlugin });
          dispatch({ type: "SET_FIELD", path: "global.defaultPrefix", value: 24 });
          dispatch({ type: "SET_FIELD", path: "global.quayBackend", value: "LocalStorage" });
          dispatch({ type: "SET_FIELD", path: "global.enabledPlugins", value: ["lvms"] });
        }

        if (pluginsResult.status === "fulfilled") {
          dispatch({ type: "SET_PLUGINS", plugins: pluginsResult.value.plugins ?? [] });
        }

        if (existingConfig.status === "fulfilled") {
          dispatch({ type: "LOAD_CONFIG", config: existingConfig.value as Record<string, unknown> });
        }
      } catch (err) {
        console.warn("Failed to load initial data:", err);
      }
      setInitDone(true);
    };
    init();
  }, [api, dispatch, initDone]);

  const currentStepId = steps[state.currentStep]?.id;
  const isWelcome = currentStepId === "welcome";
  const isFirst = state.currentStep === 0;
  const isLast = state.currentStep === steps.length - 1;

  const goBack = () => {
    setStepErrors([]);
    dispatch({ type: "SET_SHOW_VALIDATION", show: false });
    dispatch({ type: "SET_STEP", step: Math.max(0, state.currentStep - 1) });
  };

  const skipValidation = new URLSearchParams(window.location.search).has("skip_validation");

  const goNext = useCallback(() => {
    if (!skipValidation && state.schema && currentStepId) {
      const fieldsToValidate = STEP_REQUIRED_FIELDS[currentStepId];
      let errors: StepValidationError[] = [];

      if (fieldsToValidate) {
        const nonHostFields = fieldsToValidate.filter((f) => f !== "global.agent_hosts");
        errors = validateFields(state.schema, nonHostFields, state.configData as Record<string, unknown>);
      }

      if (currentStepId === "landing-zone") {
        const globalData = ((state.configData as Record<string, unknown>).global ?? {}) as Record<string, unknown>;
        const disconnected = globalData.disconnected !== false;
        if (disconnected) {
          const quayFields = ["global.quayUser", "global.quayPassword", "global.quayBackend"];
          errors.push(...validateFields(state.schema, quayFields, state.configData as Record<string, unknown>));
          const quayBackend = globalData.quayBackend as string;
          if (quayBackend === "RadosGWStorage") {
            const rgw = (globalData.quayBackendRGWConfiguration ?? {}) as Record<string, unknown>;
            for (const key of ["access_key", "secret_key", "bucket_name", "hostname"]) {
              if (!rgw[key] || (typeof rgw[key] === "string" && (rgw[key] as string).trim() === "")) {
                errors.push({ path: `global.quayBackendRGWConfiguration.${key}`, label: key, message: `${key} is required for RadosGW backend` });
              }
            }
          }
        }
      }

      if (currentStepId === "hub-cluster") {
        const globalData = ((state.configData as Record<string, unknown>).global ?? {}) as Record<string, unknown>;
        const agentHosts = Array.isArray(globalData.agent_hosts) ? (globalData.agent_hosts as Record<string, unknown>[]) : [];
        if (agentHosts.length !== 3) {
          errors.push({ path: "global.agent_hosts", label: "Control Plane Nodes", message: `Exactly 3 control plane nodes are required (currently ${agentHosts.length})` });
        } else {
          errors.push(...validateHostEntries(state.schema, agentHosts, "Node"));
        }
      }

      if (errors.length > 0) {
        setStepErrors(errors);
        dispatch({ type: "SET_SHOW_VALIDATION", show: true });
        return;
      }
    }

    setStepErrors([]);
    dispatch({ type: "SET_SHOW_VALIDATION", show: false });
    dispatch({ type: "SET_STEP", step: Math.min(steps.length - 1, state.currentStep + 1) });
  }, [currentStepId, state.schema, state.configData, state.currentStep, steps.length, skipValidation, dispatch]);

  if (schemaLoading) {
    return <Spinner aria-label="Loading wizard..." />;
  }

  return (
    <div className={styles.root}>
      <header className={styles.header}>
        <div className={styles.headerInner}>
          <RedHatLogo width={240} />
        </div>
        <Divider />
        {!isWelcome && (
          <div className={styles.headerInner}>
            <ProgressStepper aria-label="Wizard progress">
              {steps.map((step, i) => (
                <ProgressStep
                  key={step.id}
                  id={step.id}
                  titleId={`step-title-${step.id}`}
                  variant={i < state.currentStep ? "success" : i === state.currentStep ? "info" : "pending"}
                  isCurrent={i === state.currentStep}
                  aria-label={step.label}
                >
                  {step.label}
                </ProgressStep>
              ))}
            </ProgressStepper>
          </div>
        )}
      </header>

      <div className={styles.content}>
        <div className={styles.contentInner}>
          {stepErrors.length > 0 && (
            <Alert variant="danger" title="Please fill in all required fields" isInline className={styles.errorAlert}>
              <ul>
                {stepErrors.map((err) => (
                  <li key={err.path}>{err.message}</li>
                ))}
              </ul>
            </Alert>
          )}
          <Card isRounded>
            <CardBody className={styles.cardBody}>
              <StepContent stepId={currentStepId ?? "welcome"} />
            </CardBody>
          </Card>
        </div>
      </div>

      {!isWelcome && (
        <div className={styles.footer}>
          <div className={styles.footerInner}>
            <Flex justifyContent={{ default: "justifyContentSpaceBetween" }}>
              <FlexItem>
                <Button variant="secondary" onClick={goBack} isDisabled={isFirst}>Back</Button>
              </FlexItem>
              <FlexItem>
                <Button variant="primary" onClick={goNext} isDisabled={isLast}>Next</Button>
              </FlexItem>
            </Flex>
          </div>
        </div>
      )}
    </div>
  );
}

export const WizardPage: React.FC = () => {
  return (
    <WizardProvider>
      <WizardContent />
    </WizardProvider>
  );
};
