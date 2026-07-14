package api

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"
	"go.uber.org/mock/gomock"
)

func setupStreamServer(runner tasks.Runner) *httptest.Server {
	mux := http.NewServeMux()
	NewStreamHandler(runner).Register(mux)
	return httptest.NewServer(mux)
}

// parseSSEEvents reads SSE events from the response body.
type sseEvent struct {
	Type string
	Data string
}

func readSSEEvents(t *testing.T, resp *http.Response) []sseEvent {
	t.Helper()
	var events []sseEvent
	scanner := bufio.NewScanner(resp.Body)

	var currentType, currentData string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			currentType = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			currentData = strings.TrimPrefix(line, "data: ")
		case line == "":
			if currentType != "" || currentData != "" {
				events = append(events, sseEvent{Type: currentType, Data: currentData})
				currentType = ""
				currentData = ""
			}
		}
	}
	return events
}

func TestStream_SSEFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := tasks.NewMockRunner(ctrl)

	ch := make(chan tasks.Event, 3)
	ch <- tasks.Event{Type: "status", Data: `{"status":"running"}`}
	ch <- tasks.Event{Type: "log", Data: `{"line":"TASK [Install LVMS] ***"}`}
	ch <- tasks.Event{Type: "done", Data: `{"status":"successful","exitCode":0}`}
	close(ch)

	m.EXPECT().Stream("run-1").Return((<-chan tasks.Event)(ch), nil)

	srv := setupStreamServer(m)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/tasks/run-1/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)
	assertEqual(t, "content-type", "text/event-stream", resp.Header.Get("Content-Type"))
	assertEqual(t, "cache-control", "no-cache", resp.Header.Get("Cache-Control"))

	events := readSSEEvents(t, resp)
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d: %+v", len(events), events)
	}

	assertEqual(t, "event[0].type", "status", events[0].Type)
	assertEqual(t, "event[0].data", `{"status":"running"}`, events[0].Data)
	assertEqual(t, "event[1].type", "log", events[1].Type)
	assertEqual(t, "event[2].type", "done", events[2].Type)
}

func TestStream_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := tasks.NewMockRunner(ctrl)

	m.EXPECT().Stream("nonexistent").Return(nil, tasks.ErrNotFound)

	srv := setupStreamServer(m)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/tasks/nonexistent/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	assertEqual(t, "status", http.StatusNotFound, resp.StatusCode)
}

func TestStream_CompletedTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := tasks.NewMockRunner(ctrl)

	ch := make(chan tasks.Event, 4)
	ch <- tasks.Event{Type: "status", Data: `{"status":"successful"}`}
	ch <- tasks.Event{Type: "log", Data: `{"line":"PLAY RECAP"}`}
	ch <- tasks.Event{Type: "progress", Data: `{"currentTask":"Final task"}`}
	ch <- tasks.Event{Type: "done", Data: `{"status":"successful","exitCode":0}`}
	close(ch)

	m.EXPECT().Stream("completed-1").Return((<-chan tasks.Event)(ch), nil)

	srv := setupStreamServer(m)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/tasks/completed-1/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	events := readSSEEvents(t, resp)
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events (status + done), got %d", len(events))
	}

	// Last event should be "done"
	last := events[len(events)-1]
	assertEqual(t, "last.type", "done", last.Type)
}

func TestStream_ClientDisconnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := tasks.NewMockRunner(ctrl)

	// Channel that will never close (simulates a long-running task).
	ch := make(chan tasks.Event, 1)
	ch <- tasks.Event{Type: "status", Data: `{"status":"running"}`}

	m.EXPECT().Stream("running-1").Return((<-chan tasks.Event)(ch), nil)

	srv := setupStreamServer(m)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/v1/tasks/running-1/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertEqual(t, "status", http.StatusOK, resp.StatusCode)

	// Reading should terminate when the context is canceled.
	events := readSSEEvents(t, resp)
	// We should get at least the first event before disconnect.
	if len(events) < 1 {
		t.Errorf("expected at least 1 event before disconnect, got %d", len(events))
	}
}

func TestStream_MultipleEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := tasks.NewMockRunner(ctrl)

	ch := make(chan tasks.Event, 5)
	ch <- tasks.Event{Type: "status", Data: `{"status":"running","phase":"operators"}`}
	ch <- tasks.Event{Type: "progress", Data: `{"completed":10,"total":350,"percentage":3,"currentTask":"Install LVMS operator"}`}
	ch <- tasks.Event{Type: "log", Data: `{"line":"TASK [Install LVMS operator] ***"}`}
	ch <- tasks.Event{Type: "progress", Data: `{"completed":42,"total":350,"percentage":12,"currentTask":"Configure storage"}`}
	ch <- tasks.Event{Type: "done", Data: `{"status":"successful","exitCode":0}`}
	close(ch)

	m.EXPECT().Stream("run-multi").Return((<-chan tasks.Event)(ch), nil)

	srv := setupStreamServer(m)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/tasks/run-multi/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	events := readSSEEvents(t, resp)
	assertEqual(t, "event count", 5, len(events))
	assertEqual(t, "event[0].type", "status", events[0].Type)
	assertEqual(t, "event[1].type", "progress", events[1].Type)
	assertEqual(t, "event[2].type", "log", events[2].Type)
	assertEqual(t, "event[3].type", "progress", events[3].Type)
	assertEqual(t, "event[4].type", "done", events[4].Type)
}

func TestExtractTaskID(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/api/v1/tasks/abc-123/stream", "abc-123"},
		{"/api/v1/tasks/run-1/stream", "run-1"},
		{"/api/v1/tasks//stream", ""},
		{"/api/v1/tasks/stream", ""},
		{"/other/path", ""},
		{"/api/v1/tasks/a/b/stream", ""},
	}
	for _, tt := range tests {
		got := extractTaskID(tt.path)
		if got != tt.want {
			t.Errorf("extractTaskID(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
