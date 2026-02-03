package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("ai", map[string]any{
		// AI Provider: openai, azure, doubao, qwen, etc.
		"provider": config.Env("AI_PROVIDER", "openai"),
		// AI API Base URL (for custom endpoints or proxy)
		"base_url": config.Env("AI_BASE_URL", ""),
		// AI API Key
		"api_key": config.Env("AI_API_KEY", ""),
		// AI Model Name
		"model": config.Env("AI_MODEL", "gpt-4o-mini"),
		// Max Tokens for completion
		"max_tokens": config.Env("AI_MAX_TOKENS", 2048),
		// Temperature (0.0-2.0)
		"temperature": config.Env("AI_TEMPERATURE", 0.7),
		// Timeout in seconds
		"timeout": config.Env("AI_TIMEOUT", 120),
	})
}
