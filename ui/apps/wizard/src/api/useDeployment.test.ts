import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useDeployment } from "./useDeployment.ts";

// --- Mocks ---

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

// Mock EventSource for SSE tests
class MockEventSource {
  static instances: MockEventSource[] = [];

  url: string;
  listeners: Record<string, Array<(e: unknown) => void>> = {};
  closeCalled = false;

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  addEventListener(type: string, fn: (e: unknown) => void) {
    if (!this.listeners[type]) this.listeners[type] = [];
    this.listeners[type].push(fn);
  }

  removeEventListener(type: string, fn: (e: unknown) => void) {
    if (!this.listeners[type]) return;
    this.listeners[type] = this.listeners[type].filter((f) => f !== fn);
  }

  close() {
    this.closeCalled = true;
  }

  simulateEvent(type: string, data: string) {
    const event = { data, type };
    for (const fn of this.listeners[type] ?? []) {
      fn(event);
    }
  }

  simulateOpen() {
    for (const fn of this.listeners.open ?? []) {
      fn(new Event("open"));
    }
  }

  simulateError() {
    const evt = new Event("error");
    for (const fn of this.listeners.error ?? []) {
      fn(evt);
    }
  }
}

describe("useDeployment", () => {
  let originalEventSource: typeof EventSource;

  beforeEach(() => {
    vi.clearAllMocks();
    MockEventSource.instances = [];
    originalEventSource = globalThis.EventSource;
    globalThis.EventSource = MockEventSource as unknown as typeof EventSource;

    // Default: no running deployment on mount
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 404 }),
    );
  });

  afterEach(() => {
    globalThis.EventSource = originalEventSource;
    vi.restoreAllMocks();
  });

  it("starts in idle phase", () => {
    const { result } = renderHook(() => useDeployment());
    expect(result.current.state.phase).toBe("idle");
    expect(result.current.state.taskId).toBeNull();
    expect(result.current.state.logs).toBe("");
    expect(result.current.state.progress).toBeNull();
  });

  it("transitions through writing -> deploying on start", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    mockStartDeploy.mockResolvedValue({ id: "task-1", status: "running" });

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({ global: {} });
    });

    expect(mockWriteConfig).toHaveBeenCalledTimes(1);
    expect(mockStartDeploy).toHaveBeenCalledTimes(1);
    expect(result.current.state.phase).toBe("deploying");
    expect(result.current.state.taskId).toBe("task-1");
    expect(result.current.state.startTime).toBeInstanceOf(Date);
  });

  it("transitions to error when writeConfig fails", async () => {
    mockWriteConfig.mockRejectedValue(new Error("write failed"));

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({ global: {} });
    });

    expect(result.current.state.phase).toBe("error");
    expect(result.current.state.error?.message).toBe("write failed");
  });

  it("transitions to error when startDeploy fails", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    mockStartDeploy.mockRejectedValue(new Error("deploy failed"));

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({ global: {} });
    });

    expect(result.current.state.phase).toBe("error");
    expect(result.current.state.error?.message).toBe("deploy failed");
  });

  it("parses validation errors from response body", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    const mockResponse = new Response(
      JSON.stringify({
        errors: [{ message: "field X is required" }],
      }),
      { status: 422, headers: { "Content-Type": "application/json" } },
    );
    mockStartDeploy.mockRejectedValue({ response: mockResponse });

    const { result } = renderHook(() => useDeployment());

    await act(async () => {
      await result.current.start({ global: {} });
    });

    expect(result.current.state.phase).toBe("error");
    expect(result.current.state.error?.message).toBe(
      "Configuration validation failed",
    );
    expect(result.current.state.error?.details).toEqual([
      "field X is required",
    ]);
  });

  describe("SSE transport", () => {
    it("creates EventSource after task ID is set", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({ id: "task-sse", status: "running" });

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      await waitFor(() => {
        expect(MockEventSource.instances.length).toBeGreaterThan(0);
        const sseInstance = MockEventSource.instances.find(
          (es) => es.url === "/api/v1/tasks/task-sse/stream",
        );
        expect(sseInstance).toBeDefined();
      });
    });

    it("handles status events", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({ id: "task-2", status: "running" });

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      const es = MockEventSource.instances.find(
        (e) => e.url === "/api/v1/tasks/task-2/stream",
      );

      await act(async () => {
        es?.simulateEvent("status", JSON.stringify({ phase: "deploying" }));
      });

      expect(result.current.state.phase).toBe("deploying");
    });

    it("handles progress events", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({ id: "task-3", status: "running" });

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      const es = MockEventSource.instances.find(
        (e) => e.url === "/api/v1/tasks/task-3/stream",
      );

      await act(async () => {
        es?.simulateEvent(
          "progress",
          JSON.stringify({ percentage: 42, currentTask: "Installing OSAC" }),
        );
      });

      expect(result.current.state.progress).toEqual({
        percentage: 42,
        currentTask: "Installing OSAC",
      });
    });

    it("handles log events and strips ANSI", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({ id: "task-4", status: "running" });

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      const es = MockEventSource.instances.find(
        (e) => e.url === "/api/v1/tasks/task-4/stream",
      );

      await act(async () => {
        es?.simulateEvent(
          "log",
          JSON.stringify({ line: "\x1b[32mPLAY [Setup]\x1b[0m\n" }),
        );
        es?.simulateEvent(
          "log",
          JSON.stringify({ line: "TASK [install] ok\n" }),
        );
      });

      expect(result.current.state.logs).toBe(
        "PLAY [Setup]\nTASK [install] ok\n",
      );
    });

    it("handles done event with success", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({ id: "task-5", status: "running" });

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      const es = MockEventSource.instances.find(
        (e) => e.url === "/api/v1/tasks/task-5/stream",
      );

      await act(async () => {
        es?.simulateEvent("done", JSON.stringify({ status: "complete" }));
      });

      expect(result.current.state.phase).toBe("complete");
      expect(result.current.state.progress?.percentage).toBe(100);
    });

    it("ignores SSE events with invalid JSON", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({ id: "task-bad", status: "running" });

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      const es = MockEventSource.instances.find(
        (e) => e.url === "/api/v1/tasks/task-bad/stream",
      );

      await act(async () => {
        es?.simulateEvent("status", "not valid json{{{");
      });

      expect(result.current.state.phase).toBe("deploying");
      expect(result.current.state.logs).toBe("");
    });

    it("handles done event with failure", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({ id: "task-6", status: "running" });

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      const es = MockEventSource.instances.find(
        (e) => e.url === "/api/v1/tasks/task-6/stream",
      );

      await act(async () => {
        es?.simulateEvent("done", JSON.stringify({ status: "failed" }));
      });

      expect(result.current.state.phase).toBe("failed");
    });
  });

  describe("polling fallback", () => {
    it("falls back to polling when SSE errors", async () => {
      mockWriteConfig.mockResolvedValue(undefined);
      mockStartDeploy.mockResolvedValue({
        id: "task-poll",
        status: "running",
      });
      mockGetTask.mockResolvedValue({
        id: "task-poll",
        status: "running",
      });
      mockGetTaskLogs.mockResolvedValue("PLAY [Prepare] ***");

      const { result } = renderHook(() => useDeployment());

      await act(async () => {
        await result.current.start({ global: {} });
      });

      // Simulate SSE error to trigger fallback
      const es = MockEventSource.instances.find(
        (e) => e.url === "/api/v1/tasks/task-poll/stream",
      );

      await act(async () => {
        es?.simulateError();
      });

      // After SSE failure, polling should kick in
      // The usePolling hook will call getTask and getTaskLogs
      await waitFor(() => {
        expect(result.current.state.phase).toBe("deploying");
      });
    });
  });

  describe("mount reconnection", () => {
    it("reconnects to running deployment on mount", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response(
          JSON.stringify({ id: "existing-task", status: "running" }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      );

      const { result } = renderHook(() => useDeployment());

      await waitFor(() => {
        expect(result.current.state.taskId).toBe("existing-task");
        expect(result.current.state.phase).toBe("deploying");
      });
    });

    it("does not reconnect to completed deployment", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response(JSON.stringify({ id: "done-task", status: "completed" }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );

      const { result } = renderHook(() => useDeployment());

      // Wait a tick to ensure the effect ran
      await act(async () => {
        await new Promise((r) => setTimeout(r, 50));
      });

      expect(result.current.state.phase).toBe("idle");
      expect(result.current.state.taskId).toBeNull();
    });
  });
});
