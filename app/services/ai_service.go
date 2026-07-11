package services

import (
	"context"
	"fmt"

	"github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/facades"

	appfacades "goravel/app/facades"
	"goravel/app/utils"
)

type AIService interface {
	Complete(ctx context.Context, prompt string, systemPrompt string) (string, error)
}

type AIServiceImpl struct {
	ctx context.Context
}

func NewAIService(ctx context.Context) AIService {
	return &AIServiceImpl{ctx: ctx}
}

type promptAgent struct {
	instructions string
}

func (a *promptAgent) Instructions() string                 { return a.instructions }
func (a *promptAgent) Messages() []contractsai.Message      { return nil }
func (a *promptAgent) Middleware() []contractsai.Middleware { return nil }
func (a *promptAgent) Tools() []contractsai.Tool            { return nil }

func (s *AIServiceImpl) Complete(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	if !utils.AIEnabled() {
		return "", fmt.Errorf("AI API key is not configured. Please set AI_API_KEY (or OPENAI_API_KEY) in .env and restart the server")
	}

	provider := facades.Config().GetString("ai.default", "openai")
	agent := &promptAgent{instructions: systemPrompt}
	opts := []contractsai.Option{ai.WithProvider(provider)}
	if model := facades.Config().GetString("ai.model", ""); model != "" {
		opts = append(opts, ai.WithModel(model))
	}

	conversation, err := appfacades.AI().Agent(agent, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create AI conversation: %w", err)
	}

	response, err := conversation.Prompt(prompt)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}

	return response.Text(), nil
}
