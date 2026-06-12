export type FlavorId = "caas" | "vmaas";

export interface FlavorDefinition {
  id: FlavorId;
  title: string;
  subtitle: string;
  description: string;
  osacProfile: string;
  plugins: string[];
  addons?: { id: string; label: string; description: string; plugins: string[] }[];
}

const OSAC_PLUGINS = ["trust-manager", "rhbk", "authorino", "aap", "osac"];

export const FLAVORS: FlavorDefinition[] = [
  {
    id: "caas",
    title: "CaaS",
    subtitle: "Containers as a Service",
    description:
      "On-demand container clusters with built-in scalability, resilience, and lifecycle management. Provision and manage OpenShift spoke clusters from the hub.",
    osacProfile: "caas",
    plugins: [...OSAC_PLUGINS],
    addons: [
      {
        id: "gpu-ai",
        label: "GPU & AI",
        description: "NVIDIA GPU Operator and OpenShift AI for ML/AI workloads",
        plugins: ["nvidia-gpu", "openshift-ai"],
      },
    ],
  },
  {
    id: "vmaas",
    title: "VMaaS",
    subtitle: "VMs as a Service",
    description:
      "Run and manage virtual machines alongside containers using OpenShift Virtualization. Migrate existing VM workloads to a cloud-native platform.",
    osacProfile: "vmaas",
    plugins: [...OSAC_PLUGINS, "cnv"],
  },
];
