import { createContext } from "react";
import type { Container } from "./Container.ts";

export const ContainerContext = createContext<Container | null>(null);
