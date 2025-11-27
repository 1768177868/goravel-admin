package logger

import (
	"context"
	"fmt"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/utils/traceid"
)

// ErrorfHTTP logs an error and automatically attaches trace_id from the http context.
func ErrorfHTTP(ctx http.Context, format string, args ...any) {
	if ctx == nil {
		facades.Log().Errorf(format, args...)
		return
	}

	trace := traceid.FromHTTPContext(ctx)
	facades.Log().Errorf(prependTrace(trace, format), args...)
}

// ErrorfContext logs an error with a standard context's trace id (if available).
func ErrorfContext(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		facades.Log().Errorf(format, args...)
		return
	}

	trace := traceid.FromContext(ctx)
	facades.Log().Errorf(prependTrace(trace, format), args...)
}

// Errorf logs an error without any context (fallback).
func Errorf(format string, args ...any) {
	facades.Log().Errorf(format, args...)
}

func prependTrace(traceID, format string) string {
	if traceID == "" {
		return format
	}
	return fmt.Sprintf("[trace_id=%s] %s", traceID, format)
}
