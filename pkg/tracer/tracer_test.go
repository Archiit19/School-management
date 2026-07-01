package tracer

import (
	"context"
	"net/http"
	"testing"
)

func TestInitDisabledUsesNoopProvider(t *testing.T) {
	shutdown, err := Init(Config{Service: "test-service", Enabled: false})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if shutdown == nil {
		t.Fatal("Init() returned nil shutdown")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown() error = %v", err)
	}

	_, span := StartSpan(context.Background(), "test", "noop-span")
	defer span.End()
	if span == nil {
		t.Fatal("expected non-nil span")
	}
}

func TestLoadConfigFromEnvDefaults(t *testing.T) {
	t.Setenv("TRACE_ENABLED", "")
	t.Setenv("OTEL_SDK_DISABLED", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "")
	t.Setenv("TRACE_SAMPLE_RATIO", "")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "")

	cfg := LoadConfigFromEnv("user-service")
	if cfg.Service != "user-service" {
		t.Fatalf("Service = %q, want user-service", cfg.Service)
	}
	if cfg.Enabled {
		t.Fatal("Enabled = true, want false")
	}
	if cfg.Endpoint != "localhost:4318" {
		t.Fatalf("Endpoint = %q, want localhost:4318", cfg.Endpoint)
	}
	if cfg.Insecure {
		t.Fatal("Insecure = true, want false")
	}
	if cfg.SampleRatio != defaultSampleRatio {
		t.Fatalf("SampleRatio = %v, want %v", cfg.SampleRatio, defaultSampleRatio)
	}
}

func TestLoadConfigFromEnvOTELExporterEnablesTracing(t *testing.T) {
	t.Setenv("TRACE_ENABLED", "")
	t.Setenv("OTEL_SDK_DISABLED", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")

	cfg := LoadConfigFromEnv("user-service")
	if !cfg.Enabled {
		t.Fatal("Enabled = false, want true when OTEL_TRACES_EXPORTER=otlp")
	}
}

func TestGinMiddlewareDisabledIsNoop(t *testing.T) {
	shutdown, err := Init(Config{Service: "test", Enabled: false})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer shutdown(context.Background())

	mw := GinMiddleware("test")
	if mw == nil {
		t.Fatal("GinMiddleware() returned nil")
	}
}

func TestWrapTransportDisabledReturnsBase(t *testing.T) {
	shutdown, err := Init(Config{Service: "test", Enabled: false})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer shutdown(context.Background())

	base := http.DefaultTransport
	got := WrapTransport(base)
	if got != base {
		t.Fatal("WrapTransport() should return base transport when disabled")
	}
}
