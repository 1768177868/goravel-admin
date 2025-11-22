package response

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/helpers"
	"goravel/app/http/trans"
)

// Success 成功响应（支持多语言）
func Success(ctx http.Context, messageKey string, data ...http.Json) http.Response {
	message := trans.Get(ctx, messageKey)

	response := http.Json{
		"code":    200,
		"message": message,
	}
	if len(data) > 0 {
		// 转换时间字段到对应时区
		convertedData := helpers.ConvertTimesInData(ctx, data[0])
		if convertedMap, ok := convertedData.(map[string]interface{}); ok {
			response["data"] = http.Json(convertedMap)
		} else {
			response["data"] = convertedData
		}
	}
	return ctx.Response().Success().Json(response)
}

// SuccessWithHeader 成功响应（支持多语言和自定义Header）
func SuccessWithHeader(ctx http.Context, messageKey string, headerKey, headerValue string, data ...http.Json) http.Response {
	message := trans.Get(ctx, messageKey)

	response := http.Json{
		"code":    200,
		"message": message,
	}
	if len(data) > 0 {
		// 转换时间字段到对应时区
		convertedData := helpers.ConvertTimesInData(ctx, data[0])
		if convertedMap, ok := convertedData.(map[string]interface{}); ok {
			response["data"] = http.Json(convertedMap)
		} else {
			response["data"] = convertedData
		}
	}
	return ctx.Response().Header(headerKey, headerValue).Success().Json(response)
}

// Error 错误响应（支持多语言）
func Error(ctx http.Context, code int, messageKey string) http.Response {
	message := trans.Get(ctx, messageKey)

	return ctx.Response().Json(code, http.Json{
		"code":    code,
		"message": message,
	})
}

// ValidationError 验证错误响应（支持多语言）
func ValidationError(ctx http.Context, code int, messageKey string, errors map[string]map[string]string) http.Response {
	message := trans.Get(ctx, messageKey)

	return ctx.Response().Json(code, http.Json{
		"code":    code,
		"message": message,
		"errors":  errors,
	})
}

// Paginate 分页响应（支持多语言）
func Paginate(ctx http.Context, messageKey string, list interface{}, total int64, page, pageSize int) http.Response {
	message := trans.Get(ctx, messageKey)

	// 转换列表中的时间字段到对应时区
	convertedList := helpers.ConvertTimesInData(ctx, list)

	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": message,
		"data": http.Json{
			"list":      convertedList,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
