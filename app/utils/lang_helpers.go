package utils

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// ParseAcceptLanguage 解析 Accept-Language 请求头
// 格式: "zh-CN,zh;q=0.9,en;q=0.8" 或 "en-US,en;q=0.9"
// 返回: 语言代码（"cn" 或 "en"），如果无法解析则返回空字符串
func ParseAcceptLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return ""
	}

	// 分割语言列表
	languages := strings.Split(acceptLanguage, ",")
	if len(languages) == 0 {
		return ""
	}

	// 取第一个语言
	firstLang := strings.TrimSpace(languages[0])

	// 移除质量值（如果有）
	if idx := strings.Index(firstLang, ";"); idx != -1 {
		firstLang = firstLang[:idx]
	}

	// 转换为小写并提取语言代码
	firstLang = strings.ToLower(strings.TrimSpace(firstLang))

	// 处理语言代码（如 zh-CN -> cn, en-US -> en）
	if strings.HasPrefix(firstLang, "zh") {
		return "cn"
	}
	if strings.HasPrefix(firstLang, "en") {
		return "en"
	}

	return ""
}

// GetCurrentLanguage 获取当前请求的语言（从 HTTP Context）
// 优先从请求头 Accept-Language 获取，其次从查询参数获取，最后使用默认语言
func GetCurrentLanguage(ctx http.Context) string {
	// 优先从请求头 Accept-Language 获取语言
	acceptLanguage := ctx.Request().Header("Accept-Language", "")
	lang := ParseAcceptLanguage(acceptLanguage)

	// 如果请求头没有，尝试从查询参数获取
	if lang == "" {
		lang = ctx.Request().Input("lang")
	}

	// 如果都没有，使用默认语言
	if lang == "" {
		lang = facades.Config().GetString("app.locale")
	}

	// 验证并规范化语言代码
	return NormalizeLanguage(lang)
}

// NormalizeLanguage 验证并规范化语言代码
// 只支持 "cn" 和 "en"，其他值会返回默认语言
func NormalizeLanguage(lang string) string {
	// 验证语言是否支持（只支持 cn 和 en）
	if lang != "cn" && lang != "en" {
		return facades.Config().GetString("app.locale", "cn")
	}
	return lang
}

