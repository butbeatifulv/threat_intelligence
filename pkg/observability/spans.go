package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts a child span on the global tracer.
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return otel.Tracer("veil").Start(ctx, name, trace.WithAttributes(attrs...))
}

// StartWorkerSpan starts a root/child span for async worker processing.
func StartWorkerSpan(ctx context.Context, worker, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	all := append([]attribute.KeyValue{attribute.String("veil.worker", worker)}, attrs...)
	return StartSpan(ctx, name, all...)
}

// RecordError marks span status and records the error.
func RecordError(span trace.Span, err error) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SpanFromContext returns the current span.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}
