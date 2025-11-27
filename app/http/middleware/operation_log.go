package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/logger"
	"goravel/app/utils/traceid"
)

// OperationLog 操作日志中间件
func OperationLog() http.Middleware {
	return func(ctx http.Context) {
		systemLogService := services.NewSystemLogService()
		startTime := time.Now()

		// 获取请求信息
		method := ctx.Request().Method()
		path := ctx.Request().Path()
		ip := ctx.Request().Ip()
		userAgent := ctx.Request().Header("User-Agent", "")

		// 获取请求参数（排除敏感信息）
		var requestBody string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			// 获取所有输入参数
			inputs := make(map[string]interface{})
			// 记录所有非敏感参数
			allInputs := ctx.Request().All()
			for key, value := range allInputs {
				// 隐藏敏感字段
				if key == "password" || key == "old_password" || key == "new_password" || key == "confirm_password" {
					inputs[key] = "***"
				} else {
					inputs[key] = value
				}
			}
			if data, err := json.Marshal(inputs); err == nil {
				requestBody = string(data)
			}
		}

		// 获取管理员ID（从JWT中间件设置的context中获取）
		var adminID uint
		adminValue := ctx.Value("admin")
		if adminValue != nil {
			// 尝试值类型
			if admin, ok := adminValue.(models.Admin); ok {
				adminID = admin.ID
			} else if adminPtr, ok := adminValue.(*models.Admin); ok {
				// 尝试指针类型
				adminID = adminPtr.ID
			}
		}

		// 继续处理请求
		ctx.Request().Next()

		// 计算耗时
		duration := int(time.Since(startTime).Milliseconds())

		// 只记录新增、修改、删除操作（POST、PUT、PATCH、DELETE），排除 GET 请求
		// 同时排除登录和info接口
		if (method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE") &&
			path != "/api/admin/login" && path != "/api/admin/info" {
			// 在请求处理后再获取一次管理员ID（确保JWT中间件已执行）
			// 如果之前没有获取到，再次尝试从context获取
			if adminID == 0 {
				adminValue := ctx.Value("admin")
				if adminValue != nil {
					// 尝试值类型
					if admin, ok := adminValue.(models.Admin); ok {
						adminID = admin.ID
					} else if adminPtr, ok := adminValue.(*models.Admin); ok {
						// 尝试指针类型
						adminID = adminPtr.ID
					}
				}
			}

			// 默认状态为成功
			status := uint8(1)
			var errorMsg string

			// 在goroutine之前保存所有需要的数据，避免context问题
			savedAdminID := adminID
			savedMethod := method
			savedPath := path
			savedIP := ip
			savedUserAgent := userAgent
			savedRequestBody := requestBody
			savedDuration := duration

			operationLog := models.OperationLog{
				AdminID:   savedAdminID,
				Method:    savedMethod,
				Path:      savedPath,
				IP:        savedIP,
				UserAgent: savedUserAgent,
				Request:   savedRequestBody,
				Status:    status,
				ErrorMsg:  errorMsg,
				Duration:  savedDuration,
			}

			// 异步记录日志，避免影响响应速度
			traceCtx := traceid.DeriveContextFromHTTP(ctx)
			go func(ctx context.Context) {
				if err := facades.Orm().Query().Create(&operationLog); err != nil {
					_ = systemLogService.Record(ctx, "error", "operation-log", "failed to persist operation log", map[string]any{
						"error": err.Error(),
						"path":  savedPath,
					})
					logger.ErrorfContext(ctx, "Failed to create operation log: %v", err)
				}
			}(traceCtx)
		}
	}
}
