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

// RecordHTTP 同时记录文件日志和数据库日志（用于系统级错误）
// 使用场景：数据库操作失败、系统服务异常、关键业务逻辑错误等
//
// 示例：
//   if err != nil {
//       errorlog.RecordHTTP(ctx, "auth", "Failed to save admin profile", map[string]any{
//           "error": err.Error(),
//           "admin_id": admin.ID,
//       }, "Save admin profile error: %v", err)
//       return response.Error(ctx, http.StatusInternalServerError, "update_failed")
//   }
func RecordHTTP(ctx http.Context, module, message string, attributes map[string]any, format string, args ...any) {
	// 先记录到文件日志
	logger.ErrorfHTTP(ctx, format, args...)
	
	// 再记录到数据库
	if ctx != nil {
		recordToDatabaseHTTP(ctx, module, message, attributes)
	}
}

// Record 同时记录文件日志和数据库日志（用于标准 context）
// 使用场景：goroutine、后台任务等
//
// 示例：
//   go func(ctx context.Context) {
//       if err != nil {
//           errorlog.Record(ctx, "operation-log", "Failed to create operation log", map[string]any{
//               "error": err.Error(),
//           }, "Create operation log error: %v", err)
//       }
//   }(traceCtx)
func Record(ctx context.Context, module, message string, attributes map[string]any, format string, args ...any) {
	// 先记录到文件日志
	logger.ErrorfContext(ctx, format, args...)
	
	// 再记录到数据库
	if ctx != nil {
		recordToDatabase(ctx, module, message, attributes)
	}
}

// recordToDatabaseHTTP 将错误记录到数据库（HTTP context）
func recordToDatabaseHTTP(ctx http.Context, module, message string, attributes map[string]any) {
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
		Level:     "error",
		Module:    module,
		TraceID:   traceID,
		Message:   message,
		Context:   contextJSON,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent", ""),
	}

	_ = facades.Orm().Query().Create(&log)
}

// recordToDatabase 将错误记录到数据库（标准 context）
func recordToDatabase(ctx context.Context, module, message string, attributes map[string]any) {
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
		Level:   "error",
		Module:  module,
		TraceID: traceID,
		Message: message,
		Context: contextJSON,
	}

	_ = facades.Orm().Query().Create(&log)
}

