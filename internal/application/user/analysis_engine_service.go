package user

import (
	"context"
	"fmt"
	"strings"

	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/interfaces"
	"detectviz-platform/pkg/platform/contracts"
)

// AnalysisEngineService 實現了 AnalysisEngine 介面，提供增強的分析功能
// 職責: 執行數據分析並利用 LLM 進行結果解釋和歸因
type AnalysisEngineService struct {
	llmProvider            contracts.LLMProvider
	embeddingStoreProvider contracts.EmbeddingStoreProvider
	logger                 contracts.Logger
}

// AnalysisEngineConfig 定義分析引擎的配置
type AnalysisEngineConfig struct {
	LLMProviderName            string `yaml:"llm_provider_name" json:"llm_provider_name"`
	EmbeddingStoreProviderName string `yaml:"embedding_store_provider_name" json:"embedding_store_provider_name"`
	PromptTemplate             string `yaml:"prompt_template" json:"prompt_template"`
	MaxPromptLength            int    `yaml:"max_prompt_length" json:"max_prompt_length"`
	EnableLLMAnalysis          bool   `yaml:"enable_llm_analysis" json:"enable_llm_analysis"`
}

// NewAnalysisEngineService 創建新的分析引擎服務實例
func NewAnalysisEngineService(
	llmProvider contracts.LLMProvider,
	embeddingStoreProvider contracts.EmbeddingStoreProvider,
	logger contracts.Logger,
) interfaces.AnalysisEngine {
	return &AnalysisEngineService{
		llmProvider:            llmProvider,
		embeddingStoreProvider: embeddingStoreProvider,
		logger:                 logger,
	}
}

// AnalyzeData 分析原始數據
func (a *AnalysisEngineService) AnalyzeData(ctx context.Context, data []byte) (entities.AnalysisResult, error) {
	a.logger.Info("開始分析數據", "data_size", len(data))

	// 基本數據分析
	basicResult, err := a.performBasicAnalysis(ctx, data)
	if err != nil {
		return entities.AnalysisResult{}, fmt.Errorf("基本分析失敗: %w", err)
	}

	// 如果有 LLM 提供者，進行增強分析
	if a.llmProvider != nil {
		enhancedResult, err := a.performLLMAnalysis(ctx, data, basicResult)
		if err != nil {
			a.logger.Warn("LLM 分析失敗，使用基本分析結果", "error", err)
			return basicResult, nil
		}
		return enhancedResult, nil
	}

	return basicResult, nil
}

// ProcessDetection 處理偵測事件
func (a *AnalysisEngineService) ProcessDetection(ctx context.Context, detection *entities.Detection) (entities.DetectionResult, error) {
	if detection == nil {
		return entities.DetectionResult{}, fmt.Errorf("detection cannot be nil")
	}

	a.logger.Info("處理偵測事件", "detection_id", detection.ID, "type", detection.Type)

	// 基本偵測處理
	basicResult, err := a.performBasicDetectionProcessing(ctx, detection)
	if err != nil {
		return entities.DetectionResult{}, fmt.Errorf("基本偵測處理失敗: %w", err)
	}

	// 如果有 LLM 提供者，進行增強處理
	if a.llmProvider != nil {
		enhancedResult, err := a.performLLMDetectionProcessing(ctx, detection, basicResult)
		if err != nil {
			a.logger.Warn("LLM 偵測處理失敗，使用基本處理結果", "error", err)
			return basicResult, nil
		}
		return enhancedResult, nil
	}

	return basicResult, nil
}

// performBasicAnalysis 執行基本數據分析
func (a *AnalysisEngineService) performBasicAnalysis(ctx context.Context, data []byte) (entities.AnalysisResult, error) {
	// 簡單的統計分析
	dataStr := string(data)
	wordCount := len(strings.Fields(dataStr))
	charCount := len(dataStr)

	result := entities.AnalysisResult{
		// 根據實際的 AnalysisResult 結構填充
		// 這裡使用示例字段
	}

	a.logger.Debug("基本分析完成", "word_count", wordCount, "char_count", charCount)
	return result, nil
}

// performLLMAnalysis 執行 LLM 增強分析
func (a *AnalysisEngineService) performLLMAnalysis(ctx context.Context, data []byte, basicResult entities.AnalysisResult) (entities.AnalysisResult, error) {
	// 構建 LLM 提示
	prompt := a.buildAnalysisPrompt(data, basicResult)

	// 調用 LLM 生成分析
	options := map[string]any{
		"temperature": 0.3,
		"max_tokens":  1000,
	}

	llmResponse, err := a.llmProvider.GenerateText(ctx, prompt, options)
	if err != nil {
		return entities.AnalysisResult{}, fmt.Errorf("LLM 分析失敗: %w", err)
	}

	// 解析 LLM 響應並增強結果
	enhancedResult := a.enhanceAnalysisWithLLM(basicResult, llmResponse)

	a.logger.Info("LLM 分析完成", "response_length", len(llmResponse))
	return enhancedResult, nil
}

// performBasicDetectionProcessing 執行基本偵測處理
func (a *AnalysisEngineService) performBasicDetectionProcessing(ctx context.Context, detection *entities.Detection) (entities.DetectionResult, error) {
	// 基本偵測結果處理
	result := entities.DetectionResult{
		// 根據實際的 DetectionResult 結構填充
		// 這裡使用示例字段
	}

	a.logger.Debug("基本偵測處理完成", "detection_id", detection.ID)
	return result, nil
}

// performLLMDetectionProcessing 執行 LLM 增強偵測處理
func (a *AnalysisEngineService) performLLMDetectionProcessing(ctx context.Context, detection *entities.Detection, basicResult entities.DetectionResult) (entities.DetectionResult, error) {
	// 構建偵測分析提示
	prompt := a.buildDetectionPrompt(detection, basicResult)

	// 調用 LLM 進行偵測解釋
	options := map[string]any{
		"temperature": 0.2,
		"max_tokens":  800,
	}

	llmResponse, err := a.llmProvider.GenerateText(ctx, prompt, options)
	if err != nil {
		return entities.DetectionResult{}, fmt.Errorf("LLM 偵測分析失敗: %w", err)
	}

	// 增強偵測結果
	enhancedResult := a.enhanceDetectionWithLLM(basicResult, llmResponse)

	a.logger.Info("LLM 偵測處理完成", "detection_id", detection.ID, "response_length", len(llmResponse))
	return enhancedResult, nil
}

// buildAnalysisPrompt 構建數據分析提示
func (a *AnalysisEngineService) buildAnalysisPrompt(data []byte, basicResult entities.AnalysisResult) string {
	dataStr := string(data)
	if len(dataStr) > 2000 {
		dataStr = dataStr[:2000] + "..."
	}

	prompt := fmt.Sprintf(`
作為一個數據分析專家，請分析以下數據並提供洞察：

數據內容：
%s

基本統計信息：
- 數據大小：%d 字節

請提供：
1. 數據的主要特徵和模式
2. 潛在的異常或值得注意的點
3. 建議的進一步分析方向
4. 數據質量評估

請以結構化的方式回答，並保持簡潔明了。
`, dataStr, len(data))

	return prompt
}

// buildDetectionPrompt 構建偵測分析提示
func (a *AnalysisEngineService) buildDetectionPrompt(detection *entities.Detection, basicResult entities.DetectionResult) string {
	prompt := fmt.Sprintf(`
作為一個異常偵測專家，請分析以下偵測結果：

偵測信息：
- ID: %s
- 類型: %s
- 創建時間: %s

請提供：
1. 偵測結果的可能原因分析
2. 風險評估和嚴重程度
3. 建議的處理措施
4. 預防類似問題的建議

請以專業且易懂的方式回答。
`, detection.ID, detection.Type, detection.CreatedAt.Format("2006-01-02 15:04:05"))

	return prompt
}

// enhanceAnalysisWithLLM 使用 LLM 響應增強分析結果
func (a *AnalysisEngineService) enhanceAnalysisWithLLM(basicResult entities.AnalysisResult, llmResponse string) entities.AnalysisResult {
	// 將 LLM 響應集成到分析結果中
	// 這裡需要根據實際的 AnalysisResult 結構來實現
	enhancedResult := basicResult

	// 示例：將 LLM 響應添加到結果中
	// enhancedResult.LLMInsights = llmResponse
	// enhancedResult.EnhancedAnalysis = true

	return enhancedResult
}

// enhanceDetectionWithLLM 使用 LLM 響應增強偵測結果
func (a *AnalysisEngineService) enhanceDetectionWithLLM(basicResult entities.DetectionResult, llmResponse string) entities.DetectionResult {
	// 將 LLM 響應集成到偵測結果中
	// 這裡需要根據實際的 DetectionResult 結構來實現
	enhancedResult := basicResult

	// 示例：將 LLM 響應添加到結果中
	// enhancedResult.LLMExplanation = llmResponse
	// enhancedResult.EnhancedProcessing = true

	return enhancedResult
}
