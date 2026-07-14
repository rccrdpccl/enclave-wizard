import { useEffect, useRef, useState } from "react";

export interface SSEEvent {
  type: string;
  data: string;
}

export interface UseSSEOptions {
  onEvent?: (event: SSEEvent) => void;
  onError?: (error: Event) => void;
}

export interface UseSSEReturn {
  connected: boolean;
  error: Error | null;
}

export function useSSE(
  url: string | null,
  options: UseSSEOptions,
): UseSSEReturn {
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const optionsRef = useRef(options);
  optionsRef.current = options;

  useEffect(() => {
    if (!url) {
      setConnected(false);
      setError(null);
      return;
    }

    const eventSource = new EventSource(url);

    const handleOpen = () => {
      setConnected(true);
      setError(null);
    };

    const handleError = (evt: Event) => {
      setConnected(false);
      setError(new Error("SSE connection error"));
      optionsRef.current.onError?.(evt);
    };

    const eventTypes = ["status", "progress", "log", "done"];
    const listeners: Array<[string, (e: MessageEvent) => void]> = [];

    for (const type of eventTypes) {
      const listener = (e: MessageEvent) => {
        optionsRef.current.onEvent?.({ type, data: e.data });
      };
      eventSource.addEventListener(type, listener);
      listeners.push([type, listener]);
    }

    eventSource.addEventListener("open", handleOpen);
    eventSource.addEventListener("error", handleError);

    return () => {
      eventSource.removeEventListener("open", handleOpen);
      eventSource.removeEventListener("error", handleError);
      for (const [type, listener] of listeners) {
        eventSource.removeEventListener(type, listener);
      }
      eventSource.close();
      setConnected(false);
    };
  }, [url]);

  return { connected, error };
}
