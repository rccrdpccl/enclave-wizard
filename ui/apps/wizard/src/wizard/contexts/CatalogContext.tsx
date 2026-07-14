import type React from "react";
import { createContext, useContext, useState } from "react";
import type { Experience } from "../experiences.ts";
import type { FlavorDefinition } from "../flavors.ts";

export interface CatalogState {
  experiences: Experience[];
  flavors: FlavorDefinition[];
  loading: boolean;
}

const INITIAL_CATALOG: CatalogState = {
  experiences: [],
  flavors: [],
  loading: true,
};

interface CatalogContextValue {
  state: CatalogState;
  setState: React.Dispatch<React.SetStateAction<CatalogState>>;
}

const CatalogContext = createContext<CatalogContextValue | null>(null);

export const CatalogProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [state, setState] = useState<CatalogState>(INITIAL_CATALOG);
  return (
    <CatalogContext.Provider value={{ state, setState }}>
      {children}
    </CatalogContext.Provider>
  );
};

export function useCatalog(): CatalogState & {
  setState: React.Dispatch<React.SetStateAction<CatalogState>>;
} {
  const context = useContext(CatalogContext);
  if (context === null) {
    throw new Error("useCatalog must be used within a CatalogProvider.");
  }
  return { ...context.state, setState: context.setState };
}
