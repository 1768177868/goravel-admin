package admin

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
	"goravel/app/services"
	"goravel/app/utils/logger"
	"goravel/app/utils/traceid"
)

type DebugController struct {
	systemLog services.SystemLogService
}

func NewDebugController() *DebugController {
	return &DebugController{
		systemLog: services.NewSystemLogService(),
	}
}

// TraceTest 手动触发错误日志，方便校验 trace_id
func (r *DebugController) TraceTest(ctx http.Context) http.Response {
	traceID := traceid.EnsureHTTPContext(ctx, ctx.Request().Query("trace_id", ""))
	message := ctx.Request().Query("message", "manual trace log test")

	// 普通日志文件
	logger.ErrorfHTTP(ctx, "Trace test endpoint triggered: %s", message)

	// 系统日志表
	_ = r.systemLog.RecordHTTP(ctx, "error", "trace-test", message, map[string]any{
		"path":    ctx.Request().Path(),
		"method":  ctx.Request().Method(),
		"traceId": traceID,
	})

	return response.Success(ctx, "get_success", http.Json{
		"trace_id": traceID,
		"message":  message,
		"hint":     "Use this trace_id to search system logs or grep server logs.",
	})
}

