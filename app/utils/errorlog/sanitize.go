package errorlog

import (
	"strings"
	"unicode"
)

// sanitizeLogString 清理日志字符串，防止日志注入攻击
// 移除或转义危险字符：
//   - 换行符 (\n, \r) - 可能用于伪造多行日志
//   - 制表符 (\t) - 可能用于格式化攻击
//   - 控制字符 - 可能用于隐藏攻击痕迹
//
// 参数:
//   - s: 要清理的字符串
//   - maxLength: 最大长度限制（0 表示不限制）
//
// 返回:
//   - 清理后的字符串
func sanitizeLogString(s string, maxLength int) string {
	if s == "" {
		return s
	}

	// 移除控制字符和换行符
	var builder strings.Builder
	builder.Grow(len(s))

	for _, r := range s {
		// 允许打印字符和空格，移除控制字符
		if unicode.IsPrint(r) || r == ' ' {
			builder.WriteRune(r)
		} else {
			// 将控制字符替换为转义序列（可选，或直接移除）
			// 这里选择移除，因为转义序列也可能被利用
		}
	}

	result := builder.String()

	// 限制长度
	if maxLength > 0 && len(result) > maxLength {
		result = result[:maxLength] + "...[truncated]"
	}

	return result
}

// sanitizeLogLevel 验证和清理日志级别，防止伪造
// 只允许预定义的日志级别
func sanitizeLogLevel(level string) string {
	level = strings.ToLower(strings.TrimSpace(level))

	// 只允许预定义的级别
	allowedLevels := map[string]string{
		"error":   "error",
		"warning": "warning",
		"warn":    "warning", // 兼容 warn
		"info":    "info",
		"debug":   "debug",
	}

	if validLevel, ok := allowedLevels[level]; ok {
		return validLevel
	}

	// 如果级别无效，默认返回 error（最安全的级别）
	return "error"
}

// sanitizeLogArgs 清理日志参数，防止注入攻击
// 将参数转换为安全的字符串表示
func sanitizeLogArgs(args ...any) []any {
	if len(args) == 0 {
		return args
	}

	sanitized := make([]any, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			// 字符串参数：清理控制字符
			sanitized[i] = sanitizeLogString(v, 0)
		case []byte:
			// 字节数组：转换为字符串后清理
			sanitized[i] = sanitizeLogString(string(v), 0)
		case error:
			// 错误对象：清理错误消息
			if v != nil {
				sanitized[i] = sanitizeLogString(v.Error(), 0)
			} else {
				sanitized[i] = "<nil>"
			}
		default:
			// 其他类型：保持原样（数字、布尔等通常是安全的）
			sanitized[i] = arg
		}
	}

	return sanitized
}

// sanitizeLogFormat 清理日志格式字符串，防止格式字符串注入
// 移除危险的控制字符，但保留格式占位符（%v, %s 等）
func sanitizeLogFormat(format string) string {
	if format == "" {
		return format
	}

	// 移除换行符和控制字符，但保留格式占位符
	var builder strings.Builder
	builder.Grow(len(format))

	for i := 0; i < len(format); i++ {
		r := rune(format[i])

		// 允许打印字符、空格和格式占位符
		if unicode.IsPrint(r) || r == ' ' {
			builder.WriteRune(r)
		} else if r == '\n' || r == '\r' || r == '\t' {
			// 将换行符和制表符替换为空格
			builder.WriteRune(' ')
		}
		// 其他控制字符直接移除
	}

	result := builder.String()

	// 限制格式字符串长度（防止过长的格式字符串攻击）
	maxFormatLength := 1000
	if len(result) > maxFormatLength {
		result = result[:maxFormatLength] + "...[truncated]"
	}

	return result
}

// sanitizeAttributes 清理 attributes map，防止注入攻击
func sanitizeAttributes(attributes map[string]any) map[string]any {
	if attributes == nil || len(attributes) == 0 {
		return attributes
	}

	sanitized := make(map[string]any, len(attributes))
	for key, value := range attributes {
		// 清理 key
		sanitizedKey := sanitizeLogString(key, 100)

		// 清理 value
		switch v := value.(type) {
		case string:
			sanitized[sanitizedKey] = sanitizeLogString(v, 1000)
		case []byte:
			sanitized[sanitizedKey] = sanitizeLogString(string(v), 1000)
		case error:
			if v != nil {
				sanitized[sanitizedKey] = sanitizeLogString(v.Error(), 1000)
			} else {
				sanitized[sanitizedKey] = "<nil>"
			}
		case map[string]any:
			// 递归清理嵌套 map
			sanitized[sanitizedKey] = sanitizeAttributes(v)
		case []any:
			// 清理数组
			sanitizedArray := make([]any, len(v))
			for i, item := range v {
				if str, ok := item.(string); ok {
					sanitizedArray[i] = sanitizeLogString(str, 500)
				} else {
					sanitizedArray[i] = item
				}
			}
			sanitized[sanitizedKey] = sanitizedArray
		default:
			// 其他类型（数字、布尔等）保持原样
			sanitized[sanitizedKey] = value
		}
	}

	return sanitized
}
