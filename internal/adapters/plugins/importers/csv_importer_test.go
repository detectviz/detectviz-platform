package importers

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"detectviz-platform/pkg/platform/contracts"
)

// MockDBClientProvider 模擬數據庫客戶端提供者
type MockDBClientProvider struct {
	name string
}

func (m *MockDBClientProvider) GetDB(ctx context.Context) (*sql.DB, error) {
	// 返回 nil 因為這是模擬實現
	return nil, nil
}

func (m *MockDBClientProvider) GetName() string {
	return m.name
}

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

func TestNewCSVImporterPlugin(t *testing.T) {
	dbClient := &MockDBClientProvider{name: "test_db"}
	logger := &MockLogger{}

	plugin := NewCSVImporterPlugin(dbClient, logger)

	if plugin == nil {
		t.Fatal("插件創建失敗")
	}

	if plugin.GetName() != "csv_importer_plugin" {
		t.Errorf("期望插件名稱為 'csv_importer_plugin'，實際為 '%s'", plugin.GetName())
	}
}

func TestCSVImporterPlugin_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "有效配置",
			config: map[string]interface{}{
				"delimiter":     ",",
				"has_header":    true,
				"table_name":    "test_table",
				"batch_size":    500,
				"validate_data": true,
			},
			wantErr: false,
		},
		{
			name: "缺少 table_name",
			config: map[string]interface{}{
				"delimiter":  ",",
				"has_header": true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的插件實例
			dbClient := &MockDBClientProvider{name: "test_db"}
			logger := &MockLogger{}
			plugin := NewCSVImporterPlugin(dbClient, logger).(*CSVImporterPlugin)

			err := plugin.Init(context.Background(), tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSVImporterPlugin_ImportData(t *testing.T) {
	dbClient := &MockDBClientProvider{name: "test_db"}
	logger := &MockLogger{}
	plugin := NewCSVImporterPlugin(dbClient, logger).(*CSVImporterPlugin)

	ctx := context.Background()

	testData := `timestamp,cpu,memory
2024-01-15 10:00:00,45.2,62.8
2024-01-15 10:01:00,52.1,65.3`

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")
	err := os.WriteFile(csvFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("創建測試文件失敗: %v", err)
	}

	config := map[string]interface{}{
		"table_name": "test_table",
		"has_header": true,
		"batch_size": 2,
	}
	err = plugin.Init(ctx, config)
	if err != nil {
		t.Fatalf("插件初始化失敗: %v", err)
	}

	err = plugin.Start(ctx)
	if err != nil {
		t.Fatalf("插件啟動失敗: %v", err)
	}

	err = plugin.ImportData(ctx, csvFile)
	if err != nil {
		t.Errorf("數據導入失敗: %v", err)
	}

	logs := logger.GetLogs()
	hasImportLog := false
	for _, log := range logs {
		if log.Level == "INFO" && log.Message == "CSV 數據導入完成" {
			hasImportLog = true
			break
		}
	}
	if !hasImportLog {
		t.Error("期望找到數據導入完成的日誌")
	}
}
