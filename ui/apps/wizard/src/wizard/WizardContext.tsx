/**
 * Backward-compatibility shim.
 *
 * All new code should import from contexts/ConfigContext, contexts/WizardNavContext,
 * or contexts/CatalogContext directly. This module re-exports the combined
 * WizardState / WizardAction / useWizard() interfaces so existing step components
 * keep compiling during the incremental migration.
 */
import type React from "react";
import { useMemo } from "react";
import { CatalogProvider, useCatalog } from "./contexts/CatalogContext.tsx";
import {
  type ConfigData,
  ConfigProvider,
  configReducer,
  initialConfigState,
  setNestedField,
  useConfig,
} from "./contexts/ConfigContext.tsx";
import {
  useWizardNav,
  WizardNavProvider,
} from "./contexts/WizardNavContext.tsx";
import { EXPERIENCES } from "./experiences.ts";
import type { FlavorId } from "./flavors.ts";
import { FLAVORS } from "./flavors.ts";

// Re-export for buildFinalConfig and tests
export type { ConfigData };

export interface ValidationError {
  field: string;
  message: string;
}

/**
 * Legacy combined state. Step components that still call useWizard() see this shape.
 */
export interface WizardState {
  currentStep: number;
  selectedFlavors: Set<FlavorId>;
  configData: ConfigData;
  validationErrors: ValidationError[];
  showValidation: boolean;
  schema: unknown | null;
  plugins: unknown[];
}

export type WizardAction =
  | { type: "SET_STEP"; step: number }
  | { type: "TOGGLE_FLAVOR"; flavor: FlavorId }
  | { type: "SET_FIELD"; path: string; value: unknown }
  | { type: "SET_SCHEMA"; schema: unknown }
  | { type: "SET_PLUGINS"; plugins: unknown[] }
  | { type: "SET_VALIDATION_ERRORS"; errors: ValidationError[] }
  | { type: "SET_SHOW_VALIDATION"; show: boolean }
  | { type: "LOAD_CONFIG"; config: ConfigData };

export const initialWizardState: WizardState = {
  currentStep: 0,
  selectedFlavors: new Set(),
  configData: {},
  validationErrors: [],
  showValidation: false,
  schema: null,
  plugins: [],
};

// Re-export the config reducer internals used by tests
export { configReducer as wizardReducer, initialConfigState, setNestedField };

interface WizardContextValue {
  state: WizardState;
  dispatch: React.Dispatch<WizardAction>;
}

/**
 * Composite provider that wraps all three new context providers.
 */
export const WizardProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  return (
    <CatalogProvider>
      <ConfigProvider>
        <WizardNavProvider>{children}</WizardNavProvider>
      </ConfigProvider>
    </CatalogProvider>
  );
};

/**
 * Legacy hook: assembles state from the three contexts and returns a
 * dispatch function that routes actions to the appropriate context.
 */
export function useWizard(): WizardContextValue {
  const { state: config, dispatch: configDispatch } = useConfig();
  const { state: nav, dispatch: navDispatch } = useWizardNav();
  const { state: catalog, setState: setCatalog } = useCatalog();

  const state: WizardState = useMemo(
    () => ({
      currentStep: nav.currentStep,
      selectedFlavors: config.selectedFlavors as Set<FlavorId>,
      configData: config.configData,
      validationErrors: nav.validationErrors as ValidationError[],
      showValidation: nav.showValidation,
      schema: catalog.schema,
      plugins: catalog.plugins,
    }),
    [config, nav, catalog],
  );

  const dispatch = useMemo(() => {
    return (action: WizardAction) => {
      switch (action.type) {
        case "SET_STEP":
          navDispatch(action);
          break;
        case "TOGGLE_FLAVOR":
          configDispatch({
            type: "TOGGLE_FLAVOR",
            flavor: action.flavor,
            experiences: EXPERIENCES,
            flavors: FLAVORS,
          });
          break;
        case "SET_FIELD":
          configDispatch(action);
          break;
        case "SET_SCHEMA":
          setCatalog((prev) => ({ ...prev, schema: action.schema }));
          break;
        case "SET_PLUGINS":
          setCatalog((prev) => ({
            ...prev,
            plugins: action.plugins as Array<{ name: string }>,
          }));
          break;
        case "SET_VALIDATION_ERRORS":
          navDispatch({
            type: "SET_VALIDATION_ERRORS",
            errors: action.errors.map((e) => ({
              path: e.field ?? "",
              label: "",
              message: e.message,
            })),
          });
          break;
        case "SET_SHOW_VALIDATION":
          navDispatch(action);
          break;
        case "LOAD_CONFIG":
          configDispatch(action);
          break;
      }
    };
  }, [configDispatch, navDispatch, setCatalog]);

  return { state, dispatch };
}
