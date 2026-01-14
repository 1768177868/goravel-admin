package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func DevelopmentOnly() http.Middleware {
	return func(ctx http.Context) {
		env := facades.Config().Get("app.env", "production")
		if env != "local" && env != "development" {
			ctx.Response().Json(http.StatusForbidden, map[string]any{
				"code":    403,
				"message": "This feature is only available in development mode",
			})
			ctx.Request().Abort()
			return
		}
		ctx.Request().Next()
	}
}
