package middleware

import (
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/logger"
)

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Jwt() http.Middleware {
	return func(ctx http.Context) {
		// 如果路径是api/admin前缀，使用admin guard
		path := ctx.Request().Path()
		if path == "" || (!strings.HasPrefix(path, "/api/admin") && !strings.HasPrefix(path, "/admin")) {
			ctx.Request().Next()
			return
		}

		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "not_logged_in"),
			}).Abort()
			return
		}

		// 移除Bearer前缀（如果有）
		token = strings.TrimPrefix(token, "Bearer ")
		token = strings.TrimSpace(token)

		if token == "" {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "not_logged_in"),
			}).Abort()
			return
		}

		// 从数据库查找token
		tokenService := services.NewTokenServiceImpl()
		accessToken, err := tokenService.FindToken(token)
		if err != nil {
			// token查找失败或已过期
			logger.ErrorfHTTP(ctx, "JWT middleware: FindToken error: %v, token prefix: %s", err, token[:min(20, len(token))])
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "invalid_token"),
			}).Abort()
			return
		}
		if accessToken == nil {
			logger.ErrorfHTTP(ctx, "JWT middleware: accessToken is nil, token prefix: %s", token[:min(20, len(token))])
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "invalid_token"),
			}).Abort()
			return
		}

		// 检查token类型
		if accessToken.TokenableType != "admin" {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "invalid_token"),
			}).Abort()
			return
		}

		// 查询用户信息
		var admin models.Admin
		if err := facades.Orm().Query().Where("id", accessToken.TokenableID).First(&admin); err != nil {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "user_not_found"),
			}).Abort()
			return
		}

		// 更新最后使用时间
		_ = tokenService.UpdateLastUsedAt(token)

		// 滑动过期：如果token有过期时间，每次请求时自动延长过期时间
		if accessToken.ExpiresAt != nil {
			ttl := facades.Config().GetInt("jwt.ttl", 60) // 默认60分钟
			if ttl > 0 {
				newExpiresAt := time.Now().Add(time.Duration(ttl) * time.Minute)
				// 更新token的过期时间
				_, _ = facades.Orm().Query().
					Model(&models.PersonalAccessToken{}).
					Where("id", accessToken.ID).
					Update("expires_at", newExpiresAt)
			}
		}

		// 将用户信息存储到context中，供后续中间件使用
		ctx.WithValue("admin", admin)
		ctx.WithValue("token", accessToken)

		facades.Log().Debugf("JWT middleware: admin set in context, ID: %d, Username: %s", admin.ID, admin.Username)

		ctx.Request().Next()
	}
}
