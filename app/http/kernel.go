package http

import (
	"github.com/goravel/framework/contracts/http"
	httpmiddleware "github.com/goravel/framework/http/middleware"
	sessionmiddleware "github.com/goravel/framework/session/middleware"

	appmiddleware "goravel/app/http/middleware"
)

type Kernel struct {
}

// The application's global HTTP middleware stack.
// These middleware are run during every request to your application.
func (kernel Kernel) Middleware() []http.Middleware {
	return []http.Middleware{
		appmiddleware.Blacklist(), // 黑名单检查（最先执行）
		appmiddleware.Trace(),
		httpmiddleware.Throttle("global"),
		sessionmiddleware.StartSession(),
	}
}
