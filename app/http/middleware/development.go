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

// CodeGeneratorOnly restricts code-generator APIs (hidden in APP_ENV=test unless APP_ENABLE_DEV_TOOL=true).
func CodeGeneratorOnly() http.Middleware {
	return newMiddleware("code_generator_only", func(ctx http.Context) {
		if !utils.CodeGeneratorEnabled() {
			response.Error(ctx, http.StatusForbidden, "development_only")
			ctx.Request().Abort()
			return
		}

		ctx.Request().Next()
	})
}
