package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/utils/traceid"
)

// Trace middleware ensures every request carries a trace id and mirrors it in response headers.
func Trace() http.Middleware {
	return func(ctx http.Context) {
		traceID := traceid.EnsureHTTPContext(ctx, "")
		ctx.Response().Header(traceid.HeaderName(), traceID)
		ctx.Request().Next()
	}
}

