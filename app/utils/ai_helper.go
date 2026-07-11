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
