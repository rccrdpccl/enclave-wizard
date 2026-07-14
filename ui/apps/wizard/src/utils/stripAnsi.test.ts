import { describe, expect, it } from "vitest";
import { stripAnsi } from "./stripAnsi.ts";

describe("stripAnsi", () => {
  it("removes ANSI color codes", () => {
    expect(stripAnsi("\x1b[32mgreen\x1b[0m")).toBe("green");
  });

  it("removes multiple ANSI sequences", () => {
    expect(stripAnsi("\x1b[1m\x1b[31merror\x1b[0m: something failed")).toBe(
      "error: something failed",
    );
  });

  it("removes carriage returns", () => {
    expect(stripAnsi("line1\r\nline2\r")).toBe("line1\nline2");
  });

  it("handles mixed ANSI codes and carriage returns", () => {
    expect(stripAnsi("\x1b[33mwarn\x1b[0m\r\n\x1b[31merror\x1b[0m\r")).toBe(
      "warn\nerror",
    );
  });

  it("returns clean strings unchanged", () => {
    expect(stripAnsi("no ansi here")).toBe("no ansi here");
  });

  it("handles empty string", () => {
    expect(stripAnsi("")).toBe("");
  });

  it("handles ANSI codes with multiple parameters", () => {
    expect(stripAnsi("\x1b[38;5;196mred\x1b[0m")).toBe("red");
  });
});
