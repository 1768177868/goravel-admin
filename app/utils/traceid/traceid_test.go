package traceid

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestFromOTELContextReturnsTraceID(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	tr := tp.Tracer("test")
	ctx, span := tr.Start(context.Background(), "request")
	defer span.End()

	got := FromOTELContext(ctx)
	if got == "" || len(got) != 32 {
		t.Fatalf("expected 32-char OTEL trace id, got %q", got)
	}
}

func TestFromOTELContextEmptyWithoutSpan(t *testing.T) {
	if got := FromOTELContext(context.Background()); got != "" {
		t.Fatalf("expected empty trace id, got %q", got)
	}
}

func TestParseTraceParent(t *testing.T) {
	got := parseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	want := "4bf92f3577b34da6a3ce929d0e0e4736"
	if got != want {
		t.Fatalf("parseTraceParent() = %q, want %q", got, want)
	}
}

func TestEnsureContextPrefersOTELTraceID(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	tr := tp.Tracer("test")
	ctx, span := tr.Start(context.Background(), "job")
	defer span.End()

	newCtx, traceID := EnsureContext(ctx)
	if traceID == "" || len(traceID) != 32 {
		t.Fatalf("expected OTEL trace id, got %q", traceID)
	}
	if FromContext(newCtx) != traceID {
		t.Fatalf("stored trace id mismatch")
	}
}

func TestEnsureContextKeepsExistingTraceID(t *testing.T) {
	ctx := WithTrace(context.Background(), "custom-trace")
	newCtx, traceID := EnsureContext(ctx)
	if traceID != "custom-trace" {
		t.Fatalf("expected existing trace id, got %q", traceID)
	}
	if FromContext(newCtx) != "custom-trace" {
		t.Fatalf("stored trace id mismatch")
	}
}
