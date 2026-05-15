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
    title: "NVIDIA AI in a Box",
    description:
      "GPU-accelerated AI/ML platform with NVIDIA GPU Operator and OpenShift AI.",
    plugins: ["nvidia-gpu", "openshift-ai"],
  },
];
