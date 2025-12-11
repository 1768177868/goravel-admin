package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
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
			inputs := make(map[string]any)
			// 记录所有非敏感参数
			allInputs := ctx.Request().All()
			for key, value := range allInputs {
				// 使用工具函数检查是否是敏感字段
				if utils.IsSensitiveField(key) {
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
		if admin, err := helpers.GetAdminFromContext(ctx); err == nil {
			adminID = admin.ID
		}

		// 继续处理请求
		ctx.Request().Next()

		// 计算耗时
		duration := int(time.Since(startTime).Milliseconds())

		// 只记录新增、修改、删除操作（POST、PUT、PATCH、DELETE），排除 GET 请求
		// 同时排除登录和info接口，以及分片上传的进度查询（GET请求）
		// 对于分片上传，只记录 merge 操作（最终完成上传），排除 init 和 upload 操作
		if (method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE") &&
			path != "/api/admin/login" && path != "/api/admin/info" {

			// 排除分片上传的中间操作（init 和 upload），只记录 merge（最终完成上传）
			if path == "/api/admin/attachments/chunk" {
				action := ctx.Request().Input("action", "")
				if action == "" {
					action = ctx.Request().Query("action", "")
				}
				// 只记录 merge 操作，排除 init、upload 和 progress
				if action != "merge" {
					return
				}
			}
			// 在请求处理后再获取一次管理员ID（确保JWT中间件已执行）
			// 如果之前没有获取到，再次尝试从context获取
			if adminID == 0 {
				if admin, err := helpers.GetAdminFromContext(ctx); err == nil {
					adminID = admin.ID
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

			// 提前获取 traceCtx，用于日志记录
			traceCtx := traceid.DeriveContextFromHTTP(ctx)

			// 生成操作标题（只使用权限标识）
			title := utils.GetOperationTitleFromContext(ctx)
			if title == "operation.unknown" {
				// 如果无法生成标题，记录调试日志
				logger.ErrorfContext(traceCtx, "Failed to generate operation title, method: %s, path: %s", savedMethod, savedPath)
			}

			operationLog := models.OperationLog{
				AdminID:   savedAdminID,
				Method:    savedMethod,
				Path:      savedPath,
				Title:     title,
				IP:        savedIP,
				UserAgent: savedUserAgent,
				Request:   savedRequestBody,
				Status:    status,
				ErrorMsg:  errorMsg,
				Duration:  savedDuration,
			}

			// 异步记录日志，避免影响响应速度
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
