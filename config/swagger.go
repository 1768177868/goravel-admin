package config

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	appEnv := strings.ToLower(strings.TrimSpace(fmt.Sprint(config.Env("APP_ENV", "production"))))

	config.Add("swagger", map[string]any{
		// Swagger is enabled by default only in non-production environments.
		// Set SWAGGER_ENABLED=true to explicitly expose it in controlled deployments.
		"enabled": config.Env("SWAGGER_ENABLED", appEnv == "local" || appEnv == "development" || appEnv == "test"),
	})
}
