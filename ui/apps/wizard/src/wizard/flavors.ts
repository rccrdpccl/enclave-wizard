export type FlavorId = "cluster" | "gpu-ai";

export interface FlavorDefinition {
  id: FlavorId;
  title: string;
  description: string;
  plugins: string[];
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
    id: "gpu-ai",
    title: "GPU & AI Workloads",
    description:
      "GPU-accelerated computing with NVIDIA GPU Operator and OpenShift AI platform.",
    plugins: ["nvidia-gpu", "openshift-ai"],
  },
];
