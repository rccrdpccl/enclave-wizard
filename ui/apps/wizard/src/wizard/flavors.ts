import { getExperiencePlugins } from "./experiences.ts";

export type FlavorId = "caas" | "vmaas";

export interface FlavorAddon {
  id: string;
  label: string;
  description: string;
  experienceId?: string;
  plugins: string[];
}

export interface FlavorDefinition {
  id: FlavorId;
  title: string;
  subtitle: string;
  description: string;
  osacProfile: string;
  experienceId: string;
  extraPlugins: string[];
  addons?: FlavorAddon[];
}

export const FLAVORS: FlavorDefinition[] = [
  {
    id: "caas",
    title: "CaaS",
    subtitle: "Containers as a Service",
    description:
      "On-demand container clusters with built-in scalability, resilience, and lifecycle management. Provision and manage OpenShift spoke clusters from the hub.",
    osacProfile: "caas",
    experienceId: "osac",
    extraPlugins: [],
  },
  {
    id: "vmaas",
    title: "VMaaS",
    subtitle: "VMs as a Service",
    description:
      "Run and manage virtual machines alongside containers using OpenShift Virtualization. Migrate existing VM workloads to a cloud-native platform.",
    osacProfile: "vmaas",
    experienceId: "osac",
    extraPlugins: ["cnv"],
  },
];

export function getFlavorPlugins(flavor: FlavorDefinition): string[] {
  return [...getExperiencePlugins(flavor.experienceId), ...flavor.extraPlugins];
}
