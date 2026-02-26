package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GLM5Config GLM5 配置
type GLM5Config struct {
	APIKey      string        `json:"apiKey"`
	BaseURL     string        `json:"baseUrl"`
	Model       string        `json:"model"`
	MaxTokens   int           `json:"maxTokens"`
	Temperature float64       `json:"temperature"`
	Timeout     time.Duration `json:"timeout"`
}

// GLM5Client GLM5 API 客户端
type GLM5Client struct {
	config     GLM5Config
	httpClient *http.Client
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewGLM5Client 创建 GLM5 客户端
func NewGLM5Client(config GLM5Config) *GLM5Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://open.bigmodel.cn/api/paas/v4"
	}
	if config.Model == "" {
		config.Model = "glm-4"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	return &GLM5Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Chat 发送聊天请求
func (c *GLM5Client) Chat(messages []Message) (*ChatResponse, error) {
	req := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("API 错误: %s - %s", chatResp.Error.Code, chatResp.Error.Message)
	}

	return &chatResp, nil
}

// QuickChat 快速聊天
func (c *GLM5Client) QuickChat(prompt string) (string, error) {
	resp, err := c.Chat([]Message{{Role: "user", Content: prompt}})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("无响应内容")
	}
	return resp.Choices[0].Message.Content, nil
}
