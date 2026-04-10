package admin

import (
	"fmt"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/utils"
)

// getTimeInputOrQueryUTC 按 query 优先、input 兜底读取时间并转换为 UTC
func getTimeInputOrQueryUTC(ctx http.Context, paramName string) string {
	timeStr := ctx.Request().Query(paramName, "")
	if timeStr == "" {
		timeStr = ctx.Request().Input(paramName, "")
	}
	if timeStr == "" {
		return ""
	}

	return helpers.ConvertTimeToUTC(ctx, timeStr)
}

// parseOptionalTimeFromInputOrQuery 读取并解析可选时间参数，失败返回统一错误响应
func parseOptionalTimeFromInputOrQuery(ctx http.Context, paramName, invalidKey string) (time.Time, http.Response) {
	timeStr := getTimeInputOrQueryUTC(ctx, paramName)
	if timeStr == "" {
		return time.Time{}, nil
	}

	parsedTime, err := utils.ParseDateTime(timeStr)
	if err != nil {
		return time.Time{}, response.Error(ctx, nethttp.StatusBadRequest, invalidKey)
	}

	return parsedTime, nil
}

// validateTimeRangeResponse 验证时间范围，失败时返回统一错误响应
func validateTimeRangeResponse(ctx http.Context, startTime, endTime time.Time, maxMonths ...int) http.Response {
	valid, err := utils.ValidateTimeRange(startTime, endTime, maxMonths...)
	if valid {
		return nil
	}

	if timeRangeErr, ok := err.(*utils.TimeRangeError); ok {
		message := trans.Get(ctx, timeRangeErr.Key)
		if timeRangeErr.Params != nil {
			for key, value := range timeRangeErr.Params {
				placeholder := fmt.Sprintf("{%s}", key)
				message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
			}
		}
		return response.Error(ctx, nethttp.StatusBadRequest, message)
	}

	return response.Error(ctx, nethttp.StatusBadRequest, err.Error())
}
