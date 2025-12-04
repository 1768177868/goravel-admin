package helpers

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

// GetCurrentTimezone 获取当前请求的时区
// 优先从请求头 X-Timezone 或 Timezone 获取
// 如果请求头没有或时区无效，使用配置的默认时区
func GetCurrentTimezone(ctx http.Context) string {
	// 优先从 X-Timezone 请求头获取
	timezone := ctx.Request().Header("X-Timezone", "")
	if timezone == "" {
		// 尝试从 Timezone 请求头获取
		timezone = ctx.Request().Header("Timezone", "")
	}
	if timezone == "" {
		// 尝试从查询参数获取
		timezone = ctx.Request().Input("timezone")
	}

	// 如果从请求中获取到了时区，规范化并返回
	if timezone != "" {
		return NormalizeTimezone(timezone)
	}

	// 如果都没有，使用配置的默认时区
	defaultTimezone := facades.Config().GetString("app.timezone", carbon.UTC)
	return NormalizeTimezone(defaultTimezone)
}

// isValidTimezone 验证时区是否有效
func isValidTimezone(timezone string) bool {
	if timezone == "" {
		return false
	}

	// 尝试加载时区
	_, err := time.LoadLocation(timezone)
	return err == nil
}

// NormalizeTimezone 规范化时区名称（处理常见别名）
func NormalizeTimezone(timezone string) string {
	timezone = str.Of(timezone).Trim().String()
	if str.Of(timezone).IsEmpty() {
		return carbon.UTC
	}

	// 转换为标准时区名称
	timezoneMap := map[string]string{
		"UTC":       "UTC",
		"GMT":       "UTC",
		"PST":       "America/Los_Angeles",
		"PDT":       "America/Los_Angeles",
		"EST":       "America/New_York",
		"EDT":       "America/New_York",
		"CST":       "America/Chicago",
		"CDT":       "America/Chicago",
		"MST":       "America/Denver",
		"MDT":       "America/Denver",
		"Beijing":   "Asia/Shanghai",
		"Shanghai":  "Asia/Shanghai",
		"Hong Kong": "Asia/Hong_Kong",
		"Tokyo":     "Asia/Tokyo",
		"Seoul":     "Asia/Seoul",
		"Singapore": "Asia/Singapore",
		"London":    "Europe/London",
		"Paris":     "Europe/Paris",
		"Berlin":    "Europe/Berlin",
		"Moscow":    "Europe/Moscow",
		"Sydney":    "Australia/Sydney",
		"Melbourne": "Australia/Melbourne",
	}

	if normalized, ok := timezoneMap[timezone]; ok {
		return normalized
	}

	// 如果时区有效，直接返回
	if isValidTimezone(timezone) {
		return timezone
	}

	// 默认返回 UTC
	return carbon.UTC
}

// ConvertTimeToTimezone 将时间字符串转换为指定时区
// 返回转换后的时间字符串
func ConvertTimeToTimezone(timeStr string, timezone string) string {
	if timeStr == "" {
		return ""
	}
	// 规范化时区
	timezone = NormalizeTimezone(timezone)

	// 解析时间字符串
	dt := carbon.Parse(timeStr)
	if dt.IsZero() {
		return timeStr
	}
	// 转换时区并返回格式化的字符串
	return dt.SetTimezone(timezone).ToDateTimeString()
}

// ConvertTimeByContext 根据请求头中的时区转换时间字符串
// 返回转换后的时间字符串
func ConvertTimeByContext(ctx http.Context, timeStr string) string {
	if timeStr == "" {
		return ""
	}
	timezone := GetCurrentTimezone(ctx)
	return ConvertTimeToTimezone(timeStr, timezone)
}

// ConvertTimeToUTC 将本地时区的时间字符串转换为 UTC 时间字符串（用于数据库查询）
// timeStr: 前端传入的时间字符串（本地时区格式，如 "2025-11-25 14:00:00"）
// ctx: 请求上下文，用于获取当前时区
// 返回: UTC 时间字符串（如 "2025-11-25 06:00:00"）
func ConvertTimeToUTC(ctx http.Context, timeStr string) string {
	if timeStr == "" {
		return ""
	}

	// 获取当前请求的时区
	timezone := GetCurrentTimezone(ctx)

	// 如果已经是 UTC，直接返回
	if timezone == carbon.UTC || timezone == "UTC" {
		return timeStr
	}

	// 加载时区
	targetLoc, err := time.LoadLocation(timezone)
	if err != nil {
		// 如果时区无效，假设是 UTC
		return timeStr
	}
	utcLoc, _ := time.LoadLocation("UTC")

	// 解析时间字符串（假设是本地时区格式）
	// 尝试多种格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02T15:04:05.000Z07:00",
		time.RFC3339,
	}

	var t time.Time
	var parseErr error
	for _, format := range formats {
		t, parseErr = time.ParseInLocation(format, timeStr, targetLoc)
		if parseErr == nil {
			break
		}
	}

	if parseErr != nil {
		// 如果所有格式都失败，尝试使用 carbon 解析
		dt := carbon.Parse(timeStr)
		if dt.IsZero() {
			return timeStr
		}
		// 假设解析的时间是本地时区，转换为 UTC
		return dt.SetTimezone(carbon.UTC).ToDateTimeString()
	}

	// 转换为 UTC 并格式化
	return t.In(utcLoc).Format("2006-01-02 15:04:05")
}

// GetTimeQueryParam 获取并转换时间查询参数（统一处理时间查询）
// 自动将前端传入的本地时区时间转换为 UTC 时间用于数据库查询
// 支持常见的时间查询参数名称：start_time, end_time, created_at_start, created_at_end, updated_at_start, updated_at_end
func GetTimeQueryParam(ctx http.Context, paramName string) string {
	timeStr := ctx.Request().Query(paramName, "")
	if timeStr == "" {
		return ""
	}
	return ConvertTimeToUTC(ctx, timeStr)
}
