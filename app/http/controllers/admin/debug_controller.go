package admin

import (
	"fmt"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
	"goravel/app/utils/errorlog"
	"goravel/app/utils/traceid"
)

type DebugController struct {
}

func NewDebugController() *DebugController {
	return &DebugController{}
}

// TraceTest 手动触发不同级别的日志，方便校验 trace_id
// 支持 query 参数：
//   - level: 日志级别 (error/warning/info/debug)，默认 error
//   - message: 自定义消息，默认 "manual trace log test"
//   - trace_id: 自定义 trace_id（可选）
func (r *DebugController) TraceTest(ctx http.Context) http.Response {
	traceID := traceid.EnsureHTTPContext(ctx, ctx.Request().Query("trace_id", ""))
	message := ctx.Request().Query("message", "manual trace log test")
	level := ctx.Request().Query("level", "error")

	// 根据级别记录不同级别的日志
	switch level {
	case "warning":
		errorlog.RecordHTTPWithLevel(ctx, "warning", "trace-test", "Trace test warning log", map[string]any{
			"path":     ctx.Request().Path(),
			"method":   ctx.Request().Method(),
			"trace_id": traceID,
			"message":  message,
			"level":    "warning",
		}, "Trace test warning: %s", message)
		return response.Error(ctx, http.StatusBadRequest, "trace_test_warning")

	case "info":
		errorlog.RecordHTTPWithLevel(ctx, "info", "trace-test", "Trace test info log", map[string]any{
			"path":     ctx.Request().Path(),
			"method":   ctx.Request().Method(),
			"trace_id": traceID,
			"message":  message,
			"level":    "info",
		}, "Trace test info: %s", message)
		return response.Success(ctx, "trace_test_info", http.Json{
			"message": message,
			"level":   "info",
			"hint":    "Check system logs with this trace_id to see info level log",
		})

	case "debug":
		errorlog.RecordHTTPWithLevel(ctx, "debug", "trace-test", "Trace test debug log", map[string]any{
			"path":     ctx.Request().Path(),
			"method":   ctx.Request().Method(),
			"trace_id": traceID,
			"message":  message,
			"level":    "debug",
		}, "Trace test debug: %s", message)
		return response.Success(ctx, "trace_test_debug", http.Json{
			"message": message,
			"level":   "debug",
			"hint":    "Check system logs with this trace_id to see debug level log",
		})

	default: // error
		errorlog.RecordHTTP(ctx, "trace-test", "Trace test error log", map[string]any{
			"path":     ctx.Request().Path(),
			"method":   ctx.Request().Method(),
			"trace_id": traceID,
			"message":  message,
			"level":    "error",
		}, "Trace test error: %s", message)
		return response.Error(ctx, http.StatusInternalServerError, "trace_test_error")
	}

}

// TestErrorLog 测试 ErrorWithLog 日志记录功能
// 用于手动触发异常，测试日志是否正常写入
func (r *MenuController) TestErrorLog(ctx http.Context) http.Response {
	// 创建一个测试错误
	testErr := fmt.Errorf("测试错误日志记录功能 - 这是一个手动触发的异常测试")

	return response.ErrorWithLog(ctx, "menu-test", testErr)

	// return response.ErrorWithLog(ctx, "menu-test", testErr, map[string]any{
	// 	"test_type":    "manual_test",
	// 	"test_purpose": "验证 ErrorWithLog 是否能正确写入日志记录",
	// 	"test_time":    carbon.Now().ToDateTimeString(),
	// 	"controller":   "MenuController",
	// 	"method":       "TestErrorLog",
	// })
}
