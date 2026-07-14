import { describe, expect, it } from "vitest";
import { buildConfigSubSteps } from "./useSubSteps.ts";

describe("buildConfigSubSteps", () => {
  it("returns base steps when no flavors are selected", () => {
    const subs = buildConfigSubSteps(new Set(), []);
    expect(subs.map((s) => s.id)).toEqual([
      "landing-zone",
      "storage",
      "hub-cluster",
    ]);
  });

  it("adds OSAC step for caas flavor", () => {
    const subs = buildConfigSubSteps(new Set(["caas"]), []);
    const ids = subs.map((s) => s.id);
    expect(ids).toContain("osac");
    expect(ids).toContain("caas");
  });

  it("adds OSAC and gpu-ai steps for vmaas flavor", () => {
    const subs = buildConfigSubSteps(new Set(["vmaas"]), []);
    const ids = subs.map((s) => s.id);
    expect(ids).toContain("osac");
    expect(ids).toContain("gpu-ai");
  });

  it("adds AAP step when aap plugin is enabled and no OSAC flavors", () => {
    const subs = buildConfigSubSteps(new Set(), ["aap"]);
    expect(subs.map((s) => s.id)).toContain("aap");
  });

  it("does NOT add standalone AAP step when OSAC flavor is selected", () => {
    const subs = buildConfigSubSteps(new Set(["caas"]), ["aap"]);
    // OSAC step is present, standalone AAP is not
    const ids = subs.map((s) => s.id);
    expect(ids).toContain("osac");
    expect(ids).not.toContain("aap");
  });

  it("adds trust-manager step when plugin enabled and no OSAC", () => {
    const subs = buildConfigSubSteps(new Set(), ["trust-manager"]);
    expect(subs.map((s) => s.id)).toContain("trust-manager");
  });

  it("does NOT add trust-manager step when OSAC flavor selected", () => {
    const subs = buildConfigSubSteps(new Set(["caas"]), ["trust-manager"]);
    expect(subs.map((s) => s.id)).not.toContain("trust-manager");
  });

  it("includes both caas and gpu-ai for caas + vmaas", () => {
    const subs = buildConfigSubSteps(new Set(["caas", "vmaas"]), []);
    const ids = subs.map((s) => s.id);
    expect(ids).toContain("osac");
    expect(ids).toContain("gpu-ai");
    expect(ids).toContain("caas");
  });

  it("base steps always come first", () => {
    const subs = buildConfigSubSteps(new Set(["caas", "vmaas"]), []);
    expect(subs[0].id).toBe("landing-zone");
    expect(subs[1].id).toBe("storage");
    expect(subs[2].id).toBe("hub-cluster");
  });
});
