package middleware

import (
	"errors"
	"strings"

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/trans"
)

// JwtRefresh JWT刷新中间件，允许token过期但仍在刷新窗口内的请求通过
func JwtRefresh() http.Middleware {
	return func(ctx http.Context) {
		guard := "admin"

		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "unauthorized"),
			}).Abort()
			return
		}

		// 移除Bearer前缀（如果有）
		token = strings.TrimPrefix(token, "Bearer ")
		token = strings.TrimSpace(token)

		// 尝试解析token
		_, parseErr := facades.Auth(ctx).Guard(guard).Parse(token)
		if parseErr != nil {
			// Token无效或已过期
			// 检查是否是token过期错误，如果是过期且在刷新窗口内，允许通过
			if !errors.Is(parseErr, auth.ErrorTokenExpired) {
				// Token完全无效（格式错误、签名错误等），直接拒绝
				_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
					"code":    http.StatusUnauthorized,
					"message": trans.Get(ctx, "invalid_token"),
				}).Abort()
				return
			}
			// Token过期，但Refresh接口会处理刷新逻辑
		}

		// 允许通过，让Refresh接口自己处理刷新逻辑
		ctx.Request().Next()
	}
}

