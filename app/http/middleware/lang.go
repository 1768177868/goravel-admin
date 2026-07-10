package middleware

import (
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/utils"
)

// Lang 多语言中间件，从请求头获取语言
func Lang() httpcontract.Middleware {
	return newMiddleware("lang", func(ctx httpcontract.Context) {
		// 使用通用工具函数获取语言
		lang := utils.GetCurrentLanguage(ctx)
		facades.App().SetLocale(ctx, lang)
		ctx.Request().Next()
	})
}
