package importers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"detectviz-platform/pkg/domain/interfaces/plugins"
	"detectviz-platform/pkg/platform/contracts"
)

// CSVImporterPlugin 實現 CSV 數據導入功能
// 職責: 解析 CSV 文件並將數據導入到平台數據庫中
type CSVImporterPlugin struct {
	name          string
	dbClient      contracts.DBClientProvider
	logger        contracts.Logger
	config        CSVImporterConfig
	isInitialized bool
}

// CSVImporterConfig 定義 CSV 導入器的配置
type CSVImporterConfig struct {
	Delimiter      string            `yaml:"delimiter" json:"delimiter"`             // CSV 分隔符，默認為 ","
	HasHeader      bool              `yaml:"has_header" json:"has_header"`           // 是否包含標題行
	SkipRows       int               `yaml:"skip_rows" json:"skip_rows"`             // 跳過的行數
	TableName      string            `yaml:"table_name" json:"table_name"`           // 目標表名
	ColumnMapping  map[string]string `yaml:"column_mapping" json:"column_mapping"`   // CSV 列到資料庫列的映射
	BatchSize      int               `yaml:"batch_size" json:"batch_size"`           // 批量插入大小
	MaxRows        int               `yaml:"max_rows" json:"max_rows"`               // 最大導入行數，0 表示無限制
	ValidateData   bool              `yaml:"validate_data" json:"validate_data"`     // 是否驗證數據
	DateTimeFormat string            `yaml:"datetime_format" json:"datetime_format"` // 日期時間格式
}

// NewCSVImporterPlugin 創建新的 CSV 導入器插件實例
func NewCSVImporterPlugin(dbClient contracts.DBClientProvider, logger contracts.Logger) plugins.ImporterPlugin {
	return &CSVImporterPlugin{
		name:     "csv_importer_plugin",
		dbClient: dbClient,
		logger:   logger,
		config: CSVImporterConfig{
			Delimiter:      ",",
			HasHeader:      true,
			SkipRows:       0,
			BatchSize:      1000,
			MaxRows:        0,
			ValidateData:   true,
			DateTimeFormat: "2006-01-02 15:04:05",
		},
	}
}

// GetName 返回插件名稱
func (c *CSVImporterPlugin) GetName() string {
	return c.name
}

// Init 初始化插件
func (c *CSVImporterPlugin) Init(ctx context.Context, cfg map[string]interface{}) error {
	c.logger.Info("正在初始化 CSV 導入器插件", "plugin", c.name)

	// 解析配置
	if err := c.parseConfig(cfg); err != nil {
		return fmt.Errorf("解析配置失敗: %w", err)
	}

	// 驗證配置
	if err := c.validateConfig(); err != nil {
		return fmt.Errorf("配置驗證失敗: %w", err)
	}

	c.isInitialized = true
	c.logger.Info("CSV 導入器插件初始化完成", "plugin", c.name)
	return nil
}

// Start 啟動插件
func (c *CSVImporterPlugin) Start(ctx context.Context) error {
	if !c.isInitialized {
		return fmt.Errorf("插件尚未初始化")
	}
	c.logger.Info("CSV 導入器插件已啟動", "plugin", c.name)
	return nil
}

// Stop 停止插件
func (c *CSVImporterPlugin) Stop(ctx context.Context) error {
	c.logger.Info("CSV 導入器插件正在停止", "plugin", c.name)
	c.isInitialized = false
	return nil
}

// ImportData 執行 CSV 數據導入
func (c *CSVImporterPlugin) ImportData(ctx context.Context, source string) error {
	if !c.isInitialized {
		return fmt.Errorf("插件尚未初始化")
	}

	c.logger.Info("開始導入 CSV 數據", "source", source, "table", c.config.TableName)

	// 打開 CSV 文件
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("無法打開 CSV 文件 %s: %w", source, err)
	}
	defer file.Close()

	// 創建 CSV 讀取器
	reader := csv.NewReader(file)
	reader.Comma = rune(c.config.Delimiter[0])

	// 跳過指定行數
	for i := 0; i < c.config.SkipRows; i++ {
		_, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("跳過行時發生錯誤: %w", err)
		}
	}

	// 讀取標題行
	var headers []string
	if c.config.HasHeader {
		headers, err = reader.Read()
		if err != nil {
			return fmt.Errorf("讀取標題行失敗: %w", err)
		}
		c.logger.Debug("CSV 標題行", "headers", headers)
	}

	// 批量導入數據
	return c.importDataInBatches(ctx, reader, headers)
}

// parseConfig 解析插件配置
func (c *CSVImporterPlugin) parseConfig(cfg map[string]interface{}) error {
	if delimiter, ok := cfg["delimiter"].(string); ok {
		c.config.Delimiter = delimiter
	}

	if hasHeader, ok := cfg["has_header"].(bool); ok {
		c.config.HasHeader = hasHeader
	}

	if skipRows, ok := cfg["skip_rows"].(int); ok {
		c.config.SkipRows = skipRows
	}

	if tableName, ok := cfg["table_name"].(string); ok {
		c.config.TableName = tableName
	}

	if columnMapping, ok := cfg["column_mapping"].(map[string]string); ok {
		c.config.ColumnMapping = columnMapping
	}

	if batchSize, ok := cfg["batch_size"].(int); ok {
		c.config.BatchSize = batchSize
	}

	if maxRows, ok := cfg["max_rows"].(int); ok {
		c.config.MaxRows = maxRows
	}

	if validateData, ok := cfg["validate_data"].(bool); ok {
		c.config.ValidateData = validateData
	}

	if dateTimeFormat, ok := cfg["datetime_format"].(string); ok {
		c.config.DateTimeFormat = dateTimeFormat
	}

	return nil
}

// validateConfig 驗證配置
func (c *CSVImporterPlugin) validateConfig() error {
	if c.config.TableName == "" {
		return fmt.Errorf("table_name 不能為空")
	}

	if c.config.BatchSize <= 0 {
		return fmt.Errorf("batch_size 必須大於 0")
	}

	if len(c.config.Delimiter) != 1 {
		return fmt.Errorf("delimiter 必須是單個字符")
	}

	return nil
}

// importDataInBatches 批量導入數據
func (c *CSVImporterPlugin) importDataInBatches(ctx context.Context, reader *csv.Reader, headers []string) error {
	db, err := c.dbClient.GetDB(ctx)
	if err != nil {
		return fmt.Errorf("獲取數據庫連接失敗: %w", err)
	}

	var batch [][]string
	var totalRows int

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("讀取 CSV 記錄失敗: %w", err)
		}

		// 檢查最大行數限制
		if c.config.MaxRows > 0 && totalRows >= c.config.MaxRows {
			c.logger.Info("達到最大行數限制", "max_rows", c.config.MaxRows)
			break
		}

		// 驗證數據
		if c.config.ValidateData {
			if err := c.validateRecord(record, headers); err != nil {
				c.logger.Warn("數據驗證失敗，跳過此行", "row", totalRows+1, "error", err)
				continue
			}
		}

		batch = append(batch, record)
		totalRows++

		// 當批次達到指定大小時執行插入
		if len(batch) >= c.config.BatchSize {
			if err := c.insertBatch(ctx, db, batch, headers); err != nil {
				return fmt.Errorf("插入批次數據失敗: %w", err)
			}
			batch = batch[:0] // 清空批次
			c.logger.Debug("已插入批次數據", "rows", c.config.BatchSize, "total", totalRows)
		}
	}

	// 插入剩餘的數據
	if len(batch) > 0 {
		if err := c.insertBatch(ctx, db, batch, headers); err != nil {
			return fmt.Errorf("插入最後批次數據失敗: %w", err)
		}
	}

	c.logger.Info("CSV 數據導入完成", "total_rows", totalRows, "table", c.config.TableName)
	return nil
}

// validateRecord 驗證單條記錄
func (c *CSVImporterPlugin) validateRecord(record []string, headers []string) error {
	// 檢查列數是否匹配
	if c.config.HasHeader && len(record) != len(headers) {
		return fmt.Errorf("列數不匹配: 期望 %d 列，實際 %d 列", len(headers), len(record))
	}

	// 檢查空值
	for i, value := range record {
		if strings.TrimSpace(value) == "" {
			columnName := fmt.Sprintf("column_%d", i)
			if c.config.HasHeader && i < len(headers) {
				columnName = headers[i]
			}
			c.logger.Debug("發現空值", "column", columnName, "index", i)
		}
	}

	return nil
}

// insertBatch 插入批次數據
func (c *CSVImporterPlugin) insertBatch(ctx context.Context, db interface{}, batch [][]string, headers []string) error {
	if len(batch) == 0 {
		return nil
	}

	// 構建插入 SQL
	placeholders := make([]string, len(batch[0]))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	var columns []string
	if c.config.HasHeader && len(headers) > 0 {
		columns = headers
	} else {
		// 生成默認列名
		for i := 0; i < len(batch[0]); i++ {
			columns = append(columns, fmt.Sprintf("column_%d", i+1))
		}
	}

	// 應用列映射
	if len(c.config.ColumnMapping) > 0 {
		for i, col := range columns {
			if mappedCol, exists := c.config.ColumnMapping[col]; exists {
				columns[i] = mappedCol
			}
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		c.config.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	c.logger.Debug("執行插入 SQL", "sql", sql, "batch_size", len(batch))

	// 這裡應該根據實際的數據庫客戶端類型來執行 SQL
	// 由於我們使用的是通用介面，這裡只是示例
	c.logger.Info("模擬數據插入", "rows", len(batch), "table", c.config.TableName)

	return nil
}

// convertValue 轉換數據類型
func (c *CSVImporterPlugin) convertValue(value string, targetType string) (interface{}, error) {
	value = strings.TrimSpace(value)

	switch targetType {
	case "int":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	case "datetime":
		return time.Parse(c.config.DateTimeFormat, value)
	default:
		return value, nil
	}
}
