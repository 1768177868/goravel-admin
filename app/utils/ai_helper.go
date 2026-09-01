package utils

import (
	"fmt"

	"github.com/goravel/framework/facades"
)

// AIProviderAPIKey returns the configured API key for the default AI provider.
func AIProviderAPIKey() string {
	provider := facades.Config().GetString("ai.default", "openai")
	if key := facades.Config().GetString(fmt.Sprintf("ai.providers.%s.key", provider), ""); key != "" {
		return key
	}
	return ""
}

// AIEnabled reports whether AI features (code generator assistant) are available.
func AIEnabled() bool {
	if !facades.Config().GetBool("ai.enabled", true) {
		return false
	}
	return AIProviderAPIKey() != ""
}

// AILabRateLimitPerMinute returns per-admin AI lab requests allowed per minute.
func AILabRateLimitPerMinute() int {
	limit := facades.Config().GetInt("ai.lab_rate_limit_per_minute", 10)
	if limit < 1 {
		return 1
	}
	return limit
}

// AILabRateLimitPerDay returns per-admin AI lab requests allowed per day.
func AILabRateLimitPerDay() int {
	limit := facades.Config().GetInt("ai.lab_rate_limit_per_day", 200)
	if limit < 1 {
		return 1
	}
	return limit
}

// AILabMaxUploadBytes returns max upload size for AI lab vision/transcription files.
func AILabMaxUploadBytes() int64 {
	mb := facades.Config().GetInt("ai.lab_max_upload_mb", 10)
	if mb < 1 {
		mb = 1
	}
	return int64(mb) * 1024 * 1024
}
