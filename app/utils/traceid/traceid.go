package traceid

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	ContextKey         contextKey = "trace_id"
	headerKey          string     = "X-Trace-Id"
	requestIDHeader    string     = "X-Request-Id"
	traceParentHeader  string     = "traceparent"
)

// Generate returns a new trace id using ULID to ensure it sorts well and is URL safe.
func Generate() string {
	return strings.ToLower(ulid.Make().String())
}

// FromOTELContext reads the W3C trace id from an active OpenTelemetry span, if any.
func FromOTELContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() || !spanCtx.HasTraceID() {
		return ""
	}
	return spanCtx.TraceID().String()
}

// EnsureHTTPContext stores a trace id on the http.Context (and returns it).
// Resolution order: existing value > preferred > OTEL span > X-Trace-Id > X-Request-Id > traceparent > ULID.
func EnsureHTTPContext(ctx http.Context, preferred string) string {
	if ctx == nil {
		return Generate()
	}

	if existing := FromHTTPContext(ctx); existing != "" {
		return existing
	}

	traceID := resolveHTTPTraceID(ctx, preferred)
	ctx.WithValue(string(ContextKey), traceID)
	return traceID
}

func resolveHTTPTraceID(ctx http.Context, preferred string) string {
	if traceID := normalizeIncomingTraceID(preferred); traceID != "" {
		return traceID
	}
	if traceID := FromOTELContext(ctx); traceID != "" {
		return traceID
	}
	if traceID := FromOTELContext(ctx.Context()); traceID != "" {
		return traceID
	}
	if traceID := normalizeIncomingTraceID(ctx.Request().Header(headerKey, "")); traceID != "" {
		return traceID
	}
	if traceID := normalizeIncomingTraceID(ctx.Request().Header(requestIDHeader, "")); traceID != "" {
		return traceID
	}
	if traceID := parseTraceParent(ctx.Request().Header(traceParentHeader, "")); traceID != "" {
		return traceID
	}
	return Generate()
}

func normalizeIncomingTraceID(raw string) string {
	traceID := strings.TrimSpace(raw)
	if traceID == "" {
		return ""
	}
	return traceID
}

func parseTraceParent(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	parts := strings.Split(value, "-")
	if len(parts) < 2 {
		return ""
	}
	traceID := strings.TrimSpace(parts[1])
	if len(traceID) != 32 {
		return ""
	}
	return strings.ToLower(traceID)
}

// FromHTTPContext retrieves the stored trace id from http.Context.
func FromHTTPContext(ctx http.Context) string {
	if ctx == nil {
		return ""
	}
	if value := ctx.Value(string(ContextKey)); value != nil {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}
	return ""
}

// StoreHTTPContext stores an existing trace id into the http context.
func StoreHTTPContext(ctx http.Context, traceID string) {
	if ctx == nil || traceID == "" {
		return
	}
	ctx.WithValue(string(ContextKey), traceID)
}

// EnsureContext ensures a standard context carries a trace id and returns both.
func EnsureContext(ctx context.Context) (context.Context, string) {
	if traceID := FromContext(ctx); traceID != "" {
		return ctx, traceID
	}
	if traceID := FromOTELContext(ctx); traceID != "" {
		if ctx == nil {
			ctx = context.Background()
		}
		return context.WithValue(ctx, ContextKey, traceID), traceID
	}
	traceID := Generate()
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ContextKey, traceID), traceID
}

// WithTrace assigns a trace id into the provided context (or background if nil).
func WithTrace(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID == "" {
		traceID = Generate()
	}
	return context.WithValue(ctx, ContextKey, traceID)
}

// FromContext reads the trace id from a standard context.
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(ContextKey).(string); ok {
		return traceID
	}
	return ""
}

// DeriveContextFromHTTP builds a standard context containing the http trace id.
func DeriveContextFromHTTP(ctx http.Context) context.Context {
	traceID := EnsureHTTPContext(ctx, "")
	return context.WithValue(context.Background(), ContextKey, traceID)
}

// HeaderName exposes the header used to propagate the trace id.
func HeaderName() string {
	return headerKey
}

// RequestHeaderFallback exposes secondary header name for incoming trace ids.
func RequestHeaderFallback() string {
	return requestIDHeader
}
