package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New creates a configured *slog.Logger based on format and level strings.
// Unrecognized format defaults to "json"; unrecognized level defaults to "info".
// Warnings about invalid values are emitted to stderr.
func New(format, level string, w ...io.Writer) *slog.Logger {
	var out io.Writer = os.Stderr
	if len(w) > 0 {
		out = w[0]
	}

	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "text":
		handler = slog.NewTextHandler(out, opts)
	case "json", "":
		handler = slog.NewJSONHandler(out, opts)
	default:
		handler = slog.NewJSONHandler(out, opts)
		// Emit warning about unrecognized format using a temporary logger
		tmp := slog.New(handler)
		tmp.Warn("unrecognized LOG_FORMAT, defaulting to json", "format", format)
	}

	return slog.New(NewSanitizingHandler(handler))
}

// parseLevel converts a level string to slog.Level.
// Returns slog.LevelInfo for unrecognized values.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
