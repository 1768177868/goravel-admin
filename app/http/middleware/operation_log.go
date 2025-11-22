package middleware

import (
	"encoding/json"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// OperationLog 操作日志中间件
func OperationLog() http.Middleware {
	return func(ctx http.Context) {
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

		// 获取管理员ID
		var adminID uint
		var admin models.Admin
		if err := facades.Auth(ctx).Guard("admin").User(&admin); err == nil {
			adminID = admin.ID
		}

		// 继续处理请求
		ctx.Request().Next()

		// 计算耗时
		duration := int(time.Since(startTime).Milliseconds())

		// 只记录新增、修改、删除操作（POST、PUT、PATCH、DELETE），排除 GET 请求
		// 同时排除登录和info接口
		if (method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE") &&
			path != "/admin/login" && path != "/admin/info" {
			// 默认状态为成功
			status := uint8(1)
			var errorMsg string

			operationLog := models.OperationLog{
				AdminID:   adminID,
				Method:    method,
				Path:      path,
				IP:        ip,
				UserAgent: userAgent,
				Request:   requestBody,
				Status:    status,
				ErrorMsg:  errorMsg,
				Duration:  duration,
			}

			// 异步记录日志，避免影响响应速度
			go func() {
				facades.Orm().Query().Create(&operationLog)
			}()
		}
	}
}

