package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
}

type AnthropicResponse struct {
	Completion string `json:"completion"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func Chat(ctx context.Context, provider, baseURL, apiKey string, req ChatRequest) (*ChatResponse, error) {
	if provider == "anthropic" {
		return chatAnthropic(ctx, baseURL, apiKey, req)
	}

	return chatOpenAI(ctx, provider, baseURL, apiKey, req)
}

func chatOpenAI(ctx context.Context, provider, baseURL, apiKey string, req ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := trimSuffix(baseURL, "/")
	httpReqURL := url
	if provider == "azure" {
		if req.Model == "" {
			return nil, fmt.Errorf("Azure 模型/部署名称不能为空")
		}
		httpReqURL = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2023-10-01-preview", url, req.Model)
	} else {
		httpReqURL = fmt.Sprintf("%s/v1/chat/completions", url)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, httpReqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if provider == "azure" {
		httpReq.Header.Set("api-key", apiKey)
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求 AI 接口失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("AI 接口返回错误 (status=%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("AI 接口返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("解析 AI 响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("AI 未返回任何内容")
	}

	return &chatResp, nil
}

func chatAnthropic(ctx context.Context, baseURL, apiKey string, req ChatRequest) (*ChatResponse, error) {
	prompt := buildAnthropicPrompt(req.Messages)
	anthropicReq := map[string]any{
		"model":                req.Model,
		"prompt":               prompt,
		"max_tokens_to_sample": req.MaxTokens,
		"temperature":          req.Temperature,
	}
	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := trimSuffix(baseURL, "/") + "/v1/complete"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求 AI 接口失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("AI 接口返回错误 (status=%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("AI 接口返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("解析 Anthropic 响应失败: %w", err)
	}

	return &ChatResponse{Choices: []ChatChoice{{Message: ChatMessage{Role: "assistant", Content: anthropicResp.Completion}}}}, nil
}

func buildAnthropicPrompt(messages []ChatMessage) string {
	var prompt strings.Builder
	for _, msg := range messages {
		if msg.Role == "system" {
			prompt.WriteString("" + msg.Content + "\n\n")
		} else if msg.Role == "user" {
			prompt.WriteString("Human: " + msg.Content + "\n\n")
		}
	}
	prompt.WriteString("Assistant:")
	return prompt.String()
}

func trimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
