package observability

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// NewLogger returns a slog.Logger with trace_id/span_id and service name in JSON output.
func NewLogger(env, service string, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	level := parseLogLevel(activeConfig.LogLevel)
	format := activeConfig.LogFormat
	if format == "" {
		format = "json"
	}
	if env == "local" && os.Getenv("LOG_FORMAT") == "" {
		format = "text"
	}
	var inner slog.Handler
	switch format {
	case "text":
		inner = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	default:
		inner = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	}
	return slog.New(&traceContextHandler{inner: inner, service: service})
}

type traceContextHandler struct {
	inner   slog.Handler
	service string
}

func (h *traceContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *traceContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.service != "" {
		r.AddAttrs(slog.String("service", h.service))
	}
	span := trace.SpanFromContext(ctx)
	if sc := span.SpanContext(); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.inner.Handle(ctx, r)
}

func (h *traceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceContextHandler{inner: h.inner.WithAttrs(attrs), service: h.service}
}

func (h *traceContextHandler) WithGroup(name string) slog.Handler {
	return &traceContextHandler{inner: h.inner.WithGroup(name), service: h.service}
}

func parseLogLevel(level string) slog.Level {
	switch level {
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
