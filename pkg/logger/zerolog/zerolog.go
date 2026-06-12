package zerolog

import (
	"io"
	"os"
	"strings"

	"github.com/Archiit19/School-management/pkg/logger/core"
	"github.com/rs/zerolog"
)

func init() {
	core.RegisterBackend(core.BackendZerolog, New)
}

type zerologLogger struct {
	inner zerolog.Logger
}

// New builds a zerolog-backed core.Logger.
func New(cfg core.Config) (core.Logger, error) {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	writer := outputWriter(out, cfg.Format)
	root := zerolog.New(writer).Level(parseLevel(cfg.Level))
	if cfg.Service != "" {
		root = root.With().Str("service", cfg.Service).Logger()
	}
	return &zerologLogger{inner: root}, nil
}

func outputWriter(out io.Writer, format string) io.Writer {
	if strings.EqualFold(format, "json") {
		return out
	}
	return zerolog.ConsoleWriter{Out: out, TimeFormat: "2006-01-02T15:04:05Z07:00"}
}

func (l *zerologLogger) Debug(msg string, fields ...core.Field) {
	l.emit(l.inner.Debug(), msg, fields...)
}

func (l *zerologLogger) Info(msg string, fields ...core.Field) {
	l.emit(l.inner.Info(), msg, fields...)
}

func (l *zerologLogger) Warn(msg string, fields ...core.Field) {
	l.emit(l.inner.Warn(), msg, fields...)
}

func (l *zerologLogger) Error(msg string, fields ...core.Field) {
	l.emit(l.inner.Error(), msg, fields...)
}

func (l *zerologLogger) With(fields ...core.Field) core.Logger {
	ctx := l.inner.With()
	for _, f := range fields {
		ctx = applyField(ctx, f)
	}
	return &zerologLogger{inner: ctx.Logger()}
}

func (l *zerologLogger) Sync() error { return nil }

func (l *zerologLogger) emit(event *zerolog.Event, msg string, fields ...core.Field) {
	if event == nil {
		return
	}
	for _, f := range fields {
		event = applyEventField(event, f)
	}
	event.Msg(msg)
}

func applyField(ctx zerolog.Context, f core.Field) zerolog.Context {
	if f.Key == "error" {
		if err, ok := f.Val.(error); ok {
			return ctx.AnErr("error", err)
		}
	}
	return ctx.Interface(f.Key, f.Val)
}

func applyEventField(event *zerolog.Event, f core.Field) *zerolog.Event {
	if f.Key == "error" {
		if err, ok := f.Val.(error); ok {
			return event.Err(err)
		}
	}
	return event.Interface(f.Key, f.Val)
}

func parseLevel(raw string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
