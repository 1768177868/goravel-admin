package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/helpers"
)

// TimezoneQuery 时区查询参数转换中间件
// 自动将查询参数中的时间字段从本地时区转换为 UTC 时间用于数据库查询
func TimezoneQuery() http.Middleware {
	return func(ctx http.Context) {
		// 定义需要转换的时间查询参数名称
		timeParams := []string{
			"start_time",
			"end_time",
			"created_at_start",
			"created_at_end",
			"updated_at_start",
			"updated_at_end",
			"deleted_at_start",
			"deleted_at_end",
		}

		// 转换每个时间参数
		for _, paramName := range timeParams {
			timeStr := ctx.Request().Query(paramName, "")
			if timeStr != "" {
				// 转换为 UTC 时间并存储在 context 中
				utcTime := helpers.ConvertTimeToUTC(ctx, timeStr)
				// 将转换后的时间存储在 context 中，供控制器使用
				ctx.WithValue("timezone_query_"+paramName, utcTime)
			}
		}

		// 继续处理请求
		ctx.Request().Next()
	}
}

