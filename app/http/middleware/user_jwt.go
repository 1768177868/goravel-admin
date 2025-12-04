package middleware

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"

	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
)

func UserJwt() http.Middleware {
	return func(ctx http.Context) {
		// 如果路径是api/user前缀，使用user guard
		path := ctx.Request().Path()
		if str.Of(path).IsEmpty() || !str.Of(path).StartsWith("/api/user") {
			ctx.Request().Next()
			return
		}

		token := ctx.Request().Header("Authorization", "")
		if str.Of(token).IsEmpty() {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "unauthorized"),
			}).Abort()
			return
		}

		// 移除Bearer前缀（如果有）
		token = str.Of(token).ChopStart("Bearer ").Trim().String()

		// 从数据库查找token
		tokenService := services.NewTokenServiceImpl()
		accessToken, err := tokenService.FindToken(token)
		if err != nil {
			// token查找失败或已过期
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "invalid_token"),
			}).Abort()
			return
		}
		if accessToken == nil {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "invalid_token"),
			}).Abort()
			return
		}

		// 检查token类型
		if accessToken.TokenableType != "user" {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    http.StatusUnauthorized,
				"message": trans.Get(ctx, "invalid_token"),
			}).Abort()
			return
		}

		// 查询用户信息
		var user models.User
		if err := facades.Orm().Query().Where("id", accessToken.TokenableID).First(&user); err != nil {
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
		ctx.WithValue("user", user)
		ctx.WithValue("token", accessToken)

		ctx.Request().Next()
	}
}
