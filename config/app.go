package config

import (
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"

	"goravel/lang"
)

// Boot Start all init methods of the current folder to bootstrap all config.
func Boot() {}
func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		// Application Name
		//
		// This value is the name of your application. This value is used when the
		// framework needs to place the application's name in a notification or
		// any other location as required by the application or its packages.
		"name": config.Env("APP_NAME", "Goravel"),
		// Application Environment
		//
		// This value determines the "environment" your application is currently
		// running in. This may determine how you prefer to configure various
		// services the application utilizes. Set this in your ".env" file.
		"env": config.Env("APP_ENV", "production"),
		// Application Debug Mode
		"debug": config.Env("APP_DEBUG", false),
		// Enable Development Tool
		"enable_dev_tool": config.Env("APP_ENABLE_DEV_TOOL", false),
		// Application Timezone
		//
		// Here you may specify the default timezone for your application.
		// Example: UTC, Asia/Shanghai
		// More: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
		"timezone": carbon.UTC,
		// Display source timezone for datetime strings without timezone offset.
		// Used by response time conversion to interpret stored datetime strings.
		"display_source_timezone": config.Env("APP_DISPLAY_SOURCE_TIMEZONE", carbon.UTC),
		// Comma-separated response time field whitelist.
		// Example: created_at,updated_at,deleted_at
		"response_time_fields": config.Env("APP_RESPONSE_TIME_FIELDS", "created_at,updated_at,deleted_at"),
		// Application Locale Configuration
		//
		// The application locale determines the default locale that will be used
		// by the translation service provider. You are free to set this value
		// to any of the locales which will be supported by the application.
		"locale": "en",
		// Application Fallback Locale
		//
		// The fallback locale determines the locale to use when the current one
		// is not available. You may change the value to correspond to any of
		// the language folders that are provided through your application.
		"fallback_locale": "cn",
		// Application Lang Path
		//
		// The path to the language files for the application. You may change
		// the path to a different directory if you would like to customize it.
		// 框架会优先使用 lang_path 指定的文件系统，当文件不存在时才使用 lang_fs embed 文件系统
		"lang_path": "lang",
		"lang_fs":   lang.FS,
		// Encryption Key
		//
		// 32 character string, otherwise these encrypted strings
		// will not be safe. Please do this before deploying an application!
		"key": config.Env("APP_KEY", ""),
	})
}
