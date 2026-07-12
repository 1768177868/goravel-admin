package middleware

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	gingin "github.com/goravel/gin"
)

// RequestTimeout 返回与 Goravel Gin 驱动全局一致的请求超时中间件。
// 用于 SSE 等长连接路由的 WithoutMiddleware，按 Signature 排除全局 Timeout。
func RequestTimeout() http.Middleware {
	seconds := facades.Config().GetInt("http.request_timeout", 300)
	if seconds <= 0 {
		seconds = 300
	}
	return gingin.Timeout(time.Duration(seconds) * time.Second)
}
