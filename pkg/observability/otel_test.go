package observability

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestInitDisabledNoOp(t *testing.T) {
	shutdown, err := Init(context.Background(), Config{Enabled: false, ServiceName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if Enabled() {
		t.Fatal("expected Enabled() false")
	}
}

func TestStartSpanWhenDisabled(t *testing.T) {
	_, err := Init(context.Background(), Config{Enabled: false, ServiceName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, span := StartSpan(context.Background(), "test.span")
	defer span.End()
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	if ctx == nil {
		t.Fatal("expected non-nil ctx")
	}
	if otel.Tracer("veil") == nil {
		t.Fatal("expected tracer")
	}
}
