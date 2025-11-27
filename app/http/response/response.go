package response

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/helpers"
	"goravel/app/http/trans"
	"goravel/app/services"
	"goravel/app/utils/traceid"
)

// Success 成功响应（支持多语言，自动包含 trace_id）
func Success(ctx http.Context, messageKey string, data ...http.Json) http.Response {
	message := trans.Get(ctx, messageKey)

	response := http.Json{
		"code":    200,
		"message": message,
	}

	// 自动包含 trace_id，方便前端追踪
	if traceID := traceid.FromHTTPContext(ctx); traceID != "" {
		response["trace_id"] = traceID
	}

	if len(data) > 0 {
		// 转换时间字段到对应时区
		convertedData := helpers.ConvertTimesInData(ctx, data[0])
		if convertedMap, ok := convertedData.(map[string]any); ok {
			response["data"] = http.Json(convertedMap)
		} else {
			response["data"] = convertedData
		}
	}
	return ctx.Response().Success().Json(response)
}

// SuccessWithHeader 成功响应（支持多语言和自定义Header，自动包含 trace_id）
func SuccessWithHeader(ctx http.Context, messageKey string, headerKey, headerValue string, data ...http.Json) http.Response {
	message := trans.Get(ctx, messageKey)

	response := http.Json{
		"code":    200,
		"message": message,
	}

	// 自动包含 trace_id，方便前端追踪
	if traceID := traceid.FromHTTPContext(ctx); traceID != "" {
		response["trace_id"] = traceID
	}

	if len(data) > 0 {
		// 转换时间字段到对应时区
		convertedData := helpers.ConvertTimesInData(ctx, data[0])
		if convertedMap, ok := convertedData.(map[string]any); ok {
			response["data"] = http.Json(convertedMap)
		} else {
			response["data"] = convertedData
		}
	}
	return ctx.Response().Header(headerKey, headerValue).Success().Json(response)
}

// Error 错误响应（支持多语言，自动包含 trace_id）
func Error(ctx http.Context, code int, messageKey string) http.Response {
	message := trans.Get(ctx, messageKey)

	response := http.Json{
		"code":    code,
		"message": message,
	}

	// 自动包含 trace_id，方便前端显示和用户报告错误
	if traceID := traceid.FromHTTPContext(ctx); traceID != "" {
		response["trace_id"] = traceID
	}

	return ctx.Response().Json(code, response)
}

// ValidationError 验证错误响应（支持多语言，自动包含 trace_id）
func ValidationError(ctx http.Context, code int, messageKey string, errors map[string]map[string]string) http.Response {
	message := trans.Get(ctx, messageKey)

	response := http.Json{
		"code":    code,
		"message": message,
		"errors":  errors,
	}

	// 自动包含 trace_id，方便前端显示和用户报告错误
	if traceID := traceid.FromHTTPContext(ctx); traceID != "" {
		response["trace_id"] = traceID
	}

	return ctx.Response().Json(code, response)
}

// Paginate 分页响应（支持多语言，自动包含 trace_id）
func Paginate(ctx http.Context, messageKey string, list any, total int64, page, pageSize int) http.Response {
	message := trans.Get(ctx, messageKey)

	// 转换列表中的时间字段到对应时区
	convertedList := helpers.ConvertTimesInData(ctx, list)

	response := http.Json{
		"code":    200,
		"message": message,
		"data": http.Json{
			"list":      convertedList,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	}

	// 自动包含 trace_id，方便前端追踪
	if traceID := traceid.FromHTTPContext(ctx); traceID != "" {
		response["trace_id"] = traceID
	}

	return ctx.Response().Success().Json(response)
}

// Export 导出响应（支持多语言）
// headers: CSV表头（可以是翻译键数组或字符串数组）
// data: 数据行，每行是一个字符串切片
// filename: 文件名（不含扩展名）
func Export(ctx http.Context, messageKey string, headers []string, data [][]string, filename string) http.Response {
	message := trans.Get(ctx, messageKey)

	// 翻译表头（如果表头是翻译键，则翻译；如果是普通字符串，则保持原样）
	translatedHeaders := make([]string, len(headers))
	for i, header := range headers {
		// 尝试翻译，如果翻译键不存在则返回原字符串
		translated := trans.Get(ctx, header)
		if translated == header {
			// 如果翻译结果和原字符串相同，说明不是翻译键，直接使用原字符串
			translatedHeaders[i] = header
		} else {
			// 如果翻译成功，使用翻译后的文本
			translatedHeaders[i] = translated
		}
	}

	exportService := services.NewExportService()
	filePath, err := exportService.ExportToFile(translatedHeaders, data, filename)
	if err != nil {
		return Error(ctx, http.StatusInternalServerError, "export_failed")
	}

	exportURL := exportService.GetExportURL(filePath)

	response := http.Json{
		"code":    200,
		"message": message,
		"data": http.Json{
			"file_path": filePath,
			"file_url":  exportURL,
		},
	}

	// 自动包含 trace_id，方便前端追踪
	if traceID := traceid.FromHTTPContext(ctx); traceID != "" {
		response["trace_id"] = traceID
	}

	return ctx.Response().Success().Json(response)
}
