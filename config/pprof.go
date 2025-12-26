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
		// 最大失败尝试次数（防止暴力破解）
		// 超过此次数后，IP 将被临时封禁
		// 示例：PPROF_MAX_ATTEMPTS=5
		"max_attempts": config.Env("PPROF_MAX_ATTEMPTS", 5),
		// 封禁时长（秒）
		// IP 被封禁后，需要等待此时间后才能再次尝试
		// 示例：PPROF_BLOCK_DURATION=300（5分钟）
		"block_duration": config.Env("PPROF_BLOCK_DURATION", 300),
		// 失败计数重置时长（秒）
		// 如果在此时间内没有新的失败尝试，失败计数将被重置
		// 示例：PPROF_RESET_DURATION=600（10分钟）
		"reset_duration": config.Env("PPROF_RESET_DURATION", 600),
	})
}
