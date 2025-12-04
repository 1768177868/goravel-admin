package utils

import (
	"strings"

	"github.com/goravel/framework/facades"
)

// IsSensitiveField 检查字段名是否是敏感字段
func IsSensitiveField(fieldName string) bool {
	keyLower := strings.ToLower(fieldName)

	// 获取配置的敏感字段列表
	sensitiveFieldsInterface := facades.Config().Get("operation_log.sensitive_fields", []string{})
	if sensitiveFields, ok := sensitiveFieldsInterface.([]any); ok {
		for _, fieldInterface := range sensitiveFields {
			if field, ok := fieldInterface.(string); ok {
				if keyLower == strings.ToLower(field) {
					return true
				}
			}
		}
	} else if sensitiveFields, ok := sensitiveFieldsInterface.([]string); ok {
		for _, field := range sensitiveFields {
			if keyLower == strings.ToLower(field) {
				return true
			}
		}
	}

	// 检查是否包含敏感关键词
	sensitiveKeywordsInterface := facades.Config().Get("operation_log.sensitive_keywords", []string{})
	if sensitiveKeywords, ok := sensitiveKeywordsInterface.([]any); ok {
		for _, keywordInterface := range sensitiveKeywords {
			if keyword, ok := keywordInterface.(string); ok {
				if strings.Contains(keyLower, strings.ToLower(keyword)) {
					return true
				}
			}
		}
	} else if sensitiveKeywords, ok := sensitiveKeywordsInterface.([]string); ok {
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(keyLower, strings.ToLower(keyword)) {
				return true
			}
		}
	}

	return false
}
