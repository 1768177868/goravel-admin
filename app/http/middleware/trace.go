package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/utils/traceid"
)

// Trace middleware bridges OTEL trace ids into the app trace_id field and X-Trace-Id header.
func Trace() http.Middleware {
	return newMiddleware("trace", func(ctx http.Context) {
		traceID := traceid.EnsureHTTPContext(ctx, "")
		ctx.Response().Header(traceid.HeaderName(), traceID)
		ctx.Request().Next()
	})
}
