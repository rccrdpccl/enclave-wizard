import type React from "react";
import type { Container } from "./Container.ts";
import { ContainerContext } from "./Context.tsx";

interface ProviderProps {
  container: Container;
  children: React.ReactNode;
}

export const Provider: React.FC<ProviderProps> = ({ container, children }) => {
  return (
    <ContainerContext.Provider value={container}>
      {children}
    </ContainerContext.Provider>
  );
};
