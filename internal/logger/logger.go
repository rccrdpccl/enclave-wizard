package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// LevelTrace is a custom log level below DEBUG for per-request trace logging.
const LevelTrace = slog.Level(-8)

// Init configures the global slog logger with the given level string.
// Recognized values: trace, debug, info, warn, error (default: info).
func Init(levelStr string) {
	level := parseLevel(levelStr)
	opts := &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceAttr,
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
}

func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		if level, ok := a.Value.Any().(slog.Level); ok && level == LevelTrace {
			a.Value = slog.StringValue("TRACE")
		}
	}
	return a
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Trace logs at TRACE level using the default logger.
func Trace(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, LevelTrace, msg, args...)
}
