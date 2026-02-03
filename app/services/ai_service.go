package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"
)

type AIService interface {
	Complete(ctx context.Context, prompt string, systemPrompt string) (string, error)
}

type AIServiceImpl struct{}

func NewAIService() AIService {
	return &AIServiceImpl{}
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// 豆包请求结构 - 支持简单字符串格式
type DoubaoRequest struct {
	Model string      `json:"model"`
	Input interface{} `json:"input"` // 可以是字符串或数组
}

type DoubaoInput struct {
	Role    string          `json:"role"`
	Content []DoubaoContent `json:"content"`
}

type DoubaoContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// 豆包响应结构
type DoubaoResponse struct {
	Output []DoubaoOutputItem `json:"output"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

type DoubaoOutputItem struct {
	Type    string `json:"type"` // "reasoning" 或 "message"
	Role    string `json:"role"` // "assistant"
	Content []struct {
		Type string `json:"type"` // "output_text"
		Text string `json:"text"`
	} `json:"content"`
	Status string `json:"status"`
}

func (s *AIServiceImpl) Complete(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	provider := facades.Config().GetString("ai.provider")
	baseURL := facades.Config().GetString("ai.base_url")
	apiKey := facades.Config().GetString("ai.api_key")
	model := facades.Config().GetString("ai.model")
	maxTokens := cast.ToInt(facades.Config().Get("ai.max_tokens"))
	temperature := cast.ToFloat32(facades.Config().Get("ai.temperature"))
	timeout := cast.ToInt(facades.Config().Get("ai.timeout"))

	if apiKey == "" {
		return "", fmt.Errorf("AI API key is not configured. Please set AI_API_KEY in .env file and restart the server")
	}

	// 如果 timeout 为 0 或太小，设置默认值
	if timeout <= 0 {
		timeout = 300 // 默认 300 秒（5分钟）
	}

	// 构建请求 URL
	apiURL := baseURL
	if apiURL == "" {
		switch provider {
		case "openai":
			apiURL = "https://api.openai.com/v1/chat/completions"
		case "azure":
			// Azure OpenAI 需要 endpoint 和 deployment name
			apiURL = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview", baseURL, model)
		case "doubao":
			apiURL = "https://ark.cn-beijing.volces.com/api/v3/responses"
		default:
			// 默认使用 OpenAI 格式
			apiURL = "https://api.openai.com/v1/chat/completions"
		}
	} else {
		// 如果提供了自定义 base_url，根据 provider 构建 URL
		switch provider {
		case "azure":
			apiURL = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview", baseURL, model)
		case "doubao":
			// 豆包：base_url 应该包含 /api/v3，直接拼接 /responses
			if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
				apiURL = fmt.Sprintf("%sresponses", baseURL)
			} else {
				apiURL = fmt.Sprintf("%s/responses", baseURL)
			}
		default:
			apiURL = fmt.Sprintf("%s/v1/chat/completions", baseURL)
		}
	}

	var jsonData []byte
	var err error

	// 根据 provider 构建不同的请求体
	if provider == "doubao" {
		// 豆包格式：根据是否有 systemPrompt 决定使用简单字符串还是数组格式
		// 如果有 systemPrompt，使用数组格式；否则使用简单字符串格式
		if systemPrompt != "" {
			// 使用数组格式
			doubaoInput := []DoubaoInput{}
			doubaoInput = append(doubaoInput, DoubaoInput{
				Role: "system",
				Content: []DoubaoContent{
					{Type: "input_text", Text: systemPrompt},
				},
			})
			doubaoInput = append(doubaoInput, DoubaoInput{
				Role: "user",
				Content: []DoubaoContent{
					{Type: "input_text", Text: prompt},
				},
			})
			requestBody := DoubaoRequest{
				Model: model,
				Input: doubaoInput,
			}
			jsonData, err = json.Marshal(requestBody)
		} else {
			// 使用简单字符串格式（兼容 curl 示例）
			// 将 systemPrompt 和 prompt 合并
			fullPrompt := prompt
			if systemPrompt != "" {
				fullPrompt = systemPrompt + "\n\n" + prompt
			}
			requestBody := DoubaoRequest{
				Model: model,
				Input: fullPrompt,
			}
			jsonData, err = json.Marshal(requestBody)
		}
	} else {
		// OpenAI/Azure 格式：使用 messages 字段
		messages := []Message{}
		if systemPrompt != "" {
			messages = append(messages, Message{
				Role:    "system",
				Content: systemPrompt,
			})
		}
		messages = append(messages, Message{
			Role:    "user",
			Content: prompt,
		})

		requestBody := OpenAIRequest{
			Model:       model,
			Messages:    messages,
			MaxTokens:   maxTokens,
			Temperature: temperature,
		}
		jsonData, err = json.Marshal(requestBody)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if provider == "azure" {
		req.Header.Set("api-key", apiKey)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	// 创建带超时的 context
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 使用带超时的 context 重新创建请求
	req = req.WithContext(requestCtx)

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		// 检查是否是超时错误
		if requestCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("AI API request timeout after %d seconds", timeout)
		}
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误信息
		if provider == "doubao" {
			var errorResp DoubaoResponse
			if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
				return "", fmt.Errorf("AI API error: %s", errorResp.Error.Message)
			}
		} else {
			var errorResp OpenAIResponse
			if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
				return "", fmt.Errorf("AI API error: %s", errorResp.Error.Message)
			}
		}
		return "", fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 根据 provider 解析不同的响应格式
	if provider == "doubao" {
		var response DoubaoResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return "", fmt.Errorf("failed to parse response: %w. Response body: %s", err, string(body))
		}

		// 检查是否有错误
		if response.Error.Message != "" {
			return "", fmt.Errorf("AI API error [%s]: %s", response.Error.Code, response.Error.Message)
		}

		if len(response.Output) == 0 {
			return "", fmt.Errorf("no output in AI response")
		}

		// 查找 type 为 "message" 的输出项
		for _, item := range response.Output {
			if item.Type == "message" && len(item.Content) > 0 {
				// 查找 type 为 "output_text" 的内容
				for _, content := range item.Content {
					if content.Type == "output_text" && content.Text != "" {
						return content.Text, nil
					}
				}
			}
		}

		return "", fmt.Errorf("no message content found in AI response")
	} else {
		var response OpenAIResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}

		if len(response.Choices) == 0 {
			return "", fmt.Errorf("no choices in AI response")
		}

		return response.Choices[0].Message.Content, nil
	}
}
