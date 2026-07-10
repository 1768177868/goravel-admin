package providers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/foundation"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http/limit"

	"goravel/app/facades"
	"goravel/app/http"
	"goravel/app/http/helpers"
	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
)

type RouteServiceProvider struct {
}

func (receiver *RouteServiceProvider) Register(app foundation.Application) {
}

func (receiver *RouteServiceProvider) Boot(app foundation.Application) {
	// Add HTTP middleware
	facades.Route().GlobalMiddleware(http.Kernel{}.Middleware()...)
	facades.Route().Recover(func(ctx contractshttp.Context, err any) {
		systemLogService := services.NewSystemLogService(ctx)
		_ = systemLogService.RecordHTTP(ctx, "error", "recover", fmt.Sprintf("%v", err), nil)
		facades.Log().Error(err)
		_ = ctx.Response().String(contractshttp.StatusInternalServerError, "recover").Abort()
	})

	receiver.configureRateLimiting()
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

	// 登录速率限制器（IP + 账号 双维度，避免攻击者锁住其他 IP 的同名账号）
	facades.RateLimiter().For("login", func(ctx contractshttp.Context) contractshttp.Limit {
		ip := helpers.GetRealIP(ctx)
		username := resolveLoginIdentifier(ctx, ip)

		return limit.PerMinute(6).Response(func(ctx contractshttp.Context) {
			_ = ctx.Response().Json(contractshttp.StatusTooManyRequests, contractshttp.Json{
				"code":    contractshttp.StatusTooManyRequests,
				"message": trans.Get(ctx, "too_many_requests"),
			}).Abort()
		}).By(ip + ":login:" + username)
	})

	// 测试响应速率限制器（仅开发环境使用）
	facades.RateLimiter().For("testResponse", func(ctx contractshttp.Context) contractshttp.Limit {
		return limit.PerMinute(6).Response(func(ctx contractshttp.Context) {
			_ = ctx.Response().Json(contractshttp.StatusTooManyRequests, contractshttp.Json{
				"code":    contractshttp.StatusTooManyRequests,
				"message": trans.Get(ctx, "too_many_requests"),
			}).Abort()
		})
	})

	// pprof token 验证限流（按管理员 + IP）
	facades.RateLimiter().For("pprofVerify", func(ctx contractshttp.Context) contractshttp.Limit {
		ip := helpers.GetRealIP(ctx)
		identifier := resolvePprofVerifyIdentifier(ctx, ip)
		return limit.PerMinute(6).Response(func(ctx contractshttp.Context) {
			_ = ctx.Response().Json(contractshttp.StatusTooManyRequests, contractshttp.Json{
				"code":       contractshttp.StatusTooManyRequests,
				"message":    trans.Get(ctx, "too_many_requests"),
				"error_code": "pprof_verify_rate_limited",
			}).Abort()
		}).By(ip + ":pprof_verify:" + identifier)
	})

	// pprof CPU 采样限流（按管理员 + IP）
	facades.RateLimiter().For("pprofCPU", func(ctx contractshttp.Context) contractshttp.Limit {
		ip := helpers.GetRealIP(ctx)
		identifier := resolvePprofVerifyIdentifier(ctx, ip)
		return limit.PerMinute(3).Response(func(ctx contractshttp.Context) {
			_ = ctx.Response().Json(contractshttp.StatusTooManyRequests, contractshttp.Json{
				"code":       contractshttp.StatusTooManyRequests,
				"message":    trans.Get(ctx, "too_many_requests"),
				"error_code": "pprof_cpu_rate_limited",
			}).Abort()
		}).By(ip + ":pprof_cpu:" + identifier)
	})

	// pprof 内存采样限流（按管理员 + IP）
	facades.RateLimiter().For("pprofMemory", func(ctx contractshttp.Context) contractshttp.Limit {
		ip := helpers.GetRealIP(ctx)
		identifier := resolvePprofVerifyIdentifier(ctx, ip)
		return limit.PerMinute(6).Response(func(ctx contractshttp.Context) {
			_ = ctx.Response().Json(contractshttp.StatusTooManyRequests, contractshttp.Json{
				"code":       contractshttp.StatusTooManyRequests,
				"message":    trans.Get(ctx, "too_many_requests"),
				"error_code": "pprof_memory_rate_limited",
			}).Abort()
		}).By(ip + ":pprof_memory:" + identifier)
	})
}

// resolveLoginIdentifier 从请求中提取登录标识（username > email > X-Username > IP fallback）。
func resolveLoginIdentifier(ctx contractshttp.Context, fallbackIP string) string {
	for _, field := range []string{"username", "email"} {
		if v := strings.TrimSpace(ctx.Request().Input(field, "")); v != "" {
			return strings.ToLower(v)
		}
	}
	if v := strings.TrimSpace(ctx.Request().Header("X-Username", "")); v != "" {
		return strings.ToLower(v)
	}
	return fallbackIP
}

// resolvePprofVerifyIdentifier 从上下文提取管理员 ID，找不到则回退到 IP
func resolvePprofVerifyIdentifier(ctx contractshttp.Context, fallbackIP string) string {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return fallbackIP
	}

	if admin, ok := adminValue.(models.Admin); ok {
		return strconv.FormatUint(uint64(admin.ID), 10)
	}
	if adminPtr, ok := adminValue.(*models.Admin); ok && adminPtr != nil {
		return strconv.FormatUint(uint64(adminPtr.ID), 10)
	}

	return fallbackIP
}
