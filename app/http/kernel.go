package http

import (
	"github.com/goravel/framework/contracts/http"
	httpmiddleware "github.com/goravel/framework/http/middleware"

	appmiddleware "goravel/app/http/middleware"
)

type Kernel struct {
}

// The application's global HTTP middleware stack.
// These middleware are run during every request to your application.
func (kernel Kernel) Middleware() []http.Middleware {
	return []http.Middleware{
		appmiddleware.Cors(),      // CORS 跨域处理（需要在最前面处理预检请求）
		httpmiddleware.CheckForMaintenanceMode(),
		appmiddleware.Blacklist(), // 黑名单检查
		appmiddleware.SecurityHeaders(),
		appmiddleware.Trace(),
		appmiddleware.Gzip(), // 仅本地/开发环境对 JSON 等响应做 gzip 压缩
		httpmiddleware.Throttle("global"),
		// sessionmiddleware.StartSession(), // 已禁用：项目使用 JWT 认证，不需要 Session
	}
}
