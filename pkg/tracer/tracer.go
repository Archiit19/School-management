package tracer

import (
	"context"
	"fmt"
	"time"

	"github.com/Archiit19/School-management/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// ShutdownFunc flushes and shuts down the tracer provider.
type ShutdownFunc func(context.Context) error

var globalProvider trace.TracerProvider = noop.NewTracerProvider()
var globalConfig Config

// InitFromEnv configures process-wide tracing from environment variables.
func InitFromEnv(service string) (ShutdownFunc, error) {
	return Init(LoadConfigFromEnv(service))
}

// Init configures process-wide tracing and returns a shutdown function.
// When tracing is disabled, a noop provider is installed and Shutdown is a no-op.
func Init(cfg Config) (ShutdownFunc, error) {
	globalConfig = cfg

	if !cfg.Enabled {
		SetProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	exporterOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		exporterOpts = append(exporterOpts, otlptracehttp.WithInsecure())
	}
	exporter, err := otlptracehttp.New(context.Background(), exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("tracer: create otlp exporter: %w", err)
	}

	attrs := []attribute.KeyValue{semconv.ServiceName(cfg.Service)}
	if cfg.Version != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.Version))
	}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(cfg.Environment))
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("tracer: create resource: %w", err)
	}

	batcherOpts := []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithBatchTimeout(cfg.BatchTimeout),
		sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
		sdktrace.WithMaxQueueSize(cfg.MaxQueueSize),
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, batcherOpts...),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	SetProvider(provider)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Error("otel export error", logger.Err(err))
	}))

	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return provider.Shutdown(ctx)
	}
	return shutdown, nil
}

// SetProvider replaces the process-wide tracer provider and propagator.
func SetProvider(provider trace.TracerProvider) {
	if provider == nil {
		provider = noop.NewTracerProvider()
	}
	globalProvider = provider
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

// Provider returns the process-wide tracer provider.
func Provider() trace.TracerProvider {
	return globalProvider
}

// Tracer returns a named tracer from the process-wide provider.
func Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return globalProvider.Tracer(name, opts...)
}

// ActiveConfig returns the active tracer configuration.
func ActiveConfig() Config {
	return globalConfig
}

// Enabled reports whether tracing is active.
func Enabled() bool {
	return globalConfig.Enabled
}

// StartSpan starts a span on the process-wide tracer.
func StartSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer(tracerName).Start(ctx, spanName, opts...)
}
