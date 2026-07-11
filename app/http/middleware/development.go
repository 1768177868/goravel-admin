package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
	"goravel/app/utils"
)

func DevelopmentOnly() http.Middleware {
	return newMiddleware("development_only", func(ctx http.Context) {
		if !utils.DevToolsEnabled() {
			response.Error(ctx, http.StatusForbidden, "development_only")
			ctx.Request().Abort()
			return
		}

		ctx.Request().Next()
	})
}
