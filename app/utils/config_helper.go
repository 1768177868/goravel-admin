package utils

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// GetConfigValue 从数据库获取配置值
// group: 配置分组
// key: 配置键
// defaultValue: 默认值（如果配置不存在）
func GetConfigValue(group, key string, defaultValue string) string {
	var config models.Config
	err := facades.Orm().Query().Where("group", group).Where("key", key).First(&config)
	if err != nil {
		return defaultValue
	}
	if config.Value == "" {
		return defaultValue
	}
	return config.Value
}

// GetConfigValueInt 从数据库获取配置值（整数类型）
func GetConfigValueInt(group, key string, defaultValue int) int {
	var config models.Config
	err := facades.Orm().Query().Where("group", group).Where("key", key).First(&config)
	if err != nil {
		return defaultValue
	}
	if config.Value == "" {
		return defaultValue
	}
	// 简单的字符串转整数，实际可以使用更完善的转换
	value := 0
	_, err = fmt.Sscanf(config.Value, "%d", &value)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetConfigValueBool 从数据库获取配置值（布尔类型）
func GetConfigValueBool(group, key string, defaultValue bool) bool {
	var config models.Config
	err := facades.Orm().Query().Where("group", group).Where("key", key).First(&config)
	if err != nil {
		return defaultValue
	}
	if config.Value == "" {
		return defaultValue
	}
	// 支持 "1", "true", "True" 等格式
	value := config.Value
	return value == "1" || value == "true" || value == "True" || value == "TRUE"
}

