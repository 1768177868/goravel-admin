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
	imageModel := config.Env("AI_MODEL_IMAGE", "dall-e-3")
	audioModel := config.Env("AI_MODEL_AUDIO", "tts-1")
	transcriptionModel := config.Env("AI_MODEL_TRANSCRIPTION", "whisper-1")

	config.Add("ai", map[string]any{
		"default": config.Env("AI_PROVIDER", "openai"),
		"model":   config.Env("AI_MODEL", "gpt-4o-mini"),
		"enabled": config.Env("AI_ENABLED", true),

		"image_model":         imageModel,
		"audio_model":         audioModel,
		"transcription_model": transcriptionModel,

		// AI 实验室：按管理员账号限流（演示站防滥用）
		"lab_rate_limit_per_minute": config.Env("AI_LAB_RATE_LIMIT_PER_MINUTE", 10),
		"lab_rate_limit_per_day":    config.Env("AI_LAB_RATE_LIMIT_PER_DAY", 200),

		"providers": map[string]any{
			"openai": map[string]any{
				"key": aiAPIKey(config),
				"models": map[string]any{
					"text": map[string]any{
						"default": "",
					},
					"audio": map[string]any{
						"default": audioModel,
					},
					"transcription": map[string]any{
						"default": transcriptionModel,
					},
					"image": map[string]any{
						"default": imageModel,
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
