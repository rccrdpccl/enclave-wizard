import type React from "react";
import { createContext, useContext, useState } from "react";
import type { Experience } from "../experiences.ts";

export interface PluginInfo {
  name: string;
  [key: string]: unknown;
}

export interface CatalogState {
  schema: unknown | null;
  plugins: PluginInfo[];
  experiences: Experience[];
  loading: boolean;
}

const initialCatalogState: CatalogState = {
  schema: null,
  plugins: [],
  experiences: [],
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
  const [state, setState] = useState<CatalogState>(initialCatalogState);
  return (
    <CatalogContext.Provider value={{ state, setState }}>
      {children}
    </CatalogContext.Provider>
  );
};

export function useCatalog(): CatalogContextValue {
  const context = useContext(CatalogContext);
  if (context === null) {
    throw new Error("useCatalog must be used within a CatalogProvider.");
  }
  return context;
}
