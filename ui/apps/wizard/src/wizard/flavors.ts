export type FlavorId = "cluster" | "nvidia-gpu" | "openshift-ai";

export interface FlavorDefinition {
  id: FlavorId;
  title: string;
  description: string;
  plugins: string[];
  requires?: FlavorId[];
}

export const FLAVORS: FlavorDefinition[] = [
  {
    id: "cluster",
    title: "Cluster as a Service",
    description:
      "On-demand container clusters with built-in scalability, resilience, and lifecycle management.",
    plugins: [],
  },
  {
    id: "nvidia-gpu",
    title: "NVIDIA GPU Support",
    description:
      "GPU operator for accelerated computing workloads.",
    plugins: ["nvidia-gpu"],
  },
  {
    id: "openshift-ai",
    title: "OpenShift AI",
    description:
      "AI/ML platform for model training and inference. Requires NVIDIA GPU support.",
    plugins: ["openshift-ai", "nvidia-gpu"],
    requires: ["nvidia-gpu"],
  },
];
