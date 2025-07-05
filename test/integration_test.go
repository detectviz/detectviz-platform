package test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"detectviz-platform/internal/plugins/detectors"
	"detectviz-platform/internal/plugins/importers"
	"detectviz-platform/pkg/platform/contracts"
)

// 簡化的模擬實現
type TestLogger struct{}

func (t *TestLogger) Debug(msg string, fields ...interface{}) {}
func (t *TestLogger) Info(msg string, fields ...interface{})  {}
func (t *TestLogger) Warn(msg string, fields ...interface{})  {}
func (t *TestLogger) Error(msg string, fields ...interface{}) {}
func (t *TestLogger) Fatal(msg string, fields ...interface{}) {}
func (t *TestLogger) WithFields(fields ...interface{}) contracts.Logger {
	return t
}
func (t *TestLogger) WithContext(ctx interface{}) contracts.Logger {
	return t
}
func (t *TestLogger) GetName() string {
	return "test_logger"
}

type TestDBClient struct{}

func (t *TestDBClient) GetDB(ctx context.Context) (*sql.DB, error) {
	return nil, nil
}
func (t *TestDBClient) GetName() string {
	return "test_db"
}

type TestMetricsProvider struct{}

func (t *TestMetricsProvider) IncCounter(name string, tags map[string]string)                      {}
func (t *TestMetricsProvider) ObserveHistogram(name string, value float64, tags map[string]string) {}
func (t *TestMetricsProvider) SetGauge(name string, value float64, tags map[string]string)         {}
func (t *TestMetricsProvider) GetName() string {
	return "test_metrics"
}

// TestCSVImporterIntegration 測試 CSV 導入器的完整流程
func TestCSVImporterIntegration(t *testing.T) {
	logger := &TestLogger{}
	dbClient := &TestDBClient{}

	// 創建插件
	plugin := importers.NewCSVImporterPlugin(dbClient, logger)

	// 初始化插件
	config := map[string]interface{}{
		"table_name":    "test_metrics",
		"has_header":    true,
		"batch_size":    100,
		"validate_data": true,
	}

	err := plugin.Init(context.Background(), config)
	if err != nil {
		t.Fatalf("插件初始化失敗: %v", err)
	}

	// 啟動插件
	err = plugin.Start(context.Background())
	if err != nil {
		t.Fatalf("插件啟動失敗: %v", err)
	}

	// 創建測試 CSV 文件
	testData := `timestamp,cpu,memory,disk
2024-01-15 10:00:00,45.2,62.8,78.5
2024-01-15 10:01:00,95.1,85.3,78.6
2024-01-15 10:02:00,48.7,68.9,78.7`

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test_data.csv")
	err = os.WriteFile(csvFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("創建測試文件失敗: %v", err)
	}

	// 執行導入
	err = plugin.ImportData(context.Background(), csvFile)
	if err != nil {
		t.Errorf("數據導入失敗: %v", err)
	}

	// 停止插件
	err = plugin.Stop(context.Background())
	if err != nil {
		t.Errorf("插件停止失敗: %v", err)
	}

	t.Log("CSV 導入器集成測試通過")
}

// TestThresholdDetectorIntegration 測試閾值偵測器的完整流程
func TestThresholdDetectorIntegration(t *testing.T) {
	logger := &TestLogger{}
	metricsProvider := &TestMetricsProvider{}

	// 創建插件
	plugin := detectors.NewThresholdDetectorPlugin(logger, metricsProvider)

	// 初始化插件
	config := map[string]interface{}{
		"field_name":      "cpu_usage",
		"upper_threshold": 80.0,
		"lower_threshold": 20.0,
		"severity":        "high",
		"enable_upper":    true,
		"enable_lower":    true,
		"description":     "CPU 使用率監控",
	}

	err := plugin.Init(context.Background(), config)
	if err != nil {
		t.Fatalf("插件初始化失敗: %v", err)
	}

	// 啟動插件
	err = plugin.Start(context.Background())
	if err != nil {
		t.Fatalf("插件啟動失敗: %v", err)
	}

	// 測試正常值
	normalData := map[string]interface{}{
		"cpu_usage": 50.0,
		"timestamp": "2024-01-15 10:00:00",
	}

	result, err := plugin.Execute(context.Background(), normalData, map[string]interface{}{})
	if err != nil {
		t.Errorf("執行正常值檢測失敗: %v", err)
	}
	if result == nil {
		t.Error("期望返回非空結果")
	}

	// 測試異常值 - 超過上限
	anomalyData := map[string]interface{}{
		"cpu_usage": 95.0,
		"timestamp": "2024-01-15 10:01:00",
	}

	result, err = plugin.Execute(context.Background(), anomalyData, map[string]interface{}{})
	if err != nil {
		t.Errorf("執行異常值檢測失敗: %v", err)
	}
	if result == nil {
		t.Error("期望返回非空結果")
	}

	// 測試運行時配置覆蓋
	runtimeConfig := map[string]interface{}{
		"upper_threshold": 98.0, // 提高閾值
	}

	result, err = plugin.Execute(context.Background(), anomalyData, runtimeConfig)
	if err != nil {
		t.Errorf("執行運行時配置檢測失敗: %v", err)
	}
	if result == nil {
		t.Error("期望返回非空結果")
	}

	// 停止插件
	err = plugin.Stop(context.Background())
	if err != nil {
		t.Errorf("插件停止失敗: %v", err)
	}

	t.Log("閾值偵測器集成測試通過")
}

// TestPluginWorkflow 測試完整的插件工作流程
func TestPluginWorkflow(t *testing.T) {
	ctx := context.Background()
	logger := &TestLogger{}
	dbClient := &TestDBClient{}
	metricsProvider := &TestMetricsProvider{}

	// 1. 創建並配置 CSV 導入器
	importer := importers.NewCSVImporterPlugin(dbClient, logger)
	importerConfig := map[string]interface{}{
		"table_name": "system_metrics",
		"has_header": true,
		"column_mapping": map[string]string{
			"cpu":    "cpu_usage",
			"memory": "memory_usage",
		},
	}

	err := importer.Init(ctx, importerConfig)
	if err != nil {
		t.Fatalf("導入器初始化失敗: %v", err)
	}

	// 2. 創建並配置閾值偵測器
	detector := detectors.NewThresholdDetectorPlugin(logger, metricsProvider)
	detectorConfig := map[string]interface{}{
		"field_name":      "cpu_usage",
		"upper_threshold": 90.0,
		"severity":        "critical",
		"enable_upper":    true,
	}

	err = detector.Init(ctx, detectorConfig)
	if err != nil {
		t.Fatalf("偵測器初始化失敗: %v", err)
	}

	// 3. 啟動插件
	err = importer.Start(ctx)
	if err != nil {
		t.Fatalf("導入器啟動失敗: %v", err)
	}

	err = detector.Start(ctx)
	if err != nil {
		t.Fatalf("偵測器啟動失敗: %v", err)
	}

	// 4. 模擬數據導入和偵測流程
	testData := []map[string]interface{}{
		{"cpu_usage": 45.0, "memory_usage": 60.0}, // 正常
		{"cpu_usage": 95.0, "memory_usage": 85.0}, // CPU 異常
		{"cpu_usage": 50.0, "memory_usage": 70.0}, // 正常
	}

	for i, data := range testData {
		result, err := detector.Execute(ctx, data, map[string]interface{}{})
		if err != nil {
			t.Errorf("第 %d 次偵測失敗: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("第 %d 次偵測返回空結果", i+1)
		}
	}

	// 5. 停止插件
	err = detector.Stop(ctx)
	if err != nil {
		t.Errorf("偵測器停止失敗: %v", err)
	}

	err = importer.Stop(ctx)
	if err != nil {
		t.Errorf("導入器停止失敗: %v", err)
	}

	t.Log("完整插件工作流程測試通過")
}
