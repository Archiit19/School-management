package logger

import (
	"io"
	"log/slog"
	"os"
)

var defaultLogger Logger = &noopLogger{}

// InitFromEnv configures the process-wide logger from environment variables.
func InitFromEnv(service string) (Logger, error) {
	return Init(LoadConfigFromEnv(service))
}

// Init configures the process-wide logger and returns it.
func Init(cfg Config) (Logger, error) {
	l, err := New(cfg)
	if err != nil {
		return nil, err
	}
	SetDefault(l)
	return l, nil
}

// SetDefault replaces the process-wide logger.
func SetDefault(l Logger) {
	if l == nil {
		defaultLogger = &noopLogger{}
		return
	}
	defaultLogger = l
}

// L returns the process-wide logger (use when you need the Logger value, e.g. to pass around).
func L() Logger {
	return defaultLogger
}

// With returns a child of the process-wide logger with bound fields.
func With(fields ...Field) Logger {
	return defaultLogger.With(fields...)
}

// Debug logs at debug level using the process-wide logger.
func Debug(msg string, fields ...Field) { defaultLogger.Debug(msg, fields...) }

// Info logs at info level using the process-wide logger.
func Info(msg string, fields ...Field) { defaultLogger.Info(msg, fields...) }

// Warn logs at warn level using the process-wide logger.
func Warn(msg string, fields ...Field) { defaultLogger.Warn(msg, fields...) }

// Error logs at error level using the process-wide logger.
func Error(msg string, fields ...Field) { defaultLogger.Error(msg, fields...) }

// Fatal logs at error level and exits with code 1.
func Fatal(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
	os.Exit(1)
}

// NopWriter discards all written bytes (useful in tests).
var NopWriter io.Writer = io.Discard

// Discard returns a logger that writes nowhere.
func Discard() Logger {
	return &discardLogger{}
}

type noopLogger struct{}

func (n *noopLogger) Debug(msg string, fields ...Field) { n.log(slog.LevelDebug, msg, fields) }
func (n *noopLogger) Info(msg string, fields ...Field)  { n.log(slog.LevelInfo, msg, fields) }
func (n *noopLogger) Warn(msg string, fields ...Field)  { n.log(slog.LevelWarn, msg, fields) }
func (n *noopLogger) Error(msg string, fields ...Field) { n.log(slog.LevelError, msg, fields) }
func (n *noopLogger) With(fields ...Field) Logger       { return n }
func (n *noopLogger) Sync() error                         { return nil }

func (n *noopLogger) log(level slog.Level, msg string, fields []Field) {
	attrs := make([]any, 0, len(fields)*2)
	for _, f := range fields {
		attrs = append(attrs, f.Key, f.Val)
	}
	slog.Default().Log(nil, level, msg, attrs...)
}

type discardLogger struct{}

func (d *discardLogger) Debug(string, ...Field) {}
func (d *discardLogger) Info(string, ...Field)  {}
func (d *discardLogger) Warn(string, ...Field)  {}
func (d *discardLogger) Error(string, ...Field) {}
func (d *discardLogger) With(...Field) Logger   { return d }
func (d *discardLogger) Sync() error { return nil }
