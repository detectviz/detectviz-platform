# CSV Importer Plugin

## 概述

CSV Importer 插件為 Detectviz 平台提供強大的 CSV 文件數據導入功能。此插件能夠解析各種格式的 CSV 文件，並將數據批量導入到平台的數據庫中，支援靈活的列映射、數據驗證和批量處理。

## 功能特性

- **靈活的 CSV 解析**: 支援自定義分隔符、標題行處理、行跳過等功能
- **批量導入**: 高效的批量插入機制，可配置批量大小以優化性能
- **列映射**: 支援 CSV 列到數據庫列的靈活映射
- **數據驗證**: 內建數據驗證機制，確保數據完整性
- **錯誤處理**: 完善的錯誤處理和日誌記錄
- **限制控制**: 支援最大導入行數限制，防止資源耗盡

## 支援的 CSV 格式

### 標準 CSV
- 逗號分隔 (`,`)
- 雙引號包圍的字段
- 標題行支援

### 自定義分隔符
- 支援任意單字符分隔符（如 `|`, `;`, `\t` 等）
- 靈活的格式配置

### 特殊處理
- 跳過文件開頭的指定行數
- 處理包含/不包含標題行的文件
- 日期時間格式自定義解析

## 配置說明

### 基本配置

```yaml
csv_importer:
  name: "sales_data_importer"
  type: "csv_importer"
  config:
    delimiter: ","
    has_header: true
    table_name: "sales_data"
    column_mapping:
      date: "sale_date"
      amount: "sale_amount"
      customer: "customer_name"
    batch_size: 1000
    validate_data: true
  enabled: true
```

### 高級配置

```yaml
csv_importer:
  name: "log_data_importer"
  type: "csv_importer"
  config:
    delimiter: "|"
    has_header: false
    skip_rows: 2
    table_name: "log_entries"
    column_mapping:
      timestamp: "log_timestamp"
      level: "log_level"
      message: "log_message"
    batch_size: 2000
    max_rows: 100000
    datetime_format: "2006-01-02T15:04:05Z"
    validate_data: true
  enabled: true
```

### 配置參數

| 參數 | 類型 | 必需 | 默認值 | 說明 |
|------|------|------|--------|------|
| `name` | string | 是 | - | 插件的唯一標識符 |
| `type` | string | 是 | - | 插件類型，必須為 `csv_importer` |
| `config.delimiter` | string | 否 | "," | CSV 分隔符，必須是單個字符 |
| `config.has_header` | boolean | 否 | true | 是否包含標題行 |
| `config.skip_rows` | integer | 否 | 0 | 跳過的行數 |
| `config.table_name` | string | 是 | - | 目標資料庫表名 |
| `config.column_mapping` | object | 否 | {} | CSV 列到資料庫列的映射 |
| `config.batch_size` | integer | 否 | 1000 | 批量插入大小 (1-10000) |
| `config.max_rows` | integer | 否 | 0 | 最大導入行數 (0 表示無限制) |
| `config.validate_data` | boolean | 否 | true | 是否驗證數據 |
| `config.datetime_format` | string | 否 | "2006-01-02 15:04:05" | 日期時間格式 |
| `enabled` | boolean | 否 | true | 是否啟用此插件 |

## 使用範例

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "detectviz-platform/internal/adapters/plugins/importers"
    "detectviz-platform/pkg/platform/contracts"
)

func main() {
    // 創建數據庫客戶端和日誌器
    var dbClient contracts.DBClientProvider
    var logger contracts.Logger
    
    // 創建 CSV 導入器
    importer := importers.NewCSVImporterPlugin(dbClient, logger)
    
    // 配置插件
    config := map[string]interface{}{
        "delimiter":    ",",
        "has_header":   true,
        "table_name":   "sales_data",
        "column_mapping": map[string]string{
            "date":     "sale_date",
            "amount":   "sale_amount",
            "customer": "customer_name",
        },
        "batch_size":   500,
        "validate_data": true,
    }
    
    ctx := context.Background()
    
    // 初始化插件
    if err := importer.Init(ctx, config); err != nil {
        log.Fatalf("Failed to initialize CSV importer: %v", err)
    }
    
    // 啟動插件
    if err := importer.Start(ctx); err != nil {
        log.Fatalf("Failed to start CSV importer: %v", err)
    }
    
    // 導入數據
    if err := importer.ImportData(ctx, "data/sales.csv"); err != nil {
        log.Fatalf("Failed to import data: %v", err)
    }
    
    log.Println("Data imported successfully!")
}
```

### 處理不同格式的 CSV

```go
// 處理管道分隔的日誌文件
config := map[string]interface{}{
    "delimiter":      "|",
    "has_header":     false,
    "skip_rows":      1,
    "table_name":     "access_logs",
    "column_mapping": map[string]string{
        "timestamp": "access_time",
        "ip":        "client_ip",
        "method":    "http_method",
        "url":       "request_url",
        "status":    "response_status",
    },
    "batch_size":       2000,
    "max_rows":         1000000,
    "datetime_format":  "2006-01-02T15:04:05Z",
}
```

### 在服務中使用

```go
package dataservice

import (
    "context"
    "fmt"
    
    "detectviz-platform/internal/adapters/plugins/importers"
    "detectviz-platform/pkg/platform/contracts"
)

type DataImportService struct {
    csvImporter importers.CSVImporterPlugin
    logger      contracts.Logger
}

func NewDataImportService(dbClient contracts.DBClientProvider, logger contracts.Logger) *DataImportService {
    return &DataImportService{
        csvImporter: importers.NewCSVImporterPlugin(dbClient, logger),
        logger:      logger,
    }
}

func (s *DataImportService) ImportSalesData(ctx context.Context, filePath string) error {
    config := map[string]interface{}{
        "delimiter":    ",",
        "has_header":   true,
        "table_name":   "sales_transactions",
        "column_mapping": map[string]string{
            "transaction_date": "date",
            "amount":          "total_amount",
            "customer_id":     "customer",
            "product_id":      "product",
        },
        "batch_size":   1000,
        "validate_data": true,
    }
    
    if err := s.csvImporter.Init(ctx, config); err != nil {
        return fmt.Errorf("failed to initialize CSV importer: %w", err)
    }
    
    if err := s.csvImporter.Start(ctx); err != nil {
        return fmt.Errorf("failed to start CSV importer: %w", err)
    }
    
    if err := s.csvImporter.ImportData(ctx, filePath); err != nil {
        return fmt.Errorf("failed to import data: %w", err)
    }
    
    s.logger.Info("Sales data imported successfully", "file", filePath)
    return nil
}
```

## 數據驗證

### 內建驗證規則

1. **必需字段檢查**: 確保必需的列不為空
2. **數據類型驗證**: 驗證數值、日期等字段的格式
3. **長度限制**: 檢查字符串字段的長度限制
4. **範圍檢查**: 驗證數值字段的範圍

### 自定義驗證

```go
// 在配置中啟用數據驗證
config := map[string]interface{}{
    "validate_data": true,
    // 其他配置...
}
```

## 效能最佳化

### 批量大小調整

- **小文件** (< 10MB): 使用較小的批量大小 (500-1000)
- **大文件** (> 100MB): 使用較大的批量大小 (2000-5000)
- **超大文件** (> 1GB): 考慮分割文件或使用流式處理

### 記憶體管理

```go
// 對於大文件，設置最大行數限制
config := map[string]interface{}{
    "max_rows":   1000000,  // 限制最大處理行數
    "batch_size": 2000,     // 增加批量大小
}
```

## 錯誤處理

### 常見錯誤

#### 1. 配置錯誤
```
Error: table_name 不能為空
```
**解決方案**: 確保在配置中指定了有效的 `table_name`

#### 2. 文件格式錯誤
```
Error: 無法打開 CSV 文件
```
**解決方案**: 檢查文件路徑和權限

#### 3. 數據庫連接錯誤
```
Error: 資料庫連接失敗
```
**解決方案**: 確保資料庫服務正在運行且連接配置正確

### 調試技巧

1. **啟用詳細日誌**: 在開發環境中啟用 DEBUG 級別的日誌
2. **小批量測試**: 使用較小的批量大小進行測試
3. **數據抽樣**: 先導入少量數據進行驗證

## 監控和指標

### 關鍵指標

- **導入速度**: 每秒處理的行數
- **成功率**: 成功導入的行數比例
- **錯誤率**: 驗證失敗的行數比例
- **資源使用**: CPU 和記憶體使用情況

### 日誌記錄

```go
// 插件會自動記錄關鍵事件
logger.Info("開始導入 CSV 數據", "source", filePath, "table", tableName)
logger.Info("CSV 導入完成", "rows_processed", totalRows, "duration", duration)
```

## 故障排除

### 性能問題

1. **導入速度慢**: 增加 `batch_size` 或檢查資料庫性能
2. **記憶體使用過高**: 減少 `batch_size` 或設置 `max_rows`
3. **CPU 使用率高**: 關閉 `validate_data` 或優化驗證邏輯

### 數據問題

1. **字符編碼問題**: 確保 CSV 文件使用 UTF-8 編碼
2. **日期格式錯誤**: 調整 `datetime_format` 參數
3. **特殊字符處理**: 檢查分隔符和引號的使用

## 最佳實踐

1. **預處理數據**: 在導入前清理和驗證 CSV 文件
2. **監控資源**: 監控導入過程中的資源使用情況
3. **備份數據**: 在大量導入前備份目標表
4. **測試配置**: 在生產環境前充分測試配置
5. **錯誤恢復**: 實施適當的錯誤恢復機制

## 相關資源

- [Detectviz 平台數據導入指南](../data-import/guide.md)
- [數據庫最佳實踐](../database/best-practices.md)
- [插件開發指南](../development/plugin-development.md)

## 版本歷史

- **v1.0.0**: 初始版本，支援基本 CSV 導入功能
- **v1.1.0**: 添加批量處理和數據驗證功能
- **v1.2.0**: 添加列映射和錯誤處理改進
- **v1.3.0**: 添加性能優化和監控功能 