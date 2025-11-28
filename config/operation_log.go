package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("operation_log", map[string]any{
		// 敏感字段列表（用于操作日志记录时自动隐藏）
		// 这些字段的值在记录到操作日志时会被替换为 "***"
		"sensitive_fields": []string{
			"password",
			"old_password",
			"new_password",
			"confirm_password",
			"token",
			"access_token",
			"refresh_token",
			"api_key",
			"apikey",
			"secret",
			"secret_key",
			"private_key",
			"authorization",
		},
		// 敏感字段关键词（字段名包含这些关键词的也会被隐藏）
		"sensitive_keywords": []string{
			"password",
			"token",
			"secret",
			"key",
		},
		// 复数转单数映射（用于操作标题生成）
		// 如果资源名称不在这个映射中，将使用通用规则（去掉末尾的 s）
		"plural_to_singular": map[string]string{
			"admins":         "admin",
			"roles":          "role",
			"permissions":    "permission",
			"menus":          "menu",
			"departments":    "department",
			"dictionaries":   "dictionary",
			"configs":        "config",
			"blacklists":     "blacklist",
			"online_users":   "online_user",
			"operation_logs": "operation_log",
			"login_logs":     "login_log",
			"system_logs":    "system_log",
			"notifications":  "notification",
			"users":          "user",
			"tokens":         "token",
		},
		// 所有可能的操作标题列表（用于下拉选择）
		// 格式：operation.{action}_{resource}
		"all_title_keys": []string{
			// 管理员相关
			"operation.create_admin",
			"operation.update_admin",
			"operation.delete_admin",
			"operation.batch_delete_admin",
			// 角色相关
			"operation.create_role",
			"operation.update_role",
			"operation.delete_role",
			"operation.batch_delete_role",
			// 权限相关
			"operation.create_permission",
			"operation.update_permission",
			"operation.delete_permission",
			"operation.batch_delete_permission",
			// 菜单相关
			"operation.create_menu",
			"operation.update_menu",
			"operation.delete_menu",
			"operation.batch_delete_menu",
			// 部门相关
			"operation.create_department",
			"operation.update_department",
			"operation.delete_department",
			"operation.batch_delete_department",
			// 字典相关
			"operation.create_dictionary",
			"operation.update_dictionary",
			"operation.delete_dictionary",
			"operation.batch_delete_dictionary",
			// 配置相关
			"operation.create_config",
			"operation.update_config",
			"operation.delete_config",
			"operation.batch_delete_config",
			// 黑名单相关
			"operation.create_blacklist",
			"operation.update_blacklist",
			"operation.delete_blacklist",
			"operation.batch_delete_blacklist",
			// 在线用户相关
			"operation.create_online_user",
			"operation.update_online_user",
			"operation.kick_out_user",
			// 操作日志相关
			"operation.create_operation_log",
			"operation.update_operation_log",
			"operation.delete_operation_log",
			"operation.batch_delete_operation_log",
			"operation.clean_operation_log",
			// 登录日志相关
			"operation.create_login_log",
			"operation.update_login_log",
			"operation.delete_login_log",
			"operation.batch_delete_login_log",
			"operation.clean_login_log",
			// 系统日志相关
			"operation.create_system_log",
			"operation.update_system_log",
			"operation.delete_system_log",
			"operation.batch_delete_system_log",
			"operation.clean_system_log",
			// 通知相关
			"operation.create_notification",
			"operation.update_notification",
			"operation.delete_notification",
			"operation.batch_delete_notification",
			// 用户相关
			"operation.create_user",
			"operation.update_user",
			"operation.delete_user",
			// 其他
			"operation.update_profile",
			"operation.update_password",
			"operation.test_email_config",
			"operation.clean_operation_log",
			"operation.clean_login_log",
			"operation.clean_system_log",
			"operation.batch_kick_out_user",
			// "operation.unknown",
		},
	})
}
