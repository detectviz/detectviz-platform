# Threshold Detector Plugin

## 概述

Threshold Detector 插件為 Detectviz 平台提供基於閾值的異常偵測功能。此插件能夠監控數值型數據，當數值超過或低於預設的閾值時觸發異常告警，支援靈活的閾值配置、嚴重程度分級和容錯機制。

## 功能特性

- **雙向閾值檢測**: 支援上限和下限閾值檢測，可獨立啟用或禁用
- **嚴重程度分級**: 支援 low、medium、high、critical 四個嚴重程度等級
- **容錯機制**: 支援連續多次超過閾值才觸發告警，減少誤報
- **實時監控**: 提供實時的異常偵測和告警功能
- **指標統計**: 內建指標收集，支援監控和分析
- **靈活配置**: 支援運行時動態配置覆蓋

## 偵測原理

### 閾值類型

1. **上限閾值 (Upper Threshold)**: 當數值超過設定的上限時觸發告警
2. **下限閾值 (Lower Threshold)**: 當數值低於設定的下限時觸發告警
3. **雙向閾值**: 同時監控上限和下限，提供全面的異常偵測

### 觸發機制

- **即時觸發**: 單次超過閾值即觸發告警 (`tolerant_count: 1`)
- **連續觸發**: 連續多次超過閾值才觸發告警 (`tolerant_count > 1`)

## 配置說明

### 基本配置

```yaml
threshold_detector:
  name: "cpu_usage_detector"
  type: "threshold_detector"
  config:
    field_name: "cpu_usage"
    upper_threshold: 85.0
    lower_threshold: 5.0
    severity: "high"
    description: "Monitors CPU usage for abnormal values"
    enable_upper: true
    enable_lower: true
    tolerant_count: 3
  enabled: true
```

### 高級配置

```yaml
threshold_detector:
  name: "memory_usage_detector"
  type: "threshold_detector"
  config:
    field_name: "memory_usage_percent"
    upper_threshold: 90.0
    severity: "critical"
    description: "Monitors memory usage for high utilization"
    enable_upper: true
    enable_lower: false
    tolerant_count: 2
  enabled: true
```

### 配置參數

| 參數 | 類型 | 必需 | 默認值 | 說明 |
|------|------|------|--------|------|
| `name` | string | 是 | - | 插件的唯一標識符 |
| `type` | string | 是 | - | 插件類型，必須為 `threshold_detector` |
| `config.field_name` | string | 是 | - | 要監控的字段名稱 |
| `config.upper_threshold` | number | 條件* | - | 上限閾值 |
| `config.lower_threshold` | number | 條件* | - | 下限閾值 |
| `config.severity` | string | 否 | "medium" | 告警嚴重程度 (low/medium/high/critical) |
| `config.description` | string | 否 | - | 偵測器描述 |
| `config.enable_upper` | boolean | 否 | true | 是否啟用上限檢測 |
| `config.enable_lower` | boolean | 否 | true | 是否啟用下限檢測 |
| `config.tolerant_count` | integer | 否 | 1 | 容忍次數 |
| `enabled` | boolean | 否 | true | 是否啟用此插件 |

*當 `enable_upper` 為 true 時，`upper_threshold` 為必需；當 `enable_lower` 為 true 時，`lower_threshold` 為必需

## 使用範例

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "detectviz-platform/internal/adapters/plugins/detectors"
    "detectviz-platform/pkg/platform/contracts"
)

func main() {
    // 創建日誌器和指標提供者
    var logger contracts.Logger
    var metricsProvider contracts.MetricsProvider
    
    // 創建閾值偵測器
    detector := detectors.NewThresholdDetectorPlugin(logger, metricsProvider)
    
    // 配置插件
    config := map[string]interface{}{
        "field_name":       "cpu_usage",
        "upper_threshold":  85.0,
        "lower_threshold":  5.0,
        "severity":         "high",
        "description":      "CPU usage anomaly detector",
        "enable_upper":     true,
        "enable_lower":     true,
        "tolerant_count":   3,
    }
    
    ctx := context.Background()
    
    // 初始化插件
    if err := detector.Init(ctx, config); err != nil {
        log.Fatalf("Failed to initialize detector: %v", err)
    }
    
    // 啟動插件
    if err := detector.Start(ctx); err != nil {
        log.Fatalf("Failed to start detector: %v", err)
    }
    
    // 執行偵測
    data := map[string]interface{}{
        "cpu_usage": 92.5,
        "timestamp": "2024-01-01T10:00:00Z",
    }
    
    result, err := detector.Execute(ctx, data, nil)
    if err != nil {
        log.Fatalf("Failed to execute detection: %v", err)
    }
    
    log.Printf("Detection result: %+v", result)
}
```

### 系統監控場景

```go
package monitoring

import (
    "context"
    "fmt"
    
    "detectviz-platform/internal/adapters/plugins/detectors"
    "detectviz-platform/pkg/platform/contracts"
)

type SystemMonitor struct {
    cpuDetector    detectors.ThresholdDetectorPlugin
    memoryDetector detectors.ThresholdDetectorPlugin
    diskDetector   detectors.ThresholdDetectorPlugin
    logger         contracts.Logger
}

func NewSystemMonitor(logger contracts.Logger, metricsProvider contracts.MetricsProvider) *SystemMonitor {
    return &SystemMonitor{
        cpuDetector:    detectors.NewThresholdDetectorPlugin(logger, metricsProvider),
        memoryDetector: detectors.NewThresholdDetectorPlugin(logger, metricsProvider),
        diskDetector:   detectors.NewThresholdDetectorPlugin(logger, metricsProvider),
        logger:         logger,
    }
}

func (m *SystemMonitor) InitializeDetectors(ctx context.Context) error {
    // CPU 使用率偵測器
    cpuConfig := map[string]interface{}{
        "field_name":       "cpu_usage",
        "upper_threshold":  85.0,
        "severity":         "high",
        "description":      "CPU usage monitor",
        "enable_upper":     true,
        "enable_lower":     false,
        "tolerant_count":   3,
    }
    
    // 記憶體使用率偵測器
    memoryConfig := map[string]interface{}{
        "field_name":       "memory_usage",
        "upper_threshold":  90.0,
        "severity":         "critical",
        "description":      "Memory usage monitor",
        "enable_upper":     true,
        "enable_lower":     false,
        "tolerant_count":   2,
    }
    
    // 磁碟使用率偵測器
    diskConfig := map[string]interface{}{
        "field_name":       "disk_usage",
        "upper_threshold":  95.0,
        "severity":         "critical",
        "description":      "Disk usage monitor",
        "enable_upper":     true,
        "enable_lower":     false,
        "tolerant_count":   1,
    }
    
    // 初始化所有偵測器
    detectors := []struct {
        detector detectors.ThresholdDetectorPlugin
        config   map[string]interface{}
        name     string
    }{
        {m.cpuDetector, cpuConfig, "CPU"},
        {m.memoryDetector, memoryConfig, "Memory"},
        {m.diskDetector, diskConfig, "Disk"},
    }
    
    for _, d := range detectors {
        if err := d.detector.Init(ctx, d.config); err != nil {
            return fmt.Errorf("failed to initialize %s detector: %w", d.name, err)
        }
        
        if err := d.detector.Start(ctx); err != nil {
            return fmt.Errorf("failed to start %s detector: %w", d.name, err)
        }
    }
    
    return nil
}

func (m *SystemMonitor) CheckSystemMetrics(ctx context.Context, metrics map[string]interface{}) error {
    // 檢查 CPU 使用率
    if _, err := m.cpuDetector.Execute(ctx, metrics, nil); err != nil {
        m.logger.Error("CPU detection failed", "error", err)
    }
    
    // 檢查記憶體使用率
    if _, err := m.memoryDetector.Execute(ctx, metrics, nil); err != nil {
        m.logger.Error("Memory detection failed", "error", err)
    }
    
    // 檢查磁碟使用率
    if _, err := m.diskDetector.Execute(ctx, metrics, nil); err != nil {
        m.logger.Error("Disk detection failed", "error", err)
    }
    
    return nil
}
```

### 動態配置覆蓋

```go
// 運行時動態調整閾值
runtimeConfig := map[string]interface{}{
    "upper_threshold": 95.0,  // 臨時提高閾值
    "severity":        "critical",
    "tolerant_count":  1,     // 立即觸發
}

result, err := detector.Execute(ctx, data, runtimeConfig)
```

## 異常偵測邏輯

### 偵測流程

1. **數據提取**: 從輸入數據中提取指定字段的數值
2. **閾值比較**: 與配置的上下限閾值進行比較
3. **容錯判斷**: 根據容忍次數判斷是否觸發告警
4. **結果生成**: 生成包含異常信息的偵測結果

### 偵測結果

```go
type ThresholdDetectionResult struct {
    IsAnomalous   bool      `json:"is_anomalous"`    // 是否異常
    Value         float64   `json:"value"`           // 實際值
    Threshold     float64   `json:"threshold"`       // 觸發的閾值
    ThresholdType string    `json:"threshold_type"`  // "upper" 或 "lower"
    Severity      string    `json:"severity"`        // 嚴重程度
    Description   string    `json:"description"`     // 描述信息
    DetectedAt    time.Time `json:"detected_at"`     // 偵測時間
    FieldName     string    `json:"field_name"`      // 字段名稱
    Confidence    float64   `json:"confidence"`      // 置信度
}
```

## 監控和指標

### 內建指標

1. **detector_started_total**: 偵測器啟動次數
2. **detector_stopped_total**: 偵測器停止次數
3. **detector_executions_total**: 偵測執行次數
4. **detector_anomalies_total**: 異常偵測次數
5. **detector_execution_duration_seconds**: 偵測執行時間
6. **detector_extraction_errors_total**: 數據提取錯誤次數

### 指標標籤

- `detector_type`: "threshold"
- `plugin`: 插件名稱
- `anomalous`: 是否異常 ("true"/"false")
- `severity`: 嚴重程度
- `threshold_type`: 閾值類型 ("upper"/"lower")
- `field`: 字段名稱

## 最佳實踐

### 閾值設定

1. **基線建立**: 先收集歷史數據，建立正常值的基線
2. **動態調整**: 根據業務需求和系統特性動態調整閾值
3. **分級告警**: 設置不同嚴重程度的閾值，實現分級告警

### 容錯配置

```go
// 對於波動較大的指標，使用較高的容忍次數
config := map[string]interface{}{
    "field_name":       "network_latency",
    "upper_threshold":  100.0,
    "tolerant_count":   5,  // 連續5次超過才告警
}

// 對於關鍵指標，使用較低的容忍次數
config := map[string]interface{}{
    "field_name":       "disk_usage",
    "upper_threshold":  95.0,
    "tolerant_count":   1,  // 立即告警
}
```

### 性能優化

1. **字段預驗證**: 確保監控的字段存在且為數值型
2. **批量處理**: 對多個數據點進行批量偵測
3. **異步處理**: 在高頻場景下使用異步偵測

## 故障排除

### 常見問題

#### 1. 字段不存在
```
Error: 提取檢測值失敗: field not found
```
**解決方案**: 確保 `field_name` 在輸入數據中存在

#### 2. 數據類型錯誤
```
Error: 無法將字段值轉換為數值
```
**解決方案**: 確保字段值為數值型 (int, float)

#### 3. 配置驗證失敗
```
Error: 配置驗證失敗: field_name 不能為空
```
**解決方案**: 檢查配置參數的完整性和正確性

### 調試技巧

1. **啟用詳細日誌**: 設置日誌級別為 DEBUG
2. **測試數據**: 使用已知的測試數據驗證偵測邏輯
3. **指標監控**: 觀察內建指標以了解偵測器狀態

## 擴展功能

### 自定義嚴重程度

```go
// 可以根據超過閾值的程度動態調整嚴重程度
func calculateSeverity(value, threshold float64) string {
    ratio := value / threshold
    if ratio > 2.0 {
        return "critical"
    } else if ratio > 1.5 {
        return "high"
    } else if ratio > 1.2 {
        return "medium"
    }
    return "low"
}
```

### 複合條件

```go
// 結合多個閾值偵測器實現複合條件偵測
type CompositeDetector struct {
    detectors []ThresholdDetectorPlugin
}

func (c *CompositeDetector) DetectAnomalies(ctx context.Context, data map[string]interface{}) bool {
    anomalyCount := 0
    for _, detector := range c.detectors {
        result, _ := detector.Execute(ctx, data, nil)
        if result.IsAnomalous {
            anomalyCount++
        }
    }
    // 當多個偵測器都發現異常時才觸發告警
    return anomalyCount >= 2
}
```

## 相關資源

- [Detectviz 平台異常偵測指南](../detection/guide.md)
- [監控最佳實踐](../monitoring/best-practices.md)
- [插件開發指南](../development/plugin-development.md)
- [告警配置指南](../alerting/configuration.md)

## 版本歷史

- **v1.0.0**: 初始版本，支援基本閾值偵測功能
- **v1.1.0**: 添加容錯機制和嚴重程度分級
- **v1.2.0**: 添加指標統計和監控功能
- **v1.3.0**: 添加動態配置覆蓋和性能優化 