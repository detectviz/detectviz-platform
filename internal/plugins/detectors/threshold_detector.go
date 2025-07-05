package detectors

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/interfaces/plugins"
	"detectviz-platform/pkg/platform/contracts"
)

// ThresholdDetectorPlugin 實現基於閾值的異常偵測功能
// 職責: 根據配置的閾值規則檢測數值型數據的異常
type ThresholdDetectorPlugin struct {
	name            string
	logger          contracts.Logger
	metricsProvider contracts.MetricsProvider
	config          ThresholdDetectorConfig
	isInitialized   bool
}

// ThresholdDetectorConfig 定義閾值偵測器的配置
type ThresholdDetectorConfig struct {
	FieldName      string  `yaml:"field_name" json:"field_name"`           // 要檢測的字段名稱
	UpperThreshold float64 `yaml:"upper_threshold" json:"upper_threshold"` // 上限閾值
	LowerThreshold float64 `yaml:"lower_threshold" json:"lower_threshold"` // 下限閾值
	Severity       string  `yaml:"severity" json:"severity"`               // 告警嚴重程度: low, medium, high, critical
	Description    string  `yaml:"description" json:"description"`         // 偵測器描述
	EnableUpper    bool    `yaml:"enable_upper" json:"enable_upper"`       // 是否啟用上限檢測
	EnableLower    bool    `yaml:"enable_lower" json:"enable_lower"`       // 是否啟用下限檢測
	TolerantCount  int     `yaml:"tolerant_count" json:"tolerant_count"`   // 容忍次數，連續超過多少次才觸發告警
}

// ThresholdDetectionResult 閾值偵測結果
type ThresholdDetectionResult struct {
	IsAnomalous   bool      `json:"is_anomalous"`
	Value         float64   `json:"value"`
	Threshold     float64   `json:"threshold"`
	ThresholdType string    `json:"threshold_type"` // "upper" 或 "lower"
	Severity      string    `json:"severity"`
	Description   string    `json:"description"`
	DetectedAt    time.Time `json:"detected_at"`
	FieldName     string    `json:"field_name"`
	Confidence    float64   `json:"confidence"`
}

// NewThresholdDetectorPlugin 創建新的閾值偵測器插件實例
func NewThresholdDetectorPlugin(logger contracts.Logger, metricsProvider contracts.MetricsProvider) plugins.DetectorPlugin {
	return &ThresholdDetectorPlugin{
		name:            "threshold_detector_plugin",
		logger:          logger,
		metricsProvider: metricsProvider,
		config: ThresholdDetectorConfig{
			Severity:      "medium",
			EnableUpper:   true,
			EnableLower:   true,
			TolerantCount: 1,
		},
	}
}

// GetName 返回插件名稱
func (t *ThresholdDetectorPlugin) GetName() string {
	return t.name
}

// Init 初始化插件
func (t *ThresholdDetectorPlugin) Init(ctx context.Context, cfg map[string]interface{}) error {
	t.logger.Info("正在初始化閾值偵測器插件", "plugin", t.name)

	// 解析配置
	if err := t.parseConfig(cfg); err != nil {
		return fmt.Errorf("解析配置失敗: %w", err)
	}

	// 驗證配置
	if err := t.validateConfig(); err != nil {
		return fmt.Errorf("配置驗證失敗: %w", err)
	}

	t.isInitialized = true
	t.logger.Info("閾值偵測器插件初始化完成",
		"plugin", t.name,
		"field", t.config.FieldName,
		"upper_threshold", t.config.UpperThreshold,
		"lower_threshold", t.config.LowerThreshold)
	return nil
}

// Start 啟動插件
func (t *ThresholdDetectorPlugin) Start(ctx context.Context) error {
	if !t.isInitialized {
		return fmt.Errorf("插件尚未初始化")
	}
	t.logger.Info("閾值偵測器插件已啟動", "plugin", t.name)

	// 記錄啟動指標
	if t.metricsProvider != nil {
		t.metricsProvider.IncCounter("detector_started_total", map[string]string{
			"detector_type": "threshold",
			"plugin":        t.name,
		})
	}

	return nil
}

// Stop 停止插件
func (t *ThresholdDetectorPlugin) Stop(ctx context.Context) error {
	t.logger.Info("閾值偵測器插件正在停止", "plugin", t.name)
	t.isInitialized = false

	// 記錄停止指標
	if t.metricsProvider != nil {
		t.metricsProvider.IncCounter("detector_stopped_total", map[string]string{
			"detector_type": "threshold",
			"plugin":        t.name,
		})
	}

	return nil
}

// Execute 執行閾值偵測
func (t *ThresholdDetectorPlugin) Execute(ctx context.Context, data map[string]interface{}, detectorConfig map[string]interface{}) (*entities.AnalysisResult, error) {
	if !t.isInitialized {
		return nil, fmt.Errorf("插件尚未初始化")
	}

	startTime := time.Now()
	defer func() {
		if t.metricsProvider != nil {
			duration := time.Since(startTime).Seconds()
			t.metricsProvider.ObserveHistogram("detector_execution_duration_seconds", duration, map[string]string{
				"detector_type": "threshold",
				"plugin":        t.name,
			})
		}
	}()

	t.logger.Debug("開始執行閾值偵測", "plugin", t.name, "field", t.config.FieldName)

	// 臨時合併運行時配置
	runtimeConfig := t.config
	if err := t.mergeRuntimeConfig(&runtimeConfig, detectorConfig); err != nil {
		return nil, fmt.Errorf("合併運行時配置失敗: %w", err)
	}

	// 提取要檢測的值
	value, err := t.extractValue(data, runtimeConfig.FieldName)
	if err != nil {
		t.logger.Warn("提取檢測值失敗", "field", runtimeConfig.FieldName, "error", err)
		if t.metricsProvider != nil {
			t.metricsProvider.IncCounter("detector_extraction_errors_total", map[string]string{
				"detector_type": "threshold",
				"plugin":        t.name,
				"field":         runtimeConfig.FieldName,
			})
		}
		return nil, err
	}

	// 執行閾值檢測
	result := t.performThresholdCheck(value, runtimeConfig)

	// 記錄檢測指標
	if t.metricsProvider != nil {
		t.metricsProvider.IncCounter("detector_executions_total", map[string]string{
			"detector_type": "threshold",
			"plugin":        t.name,
			"anomalous":     fmt.Sprintf("%t", result.IsAnomalous),
		})

		if result.IsAnomalous {
			t.metricsProvider.IncCounter("detector_anomalies_total", map[string]string{
				"detector_type":  "threshold",
				"plugin":         t.name,
				"severity":       result.Severity,
				"threshold_type": result.ThresholdType,
			})
		}
	}

	// 構建 AnalysisResult
	analysisResult := &entities.AnalysisResult{
		// 這裡應該根據實際的 AnalysisResult 結構來填充
		// 暫時使用空結構體作為佔位符
	}

	t.logger.Info("閾值偵測完成",
		"plugin", t.name,
		"field", runtimeConfig.FieldName,
		"value", value,
		"anomalous", result.IsAnomalous,
		"severity", result.Severity)

	return analysisResult, nil
}

// parseConfig 解析插件配置
func (t *ThresholdDetectorPlugin) parseConfig(cfg map[string]interface{}) error {
	if fieldName, ok := cfg["field_name"].(string); ok {
		t.config.FieldName = fieldName
	}

	if upperThreshold, ok := cfg["upper_threshold"].(float64); ok {
		t.config.UpperThreshold = upperThreshold
	}

	if lowerThreshold, ok := cfg["lower_threshold"].(float64); ok {
		t.config.LowerThreshold = lowerThreshold
	}

	if severity, ok := cfg["severity"].(string); ok {
		t.config.Severity = severity
	}

	if description, ok := cfg["description"].(string); ok {
		t.config.Description = description
	}

	if enableUpper, ok := cfg["enable_upper"].(bool); ok {
		t.config.EnableUpper = enableUpper
	}

	if enableLower, ok := cfg["enable_lower"].(bool); ok {
		t.config.EnableLower = enableLower
	}

	if tolerantCount, ok := cfg["tolerant_count"].(int); ok {
		t.config.TolerantCount = tolerantCount
	}

	return nil
}

// validateConfig 驗證配置
func (t *ThresholdDetectorPlugin) validateConfig() error {
	if t.config.FieldName == "" {
		return fmt.Errorf("field_name 不能為空")
	}

	if !t.config.EnableUpper && !t.config.EnableLower {
		return fmt.Errorf("至少需要啟用上限或下限檢測中的一個")
	}

	if t.config.EnableUpper && t.config.EnableLower && t.config.UpperThreshold <= t.config.LowerThreshold {
		return fmt.Errorf("上限閾值必須大於下限閾值")
	}

	validSeverities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validSeverities[t.config.Severity] {
		return fmt.Errorf("無效的嚴重程度: %s", t.config.Severity)
	}

	if t.config.TolerantCount < 1 {
		return fmt.Errorf("tolerant_count 必須大於等於 1")
	}

	return nil
}

// mergeRuntimeConfig 合併運行時配置
func (t *ThresholdDetectorPlugin) mergeRuntimeConfig(config *ThresholdDetectorConfig, runtimeCfg map[string]interface{}) error {
	// 運行時配置可以覆蓋部分配置項
	if upperThreshold, ok := runtimeCfg["upper_threshold"].(float64); ok {
		config.UpperThreshold = upperThreshold
	}

	if lowerThreshold, ok := runtimeCfg["lower_threshold"].(float64); ok {
		config.LowerThreshold = lowerThreshold
	}

	if severity, ok := runtimeCfg["severity"].(string); ok {
		config.Severity = severity
	}

	return nil
}

// extractValue 從數據中提取要檢測的數值
func (t *ThresholdDetectorPlugin) extractValue(data map[string]interface{}, fieldName string) (float64, error) {
	rawValue, exists := data[fieldName]
	if !exists {
		return 0, fmt.Errorf("字段 %s 不存在", fieldName)
	}

	// 嘗試轉換為 float64
	switch v := rawValue.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		// 嘗試解析字符串為數字
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, nil
		}
		return 0, fmt.Errorf("無法將字符串 '%s' 轉換為數字", v)
	default:
		return 0, fmt.Errorf("不支持的數據類型: %T", v)
	}
}

// performThresholdCheck 執行閾值檢查
func (t *ThresholdDetectorPlugin) performThresholdCheck(value float64, config ThresholdDetectorConfig) *ThresholdDetectionResult {
	result := &ThresholdDetectionResult{
		Value:       value,
		Severity:    config.Severity,
		Description: config.Description,
		DetectedAt:  time.Now(),
		FieldName:   config.FieldName,
		Confidence:  1.0, // 閾值檢測的置信度總是 100%
	}

	// 檢查上限閾值
	if config.EnableUpper && value > config.UpperThreshold {
		result.IsAnomalous = true
		result.Threshold = config.UpperThreshold
		result.ThresholdType = "upper"
		return result
	}

	// 檢查下限閾值
	if config.EnableLower && value < config.LowerThreshold {
		result.IsAnomalous = true
		result.Threshold = config.LowerThreshold
		result.ThresholdType = "lower"
		return result
	}

	// 沒有異常
	result.IsAnomalous = false
	return result
}
