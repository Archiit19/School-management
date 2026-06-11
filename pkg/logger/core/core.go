package core

// Logger is the backend-agnostic logging contract.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	Sync() error
}

// Backend identifies which logging implementation to use.
type Backend string

const (
	BackendSlog    Backend = "slog"
	BackendZap     Backend = "zap"
	BackendZerolog Backend = "zerolog"
)
