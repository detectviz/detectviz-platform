package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"detectviz-platform/pkg/platform/contracts"
)

// GeminiLLMProvider 實現了 LLMProvider 介面，提供 Google Gemini API 集成
// 職責: 與 Google Gemini API 交互，執行文本生成任務
type GeminiLLMProvider struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	logger     contracts.Logger
}

// GeminiConfig 定義 Gemini LLM 提供者的配置
type GeminiConfig struct {
	APIKey      string  `yaml:"api_key" json:"api_key"`
	BaseURL     string  `yaml:"base_url" json:"base_url"`
	Model       string  `yaml:"model" json:"model"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	Timeout     string  `yaml:"timeout" json:"timeout"`
}

// GeminiRequest 定義 Gemini API 請求結構
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent 定義內容結構
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart 定義內容部分
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig 定義生成配置
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// GeminiResponse 定義 Gemini API 響應結構
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

// GeminiCandidate 定義候選回應
type GeminiCandidate struct {
	Content       GeminiContent  `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

// SafetyRating 定義安全評級
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GeminiError 定義錯誤結構
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewGeminiLLMProvider 創建新的 Gemini LLM 提供者實例
func NewGeminiLLMProvider(config GeminiConfig, logger contracts.Logger) (contracts.LLMProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	if config.Model == "" {
		config.Model = "gemini-pro"
	}

	timeout := 30 * time.Second
	if config.Timeout != "" {
		if t, err := time.ParseDuration(config.Timeout); err == nil {
			timeout = t
		}
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	logger.Info("初始化 Gemini LLM 提供者", "model", config.Model, "base_url", config.BaseURL)

	return &GeminiLLMProvider{
		apiKey:     config.APIKey,
		baseURL:    config.BaseURL,
		model:      config.Model,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// GenerateText 生成文本內容
func (g *GeminiLLMProvider) GenerateText(ctx context.Context, prompt string, options map[string]any) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// 構建請求
	request := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	// 解析選項
	if options != nil {
		genConfig := GeminiGenerationConfig{}

		if temp, ok := options["temperature"].(float64); ok {
			genConfig.Temperature = temp
		}

		if maxTokens, ok := options["max_tokens"].(int); ok {
			genConfig.MaxOutputTokens = maxTokens
		}

		if topP, ok := options["top_p"].(float64); ok {
			genConfig.TopP = topP
		}

		if topK, ok := options["top_k"].(int); ok {
			genConfig.TopK = topK
		}

		request.GenerationConfig = genConfig
	}

	// 序列化請求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 構建 URL
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.baseURL, g.model, g.apiKey)

	// 創建 HTTP 請求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 發送請求
	g.logger.Debug("發送 Gemini API 請求", "url", url, "prompt_length", len(prompt))

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析響應
	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// 檢查錯誤
	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s (code: %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// 提取生成的文本
	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := geminiResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in candidate")
	}

	generatedText := candidate.Content.Parts[0].Text

	g.logger.Info("Gemini API 請求成功",
		"prompt_length", len(prompt),
		"response_length", len(generatedText),
		"finish_reason", candidate.FinishReason)

	return generatedText, nil
}

// GetName 返回提供者名稱
func (g *GeminiLLMProvider) GetName() string {
	return "gemini_llm_provider"
}
