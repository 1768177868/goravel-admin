package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("export", map[string]any{
		// Export Disk
		//
		// 指定导出文件存储的磁盘，可选值: "local", "s3", "oss", "cos", "minio"
		// 默认使用 "local" 存储到本地
		"disk": config.Env("EXPORT_DISK", "local"),

		// Export Path
		//
		// 导出文件的存储路径（相对于磁盘根目录）
		"path": config.Env("EXPORT_PATH", "exports"),

		// Export Format
		//
		// 导出文件格式，可选值: "csv", "xlsx"
		// 默认使用 "csv"
		"format": config.Env("EXPORT_FORMAT", "csv"),

		// Export URL Prefix
		//
		// 导出文件的访问URL前缀（用于生成下载链接）
		// 如果为空，则返回文件路径
		"url_prefix": config.Env("EXPORT_URL_PREFIX", ""),
	})
}
