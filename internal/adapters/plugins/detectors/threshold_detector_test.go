package detectors

import (
	"context"
	"testing"

	"detectviz-platform/pkg/platform/contracts"
)

// MockLogger 模擬日誌記錄器
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Fields  []interface{}
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "DEBUG", Message: msg, Fields: fields})
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "INFO", Message: msg, Fields: fields})
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "WARN", Message: msg, Fields: fields})
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "ERROR", Message: msg, Fields: fields})
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "FATAL", Message: msg, Fields: fields})
}

func (m *MockLogger) WithFields(fields ...interface{}) contracts.Logger {
	return m
}

func (m *MockLogger) WithContext(ctx interface{}) contracts.Logger {
	return m
}

func (m *MockLogger) GetName() string {
	return "mock_logger"
}

func (m *MockLogger) GetLogs() []LogEntry {
	return m.logs
}

// MockMetricsProvider 模擬指標提供者
type MockMetricsProvider struct {
	counters   map[string]int
	histograms map[string][]float64
	gauges     map[string]float64
}

func NewMockMetricsProvider() *MockMetricsProvider {
	return &MockMetricsProvider{
		counters:   make(map[string]int),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
	}
}

func (m *MockMetricsProvider) IncCounter(name string, tags map[string]string) {
	m.counters[name]++
}

func (m *MockMetricsProvider) ObserveHistogram(name string, value float64, tags map[string]string) {
	m.histograms[name] = append(m.histograms[name], value)
}

func (m *MockMetricsProvider) SetGauge(name string, value float64, tags map[string]string) {
	m.gauges[name] = value
}

func (m *MockMetricsProvider) GetName() string {
	return "mock_metrics"
}

func (m *MockMetricsProvider) GetCounterValue(name string) int {
	return m.counters[name]
}

func TestNewThresholdDetectorPlugin(t *testing.T) {
	logger := &MockLogger{}
	metricsProvider := NewMockMetricsProvider()

	plugin := NewThresholdDetectorPlugin(logger, metricsProvider)

	if plugin == nil {
		t.Fatal("插件創建失敗")
	}

	if plugin.GetName() != "threshold_detector_plugin" {
		t.Errorf("期望插件名稱為 'threshold_detector_plugin'，實際為 '%s'", plugin.GetName())
	}
}

func TestThresholdDetectorPlugin_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "有效配置",
			config: map[string]interface{}{
				"field_name":      "cpu_usage",
				"upper_threshold": 90.0,
				"lower_threshold": 10.0,
				"severity":        "high",
				"enable_upper":    true,
				"enable_lower":    true,
			},
			wantErr: false,
		},
		{
			name: "缺少 field_name",
			config: map[string]interface{}{
				"upper_threshold": 90.0,
				"severity":        "high",
			},
			wantErr: true,
		},
		{
			name: "無效的嚴重程度",
			config: map[string]interface{}{
				"field_name":      "cpu_usage",
				"upper_threshold": 90.0,
				"severity":        "invalid",
			},
			wantErr: true,
		},
		{
			name: "上限小於下限",
			config: map[string]interface{}{
				"field_name":      "cpu_usage",
				"upper_threshold": 10.0,
				"lower_threshold": 90.0,
				"enable_upper":    true,
				"enable_lower":    true,
			},
			wantErr: true,
		},
		{
			name: "未啟用任何檢測",
			config: map[string]interface{}{
				"field_name":   "cpu_usage",
				"enable_upper": false,
				"enable_lower": false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的插件實例
			logger := &MockLogger{}
			metricsProvider := NewMockMetricsProvider()
			plugin := NewThresholdDetectorPlugin(logger, metricsProvider).(*ThresholdDetectorPlugin)

			err := plugin.Init(context.Background(), tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThresholdDetectorPlugin_Execute(t *testing.T) {
	logger := &MockLogger{}
	metricsProvider := NewMockMetricsProvider()
	plugin := NewThresholdDetectorPlugin(logger, metricsProvider).(*ThresholdDetectorPlugin)

	ctx := context.Background()

	// 初始化插件
	config := map[string]interface{}{
		"field_name":      "cpu_usage",
		"upper_threshold": 80.0,
		"lower_threshold": 20.0,
		"severity":        "high",
		"enable_upper":    true,
		"enable_lower":    true,
	}
	err := plugin.Init(ctx, config)
	if err != nil {
		t.Fatalf("插件初始化失敗: %v", err)
	}

	err = plugin.Start(ctx)
	if err != nil {
		t.Fatalf("插件啟動失敗: %v", err)
	}

	tests := []struct {
		name           string
		data           map[string]interface{}
		detectorConfig map[string]interface{}
		wantErr        bool
		expectAnomaly  bool
	}{
		{
			name: "正常值",
			data: map[string]interface{}{
				"cpu_usage": 50.0,
			},
			detectorConfig: map[string]interface{}{},
			wantErr:        false,
			expectAnomaly:  false,
		},
		{
			name: "超過上限",
			data: map[string]interface{}{
				"cpu_usage": 95.0,
			},
			detectorConfig: map[string]interface{}{},
			wantErr:        false,
			expectAnomaly:  true,
		},
		{
			name: "低於下限",
			data: map[string]interface{}{
				"cpu_usage": 5.0,
			},
			detectorConfig: map[string]interface{}{},
			wantErr:        false,
			expectAnomaly:  true,
		},
		{
			name: "運行時配置覆蓋",
			data: map[string]interface{}{
				"cpu_usage": 85.0,
			},
			detectorConfig: map[string]interface{}{
				"upper_threshold": 90.0, // 運行時提高閾值
			},
			wantErr:       false,
			expectAnomaly: false,
		},
		{
			name: "缺少字段",
			data: map[string]interface{}{
				"memory_usage": 50.0, // 缺少 cpu_usage 字段
			},
			detectorConfig: map[string]interface{}{},
			wantErr:        true,
			expectAnomaly:  false,
		},
		{
			name: "無效的數據類型",
			data: map[string]interface{}{
				"cpu_usage": "invalid", // 字符串無法轉換為數字
			},
			detectorConfig: map[string]interface{}{},
			wantErr:        true,
			expectAnomaly:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := plugin.Execute(ctx, tt.data, tt.detectorConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("期望返回非空的 AnalysisResult")
			}
		})
	}

	// 驗證指標是否被正確記錄
	if metricsProvider.GetCounterValue("detector_started_total") != 1 {
		t.Error("期望 detector_started_total 指標為 1")
	}

	if metricsProvider.GetCounterValue("detector_executions_total") < 1 {
		t.Error("期望 detector_executions_total 指標大於 0")
	}
}

func TestThresholdDetectorPlugin_ExtractValue(t *testing.T) {
	plugin := &ThresholdDetectorPlugin{}

	tests := []struct {
		name      string
		data      map[string]interface{}
		fieldName string
		want      float64
		wantErr   bool
	}{
		{
			name:      "提取 float64",
			data:      map[string]interface{}{"value": 123.45},
			fieldName: "value",
			want:      123.45,
			wantErr:   false,
		},
		{
			name:      "提取 int",
			data:      map[string]interface{}{"value": 123},
			fieldName: "value",
			want:      123.0,
			wantErr:   false,
		},
		{
			name:      "提取字符串數字",
			data:      map[string]interface{}{"value": "123.45"},
			fieldName: "value",
			want:      123.45,
			wantErr:   false,
		},
		{
			name:      "字段不存在",
			data:      map[string]interface{}{"other": 123},
			fieldName: "value",
			want:      0,
			wantErr:   true,
		},
		{
			name:      "無效的字符串",
			data:      map[string]interface{}{"value": "invalid"},
			fieldName: "value",
			want:      0,
			wantErr:   true,
		},
		{
			name:      "不支持的類型",
			data:      map[string]interface{}{"value": []int{1, 2, 3}},
			fieldName: "value",
			want:      0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := plugin.extractValue(tt.data, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThresholdDetectorPlugin_PerformThresholdCheck(t *testing.T) {
	plugin := &ThresholdDetectorPlugin{}

	tests := []struct {
		name           string
		value          float64
		config         ThresholdDetectorConfig
		expectAnomaly  bool
		expectType     string
		expectSeverity string
	}{
		{
			name:  "正常值",
			value: 50.0,
			config: ThresholdDetectorConfig{
				UpperThreshold: 80.0,
				LowerThreshold: 20.0,
				EnableUpper:    true,
				EnableLower:    true,
				Severity:       "medium",
				FieldName:      "test_field",
			},
			expectAnomaly:  false,
			expectType:     "",
			expectSeverity: "medium",
		},
		{
			name:  "超過上限",
			value: 90.0,
			config: ThresholdDetectorConfig{
				UpperThreshold: 80.0,
				LowerThreshold: 20.0,
				EnableUpper:    true,
				EnableLower:    true,
				Severity:       "high",
				FieldName:      "test_field",
			},
			expectAnomaly:  true,
			expectType:     "upper",
			expectSeverity: "high",
		},
		{
			name:  "低於下限",
			value: 10.0,
			config: ThresholdDetectorConfig{
				UpperThreshold: 80.0,
				LowerThreshold: 20.0,
				EnableUpper:    true,
				EnableLower:    true,
				Severity:       "critical",
				FieldName:      "test_field",
			},
			expectAnomaly:  true,
			expectType:     "lower",
			expectSeverity: "critical",
		},
		{
			name:  "僅啟用上限檢測",
			value: 10.0,
			config: ThresholdDetectorConfig{
				UpperThreshold: 80.0,
				LowerThreshold: 20.0,
				EnableUpper:    true,
				EnableLower:    false,
				Severity:       "medium",
				FieldName:      "test_field",
			},
			expectAnomaly:  false,
			expectType:     "",
			expectSeverity: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plugin.performThresholdCheck(tt.value, tt.config)

			if result.IsAnomalous != tt.expectAnomaly {
				t.Errorf("performThresholdCheck() IsAnomalous = %v, want %v", result.IsAnomalous, tt.expectAnomaly)
			}

			if result.IsAnomalous && result.ThresholdType != tt.expectType {
				t.Errorf("performThresholdCheck() ThresholdType = %v, want %v", result.ThresholdType, tt.expectType)
			}

			if result.Severity != tt.expectSeverity {
				t.Errorf("performThresholdCheck() Severity = %v, want %v", result.Severity, tt.expectSeverity)
			}

			if result.Value != tt.value {
				t.Errorf("performThresholdCheck() Value = %v, want %v", result.Value, tt.value)
			}

			if result.Confidence != 1.0 {
				t.Errorf("performThresholdCheck() Confidence = %v, want 1.0", result.Confidence)
			}
		})
	}
}
