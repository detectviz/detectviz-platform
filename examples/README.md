# Detectviz 平台插件使用示例

本目錄包含了 Detectviz 平台插件的使用示例，展示了如何配置和使用數據導入器和偵測器插件。

## 里程碑 0.5 完成功能

### 🔌 已實現的插件

#### 1. CSV 導入器插件 (CSVImporterPlugin)
- **位置**: `internal/adapters/plugins/importers/csv_importer.go`
- **功能**: 將 CSV 文件數據導入到數據庫中
- **特性**:
  - 支持自定義分隔符
  - 支持跳過指定行數
  - 支持列名映射
  - 批量插入優化
  - 數據驗證
  - 最大行數限制

#### 2. 閾值偵測器插件 (ThresholdDetectorPlugin)
- **位置**: `internal/adapters/plugins/detectors/threshold_detector.go`
- **功能**: 基於閾值的異常偵測
- **特性**:
  - 支持上限和下限閾值
  - 可配置告警嚴重程度
  - 支持容忍次數設置
  - 集成指標監控
  - 靈活的運行時配置

### 📊 示例數據

#### sample_data.csv
包含系統監控數據的示例 CSV 文件，包含以下欄位：
- `timestamp`: 時間戳
- `cpu`: CPU 使用率 (%)
- `memory`: 記憶體使用率 (%)
- `disk`: 磁碟使用率 (%)
- `response_time`: API 響應時間 (ms)

數據中包含了一些異常值，用於測試偵測器功能：
- CPU 使用率超過 90%
- 記憶體使用率超過 85%
- API 響應時間超過 5000ms

### ⚙️ 配置示例

#### 插件配置 (configs/plugins_config.yaml)
完整的插件配置示例，包含：
- 導入器配置
- 偵測器配置
- 偵測器實例配置
- 導入任務配置
- 偵測流程配置

#### 使用場景示例

1. **系統監控場景**:
   - 導入系統指標數據
   - 監控 CPU、記憶體、磁碟使用率
   - 當指標超過閾值時觸發告警

2. **API 性能監控場景**:
   - 導入 API 日誌數據
   - 監控響應時間
   - 檢測性能異常

### 🚀 快速開始

1. **配置插件**:
   ```yaml
   # 在 configs/plugins_config.yaml 中配置插件
   importers:
     csv_importer:
       name: "csv_importer_plugin"
       enabled: true
   
   detectors:
     threshold_detector:
       name: "threshold_detector_plugin"
       enabled: true
   ```

2. **創建偵測器實例**:
   ```yaml
   detector_instances:
     - name: "cpu_usage_detector"
       type: "threshold_detector"
       config:
         field_name: "cpu_usage"
         upper_threshold: 90.0
         severity: "high"
   ```

3. **配置導入任務**:
   ```yaml
   import_tasks:
     - name: "system_metrics_import"
       importer: "csv_importer"
       config:
         table_name: "system_metrics"
         column_mapping:
           "cpu": "cpu_usage"
           "memory": "memory_usage"
   ```

4. **設置偵測流程**:
   ```yaml
   detection_workflows:
     - name: "system_monitoring"
       data_source: "system_metrics"
       detectors:
         - "cpu_usage_detector"
       schedule: "*/2 * * * *"
   ```

### 📈 可觀察性集成

插件已集成 Prometheus 指標監控：
- `detector_started_total`: 偵測器啟動次數
- `detector_stopped_total`: 偵測器停止次數
- `detector_executions_total`: 偵測器執行次數
- `detector_anomalies_total`: 檢測到的異常次數
- `detector_execution_duration_seconds`: 偵測器執行時間
- `detector_extraction_errors_total`: 數據提取錯誤次數

### 🔧 擴展開發

要開發新的插件，請參考：
1. 實現對應的插件介面 (Plugin, Importer, DetectorPlugin)
2. 在 `internal/adapters/plugins/` 下創建實現
3. 在配置文件中註冊插件
4. 更新 `docs/architecture/interface_spec.md` 中的進度標記

### 📋 TODO

里程碑 0.5 的後續工作：
- [ ] 實現插件註冊和載入機制
- [ ] 添加更多類型的偵測器 (統計異常偵測、模式匹配等)
- [ ] 實現通知和告警插件
- [ ] 添加插件性能監控和健康檢查
- [ ] 完善錯誤處理和重試機制 