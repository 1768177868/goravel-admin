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

		// ========== 操作日志过滤规则 ==========

		// 需要记录操作日志的 HTTP 方法（仅写操作）
		"allowed_methods": []string{"POST", "PUT", "PATCH", "DELETE"},

		// 精确匹配排除的路径（这些路径的请求不记录操作日志）
		"excluded_paths": []string{
			"/api/admin/login",
			"/api/admin/info",
		},

		// 前缀匹配排除的路径（以这些前缀开头的请求不记录操作日志）
		"excluded_path_prefixes": []string{
			"/api/admin/code-generator/",
		},

		// ========== 审计变更对比 ==========

		// 需要记录变更详情的数据表（PUT 时自动对比修改前后差异，DELETE 时记录删除快照）
		// 新增表只需在此追加表名即可
		"auditable_tables": []string{
			"admins",
			"roles",
			"menus",
			"departments",
			"blacklists",
		},
	})
}
