import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useSSE } from "./useSSE.ts";

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
    if (!this.listeners[type]) {
      this.listeners[type] = [];
    }
    this.listeners[type].push(fn);
  }

  removeEventListener(type: string, fn: (e: unknown) => void) {
    if (!this.listeners[type]) return;
    this.listeners[type] = this.listeners[type].filter((f) => f !== fn);
  }

  close() {
    this.closeCalled = true;
  }

  // Test helpers
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

describe("useSSE", () => {
  let originalEventSource: typeof EventSource;

  beforeEach(() => {
    MockEventSource.instances = [];
    originalEventSource = globalThis.EventSource;
    globalThis.EventSource = MockEventSource as unknown as typeof EventSource;
  });

  afterEach(() => {
    globalThis.EventSource = originalEventSource;
  });

  it("creates EventSource with correct URL", () => {
    renderHook(() => useSSE("/api/v1/tasks/123/stream", {}));

    expect(MockEventSource.instances).toHaveLength(1);
    expect(MockEventSource.instances[0].url).toBe("/api/v1/tasks/123/stream");
  });

  it("does not create EventSource when URL is null", () => {
    renderHook(() => useSSE(null, {}));

    expect(MockEventSource.instances).toHaveLength(0);
  });

  it("dispatches events to onEvent callback", () => {
    const onEvent = vi.fn();
    renderHook(() => useSSE("/api/v1/tasks/123/stream", { onEvent }));

    const es = MockEventSource.instances[0];
    act(() => {
      es.simulateEvent("status", '{"phase":"deploying"}');
    });

    expect(onEvent).toHaveBeenCalledWith({
      type: "status",
      data: '{"phase":"deploying"}',
    });
  });

  it("dispatches different event types", () => {
    const onEvent = vi.fn();
    renderHook(() => useSSE("/api/v1/tasks/123/stream", { onEvent }));

    const es = MockEventSource.instances[0];
    act(() => {
      es.simulateEvent("progress", '{"percentage":50}');
      es.simulateEvent("log", '{"line":"hello"}');
      es.simulateEvent("done", '{"status":"complete"}');
    });

    expect(onEvent).toHaveBeenCalledTimes(3);
    expect(onEvent).toHaveBeenCalledWith({
      type: "progress",
      data: '{"percentage":50}',
    });
    expect(onEvent).toHaveBeenCalledWith({
      type: "log",
      data: '{"line":"hello"}',
    });
    expect(onEvent).toHaveBeenCalledWith({
      type: "done",
      data: '{"status":"complete"}',
    });
  });

  it("sets connected to true on open", () => {
    const { result } = renderHook(() => useSSE("/api/v1/tasks/123/stream", {}));

    expect(result.current.connected).toBe(false);

    const es = MockEventSource.instances[0];
    act(() => {
      es.simulateOpen();
    });

    expect(result.current.connected).toBe(true);
  });

  it("sets error on error event", () => {
    const onError = vi.fn();
    const { result } = renderHook(() =>
      useSSE("/api/v1/tasks/123/stream", { onError }),
    );

    const es = MockEventSource.instances[0];
    act(() => {
      es.simulateError();
    });

    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe("SSE connection error");
    expect(result.current.connected).toBe(false);
    expect(onError).toHaveBeenCalled();
  });

  it("closes EventSource on unmount", () => {
    const { unmount } = renderHook(() =>
      useSSE("/api/v1/tasks/123/stream", {}),
    );

    const es = MockEventSource.instances[0];
    expect(es.closeCalled).toBe(false);

    unmount();

    expect(es.closeCalled).toBe(true);
  });

  it("closes old EventSource when URL changes", () => {
    const { rerender } = renderHook(
      ({ url }: { url: string | null }) => useSSE(url, {}),
      { initialProps: { url: "/api/v1/tasks/1/stream" } },
    );

    const firstEs = MockEventSource.instances[0];

    rerender({ url: "/api/v1/tasks/2/stream" });

    expect(firstEs.closeCalled).toBe(true);
    expect(MockEventSource.instances).toHaveLength(2);
    expect(MockEventSource.instances[1].url).toBe("/api/v1/tasks/2/stream");
  });

  it("closes EventSource when URL becomes null", () => {
    const { rerender, result } = renderHook(
      ({ url }: { url: string | null }) => useSSE(url, {}),
      { initialProps: { url: "/api/v1/tasks/1/stream" as string | null } },
    );

    const es = MockEventSource.instances[0];

    rerender({ url: null });

    expect(es.closeCalled).toBe(true);
    expect(result.current.connected).toBe(false);
    expect(result.current.error).toBeNull();
  });
});
