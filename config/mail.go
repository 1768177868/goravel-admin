package config

import "github.com/goravel/framework/facades"

func init() {
	config := facades.Config()
	// 注意：邮件配置已迁移到数据库存储（configs表，group='email'）
	// 后台管理系统 -> 系统管理 -> 配置管理 -> 邮箱配置
	// 如需在代码中使用邮件功能，请从数据库读取配置，而不是从环境变量读取
	// 保留此配置仅用于框架初始化，实际邮件配置请使用数据库中的配置
	config.Add("mail", map[string]any{
		"host": config.Env("MAIL_HOST", ""),
		"port": config.Env("MAIL_PORT", 587),
		"from": map[string]any{
			"address": config.Env("MAIL_FROM_ADDRESS", "hello@example.com"),
			"name":    config.Env("MAIL_FROM_NAME", "Example"),
		},
		"username": config.Env("MAIL_USERNAME"),
		"password": config.Env("MAIL_PASSWORD"),
	})
}
