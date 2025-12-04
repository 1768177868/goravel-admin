package utils

import (
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
)

// IsSensitiveField 检查字段名是否是敏感字段
func IsSensitiveField(fieldName string) bool {
	keyLower := str.Of(fieldName).Lower().String()

	// 获取配置的敏感字段列表
	sensitiveFieldsInterface := facades.Config().Get("operation_log.sensitive_fields", []string{})
	if sensitiveFields, ok := sensitiveFieldsInterface.([]any); ok {
		for _, fieldInterface := range sensitiveFields {
			if field, ok := fieldInterface.(string); ok {
				if keyLower == str.Of(field).Lower().String() {
					return true
				}
			}
		}
	} else if sensitiveFields, ok := sensitiveFieldsInterface.([]string); ok {
		for _, field := range sensitiveFields {
			if keyLower == str.Of(field).Lower().String() {
				return true
			}
		}
	}

	// 检查是否包含敏感关键词
	sensitiveKeywordsInterface := facades.Config().Get("operation_log.sensitive_keywords", []string{})
	if sensitiveKeywords, ok := sensitiveKeywordsInterface.([]any); ok {
		for _, keywordInterface := range sensitiveKeywords {
			if keyword, ok := keywordInterface.(string); ok {
				if str.Of(keyLower).Contains(str.Of(keyword).Lower().String()) {
					return true
				}
			}
		}
	} else if sensitiveKeywords, ok := sensitiveKeywordsInterface.([]string); ok {
		for _, keyword := range sensitiveKeywords {
			if str.Of(keyLower).Contains(str.Of(keyword).Lower().String()) {
				return true
			}
		}
	}

	return false
}
