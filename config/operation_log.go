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
	})
}
