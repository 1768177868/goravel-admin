package providers

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/foundation"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/http/limit"

	"goravel/app/http"
	"goravel/app/services"
	"goravel/routes"
)

type RouteServiceProvider struct {
}

func (receiver *RouteServiceProvider) Register(app foundation.Application) {
}

func (receiver *RouteServiceProvider) Boot(app foundation.Application) {
	systemLogService := services.NewSystemLogService()

	// Add HTTP middleware
	facades.Route().GlobalMiddleware(http.Kernel{}.Middleware()...)
	facades.Route().Recover(func(ctx contractshttp.Context, err any) {
		_ = systemLogService.RecordHTTP(ctx, "error", "recover", fmt.Sprintf("%v", err), nil)
		facades.Log().Error(err)
		_ = ctx.Response().String(contractshttp.StatusInternalServerError, "recover").Abort()
	})

	receiver.configureRateLimiting()

	// Add routes
	routes.Web()
	routes.Api()
	routes.Admin()
	routes.Pprof() // 性能分析路由（仅在调试模式下启用）
}

func (receiver *RouteServiceProvider) configureRateLimiting() {
	// 全局速率限制器
	facades.RateLimiter().For("global", func(ctx contractshttp.Context) contractshttp.Limit {
		return limit.PerMinute(1000)
	})

	// IP 速率限制器
	facades.RateLimiter().ForWithLimits("ip", func(ctx contractshttp.Context) []contractshttp.Limit {
		return []contractshttp.Limit{
			limit.PerDay(1000),
			limit.PerMinute(2).By(ctx.Request().Ip()),
		}
	})

	// 登录速率限制器
	facades.RateLimiter().For("login", func(ctx contractshttp.Context) contractshttp.Limit {
		username := ctx.Request().Input("username", "")
		if username == "" {
			username = ctx.Request().Input("email", "")
		}
		if username == "" {
			username = ctx.Request().Header("X-Username", "")
		}
		username = strings.ToLower(strings.TrimSpace(username))
		if username == "" {
			username = ctx.Request().Ip()
		}
		return limit.PerMinute(10).By("login:" + username)
	})

	// 测试响应速率限制器
	facades.RateLimiter().For("testResponse", func(ctx contractshttp.Context) contractshttp.Limit {
		return limit.PerMinute(6).Response(func(ctx contractshttp.Context) {

			method := ctx.Request().Method()
			ip := ctx.Request().Ip()

			// 示例: 根据请求方法判断
			if method == "GET" {
				ctx.Response().Json(contractshttp.StatusTooManyRequests, contractshttp.Json{
					"message": "GET请求过于频繁",
					"ip":      ip,
				}).Abort()
				return
			}

			// 默认响应
			ctx.Response().Json(contractshttp.StatusTooManyRequests, contractshttp.Json{
				"ip": ip,
			}).Abort()
		})
	})

}
