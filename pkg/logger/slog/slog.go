package slog

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/Archiit19/School-management/pkg/logger/core"
)

func init() {
	core.RegisterBackend(core.BackendSlog, New)
}

type slogLogger struct {
	inner *slog.Logger
}

// New builds a slog-backed core.Logger.
func New(cfg core.Config) (core.Logger, error) {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "json") {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	root := slog.New(handler)
	if cfg.Service != "" {
		root = root.With("service", cfg.Service)
	}
	return &slogLogger{inner: root}, nil
}

func (l *slogLogger) Debug(msg string, fields ...core.Field) {
	l.inner.Log(nil, slog.LevelDebug, msg, toAttrs(fields)...)
}

func (l *slogLogger) Info(msg string, fields ...core.Field) {
	l.inner.Info(msg, toAttrs(fields)...)
}

func (l *slogLogger) Warn(msg string, fields ...core.Field) {
	l.inner.Warn(msg, toAttrs(fields)...)
}

func (l *slogLogger) Error(msg string, fields ...core.Field) {
	l.inner.Error(msg, toAttrs(fields)...)
}

func (l *slogLogger) With(fields ...core.Field) core.Logger {
	return &slogLogger{inner: l.inner.With(toAttrs(fields)...)}
}

func (l *slogLogger) Sync() error { return nil }

func toAttrs(fields []core.Field) []any {
	if len(fields) == 0 {
		return nil
	}
	attrs := make([]any, 0, len(fields)*2)
	for _, f := range fields {
		attrs = append(attrs, f.Key, f.Val)
	}
	return attrs
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
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

var _ io.Writer = os.Stdout
