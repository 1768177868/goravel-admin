package middleware

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// Timeout 创建一个超时中间件，支持自定义超时时间
// timeoutSeconds: 超时时间（秒），0 或负数表示使用配置中的默认值
func Timeout(timeoutSeconds int) http.Middleware {
	return func(ctx http.Context) {
		// 确定超时时间
		timeout := time.Duration(timeoutSeconds) * time.Second
		if timeoutSeconds <= 0 {
			// 使用配置中的默认超时时间
			defaultTimeout := facades.Config().GetInt("http.request_timeout", 60)
			timeout = time.Duration(defaultTimeout) * time.Second
		}

		// 获取底层请求的 context
		reqCtx := ctx.Request().Origin().Context()
		if reqCtx == nil {
			reqCtx = context.Background()
		}

		// 创建带超时的 context
		timeoutCtx, cancel := context.WithTimeout(reqCtx, timeout)
		defer cancel()

		// 创建一个 channel 来监听请求完成
		done := make(chan struct{}, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// 如果发生 panic，也要通知完成
					select {
					case done <- struct{}{}:
					default:
					}
				}
			}()
			ctx.Request().Next()
			select {
			case done <- struct{}{}:
			default:
			}
		}()

		// 等待请求完成或超时
		select {
		case <-done:
			// 请求正常完成
			return
		case <-timeoutCtx.Done():
			// 请求超时
			if timeoutCtx.Err() == context.DeadlineExceeded {
				_ = ctx.Response().Json(http.StatusRequestTimeout, http.Json{
					"code":    http.StatusRequestTimeout,
					"message": "请求超时，请稍后再试",
				}).Abort()
			}
			return
		}
	}
}

// TimeoutForUpload 文件上传专用超时中间件（30分钟）
func TimeoutForUpload() http.Middleware {
	return Timeout(1800) // 30分钟 = 1800秒
}

