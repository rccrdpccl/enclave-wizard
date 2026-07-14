import { describe, expect, it, vi } from "vitest";
import { formatDuration } from "./formatDuration.ts";

describe("formatDuration", () => {
  it("returns dash when start is null", () => {
    expect(formatDuration(null)).toBe("—");
  });

  it("returns dash when start is undefined", () => {
    expect(formatDuration(undefined)).toBe("—");
  });

  it("formats seconds only", () => {
    const start = new Date("2024-01-01T00:00:00Z");
    const end = new Date("2024-01-01T00:00:42Z");
    expect(formatDuration(start, end)).toBe("42s");
  });

  it("formats zero seconds", () => {
    const start = new Date("2024-01-01T00:00:00Z");
    const end = new Date("2024-01-01T00:00:00Z");
    expect(formatDuration(start, end)).toBe("0s");
  });

  it("formats minutes and seconds", () => {
    const start = new Date("2024-01-01T00:00:00Z");
    const end = new Date("2024-01-01T00:03:15Z");
    expect(formatDuration(start, end)).toBe("3m 15s");
  });

  it("formats hours and minutes", () => {
    const start = new Date("2024-01-01T00:00:00Z");
    const end = new Date("2024-01-01T02:30:45Z");
    expect(formatDuration(start, end)).toBe("2h 30m");
  });

  it("formats exactly one hour", () => {
    const start = new Date("2024-01-01T00:00:00Z");
    const end = new Date("2024-01-01T01:00:00Z");
    expect(formatDuration(start, end)).toBe("1h 0m");
  });

  it("uses current time when end is not provided", () => {
    const now = new Date("2024-01-01T00:05:00Z");
    vi.setSystemTime(now);
    const start = new Date("2024-01-01T00:02:00Z");
    expect(formatDuration(start)).toBe("3m 0s");
    vi.useRealTimers();
  });
});
