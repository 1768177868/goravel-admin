package traceid

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/oklog/ulid/v2"
)

type contextKey string

const (
	ContextKey      contextKey = "trace_id"
	headerKey       string     = "X-Trace-Id"
	requestIDHeader string     = "X-Request-Id"
)

// Generate returns a new trace id using ULID to ensure it sorts well and is URL safe.
func Generate() string {
	return strings.ToLower(ulid.Make().String())
}

// EnsureHTTPContext stores a trace id on the http.Context (and returns it).
func EnsureHTTPContext(ctx http.Context, preferred string) string {
	if ctx == nil {
		return Generate()
	}

	traceID := preferred
	if traceID == "" {
		traceID = ctx.Request().Header(headerKey, "")
	}
	if traceID == "" {
		traceID = ctx.Request().Header(requestIDHeader, "")
	}
	if traceID == "" {
		traceID = Generate()
	}

	ctx.WithValue(string(ContextKey), traceID)
	return traceID
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
	traceID := FromContext(ctx)
	if traceID == "" {
		traceID = Generate()
	}
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
	traceID := FromHTTPContext(ctx)
	if traceID == "" {
		traceID = Generate()
		StoreHTTPContext(ctx, traceID)
	}
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
