package middleware

import (
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// Lang 多语言中间件，从请求头获取语言
func Lang() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		// 优先从请求头 Accept-Language 获取语言
		acceptLanguage := ctx.Request().Header("Accept-Language", "")
		lang := parseAcceptLanguage(acceptLanguage)

		// 如果请求头没有，尝试从查询参数获取
		if lang == "" {
			lang = ctx.Request().Input("lang")
		}

		// 如果都没有，使用默认语言
		if lang == "" {
			lang = facades.Config().GetString("app.locale")
		}

		// 验证语言是否支持（只支持 cn 和 en）
		if lang != "cn" && lang != "en" {
			lang = facades.Config().GetString("app.locale")
		}

		facades.App().SetLocale(ctx, lang)

		ctx.Request().Next()
	}
}

// parseAcceptLanguage 解析 Accept-Language 请求头
// 格式: "zh-CN,zh;q=0.9,en;q=0.8" 或 "en-US,en;q=0.9"
func parseAcceptLanguage(acceptLanguage string) string {
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
	
	// 处理语言代码（如 zh-CN -> zh, en-US -> en）
	if strings.HasPrefix(firstLang, "zh") {
		return "cn"
	}
	if strings.HasPrefix(firstLang, "en") {
		return "en"
	}

	return ""
}
