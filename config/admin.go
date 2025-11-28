package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("admin", map[string]any{
		// Developer Admin IDs
		//
		// The developer admin IDs that should be hidden from the admin list.
		// These admins cannot be deleted or disabled, and will not appear in the list.
		// Usually these are robot accounts, script accounts, etc.
		// You can set multiple IDs separated by commas in the .env file.
		//
		// Examples in .env file:
		//   Single ID: ADMIN_DEVELOPER_IDS=2
		//   Multiple IDs: ADMIN_DEVELOPER_IDS=2,3,4
		//   Multiple IDs with spaces: ADMIN_DEVELOPER_IDS=2, 3, 4
		"developer_ids": config.Env("ADMIN_DEVELOPER_IDS", "2"), // Default: ID 2 (developer admin)
		// 注意：验证码配置已迁移到数据库存储（configs表，group='captcha'）
		// 后台管理系统 -> 系统管理 -> 配置管理 -> 验证码配置
		// 如需在代码中使用验证码配置，请从数据库读取，而不是从环境变量读取
		"login_captcha_enabled": config.Env("ADMIN_LOGIN_CAPTCHA_ENABLED", false),
		"login_captcha_expire":  config.Env("ADMIN_LOGIN_CAPTCHA_EXPIRE", 120), // seconds
	})
	config.Add("role", map[string]any{
		// Protected Role Slugs
		//
		// These role slugs cannot be deleted or disabled. Usually these are system roles
		// or super administrator roles that are critical to the system operation.
		// You can add multiple slugs separated by commas in the .env file.
		//
		// Examples in .env file:
		//   Single slug: ROLE_PROTECTED_SLUGS=super-admin
		//   Multiple slugs: ROLE_PROTECTED_SLUGS=super-admin,admin
		//   Multiple slugs with spaces: ROLE_PROTECTED_SLUGS=super-admin, admin
		"protected_slugs": config.Env("ROLE_PROTECTED_SLUGS", "super-admin"), // Default: super-admin
	})
}
