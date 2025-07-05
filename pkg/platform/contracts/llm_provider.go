package contracts

import "context"

// LLMProvider 定義了大型語言模型推論功能的通用介面。
// 職責: 將 prompt 傳入 LLM 並取得模型輸出。
// AI_PLUGIN_TYPE: "llm_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/llm/gemini_llm"
// AI_IMPL_CONSTRUCTOR: "NewGeminiLLMProvider"
type LLMProvider interface {
	GenerateText(ctx context.Context, prompt string, options map[string]any) (string, error)
	GetName() string
}
