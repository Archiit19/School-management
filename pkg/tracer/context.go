package tracer

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// LogFields returns trace_id and span_id key-value pairs for structured logging.
// Returns nil when no valid span is present in ctx.
func LogFields(ctx context.Context) []any {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return nil
	}
	sc := span.SpanContext()
	if !sc.IsValid() {
		return nil
	}
	return []any{
		"trace_id", sc.TraceID().String(),
		"span_id", sc.SpanID().String(),
	}
}

// RecordError marks the span in ctx as failed when one is recording.
func RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
