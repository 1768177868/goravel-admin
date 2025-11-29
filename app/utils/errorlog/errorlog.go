package errorlog

import (
	"context"
	"encoding/json"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/utils/logger"
	"goravel/app/utils/traceid"
)

// RecordHTTP 同时记录文件日志和数据库日志（用于系统级错误，默认 error 级别）
// 使用场景：数据库操作失败、系统服务异常、关键业务逻辑错误等
//
// 示例：
//
//	if err != nil {
//	    errorlog.RecordHTTP(ctx, "auth", "Failed to save admin profile", map[string]any{
//	        "error": err.Error(),
//	        "admin_id": admin.ID,
//	    }, "Save admin profile error: %v", err)
//	    return response.Error(ctx, http.StatusInternalServerError, "update_failed")
//	}
func RecordHTTP(ctx http.Context, module, message string, attributes map[string]any, format string, args ...any) {
	RecordHTTPWithLevel(ctx, "error", module, message, attributes, format, args...)
}

// RecordHTTPWithLevel 同时记录文件日志和数据库日志（可指定日志级别）
// level: 日志级别，支持 "error", "warning", "info", "debug"
// 使用场景：需要记录不同级别的系统日志
//
// 示例：
//
//	// 记录警告
//	errorlog.RecordHTTPWithLevel(ctx, "warning", "auth", "Unusual login pattern detected", map[string]any{
//	    "admin_id": admin.ID,
//	    "ip": ctx.Request().Ip(),
//	}, "Unusual login pattern: %s", pattern)
//
//	// 记录信息
//	errorlog.RecordHTTPWithLevel(ctx, "info", "payment", "Payment processed successfully", map[string]any{
//	    "order_id": order.ID,
//	    "amount": order.Amount,
//	}, "Payment processed: %d", order.ID)
func RecordHTTPWithLevel(ctx http.Context, level, module, message string, attributes map[string]any, format string, args ...any) {
	// 根据级别选择不同的日志函数
	switch level {
	case "error":
		logger.ErrorfHTTP(ctx, format, args...)
	case "warning":
		// warning 级别也使用 ErrorfHTTP（因为 logger 包目前只有 ErrorfHTTP）
		logger.ErrorfHTTP(ctx, format, args...)
	case "info", "debug":
		// info 和 debug 级别暂时也使用 ErrorfHTTP
		// 如果需要区分，可以扩展 logger 包
		logger.ErrorfHTTP(ctx, format, args...)
	default:
		logger.ErrorfHTTP(ctx, format, args...)
	}

	// 记录到数据库（所有级别都记录 trace_id）
	if ctx != nil {
		recordToDatabaseHTTPWithLevel(ctx, level, module, message, attributes)
	}
}

// Record 同时记录文件日志和数据库日志（用于标准 context，默认 error 级别）
// 使用场景：goroutine、后台任务等
//
// 示例：
//
//	go func(ctx context.Context) {
//	    if err != nil {
//	        errorlog.Record(ctx, "operation-log", "Failed to create operation log", map[string]any{
//	            "error": err.Error(),
//	        }, "Create operation log error: %v", err)
//	    }
//	}(traceCtx)
func Record(ctx context.Context, module, message string, attributes map[string]any, format string, args ...any) {
	RecordWithLevel(ctx, "error", module, message, attributes, format, args...)
}

// RecordWithLevel 同时记录文件日志和数据库日志（可指定日志级别，用于标准 context）
// level: 日志级别，支持 "error", "warning", "info", "debug"
// 使用场景：goroutine、后台任务中需要记录不同级别的日志
//
// 示例：
//
//	go func(ctx context.Context) {
//	    errorlog.RecordWithLevel(ctx, "info", "background-task", "Task completed", map[string]any{
//	        "task_id": taskID,
//	    }, "Background task completed: %s", taskID)
//	}(traceCtx)
func RecordWithLevel(ctx context.Context, level, module, message string, attributes map[string]any, format string, args ...any) {
	// 根据级别选择不同的日志函数
	switch level {
	case "error":
		logger.ErrorfContext(ctx, format, args...)
	case "warning", "info", "debug":
		// warning, info, debug 级别暂时也使用 ErrorfContext
		// 如果需要区分，可以扩展 logger 包
		logger.ErrorfContext(ctx, format, args...)
	default:
		logger.ErrorfContext(ctx, format, args...)
	}

	// 记录到数据库（所有级别都记录 trace_id）
	if ctx != nil {
		recordToDatabaseWithLevel(ctx, level, module, message, attributes)
	}
}

// recordToDatabaseHTTPWithLevel 将日志记录到数据库（HTTP context，支持所有级别）
func recordToDatabaseHTTPWithLevel(ctx http.Context, level, module, message string, attributes map[string]any) {
	var contextJSON string
	if len(attributes) > 0 {
		if data, err := json.Marshal(attributes); err == nil {
			contextJSON = string(data)
		}
	}

	traceID := traceid.FromHTTPContext(ctx)
	if traceID == "" {
		traceID = traceid.EnsureHTTPContext(ctx, "")
	}

	log := models.SystemLog{
		Level:     level,
		Module:    module,
		TraceID:   traceID,
		Message:   message,
		Context:   contextJSON,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent", ""),
	}

	_ = facades.Orm().Query().Create(&log)
}

// recordToDatabaseWithLevel 将日志记录到数据库（标准 context，支持所有级别）
func recordToDatabaseWithLevel(ctx context.Context, level, module, message string, attributes map[string]any) {
	if ctx == nil {
		ctx = context.Background()
	}

	var contextJSON string
	if len(attributes) > 0 {
		if data, err := json.Marshal(attributes); err == nil {
			contextJSON = string(data)
		}
	}

	traceID := traceid.FromContext(ctx)
	if traceID == "" {
		var newCtx context.Context
		newCtx, traceID = traceid.EnsureContext(ctx)
		ctx = newCtx
	}

	log := models.SystemLog{
		Level:   level,
		Module:  module,
		TraceID: traceID,
		Message: message,
		Context: contextJSON,
	}

	_ = facades.Orm().Query().Create(&log)
}
