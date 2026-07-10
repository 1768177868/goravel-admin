package utils

import (
	"context"
	"fmt"

	appfacades "goravel/app/facades"

	"goravel/app/models"
)

func configCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// GetConfigValue 从数据库获取配置值
func GetConfigValue(ctx context.Context, group, key string, defaultValue string) string {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if appfacades.Orm() == nil {
		return defaultValue
	}

	var config models.Config
	err := appfacades.OrmQuery(configCtx(ctx)).Where("group", group).Where("key", key).First(&config)
	if err != nil {
		return defaultValue
	}
	if config.Value == "" {
		return defaultValue
	}
	return config.Value
}

// GetConfigValueInt 从数据库获取配置值（整数类型）
func GetConfigValueInt(ctx context.Context, group, key string, defaultValue int) int {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if appfacades.Orm() == nil {
		return defaultValue
	}

	var config models.Config
	err := appfacades.OrmQuery(configCtx(ctx)).Where("group", group).Where("key", key).First(&config)
	if err != nil {
		return defaultValue
	}
	if config.Value == "" {
		return defaultValue
	}
	value := 0
	_, err = fmt.Sscanf(config.Value, "%d", &value)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetConfigValueBool 从数据库获取配置值（布尔类型）
func GetConfigValueBool(ctx context.Context, group, key string, defaultValue bool) bool {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if appfacades.Orm() == nil {
		return defaultValue
	}

	var config models.Config
	err := appfacades.OrmQuery(configCtx(ctx)).Where("group", group).Where("key", key).First(&config)
	if err != nil {
		return defaultValue
	}
	if config.Value == "" {
		return defaultValue
	}
	value := config.Value
	return value == "1" || value == "true" || value == "True" || value == "TRUE"
}
