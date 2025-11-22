package middleware

import (
	"errors"
	"strings"

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/trans"
	"goravel/app/models"
)

func Jwt() http.Middleware {
	return func(ctx http.Context) {
		guard := facades.Config().GetString("auth.defaults.guard")
		if ctx.Request().Header("Guard") != "" {
			guard = ctx.Request().Header("Guard")
		}
		// 如果路径是api/admin前缀，使用admin guard
		path := ctx.Request().Path()
		if path != "" && (strings.HasPrefix(path, "/api/admin") || strings.HasPrefix(path, "/admin")) {
			guard = "admin"
		}

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

		// 解析token，检查是否有效
		_, parseErr := facades.Auth(ctx).Guard(guard).Parse(token)
		if parseErr != nil {
			// 检查是否是token过期错误
			if errors.Is(parseErr, auth.ErrorTokenExpired) {
				// Token已过期，尝试从token中获取用户ID（即使token过期，JWT payload仍然可以解析）
				var admin models.Admin
				if userID, err := facades.Auth(ctx).Guard(guard).ID(); err == nil {
					// 查询用户配置，检查是否是永久token用户
					if err := facades.Orm().Query().Where("id", userID).First(&admin); err == nil {
						if admin.TokenNeverExpires {
							// 永久token用户，即使token过期也允许通过
							// 重新生成token（使用永久配置）
							if newToken, err := facades.Auth(ctx).Guard(guard).LoginUsingID(userID); err == nil {
								token = newToken
							} else {
								// 如果重新生成失败，仍然允许通过（因为是永久token用户）
							}
						} else {
							// 非永久token用户，token过期需要重新登录
							_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
								"code":    http.StatusUnauthorized,
								"message": trans.Get(ctx, "token_expired"),
							}).Abort()
							return
						}
					} else {
						// 无法查询用户信息，拒绝访问
						_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
							"code":    http.StatusUnauthorized,
							"message": trans.Get(ctx, "user_not_found"),
						}).Abort()
						return
					}
				} else {
					// 无法从token中获取用户ID，拒绝访问
					_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
						"code":    http.StatusUnauthorized,
						"message": trans.Get(ctx, "invalid_token"),
					}).Abort()
					return
				}
			} else {
				// Token完全无效（格式错误、签名错误等），直接拒绝
				_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
					"code":    http.StatusUnauthorized,
					"message": trans.Get(ctx, "invalid_token"),
				}).Abort()
				return
			}
		} else {
			// Token有效，检查用户是否需要永久token
			// 如果是永久token，不需要滑动过期
			var admin models.Admin
			if err := facades.Auth(ctx).Guard(guard).User(&admin); err == nil {
				// 如果不是永久token，才进行滑动过期
				if !admin.TokenNeverExpires {
					// 重新生成token以延长过期时间（滑动过期）
					if userID, err := facades.Auth(ctx).Guard(guard).ID(); err == nil {
						// 重新生成token，延长过期时间
						if newToken, err := facades.Auth(ctx).Guard(guard).LoginUsingID(userID); err == nil {
							token = newToken
						}
						// 如果重新生成失败，继续使用原token
					}
				}
				// 如果是永久token，直接使用原token，不进行滑动过期
			} else {
				// 如果无法获取用户信息，尝试滑动过期（兼容性处理）
				if userID, err := facades.Auth(ctx).Guard(guard).ID(); err == nil {
					if newToken, err := facades.Auth(ctx).Guard(guard).LoginUsingID(userID); err == nil {
						token = newToken
					}
				}
			}
		}

		// You can get User in DB and set it to ctx

		//var user models.User
		//if err := facades.Auth().User(ctx, &user); err != nil {
		//	ctx.Request().AbortWithStatus(http.StatusUnauthorized)
		//  return
		//}
		//ctx.WithValue("user", user)

		ctx.Response().Header("Authorization", "Bearer "+token)
		ctx.Request().Next()
	}
}
