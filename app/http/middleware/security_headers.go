package middleware

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
)

// SecurityHeaders adds baseline browser security headers for HTML/API responses.
func SecurityHeaders() http.Middleware {
	return newMiddleware("security_headers", func(ctx http.Context) {
		path := ctx.Request().Path()
		isWebSocket := strings.EqualFold(ctx.Request().Header("Upgrade", ""), "websocket")

		ctx.Response().Header("X-Content-Type-Options", "nosniff")
		ctx.Response().Header("X-Frame-Options", "SAMEORIGIN")
		ctx.Response().Header("Referrer-Policy", "strict-origin-when-cross-origin")
		ctx.Response().Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Avoid breaking SPA assets / websocket upgrades with a strict CSP.
		if !isWebSocket && !strings.HasPrefix(path, "/api/") {
			ctx.Response().Header("Content-Security-Policy", "frame-ancestors 'self'")
		}

		ctx.Request().Next()
	})
}
