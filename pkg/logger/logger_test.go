package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Archiit19/School-management/pkg/logger"
)

func TestNewSlogJSON(t *testing.T) {
	var buf bytes.Buffer
	l, err := logger.New(logger.Config{
		Service: "test-service",
		Level:   "info",
		Format:  "json",
		Backend: logger.BackendSlog,
		Output:  &buf,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	l.Info("hello", logger.String("key", "value"))
	out := buf.String()
	if !strings.Contains(out, `"msg":"hello"`) {
		t.Fatalf("expected msg in output, got %q", out)
	}
	if !strings.Contains(out, `"service":"test-service"`) {
		t.Fatalf("expected service in output, got %q", out)
	}
	if !strings.Contains(out, `"key":"value"`) {
		t.Fatalf("expected field in output, got %q", out)
	}
}

func TestInitFromEnvDefaultsToSlog(t *testing.T) {
	t.Setenv("LOG_BACKEND", "")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")

	l, err := logger.InitFromEnv("user-service")
	if err != nil {
		t.Fatalf("InitFromEnv: %v", err)
	}
	if l == nil {
		t.Fatal("expected logger")
	}
	child := l.With(logger.String("component", "db"))
	child.Debug("connected")
}

func TestKV(t *testing.T) {
	fields := logger.KV("a", 1, "b", "two")
	if len(fields) != 2 || fields[0].Key != "a" || fields[1].Val != "two" {
		t.Fatalf("unexpected fields: %+v", fields)
	}
}

func TestZapBackendNotImplemented(t *testing.T) {
	_, err := logger.New(logger.Config{
		Backend: logger.BackendZap,
		Output:  logger.NopWriter,
	})
	if err == nil {
		t.Fatal("expected error for zap backend")
	}
}

func TestGlobalInfo(t *testing.T) {
	var buf bytes.Buffer
	_, err := logger.Init(logger.Config{
		Service: "test-service",
		Level:   "info",
		Format:  "json",
		Backend: logger.BackendSlog,
		Output:  &buf,
	})
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("started", logger.String("component", "main"))
	if !strings.Contains(buf.String(), `"msg":"started"`) {
		t.Fatalf("expected global Info output, got %q", buf.String())
	}
}

func TestWithChaining(t *testing.T) {
	var buf bytes.Buffer
	root, err := logger.New(logger.Config{
		Service: "svc",
		Level:   "info",
		Format:  "json",
		Backend: logger.BackendSlog,
		Output:  &buf,
	})
	if err != nil {
		t.Fatal(err)
	}
	root.With(logger.String("request_id", "abc")).Info("done")
	if !strings.Contains(buf.String(), `"request_id":"abc"`) {
		t.Fatalf("expected chained field, got %q", buf.String())
	}
}
