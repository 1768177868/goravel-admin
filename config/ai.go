package config

import (
	"github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/facades"
	openaifacades "github.com/goravel/openai/facades"
)

func init() {
	config := facades.Config()
	config.Add("ai", map[string]any{
		"default": config.Env("AI_PROVIDER", "openai"),
		"model":   config.Env("AI_MODEL", "gpt-4o-mini"),

		"providers": map[string]any{
			"openai": map[string]any{
				"key": config.Env("AI_API_KEY", ""),
				"url": config.Env("AI_BASE_URL", ""),
				"models": map[string]any{
					"text": map[string]any{
						"default": config.Env("AI_MODEL", "gpt-4o-mini"),
					},
				},
				"via": func() (ai.Provider, error) {
					return openaifacades.OpenAI("openai")
				},
			},
		},
	})
}
