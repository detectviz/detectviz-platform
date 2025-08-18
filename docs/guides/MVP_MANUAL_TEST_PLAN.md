# DetectViz 空實作 MVP 手動測試指南

## 📋 概述

本文件提供完整的手動測試指南，用於驗證 DetectViz 平台的空實作 MVP。此版本使用模擬數據替代真實後端調用，讓你能夠測試完整的 Agent 協調工作流程。

## 🎯 測試目標

驗證以下核心功能：
- ✅ `postmortem_orchestrator` Agent 團隊協調
- ✅ `data_collector` → `root_cause_analyzer` → `report_writer` 數據流
- ✅ 模擬工具調用（`get_health_metrics`, `generate_report`）
- ✅ 端到端事後複盤報告生成

## 🚀 啟動步驟

### 1. 啟動 ADK Web 伺服器

```bash
# 切換到專案根目錄
cd /Users/zoe/Documents/detectviz-platform

# 啟動 ADK Web 伺服器
adk web python-adk-runtime/agents --port 8080
```

**預期輸出：**
```
+-----------------------------------------------------------------------------+
| ADK Web Server started                                                      |
|                                                                             |
| For local testing, access at http://localhost:8080.                         |
+-----------------------------------------------------------------------------+
```

### 2. 訪問 Web 介面

1. 打開瀏覽器
2. 訪問：`http://localhost:8080`
3. 選擇 `postmortem` 應用
4. 創建新的對話 session

## 🧪 測試腳本

### 階段一：觸發數據收集

**輸入提示詞：**
```
請為事件 INC-MOCK-2025-001 進行事後複盤，受影響服務為 mock-service
```

**預期行為：**
- `postmortem_orchestrator` 接收請求
- 委派任務給 `data_collector` Agent
- `data_collector` 要求提供事件時間範圍

**預期回應：**
```
好的，我來收集事件 INC-MOCK-2025-001 的相關資料，受影響服務為 mock-service。

請提供事件的時間範圍，以利我查詢相關指標。
```

---

### 階段二：提供時間範圍並觸發工具調用

**輸入提示詞：**
```
事件時間範圍是 2025-08-16 14:00:00 到 2025-08-16 16:00:00
```

**預期行為：**
- `data_collector` 調用 `get_health_metrics` 工具
- 後端日誌顯示：`--- [TOOL EXECUTED] get_health_metrics_func: Returning mock anomaly data for service 'mock-service'. ---`
- 返回模擬的異常指標數據

**預期回應：**
```
我已收集到以下指標：

* **CPU 使用率：** 從 20% 飆升至 95%。
* **錯誤率：** 從 1% 上升至 15%。
* **延遲：** P99 請求延遲從 150ms 增加到 2500ms。

請問是否還需要收集其他指標或資料？
```

---

### 階段三：觸發根因分析

**輸入提示詞：**
```
數據收集完成，請進行根因分析
```

**預期行為：**
- 任務轉交給 `root_cause_analyzer` Agent
- 基於收集的數據進行分析
- 生成根因分析結論

**預期回應：**
```
初步分析顯示，事件的根本原因可能與服務器資源耗盡有關，CPU 使用率飆升導致了請求處理延遲和錯誤率上升。
```

---

### 階段四：生成最終報告

**輸入提示詞：**
```
分析完成，請生成完整的事後複盤報告
```

**預期行為：**
- 任務轉交給 `report_writer` Agent
- 調用 `generate_report` 工具
- 後端日誌顯示：`--- [TOOL EXECUTED] generate_report_func: Generating mock report. ---`
- 生成完整的 Markdown 格式報告

**預期回應：**
```markdown
**事後檢討報告 (模擬)**

**事件資訊**

* **事件 ID**: INC-MOCK-2025-001
* **受影響服務**: mock-service
* **報告生成時間**: 模擬時間戳

**根本原因分析**

服務器資源耗盡，CPU 使用率飆升導致請求處理延遲和錯誤率上升。

**改善建議**

1. 增加服務器資源監控告警。
2. 對相關服務進行壓力測試。
3. 建立自動擴容機制，防止資源耗盡。
4. 實施更嚴格的負載均衡策略。

**預防措施**

* 設定 CPU 使用率告警閾值為 80%
* 建立延遲監控儀表板
* 定期進行災難復原演練

---

*本報告由 DetectViz 平台自動生成*
```

## ✅ 驗證檢查清單

### Agent 協調驗證
- [ ] `postmortem_orchestrator` 正確識別請求
- [ ] 任務順序委派：`data_collector` → `root_cause_analyzer` → `report_writer`
- [ ] 數據在 Agent 之間正確傳遞
- [ ] 每個 Agent 完成後正確返回控制權

### 工具調用驗證
- [ ] `get_health_metrics` 被成功調用
- [ ] 模擬數據包含 CPU、延遲、錯誤率指標
- [ ] `generate_report` 被成功調用
- [ ] 報告包含分析結論和建議

### 後端日誌驗證
在終端中確認看到以下日誌：
```
--- [TOOL EXECUTED] get_health_metrics_func: Returning mock anomaly data for service 'mock-service'. ---
--- [TOOL EXECUTED] generate_report_func: Generating mock report. ---
```

### 數據流驗證
- [ ] 階段一：請求識別和委派
- [ ] 階段二：數據收集和工具調用
- [ ] 階段三：分析結論生成
- [ ] 階段四：最終報告整合

## 🎨 高級測試案例

### 測試變化一：不同服務名稱
```
請為事件 INC-TEST-2025-002 進行事後複盤，受影響服務為 api-gateway
```

**預期：** 模擬數據會自動適應 `api-gateway` 服務名稱

### 測試變化二：不同時間範圍
```
事件時間範圍是 2025-08-15 10:00:00 到 2025-08-15 12:00:00
```

**預期：** 時間範圍會反映在收集的數據中

### 測試變化三：完整一次性請求
```
請為事件 INC-BATCH-001 進行事後複盤，受影響服務為 payment-service，時間範圍是 2025-08-16 09:00:00 到 2025-08-16 11:00:00
```

**預期：** Agent 團隊會自動處理完整流程

## 🔧 故障排除

### 常見問題

#### 問題：ADK Web 無法啟動
**解決方案：**
```bash
# 檢查目錄結構
ls -la python-adk-runtime/agents/postmortem/
# 應該看到：agent.py, __init__.py

# 檢查 root_agent 變數
grep -n "root_agent" python-adk-runtime/agents/postmortem/agent.py
```

#### 問題：工具未被調用
**檢查：**
1. 後端日誌中是否有錯誤訊息
2. Agent 是否正確導入工具
3. 模擬數據函式是否正確實作

#### 問題：Agent 協調失敗
**檢查：**
1. 每個 Agent 的 `instruction` 是否正確
2. `sub_agents` 是否正確配置
3. Agent 間的委派是否成功

### 重新啟動流程
```bash
# 停止伺服器
Ctrl + C

# 重新啟動
adk web python-adk-runtime/agents --port 8080

# 清除瀏覽器快取並重新載入頁面
```

## 📊 性能指標

### 預期回應時間
- 階段一（委派）：< 2 秒
- 階段二（數據收集）：< 5 秒（包含 1 秒模擬延遲）
- 階段三（分析）：< 3 秒
- 階段四（報告生成）：< 5 秒（包含 1 秒模擬延遲）

### Token 使用統計
每個階段的約略 token 消耗：
- 階段一：~500 tokens
- 階段二：~700 tokens
- 階段三：~1200 tokens
- 階段四：~2500 tokens

## 🎯 測試成功標準

完成測試後，確認：

✅ **功能性**：所有 4 個階段成功完成  
✅ **正確性**：生成的報告包含正確的事件資訊和分析  
✅ **穩定性**：多次測試結果一致  
✅ **可觀察性**：後端日誌清楚顯示工具執行  
✅ **用戶體驗**：Web 介面回應流暢，訊息清晰  

## 📝 測試記錄模板

```
測試日期：____
測試人員：____
ADK 版本：____

階段一 - 請求識別：□ 通過 □ 失敗
階段二 - 數據收集：□ 通過 □ 失敗  
階段三 - 根因分析：□ 通過 □ 失敗
階段四 - 報告生成：□ 通過 □ 失敗

備註：
_________________________________
_________________________________
```

---

## 🔄 下一步

此空實作 MVP 驗證成功後，可以開始：
1. 實作真實的 Go 端 HealthAggregator 
2. 整合真實的 InfluxDB 查詢
3. 添加更多業務邏輯和錯誤處理
4. 實作完整的測試覆蓋

**測試愉快！** 🚀