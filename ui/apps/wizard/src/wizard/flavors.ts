import type { ReactNode } from "react";

export type FlavorId = "cluster";

export interface FlavorDefinition {
  id: FlavorId;
  title: string;
  description: string;
  hostStepPaths: string[];
  defaultPlugins: string[];
  icon?: ReactNode;
}

export const FLAVORS: FlavorDefinition[] = [
  {
    id: "cluster",
    title: "Cluster as a Service",
    description:
      "On-demand container clusters with built-in scalability, resilience, and lifecycle management.",
    hostStepPaths: ["global.agent_hosts", "cloudInfra.discovery_hosts"],
    defaultPlugins: ["lvms"],
  },
];
