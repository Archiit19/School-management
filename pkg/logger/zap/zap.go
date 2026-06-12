package zap

import (
	"os"
	"strings"

	"github.com/Archiit19/School-management/pkg/logger/core"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	core.RegisterBackend(core.BackendZap, New)
}

type zapLogger struct {
	inner *zap.Logger
}

// New builds a zap-backed core.Logger.
func New(cfg core.Config) (core.Logger, error) {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	level := parseLevel(cfg.Level)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if strings.EqualFold(cfg.Format, "json") {
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	zapCore := zapcore.NewCore(encoder, zapcore.AddSync(out), level)

	opts := []zap.Option{}
	if level == zapcore.DebugLevel {
		opts = append(opts, zap.AddCaller())
	}

	root := zap.New(zapCore, opts...)
	if cfg.Service != "" {
		root = root.With(zap.String("service", cfg.Service))
	}
	return &zapLogger{inner: root}, nil
}

func (l *zapLogger) Debug(msg string, fields ...core.Field) {
	l.inner.Debug(msg, toZapFields(fields)...)
}

func (l *zapLogger) Info(msg string, fields ...core.Field) {
	l.inner.Info(msg, toZapFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...core.Field) {
	l.inner.Warn(msg, toZapFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...core.Field) {
	l.inner.Error(msg, toZapFields(fields)...)
}

func (l *zapLogger) With(fields ...core.Field) core.Logger {
	return &zapLogger{inner: l.inner.With(toZapFields(fields)...)}
}

func (l *zapLogger) Sync() error {
	return l.inner.Sync()
}

func toZapFields(fields []core.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	out := make([]zap.Field, len(fields))
	for i, f := range fields {
		if f.Key == "error" {
			if err, ok := f.Val.(error); ok {
				out[i] = zap.Error(err)
				continue
			}
		}
		out[i] = zap.Any(f.Key, f.Val)
	}
	return out
}

func parseLevel(raw string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
