package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	// 注意：导出配置已迁移到数据库存储（configs表，group='storage'）
	// 后台管理系统 -> 系统管理 -> 配置管理 -> 文件存储配置
	// 如需在代码中使用导出配置，请从数据库读取，而不是从环境变量读取
	// 保留此配置仅用于框架初始化，实际导出配置请使用数据库中的配置
	config.Add("export", map[string]any{
		"disk":      config.Env("EXPORT_DISK", "public"),
		"path":      config.Env("EXPORT_PATH", "exports"),
		"format":    config.Env("EXPORT_FORMAT", "csv"),
		"url_prefix": config.Env("EXPORT_URL_PREFIX", ""),
	})
}
