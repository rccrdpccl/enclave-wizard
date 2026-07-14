export interface ExperiencePlugin {
  name: string;
  order: number;
}

export interface Experience {
  id: string;
  name: string;
  description: string;
  plugins: ExperiencePlugin[];
}

/** Hardcoded fallback used when GET /api/v1/experiences is unavailable */
export const FALLBACK_EXPERIENCES: Experience[] = [
  {
    id: "caas",
    name: "Containers as a Service",
    description:
      "Self-service OpenShift cluster provisioning with Hosted Control Planes on bare metal infrastructure",
    plugins: [
      { name: "trust-manager", order: 100 },
      { name: "rhbk", order: 101 },
      { name: "authorino", order: 102 },
      { name: "aap", order: 103 },
      { name: "osac", order: 200 },
    ],
  },
  {
    id: "vmaas",
    name: "VMs as a Service",
    description:
      "Self-service virtual machine provisioning on OpenShift Virtualization",
    plugins: [
      { name: "trust-manager", order: 100 },
      { name: "rhbk", order: 101 },
      { name: "authorino", order: 102 },
      { name: "aap", order: 103 },
      { name: "cnv", order: 104 },
      { name: "osac", order: 200 },
    ],
  },
  {
    id: "bmaas",
    name: "Bare Metal as a Service",
    description:
      "Self-service bare metal server provisioning and lifecycle management",
    plugins: [
      { name: "trust-manager", order: 100 },
      { name: "rhbk", order: 101 },
      { name: "authorino", order: 102 },
      { name: "aap", order: 103 },
      { name: "osac", order: 200 },
    ],
  },
  {
    id: "gpu",
    name: "GPU Compute",
    description: "NVIDIA GPU Operator for GPU-accelerated workloads",
    plugins: [{ name: "nvidia-gpu", order: 110 }],
  },
];

/** @deprecated Use FALLBACK_EXPERIENCES instead */
export const EXPERIENCES = FALLBACK_EXPERIENCES;

export function getExperiencePlugins(
  experienceId: string,
  experiences: Experience[] = FALLBACK_EXPERIENCES,
): string[] {
  const exp = experiences.find((e) => e.id === experienceId);
  if (!exp) return [];
  return exp.plugins.sort((a, b) => a.order - b.order).map((p) => p.name);
}
