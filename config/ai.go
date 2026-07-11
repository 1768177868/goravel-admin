package config

import (
	"fmt"

	"github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/facades"
	openaifacades "github.com/goravel/openai/facades"
)

func aiAPIKey(cfg config.Config) string {
	if key := fmt.Sprint(cfg.Env("AI_API_KEY", "")); key != "" {
		return key
	}
	return fmt.Sprint(cfg.Env("OPENAI_API_KEY", ""))
}

func aiBaseURL(cfg config.Config) string {
	if url := fmt.Sprint(cfg.Env("AI_BASE_URL", "")); url != "" {
		return url
	}
	return fmt.Sprint(cfg.Env("OPENAI_BASE_URL", ""))
}

func init() {
	config := facades.Config()
	config.Add("ai", map[string]any{
		"default": config.Env("AI_PROVIDER", "openai"),
		"model":   config.Env("AI_MODEL", "gpt-4o-mini"),
		"enabled": config.Env("AI_ENABLED", true),

		"providers": map[string]any{
			"openai": map[string]any{
				"key": aiAPIKey(config),
				"models": map[string]any{
					"text": map[string]any{
						"default": "",
					},
					"audio": map[string]any{
						"default": "",
					},
					"transcription": map[string]any{
						"default": "",
					},
					"image": map[string]any{
						"default": "",
					},
				},
				"failover": map[string][]string{},
				"url":      aiBaseURL(config),
				"via": func() (ai.Provider, error) {
					return openaifacades.OpenAI("openai")
				},
			},
		},
	})
}
