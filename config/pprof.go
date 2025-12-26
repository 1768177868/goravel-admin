package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("pprof", map[string]any{
		// 是否启用 pprof 性能分析
		// 默认：仅在 APP_DEBUG=true 时启用
		// 生产环境可通过设置 PPROF_ENABLED=true 来启用（需要配合 IP 白名单或 token 使用）
		"enabled": config.Env("PPROF_ENABLED", false),
		// 允许访问的 IP 地址列表，逗号分隔
		// 支持单个 IP（如：127.0.0.1）和 CIDR 格式（如：192.168.1.0/24）
		// 示例：PPROF_ALLOWED_IPS=127.0.0.1,192.168.1.100,10.0.0.0/8
		// 如果为空，则不限制 IP（不推荐在生产环境使用）
		"allowed_ips": config.Env("PPROF_ALLOWED_IPS", ""),
		// 访问 token（可选）
		// 如果设置，需要在请求头 X-Pprof-Token 或查询参数 token 中提供
		// 示例：PPROF_TOKEN=your-secret-token-here
		"token": config.Env("PPROF_TOKEN", ""),
	})
}

