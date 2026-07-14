import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useDeployment } from "./useDeployment.ts";

const mockWriteConfig = vi.fn();
const mockStartDeploy = vi.fn();
const mockGetTask = vi.fn();
const mockGetTaskLogs = vi.fn();

vi.mock("./useEnclaveApi.ts", () => ({
  useEnclaveApi: () => ({ writeConfig: mockWriteConfig }),
}));

vi.mock("./useTasksApi.ts", () => ({
  useTasksApi: () => ({
    startDeploy: mockStartDeploy,
    getTask: mockGetTask,
    getTaskLogs: mockGetTaskLogs,
  }),
}));

describe("useDeployment", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("starts in idle phase", () => {
    const { result } = renderHook(() => useDeployment());
    expect(result.current.state.phase).toBe("idle");
    expect(result.current.state.task).toBeNull();
    expect(result.current.state.error).toBeNull();
  });

  it("transitions to writing then deploying on start", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    mockStartDeploy.mockResolvedValue({ id: "task-1", status: "running" });

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({} as never);
    });

    expect(mockWriteConfig).toHaveBeenCalledTimes(1);
    expect(mockStartDeploy).toHaveBeenCalledTimes(1);
    expect(result.current.state.phase).toBe("deploying");
    expect(result.current.state.taskId).toBe("task-1");
  });

  it("transitions to error when writeConfig fails", async () => {
    mockWriteConfig.mockRejectedValue(new Error("write failed"));

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({} as never);
    });

    expect(result.current.state.phase).toBe("error");
    expect(result.current.state.error?.message).toBe("write failed");
  });

  it("transitions to error when startDeploy fails", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    mockStartDeploy.mockRejectedValue(new Error("deploy failed"));

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({} as never);
    });

    expect(result.current.state.phase).toBe("error");
    expect(result.current.state.error?.message).toBe("deploy failed");
  });

  it("parses error details from response body", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    const errorResponse = new Response(
      JSON.stringify({ errors: [{ message: "field X is required" }] }),
      { status: 400 },
    );
    mockStartDeploy.mockRejectedValue({ response: errorResponse });

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({} as never);
    });

    expect(result.current.state.phase).toBe("error");
    expect(result.current.state.error?.message).toBe(
      "Configuration validation failed",
    );
    expect(result.current.state.error?.details).toContain(
      "field X is required",
    );
  });
});
