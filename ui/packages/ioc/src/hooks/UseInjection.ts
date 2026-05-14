import { useContext } from "react";
import { ContainerContext } from "../Context.tsx";

export function useInjection<T>(key: symbol): T {
  const container = useContext(ContainerContext);
  if (container === null) {
    throw new Error(
      "useInjection must be used within a DI Provider. Wrap your app with <Provider container={...}>.",
    );
  }
  return container.resolve<T>(key);
}
