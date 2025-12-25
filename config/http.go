package config

import (
	"github.com/gin-gonic/gin/render"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	fiberfacades "github.com/goravel/fiber/facades"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/gin"
	ginfacades "github.com/goravel/gin/facades"
)

func init() {
	config := facades.Config()
	config.Add("http", map[string]any{
		// HTTP Driver
		"default": "gin",
		// HTTP Drivers
		"drivers": map[string]any{
			"gin": map[string]any{
				// Optional, default is 4096 KB
				"body_limit":   4096,
				"header_limit": 4096,
				"route": func() (route.Route, error) {
					return ginfacades.Route("gin"), nil
				},
				// Optional, default is http/template
				"template": func() (render.HTMLRender, error) {
					return gin.DefaultTemplate()
				},
			},
			"fiber": map[string]any{
				// immutable mode, see https://docs.gofiber.io/#zero-allocation
				// WARNING: This option is dangerous. Only change it if you fully understand the potential consequences.
				"immutable": true,
				// prefork mode, see https://docs.gofiber.io/api/fiber/#config
				"prefork": false,
				// Optional, default is 4096 KB
				"body_limit":   4096,
				"header_limit": 4096,
				"route": func() (route.Route, error) {
					return fiberfacades.Route("fiber"), nil
				},
				// Optional, default is "html/template"
				"template": func() (fiber.Views, error) {
					return html.New(path.Resource("views"), ".tmpl"), nil
				},
			},
		},
		// HTTP URL
		"url": config.Env("APP_URL", "http://localhost"),
		// HTTP Host
		"host": config.Env("APP_HOST", "127.0.0.1"),
		// HTTP Port
		"port": config.Env("APP_PORT", "3000"),
		// HTTP Timeout, default is 1800 seconds (30 minutes)
		// 注意：这是全局超时时间，适用于所有接口
		// 大文件上传接口（/attachments/chunk）需要更长的超时时间，特别是 merge 操作
		// 可以通过环境变量 HTTP_REQUEST_TIMEOUT 自定义（单位：秒）
		// 建议值：
		//   - 普通接口：30-60 秒
		//   - 支持大文件上传：1800 秒（30 分钟）或更长
		// 如果使用反向代理（如 Nginx），建议在 Nginx 层面为特定接口单独配置超时
		"request_timeout": config.GetInt("HTTP_REQUEST_TIMEOUT", 60),
		// HTTPS Configuration
		"tls": map[string]any{
			// HTTPS Host
			"host": config.Env("APP_HOST", "127.0.0.1"),
			// HTTPS Port
			"port": config.Env("APP_PORT", "3000"),
			// SSL Certificate, you can put the certificate in /public folder
			"ssl": map[string]any{
				// ca.pem
				"cert": "",
				// ca.key
				"key": "",
			},
		},
		"client": map[string]any{
			"base_url":                config.GetString("HTTP_CLIENT_BASE_URL", "http://127.0.0.1:3000"),
			"timeout":                 config.GetDuration("HTTP_CLIENT_TIMEOUT"),
			"max_idle_conns":          config.GetInt("HTTP_CLIENT_MAX_IDLE_CONNS"),
			"max_idle_conns_per_host": config.GetInt("HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST"),
			"max_conns_per_host":      config.GetInt("HTTP_CLIENT_MAX_CONN_PER_HOST"),
			"idle_conn_timeout":       config.GetDuration("HTTP_CLIENT_IDLE_CONN_TIMEOUT"),
		},
	})
}
