export interface ExperiencePlugin {
  name: string;
  order: number;
  required: boolean;
}

export interface Experience {
  id: string;
  name: string;
  description: string;
  plugins: ExperiencePlugin[];
}

export const EXPERIENCES: Experience[] = [
  {
    id: "osac",
    name: "Open Sovereign AI Cloud",
    description: "Full OSAC platform with fulfillment, identity, and authorization",
    plugins: [
      { name: "trust-manager", order: 100, required: true },
      { name: "rhbk", order: 101, required: true },
      { name: "authorino", order: 102, required: true },
      { name: "aap", order: 103, required: true },
      { name: "osac", order: 200, required: true },
    ],
  },
  {
    id: "aiaas",
    name: "AI as a Service",
    description: "GPU-accelerated AI/ML workloads with OpenShift AI",
    plugins: [
      { name: "nvidia-gpu", order: 110, required: true },
      { name: "openshift-ai", order: 100, required: true },
    ],
  },
];

export function getExperiencePlugins(experienceId: string): string[] {
  const exp = EXPERIENCES.find((e) => e.id === experienceId);
  if (!exp) return [];
  return exp.plugins
    .sort((a, b) => a.order - b.order)
    .map((p) => p.name);
}
