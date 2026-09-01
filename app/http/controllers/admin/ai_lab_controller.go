package admin

import (
	"encoding/base64"
	"strings"

	"github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/response"
	"goravel/app/services"
	"goravel/app/utils"
)

type AiLabController struct{}

func NewAiLabController() *AiLabController {
	return &AiLabController{}
}

func (c *AiLabController) aiService(ctx http.Context) services.AIService {
	return services.NewAIService(ctx)
}

func (c *AiLabController) ensureAI(ctx http.Context) http.Response {
	if !utils.AIEnabled() {
		return response.Error(ctx, http.StatusBadRequest, "ai_not_configured")
	}
	return nil
}

// Status 返回 AI 实验室配置与限流信息
func (c *AiLabController) Status(ctx http.Context) http.Response {
	return response.Success(ctx, http.Json{
		"ai_enabled":            utils.AIEnabled(),
		"text_model":            strings.TrimSpace(facades.Config().GetString("ai.model", "")),
		"image_model":           strings.TrimSpace(facades.Config().GetString("ai.image_model", "")),
		"audio_model":           strings.TrimSpace(facades.Config().GetString("ai.audio_model", "")),
		"transcription_model":   strings.TrimSpace(facades.Config().GetString("ai.transcription_model", "")),
		"rate_limit_per_minute": utils.AILabRateLimitPerMinute(),
		"rate_limit_per_day":    utils.AILabRateLimitPerDay(),
	})
}

type aiLabTextRequest struct {
	Prompt       string `json:"prompt" form:"prompt"`
	SystemPrompt string `json:"system_prompt" form:"system_prompt"`
}

// Text 文本对话
func (c *AiLabController) Text(ctx http.Context) http.Response {
	if resp := c.ensureAI(ctx); resp != nil {
		return resp
	}

	var req aiLabTextRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_fields")
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		return response.Error(ctx, http.StatusBadRequest, "prompt_required")
	}

	text, err := c.aiService(ctx).Complete(ctx, req.Prompt, strings.TrimSpace(req.SystemPrompt))
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{"text": text})
}

// Vision 图片理解（multipart: file + prompt）
func (c *AiLabController) Vision(ctx http.Context) http.Response {
	if resp := c.ensureAI(ctx); resp != nil {
		return resp
	}

	prompt := strings.TrimSpace(ctx.Request().Input("prompt", ""))
	if prompt == "" {
		return response.Error(ctx, http.StatusBadRequest, "prompt_required")
	}

	file, err := ctx.Request().File("file")
	if err != nil || file == nil {
		return response.Error(ctx, http.StatusBadRequest, "file_required")
	}

	attachment := ai.ImageFromUpload(file)
	text, err := c.aiService(ctx).CompleteWithAttachments(ctx, prompt, "", []contractsai.Attachment{attachment})
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{"text": text})
}

type aiLabImageRequest struct {
	Prompt string `json:"prompt" form:"prompt"`
	Size   string `json:"size" form:"size"`
}

// Image 图片生成
func (c *AiLabController) Image(ctx http.Context) http.Response {
	if resp := c.ensureAI(ctx); resp != nil {
		return resp
	}

	var req aiLabImageRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_fields")
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		return response.Error(ctx, http.StatusBadRequest, "prompt_required")
	}

	content, mimeType, err := c.aiService(ctx).GenerateImage(ctx, req.Prompt, req.Size)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"mime_type":      mimeType,
		"content_base64": base64.StdEncoding.EncodeToString(content),
	})
}

type aiLabAudioRequest struct {
	Prompt string `json:"prompt" form:"prompt"`
	Voice  string `json:"voice" form:"voice"`
}

// Audio 语音合成
func (c *AiLabController) Audio(ctx http.Context) http.Response {
	if resp := c.ensureAI(ctx); resp != nil {
		return resp
	}

	var req aiLabAudioRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_fields")
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		return response.Error(ctx, http.StatusBadRequest, "prompt_required")
	}

	content, mimeType, err := c.aiService(ctx).GenerateAudio(ctx, req.Prompt, req.Voice)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"mime_type":      mimeType,
		"content_base64": base64.StdEncoding.EncodeToString(content),
	})
}

// Transcription 语音转写（multipart: file + optional language）
func (c *AiLabController) Transcription(ctx http.Context) http.Response {
	if resp := c.ensureAI(ctx); resp != nil {
		return resp
	}

	file, err := ctx.Request().File("file")
	if err != nil || file == nil {
		return response.Error(ctx, http.StatusBadRequest, "file_required")
	}

	language := strings.TrimSpace(ctx.Request().Input("language", ""))
	text, err := c.aiService(ctx).Transcribe(ctx, file, language)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{"text": text})
}
