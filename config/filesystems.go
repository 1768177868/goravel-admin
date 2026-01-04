package config

import (
	cosfacades "github.com/goravel/cos/facades"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/path"
	miniofacades "github.com/goravel/minio/facades"
	ossfacades "github.com/goravel/oss/facades"
	s3facades "github.com/goravel/s3/facades"
)

func init() {
	config := facades.Config()
	// 注意：文件存储配置说明
	// 1. 所有配置（密钥、bucket、region等）都从 .env 读取
	// 2. 后台管理系统 -> 系统管理 -> 配置管理 -> 文件存储配置 只用于选择驱动（file_disk）
	// 3. 应用启动后，AppServiceProvider 会从数据库读取 file_disk（如果没有则读取 storage_disk）并更新框架的默认磁盘
	// 4. 如需在代码中选择使用哪个磁盘，请使用 utils.GetConfigValue("storage", "file_disk", "local")
	//    如果 file_disk 未配置，会自动回退到 storage_disk，最后使用默认值 "local"
	config.Add("filesystems", map[string]any{
		// Default Filesystem Disk
		//
		// Here you may specify the default filesystem disk that should be used
		// by the framework. The "local" disk, as well as a variety of cloud
		// based disks are available to your application. Just store away!
		"default": config.Env("FILESYSTEM_DISK", "local"),
		// Filesystem Disks
		//
		// Here you may configure as many filesystem "disks" as you wish, and you
		// may even configure multiple disks of the same driver. Defaults have
		// been set up for each driver as an example of the required values.
		//
		// Supported Drivers: "local", "custom"
		"disks": map[string]any{
			"local": map[string]any{
				"driver": "local",
				"root":   path.Storage("app"),
			},
			"public": map[string]any{
				"driver": "local",
				"root":   path.Storage("app/public"),
				"url":    config.Env("APP_URL", "").(string) + "/storage",
			},
			"s3": map[string]any{
				"driver": "custom",
				"key":    config.Env("AWS_ACCESS_KEY_ID"),
				"secret": config.Env("AWS_ACCESS_KEY_SECRET"),
				"region": config.Env("AWS_REGION"),
				"bucket": config.Env("AWS_BUCKET"),
				"url":    config.Env("AWS_URL"),
				"via": func() (filesystem.Driver, error) {
					return s3facades.S3("s3") // The `s3` value is the `disks` key
				},
			},
			"cos": map[string]any{
				"driver": "custom",
				"key":    config.Env("TENCENT_ACCESS_KEY_ID"),
				"secret": config.Env("TENCENT_ACCESS_KEY_SECRET"),
				"url":    config.Env("TENCENT_URL"),
				"via": func() (filesystem.Driver, error) {
					return cosfacades.Cos("cos") // The `cos` value is the `disks` key
				},
			},
			"oss": map[string]any{
				"driver":   "custom",
				"key":      config.Env("ALIYUN_ACCESS_KEY_ID"),
				"secret":   config.Env("ALIYUN_ACCESS_KEY_SECRET"),
				"bucket":   config.Env("ALIYUN_BUCKET"),
				"url":      config.Env("ALIYUN_URL"),
				"endpoint": config.Env("ALIYUN_ENDPOINT"),
				"via": func() (filesystem.Driver, error) {
					return ossfacades.Oss("oss") // The `oss` value is the `disks` key
				},
			},
			"minio": map[string]any{
				"driver":   "custom",
				"key":      config.Env("MINIO_ACCESS_KEY_ID"),
				"secret":   config.Env("MINIO_ACCESS_KEY_SECRET"),
				"region":   config.Env("MINIO_REGION"),
				"bucket":   config.Env("MINIO_BUCKET"),
				"url":      config.Env("MINIO_URL"),
				"endpoint": config.Env("MINIO_ENDPOINT"),
				"ssl":      config.Env("MINIO_SSL", false),
				"via": func() (filesystem.Driver, error) {
					return miniofacades.Minio("minio") // The `minio` value is the `disks` key
				},
			},
		},
	})
}
