package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"

	appfacades "goravel/app/facades"
	"goravel/app/utils"
)

type AIService interface {
	Complete(ctx context.Context, prompt string, systemPrompt string) (string, error)
	CompleteWithAttachments(ctx context.Context, prompt string, systemPrompt string, attachments []contractsai.Attachment) (string, error)
	GenerateImage(ctx context.Context, prompt string, size string) ([]byte, string, error)
	GenerateAudio(ctx context.Context, prompt string, voice string) ([]byte, string, error)
	Transcribe(ctx context.Context, file contractsfilesystem.File, language string) (string, error)
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
	return s.CompleteWithAttachments(ctx, prompt, systemPrompt, nil)
}

func (s *AIServiceImpl) CompleteWithAttachments(ctx context.Context, prompt string, systemPrompt string, attachments []contractsai.Attachment) (string, error) {
	if err := ensureAIConfigured(); err != nil {
		return "", err
	}

	provider := aiProvider()
	agent := &promptAgent{instructions: systemPrompt}
	opts := []contractsai.Option{ai.WithProvider(provider)}
	if model := textModel(); model != "" {
		opts = append(opts, ai.WithModel(model))
	}

	conversation, err := appfacades.AI().WithContext(ctx).Agent(agent, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create AI conversation: %w", err)
	}

	promptOpts := make([]contractsai.ConversationOption, 0, 1)
	if len(attachments) > 0 {
		promptOpts = append(promptOpts, ai.WithAttachments(attachments...))
	}

	response, err := conversation.Prompt(prompt, promptOpts...)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}

	return response.Text(), nil
}

func (s *AIServiceImpl) GenerateImage(ctx context.Context, prompt string, size string) ([]byte, string, error) {
	if err := ensureAIConfigured(); err != nil {
		return nil, "", err
	}

	req := appfacades.AI().WithContext(ctx).Image(prompt).Provider(aiProvider())
	if model := imageModel(); model != "" {
		req = req.Model(model)
	}
	switch strings.ToLower(strings.TrimSpace(size)) {
	case "portrait":
		req = req.Portrait()
	case "landscape":
		req = req.Landscape()
	default:
		req = req.Square()
	}

	resp, err := req.Generate()
	if err != nil {
		return nil, "", fmt.Errorf("AI image generation failed: %w", err)
	}

	content, err := resp.Content()
	if err != nil {
		return nil, "", fmt.Errorf("AI image content failed: %w", err)
	}

	mimeType := resp.MimeType()
	if mimeType == "" {
		mimeType = "image/png"
	}

	return content, mimeType, nil
}

func (s *AIServiceImpl) GenerateAudio(ctx context.Context, prompt string, voice string) ([]byte, string, error) {
	if err := ensureAIConfigured(); err != nil {
		return nil, "", err
	}

	req := appfacades.AI().WithContext(ctx).Audio(prompt).Provider(aiProvider())
	if model := audioModel(); model != "" {
		req = req.Model(model)
	}
	voice = strings.TrimSpace(voice)
	switch strings.ToLower(voice) {
	case "", "female", "default-female":
		req = req.Female()
	case "male", "default-male":
		req = req.Male()
	default:
		req = req.Voice(voice)
	}

	resp, err := req.Generate()
	if err != nil {
		return nil, "", fmt.Errorf("AI audio generation failed: %w", err)
	}

	content, err := resp.Content()
	if err != nil {
		return nil, "", fmt.Errorf("AI audio content failed: %w", err)
	}

	mimeType := resp.MimeType()
	if mimeType == "" {
		mimeType = "audio/mpeg"
	}

	return content, mimeType, nil
}

func (s *AIServiceImpl) Transcribe(ctx context.Context, file contractsfilesystem.File, language string) (string, error) {
	if err := ensureAIConfigured(); err != nil {
		return "", err
	}

	storable := ai.DocumentFromUpload(file)
	req := appfacades.AI().WithContext(ctx).Transcription(storable).Provider(aiProvider())
	if model := transcriptionModel(); model != "" {
		req = req.Model(model)
	}
	if language = strings.TrimSpace(language); language != "" {
		req = req.Language(language)
	}

	resp, err := req.Generate()
	if err != nil {
		return "", fmt.Errorf("AI transcription failed: %w", err)
	}

	return resp.Text(), nil
}

func ensureAIConfigured() error {
	if !utils.AIEnabled() {
		return fmt.Errorf("AI API key is not configured. Please set AI_API_KEY (or OPENAI_API_KEY) in .env and restart the server")
	}
	return nil
}

func aiProvider() string {
	return facades.Config().GetString("ai.default", "openai")
}

func textModel() string {
	return facades.Config().GetString("ai.model", "")
}

func imageModel() string {
	return facades.Config().GetString("ai.image_model", "")
}

func audioModel() string {
	return facades.Config().GetString("ai.audio_model", "")
}

func transcriptionModel() string {
	return facades.Config().GetString("ai.transcription_model", "")
}
