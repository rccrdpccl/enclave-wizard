package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		input string
		want  slog.Level
	}{
		{"trace", LevelTrace},
		{"TRACE", LevelTrace},
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
	}
	for _, c := range cases {
		got := parseLevel(c.input)
		if got != c.want {
			t.Errorf("parseLevel(%q): want %v, got %v", c.input, c.want, got)
		}
	}
}

func TestInit_DoesNotPanic(t *testing.T) {
	for _, level := range []string{"trace", "debug", "info", "warn", "error", ""} {
		Init(level)
	}
}

func TestTrace_DoesNotPanic(t *testing.T) {
	Init("trace")
	Trace(context.Background(), "test message", "key", "value")
}

func TestReplaceAttr_TraceLevel(t *testing.T) {
	a := replaceAttr(nil, slog.Any(slog.LevelKey, LevelTrace))
	if a.Value.String() != "TRACE" {
		t.Errorf("expected TRACE, got %q", a.Value.String())
	}
}

func TestReplaceAttr_NonTraceLevel(t *testing.T) {
	a := replaceAttr(nil, slog.Any(slog.LevelKey, slog.LevelInfo))
	if a.Value.String() == "TRACE" {
		t.Error("non-trace level should not be replaced with TRACE")
	}
}

func TestReplaceAttr_NonLevelKey(t *testing.T) {
	original := slog.String("msg", "hello")
	got := replaceAttr(nil, original)
	if got.Key != original.Key || got.Value.String() != original.Value.String() {
		t.Error("non-level attributes should pass through unchanged")
	}
}
