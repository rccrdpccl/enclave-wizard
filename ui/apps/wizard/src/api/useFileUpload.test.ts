import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useFileUpload } from "./useFileUpload.ts";

vi.mock("../auth/AuthContext.tsx", () => ({
  useAuth: () => ({ token: "test-token" }),
}));

describe("useFileUpload", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("starts in idle state", () => {
    const { result } = renderHook(() => useFileUpload());
    expect(result.current.uploading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it("uploads a file successfully", async () => {
    const mockResponse = { path: "/uploads/test.zip" };
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify(mockResponse), { status: 200 }),
    );

    const { result } = renderHook(() => useFileUpload());
    const file = new File(["content"], "test.zip", { type: "application/zip" });

    let uploadResult: { path: string } | undefined;
    await act(async () => {
      uploadResult = await result.current.upload(file, "plugins");
    });

    expect(uploadResult?.path).toBe("/uploads/test.zip");
    expect(result.current.uploading).toBe(false);
    expect(result.current.error).toBeNull();

    // Verify fetch was called with correct params
    const fetchCall = vi.mocked(globalThis.fetch).mock.calls[0];
    expect(fetchCall[0]).toBe("/api/v1/files");
    expect((fetchCall[1] as RequestInit).method).toBe("POST");
    expect(
      (fetchCall[1] as RequestInit).headers as Record<string, string>,
    ).toEqual({ Authorization: "Bearer test-token" });
  });

  it("sets error on upload failure", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response("Server error", { status: 500 }),
    );

    const { result } = renderHook(() => useFileUpload());
    const file = new File(["content"], "test.zip");

    await act(async () => {
      try {
        await result.current.upload(file, "plugins");
      } catch {
        // expected
      }
    });

    expect(result.current.uploading).toBe(false);
    expect(result.current.error).toBe("Server error");
  });

  it("clears error", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response("fail", { status: 500 }),
    );

    const { result } = renderHook(() => useFileUpload());
    const file = new File(["content"], "test.zip");

    await act(async () => {
      try {
        await result.current.upload(file, "plugins");
      } catch {
        // expected
      }
    });

    expect(result.current.error).not.toBeNull();

    act(() => {
      result.current.clearError();
    });

    expect(result.current.error).toBeNull();
  });
});
