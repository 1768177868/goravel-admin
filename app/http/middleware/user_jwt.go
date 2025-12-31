package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"

	"goravel/app/http/trans"
	"goravel/app/models"
)

// UserJwt C端用户JWT认证中间件（使用Goravel标准Auth）
func UserJwt() http.Middleware {
	return func(ctx http.Context) {
		// 如果路径是api/user前缀，使用user guard
		path := ctx.Request().Path()
		pathStr := str.Of(path)
		if pathStr.IsEmpty() || !pathStr.StartsWith("/api/user") {
			ctx.Request().Next()
			return
		}

		// 使用Goravel标准Auth解析token
		if _, err := facades.Auth(ctx).Guard("user").Parse(ctx.Request().Header("Authorization", "")); err != nil {
			// 如果Header中没有token，尝试从URL参数中获取
			if token := ctx.Request().Query("_token", ""); token != "" {
				if _, err := facades.Auth(ctx).Guard("user").Parse(token); err != nil {
					_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
						"code":    http.StatusUnauthorized,
						"message": trans.Get(ctx, "invalid_token"),
					}).Abort()
					return
				}
			} else {
				_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
					"code":    http.StatusUnauthorized,
					"message": trans.Get(ctx, "not_logged_in"),
				}).Abort()
				return
			}
		}

		// 获取用户信息
		var user models.User
		if err := facades.Auth(ctx).Guard("user").User(&user); err != nil {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "user_not_found"),
			}).Abort()
			return
		}

		if user.ID == 0 {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "user_not_found"),
			}).Abort()
			return
		}

		// 检查用户状态
		if user.Status == 0 {
			_ = ctx.Response().Json(http.StatusForbidden, http.Json{
				"code":    http.StatusForbidden,
				"message": trans.Get(ctx, "account_disabled"),
			}).Abort()
			return
		}

		// 将用户信息存储到context中，供后续中间件使用
		ctx.WithValue("user", user)

		ctx.Request().Next()
	}
}
