package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"
)

// StreamHandler serves SSE streams for task runs. It is registered as a raw
// http.Handler rather than through huma because SSE connections are long-lived
// and do not fit huma's request/response model.
type StreamHandler struct {
	runner runner.Runner
}

// NewStreamHandler creates a new SSE stream handler.
func NewStreamHandler(runner runner.Runner) *StreamHandler {
	return &StreamHandler{runner: runner}
}

// Register adds the SSE stream endpoint to the mux.
func (h *StreamHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/tasks/{id}/stream", h.serveHTTP)
}

func (h *StreamHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		// Fallback: extract from URL path for Go < 1.22 compatibility.
		id = extractTaskID(r.URL.Path)
	}
	if id == "" {
		http.Error(w, `{"title":"Bad Request","status":400,"detail":"missing task id"}`, http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"title":"Internal Server Error","status":500,"detail":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	events, err := h.runner.Stream(id)
	if err != nil {
		if isNotFound(err) {
			http.Error(w, `{"title":"Not Found","status":404,"detail":"run not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"title":"Internal Server Error","status":500,"detail":"stream error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			slog.Debug("SSE client disconnected", "task_id", id)
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, ev.Data)
			flusher.Flush()
		}
	}
}

// extractTaskID extracts the task ID from a path like /api/v1/tasks/{id}/stream.
func extractTaskID(path string) string {
	const prefix = "/api/v1/tasks/"
	const suffix = "/stream"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	trimmed := path[len(prefix):]
	if len(trimmed) <= len(suffix) {
		return ""
	}
	id := trimmed[:len(trimmed)-len(suffix)]
	if id == "" || strings.Contains(id, "/") {
		return ""
	}
	return id
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
