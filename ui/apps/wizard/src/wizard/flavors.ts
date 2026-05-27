export type FlavorId = "cluster";

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
];
