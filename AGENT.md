# AI 開發守則與協作指南

本文檔提供 AI 與人類開發者在 Detectviz 平台上協作的完整指南，整合了 AI 開發守則、協作流程和貢獻準則。

## 文檔層級關係與執行機制

```
docs/sre-services-map.md (架構憲法 - 業務邏輯與決策)
         ↓ 指導
    spec.md (技術規格 - 實作細節與系統設計)
         ↓ 實現
   AGENT.md (本文檔 - AI 開發守則與協作規範)
         ↓ 具體落實
   各模組 llm.txt (模組專用開發檢查清單)
         ↓ 指引
   MVP Implementation (實際開發工作)
```

### 模組 llm.txt 執行機制
本指南的有效性依賴於各模組 llm.txt 的嚴格執行：
- **contracts/llm.txt**：SSOT 契約維護檢查清單
- **go-platform/llm.txt**：Go 執行層開發檢查清單  
- **python-adk-runtime/llm.txt**：Python 決策層開發檢查清單

**重要原則**：每次開發必須先查閱對應模組的 llm.txt，完成後必須逐項檢查

### Agent vs Tool 職責劃分原則（黃金準則）
**核心設計原則**：決策與執行分離

#### Agent 職責（決策層 - WHY/WHAT/WHEN）
- **決策制定**：WHY（為什麼）、WHAT（做什麼）、WHEN（何時做）
- **工作流編排**：協調多個 Tool 完成複雜任務
- **知識整合**：結合歷史經驗和當前數據進行智能決策
- **業務邏輯處理**：分析、判斷、策略制定
- **重要限制**：**Agent 不直接操作數據**，所有數據查詢、API 調用、文件生成都通過 Tool

#### Tool 職責（執行層 - HOW/WHERE/WITH）
- **具體執行**：HOW（如何做）、WHERE（在哪做）、WITH（用什麼）
- **數據操作**：查詢、轉換、存儲、生成
- **外部系統集成**：API 調用、文件操作、網路通訊
- **重要特性**：**Tool 不包含業務邏輯**，是無狀態、冪等、原子性的執行單元

#### 混合架構決策指引
- **高性能數據查詢** → Go 端實現（如 HealthAggregator 核心）
- **業務邏輯處理** → Python 端實現（Agent 決策）
- **跨語言橋接** → 通過 gRPC RemoteTool 連接兩端

### MVP 聚焦：Phase 3（事後複盤）
當前開發重點為 **Phase 3: 事後複盤系統**，包含：
- postmortem_orchestrator（ADK Root Agent 決策協調）
- HealthAggregator（數據查詢）
- ReportGenerator（報告生成）
- ResponseHistoryStore（知識存儲）

---

## 核心開發原則

所有開發工作必須嚴格遵守以下三大原則：

### 1. SSOT 契約優先 (Contracts-First)
任何跨語言介面、API 或組態結構的變更，**必須**優先在 `contracts/` 目錄下完成。下游的 Go 和 Python 專案僅能同步對齊生成的程式碼與規範，禁止手動修改。

**具體操作**：
- Proto：`contracts/proto/detectviz/contracts/v1/adk_bridge.proto`
- Schema：`contracts/schemas/{config.schema.json,module.card.schema.json}`
- 樣本：`contracts/samples/config.yaml`
- 生成與驗證：
  ```bash
  cd contracts
  buf lint && buf generate
  make validate            # 若已提供
  make validate-cards      # 驗證所有 module.card.json
  ```
- **禁止**手動編輯任何生成碼（例如 `*.pb.go`、`*_pb2.py`、`*_pb2_grpc.py`）

### 2. 文件同步 (Docs-as-Code)
任何影響使用者行為、系統架構或配置方式的程式碼變更，都必須同步更新相關文件。文件是交付成果的一部分。

### 3. 安全第一 (Security-First)
嚴禁在版本控制中提交任何真實的密鑰、Token 或其他敏感憑證。請一律使用環境變數或 Secret 管理機制。

### 程式碼品質要求
- **類型標註**：使用 Python typing 和 Go 明確類型
- **文檔字串**：每個函數都必須有完整的 docstring
- **錯誤處理**：完善的異常處理和錯誤回復機制
- **測試覆蓋率**：單元測試覆蓋率 > 90%
- **模組卡片**：每個 Agent/Tool 必須有對應的 `module.card.json`

### 工作流程規範

#### 開發前準備
1. **文檔研讀**：仔細閱讀相關章節，確認理解 Agent 的決策職責
2. **設計決策**：設計決策樹或決策矩陣，明確業務邏輯
3. **技術選型**：根據混合架構原則選擇 Go 或 Python 實現

#### 開發過程
1. **職責分離**：嚴格遵守 Agent/Tool 職責分離，Agent 不直接操作數據
2. **跨語言整合**：使用 RemoteTool 調用 Go 服務，確保類型安全
3. **監控埋點**：添加適當的日誌、指標和追蹤點

#### 完成驗證
1. **測試完備性**：確保測試覆蓋率達標，包含異常處理測試
2. **文檔同步**：更新相關文檔，確保一致性
3. **性能驗證**：特別關注 Go 端的高性能查詢和數據處理

### 文檔與溝通風格規範
- **禁用 Emoji**：所有文檔、程式碼註解、提示詞和溝通內容都不使用 emoji 符號
- **專業文字表達**：使用清晰的文字描述代替視覺符號，保持專業和正式的技術文檔風格

---

## AI 開發工作流程（強制執行機制）

### 0. 開發前必讀（強制）
**每次開發前必須執行以下步驟**：
- [ ] **第一步**：閱讀本文檔相關章節
- [ ] **第二步**：查閱對應模組的 llm.txt 檢查清單
  - 契約變更 → [`contracts/llm.txt`](./contracts/llm.txt)
  - Go 平台開發 → [`go-platform/llm.txt`](./go-platform/llm.txt)  
  - Python Agent 開發 → [`python-adk-runtime/llm.txt`](./python-adk-runtime/llm.txt)
- [ ] **第三步**：確認理解 Agent vs Tool 職責分離原則
- [ ] **第四步**：制定具體實施計畫，標明使用哪個模組的檢查清單

### 1. 變更前規劃
- [ ] 識別變更類型：SSOT 契約、核心邏輯、介面變更、內部重構
- [ ] 評估影響範圍：使用者介面、系統行為、文檔、範例
- [ ] **模組檢查清單對照**：確認對應模組 llm.txt 的適用項目
- [ ] 建立 TODO 清單，**必須包含文檔更新任務**
- [ ] **MVP 檢查**：確認變更符合 Phase 3 事後複盤範圍

### 2. 實作過程中
- [ ] 遵循 SSOT 原則，契約變更優先
- [ ] **模組規範遵循**：嚴格按照對應模組 llm.txt 的必守規範
- [ ] 保持向後相容性，除非明確說明破壞性變更
- [ ] 記錄重要的設計決策和權衡考量
- [ ] **MVP 聚焦**：避免添加非 Phase 3 範圍的功能

### 3. 完成後檢查（模組特定）
- [ ] **模組檢查清單執行**：逐項完成對應模組 llm.txt 的提交前檢查
- [ ] 編譯和基本功能測試
- [ ] 檢查是否需要更新文檔（參考下述檢查清單）
- [ ] 驗證所有範例指令和配置仍然有效
- [ ] 確認變更符合平台設計原則
- [ ] **MVP 驗證**：確認符合 8 週交付計畫

### 4. 文檔同步更新
- [ ] 根據變更類型更新相應文檔
- [ ] 更新快速開始指南中的指令
- [ ] **模組文檔一致性**：確保模組 README 與 llm.txt 保持同步
- [ ] 檢查所有文檔間的一致性
- [ ] **MVP 文檔**：確保文檔反映當前 MVP 狀態

### 5. 最終驗證（強制）
**提交前必須確認**：
- [ ] **模組檢查清單 100% 完成**：對應 llm.txt 的所有檢查項都已通過
- [ ] **跨模組整合測試**：如果涉及多個模組，確保整合功能正常
- [ ] **文檔更新確認**：相關模組的 README.md 已同步更新

---

## 文檔更新檢查清單

### P0（必須更新）
當變更影響以下內容時，必須同步更新文檔：
- [ ] CLI 參數或環境變數
- [ ] 啟動流程或配置方式  
- [ ] 用戶介面或操作步驟
- [ ] 快速開始指南中的指令

### P1（重要更新）
當變更影響以下內容時，應該更新文檔：
- [ ] 系統架構或行為變更
- [ ] 錯誤處理流程
- [ ] 性能優化說明
- [ ] 新功能介紹

### P2（建議更新）
當變更影響以下內容時，可選更新文檔：
- [ ] 內部代碼結構
- [ ] 函數重構細節
- [ ] 開發者注意事項

### 需要更新的文檔清單

**核心文檔**：
- [`README.md`](./README.md) - 項目總覽和快速開始
- [`spec.md`](./spec.md) - 平台技術規格
- [`AGENT.md`](./AGENT.md) - AI 開發守則（本文檔）

**子項目文檔**：
- [`go-platform/README.md`](./go-platform/README.md) - Go 平台說明
- [`python-adk-runtime/README.md`](./python-adk-runtime/README.md) - Python 運行時說明
- [`contracts/README.md`](./contracts/README.md) - 契約管理說明

**開發指南**：
- [`docs/`](./docs/) 目錄下的相關指南
- 各子目錄的 README 文件

---

## MVP 專用開發守則

### postmortem_orchestrator ADK Agent 開發規範

**核心原則**：遵循 Google ADK 標準，使用 Agent 團隊協作模式

#### 實作要求

1. **使用 ADK Agent 定義**：
   ```python
   from google import adk
   from detectviz_adk.tools.adk_tools import get_health_metrics, generate_report
   
   postmortem_orchestrator = adk.Agent(
       name="postmortem_orchestrator",
       model="gemini-2.0-flash",
       instruction="你是事後檢討協調器...",
       sub_agents=[data_collector_agent, root_cause_analyzer, report_writer]
   )
   ```

2. **ADK Runner 使用範例**：
   ```python
   from detectviz_adk.runners.postmortem_runner import PostmortemRunner
   
   async def run_postmortem_analysis(incident_request):
       # 使用 ADK Runner 執行
       runner = PostmortemRunner()
       
       # ADK Agent 團隊會自動處理：
       # 1. data_collector: 根據事件決定收集策略
       # 2. root_cause_analyzer: 分析數據並制定報告策略
       # 3. report_writer: 生成結構化報告
       
       result = await runner.execute_postmortem(incident_request)
       return result
   ```

3. **混合架構決策指引**：
   - **Python 端**：業務邏輯、決策制定、工作流編排
   - **Go 端**：高性能查詢、數據處理、外部系統集成
   - **gRPC 通訊**：使用 RemoteTool 進行跨語言調用

### 正確的 Agent 實現範例

```python
# ADK Root Agent 實作範例
from google import adk
from detectviz_adk.tools.adk_tools import get_health_metrics, generate_report

postmortem_orchestrator = adk.Agent(
    name="postmortem_orchestrator",
    model="gemini-2.0-flash",
    instruction="""你是事後檢討協調器，負責管理整個檢討流程。

你有以下子代理可以委派任務：
1. 'data_collector': 收集事故相關資料和指標
2. 'root_cause_analyzer': 分析根本原因和相關性
3. 'report_writer': 產生完整報告和文件

重要：你不直接使用工具，而是透過委派給專門的子代理來完成任務。""",
    description="協調事後檢討流程的主代理",
    tools=[],  # Root Agent 不直接使用工具
    sub_agents=[data_collector_agent, root_cause_analyzer, report_writer]
)

# 使用 PostmortemRunner 執行
from detectviz_adk.runners.postmortem_runner import PostmortemRunner

async def run_postmortem_analysis(incident_request):
    runner = PostmortemRunner()
    result = await runner.execute_postmortem(incident_request)
    return result
```

### MVP 檢查清單

#### 設計階段（Week 1-2）
- [ ] **P0** 創建 `python-adk-runtime/src/detectviz_adk/agents/postmortem/` 目錄結構
- [ ] **P0** 實現 postmortem_orchestrator ADK Agent 團隊基本架構
- [ ] **P0** 定義 HealthAggregator Go 端插件接口
- [ ] **P1** 創建 module.card.json 並通過驗證
- [ ] **P1** 建立基本測試框架

#### 實作階段（Week 3-6）
- [ ] **P0** 實現 postmortem_orchestrator ADK Agent 團隊核心協作邏輯
- [ ] **P0** 完成 HealthAggregator Go 端插件實作
- [ ] **P0** 實現 ReportGenerator 基本功能
- [ ] **P1** 集成 ResponseHistoryStore 知識存儲
- [ ] **P1** 添加錯誤處理和重試邏輯
- [ ] **P2** 性能優化和資源監控

#### 文檔階段（Week 7-8）
- [ ] **P0** 更新 README.md 反映 MVP 功能
- [ ] **P0** 創建使用說明和範例
- [ ] **P1** 完善 API 文檔
- [ ] **P1** 更新架構圖和流程圖
- [ ] **P2** 創建故障排查指南

### MVP 里程碑與交付標準（8 週計畫）
- **Week 2**：基本架構搭建完成，可以啟動 Agent
- **Week 4**：核心功能實現，可以執行簡單的事後複盤流程
- **Week 6**：完整功能實現，包含錯誤處理和優化
- **Week 8**：文檔完善，準備交付生產環境

---

## Go 平台實作守則

（詳細技術操作請參考 [`README.md`](./README.md) 快速開始部分）

- **設定載入**：一律透過 `internal/configx/loader.go`；**禁止**在其他模組自行讀檔或解析 YAML
- **日誌**：使用 `zap`，寫入純文字檔案（非 JSON），預設 `./var/log/detectviz/detectviz.log`
- **OTEL**：依 `observability.otlp` 初始化 Traces/Metrics
- **健康檢查**：提供 `/livez`、`/readyz` 與 gRPC Health
- **優雅關機**：收到 SIGTERM/SIGINT → 先標記 not-ready → 停 HTTP demo → `GracefulStop` gRPC → 關閉 OTel
- **Plugin Registry（並發安全）**：提供 `RegisterStrict(toolID, h) error`，同名即報錯，不覆蓋

詳細啟動指令和配置請參考：[`README.md - 快速開始`](./README.md#快速開始)

---

## Python ADK Runtime 守則

（詳細技術操作請參考 [`README.md`](./README.md) 快速開始部分）

- **ADK 對齊**：遵守 Agent / Workflow / MemoryBank / Tools / Capabilities 模組邊界
- **RemoteTool**：透過 `grpc.aio` 呼叫 ToolBridge；端點以 `DETECTVIZ_TOOLBRIDGE_ADDR` 設定
- **設定載入**：使用 `detectviz_adk/config/loader.py`，與 Go 相同的搜尋序與環境覆蓋
- **模組卡**：新增/擴增元件需附 `module.card.json` 並通過驗證
- **安全**：Python 不持有雲端憑證；外部系統交互集中於 Go 插件（Tools）

### RemoteTool 使用規範

**MVP 實作要求**：
```python
# 正確使用方式
from detectviz_adk.tools.remote_tool import RemoteTool
from detectviz_adk.runners.postmortem_runner import PostmortemRunner

# 使用 ADK Agent 團隊
runner = PostmortemRunner()

async def execute_postmortem_analysis(incident_request):
    """使用 ADK Agent 團隊執行事後檢討分析"""
    try:
        # ADK Runner 會自動協調 Agent 團隊：
        # - data_collector: 收集相關資料
        # - root_cause_analyzer: 分析根本原因 
        # - report_writer: 產生完整報告
        
        result = await runner.execute_postmortem(incident_request)
        return result
    except Exception as e:
        logger.error(f"Postmortem analysis failed: {e}")
        raise
```

詳細環境設置和配置請參考：[`README.md - 快速開始`](./README.md#快速開始)

---

## 測試與驗證清單

- **契約**：`cd contracts && buf lint && buf generate && make validate-cards`
- **啟動**：詳細啟動指令參考 [`README.md`](./README.md)
- **健康**：檢查 `GET /livez` 與 `GET /readyz`（Ready 需等 ToolBridge 成功啟動）
- **Drilldown**：Grafana Explore 檢查 Logs ↔ Traces ↔ Profiles 是否關聯
- **關機**：發送 SIGTERM，確認優雅關機順序與 OTel flush 成功
- **插件**：測試 `RegisterStrict` 重複註冊回錯；`RegisterOrReplace` 能熱替換且釋放資源

---

## 常見錯誤與排查

- `config validation failed: ... profiling.* not allowed`：`config.yaml` 殘留舊欄位；請以最新 Schema 修正
- `plugin.paths: Invalid type. Expected: array`：請確認是 YAML 陣列（即使只有一個路徑也需 `[...]`）
- `could not import grpc/codes`：在 `contracts/` 重新 `buf generate`，並於 Go 端 `go mod tidy`
- **Logs 無輸出**：確認 `zap.ReplaceGlobals`、日誌目錄存在、且未使用 `log.Printf`
- **Traces 缺失**：檢查 OTLP endpoint/協定（grpc/http）與 Alloy 狀態

詳細故障排查請參考：[`README.md - 故障排查`](./README.md#故障排查)

---

## 模組 llm.txt 強制執行指南

### 執行模式：三級檢查制度

#### 第一級：開發前強制檢查
**AI 開發者在開始任何開發工作前，必須完成以下動作**：

1. **模組識別**：明確本次開發涉及的模組
   ```bash
   # 範例：如果要開發 postmortem_orchestrator
   涉及模組：python-adk-runtime (主要), go-platform (插件), contracts (可能)
   ```

2. **檢查清單預覽**：完整閱讀對應的 llm.txt
   ```bash
   # 必讀文件
   cat python-adk-runtime/llm.txt  # 主要模組
   cat go-platform/llm.txt         # 如果需要 Go 端支持
   cat contracts/llm.txt           # 如果涉及契約變更
   ```

3. **能力確認清單**：
   - [ ] 我理解該模組的 Agent vs Tool 職責分離要求
   - [ ] 我知道該模組的必守規範和禁止事項
   - [ ] 我明確該模組的 MVP Phase 3 特定要求
   - [ ] 我理解該模組的品質標準和測試要求

#### 第二級：開發過程中強制檢查
**開發過程中每達成一個里程碑，必須進行以下檢查**：

1. **規範遵循確認**：
   ```markdown
   ## 開發進度檢查（模組：{模組名稱}）
   
   ### 已完成工作
   - [ ] 功能 A 已完成，符合 llm.txt 第 X 項要求
   - [ ] 功能 B 已完成，符合 llm.txt 第 Y 項要求
   
   ### llm.txt 規範檢查
   - [ ] Agent vs Tool 職責分離：✓ 已遵循
   - [ ] 技術實現規範：✓ 已遵循  
   - [ ] MVP 聚焦要求：✓ 已遵循
   
   ### 下一步計畫
   - [ ] 接下來要完成的 llm.txt 檢查項目
   ```

#### 第三級：提交前強制檢查
**任何程式碼提交前，必須完成 100% 的模組檢查清單**：

```bash
# 提交前強制執行腳本（範例）
#!/bin/bash
echo "=== 模組 llm.txt 檢查清單驗證 ==="

# 確認涉及的模組
read -p "涉及的模組 (contracts/go-platform/python-adk-runtime): " modules

for module in $modules; do
    echo "正在檢查 $module/llm.txt..."
    
    # 這裡可以實施自動化檢查
    # 例如：檢查是否有生成碼被修改、測試覆蓋率等
    
    read -p "$module 的 llm.txt 檢查清單是否 100% 完成？ (y/N): " completed
    if [[ $completed != "y" ]]; then
        echo "❌ $module 檢查清單未完成，禁止提交"
        exit 1
    fi
done

echo "✅ 所有模組檢查清單已完成，可以提交"
```

### 問責與改進機制

#### 違規處理
如果發現開發過程中未遵循模組 llm.txt：
1. **輕微違規**：提醒並要求補充完成檢查清單
2. **嚴重違規**：要求重新開發，嚴格按照檢查清單執行
3. **重複違規**：檢討並加強 llm.txt 的具體性和可執行性

#### 持續改進
**定期檢討機制**：
- **每週檢討**：llm.txt 執行情況和遇到的問題
- **每月更新**：根據實際開發經驗更新 llm.txt 內容
- **每季評估**：評估 llm.txt 對開發品質和效率的影響

**改進回饋循環**：
```
實際開發問題 → 更新 llm.txt → 驗證有效性 → 推廣執行
```

---

## 貢獻指南

### 基本原則
- **SSOT 契約優先**：跨語言介面或設定變更，請先修改 `contracts/`（proto/schema/samples），再更新下游（Go/Python/Docs）
- **文件同步**：任何影響使用者或行為的變更，必須同步更新對應文件
- **安全**：禁止提交真實密鑰/Token；請使用環境變數或 Secret

### 提交前檢查清單
- [ ] 若變更觸及契約，已在 `contracts/` 執行 `buf lint && buf generate` 與驗證腳本
- [ ] 若變更觸及 Go/Python，已對齊生成碼與設定載入，端到端測過（含健康檢查與觀測）
- [ ] 變更摘要、風險、驗證方式已寫入 PR 描述
- [ ] **強制文檔更新檢查清單**：
  - [ ] 檢查是否影響使用者介面（CLI 參數、環境變數、啟動流程）
  - [ ] 檢查是否影響架構或系統行為（啟動序列、錯誤處理、關機流程）
  - [ ] 檢查是否需要更新：`spec.md`、根 `README.md`、子專案 `README.md`、本檔
  - [ ] 驗證所有文檔中的範例指令和配置是否仍然有效
  - [ ] 確認新功能在快速開始指南中有適當說明

### 變更要求（PR 模板建議）
1. 說明對 SSOT 的變更（proto/schema/samples）與影響面
2. 附端到端驗證步驟（Logs/Traces/Profiles 與健康/關機）
3. **文檔更新優先級**：
   - **P0（必須）**：影響使用者操作的變更（CLI、環境變數、啟動指令）
   - **P1（重要）**：架構或行為變更（錯誤處理、關機流程、性能優化）
   - **P2（建議）**：內部實作細節（代碼結構、函數重構）

---

## 協作最佳實務

### 溝通原則
1. **主動詢問**：不確定是否需要更新文檔時，主動詢問
2. **詳細說明**：清楚說明變更的影響範圍
3. **提供建議**：基於變更類型提供文檔更新建議

### 品質保證
1. **一致性檢查**：確保所有文檔間的一致性
2. **範例驗證**：驗證文檔中的所有範例和指令
3. **連結檢查**：確認文檔間的連結正確

### 持續改進
1. **經驗總結**：記錄協作過程中的問題和改進
2. **流程優化**：定期檢討和優化協作流程
3. **知識分享**：將最佳實務更新到指南中

---

## 重要注意事項

### 開發禁忌
1. **不要違反 Agent/Tool 職責分離**：Agent 負責決策，Tool 負責執行
2. **不要在 Python 端直接查詢外部系統**：應通過 Go 端 RemoteTool
3. **不要忽略錯誤處理**：必須有完善的異常處理和降級機制
4. **不要跳過測試直接提交**：測試覆蓋率必須達到 90% 以上
5. **不要修改已定義的契約**：除非經過團隊討論和 SSOT 更新

### 問題解決指引
如果遇到以下情況，請先查閱對應文檔：
- **不確定功能歸屬**：查看 [`docs/sre-services-map.md`](./docs/sre-services-map.md) 了解 Agent vs Tool 職責
- **技術實現細節**：查看 [`spec.md`](./spec.md) 了解完整技術規格
- **開發規範問題**：查看本文檔的相關章節
- **配置參考**：查看 [`contracts/samples/config.yaml`](./contracts/samples/config.yaml)

### 期望成果
在協作開發中，AI 開發者應該：
1. **產出高品質程式碼**：可讀性高、邏輯清晰、性能優化
2. **遵守架構原則**：嚴格按照混合架構和職責分離設計
3. **確保交付時程**：MVP 能在 8 週內順利交付
4. **預留擴展空間**：為未來 Phase 1/2 功能預留良好介面
5. **同步更新文檔**：特別是 P0 級別的使用者相關文檔
6. **嚴格執行檢查清單**：100% 完成對應模組 llm.txt 的所有檢查項目

### 模組 llm.txt 落實監督機制

#### 自動化檢查建議
為確保各模組 llm.txt 被有效執行，建議實施以下機制：

**AI 開發者自檢機制**：
```markdown
在每次提交前，AI 必須在提交訊息中包含：
- 使用的模組檢查清單：[contracts/go-platform/python-adk-runtime]
- 檢查清單完成度：X/Y 項已完成
- 跨模組影響：是否涉及多個模組整合
```

**團隊審查機制**：
```markdown
在代碼審查時，審查者應驗證：
- 提交者是否完成了對應模組的 llm.txt 檢查清單
- 是否有跨模組的一致性問題
- 文檔更新是否與模組檢查清單要求一致
```

#### 模組 llm.txt 有效性指標
**定期評估以下指標以確保 llm.txt 發揮最大作用**：
1. **遵循率**：實際執行檢查清單的比例
2. **缺陷減少率**：遵循檢查清單後的 bug 減少程度  
3. **交付品質**：符合模組規範的程式碼品質評分
4. **文檔一致性**：各層級文檔的同步更新成功率
5. **整合成功率**：跨模組功能的一次性整合成功比例

---

## 相關文檔

### 核心規範
- [`spec.md`](./spec.md) - 平台技術規格（完整架構定義）
- [`docs/sre-services-map.md`](./docs/sre-services-map.md) - SRE 三階段架構設計

### 組件文檔
- [`contracts/README.md`](./contracts/README.md) - SSOT 契約管理
- [`go-platform/README.md`](./go-platform/README.md) - Go 平台核心
- [`python-adk-runtime/README.md`](./python-adk-runtime/README.md) - Python ADK 運行時

### 配置參考
- [`grafana-alloy/config.alloy`](./grafana-alloy/config.alloy) - 可觀測性配置
- [`contracts/samples/config.yaml`](./contracts/samples/config.yaml) - 配置樣本