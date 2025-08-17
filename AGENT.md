# AGENT.md - Detectviz 平台開發貢獻指南

本文檔為 Detectviz 平台所有貢獻者（包含 AI 與人類協作者）的統一開發指南，旨在確保程式碼品質、文件同步與高效協作。

## 1. 核心開發原則與架構

### 1.1. 文檔層級關係
本專案所有開發工作遵循一個清晰的文檔層級，本文件是其中的核心協作規範：

```bash
sre-services-map.md (架構憲法 - 業務邏輯)
↓ 指導
spec.md (技術規格 - 實作細節) 
↓ 實現
AGENT.md (本文檔 - 開發守則與協作規範)
↓ 具體落實
各模組 llm.txt (模組專用開發檢查清單)
```

### 1.2. Agent vs. Tool 黃金準則
所有開發必須嚴格遵守「決策」與「執行」分離的黃金準則：
* **Agent (決策層)**：負責 **WHY** (為什麼), **WHAT** (做什麼), **WHEN** (何時做)。Agent 不直接操作數據，專注於業務邏輯和工作流編排。
* **Tool (執行層)**：負責 **HOW** (如何做), **WHERE** (在哪做), **WITH** (用什麼)。Tool 是無狀態、冪等的執行單元，負責數據操作和與外部系統集成。

### 1.3. 三大基本原則
1.  **SSOT 契約優先 (Contracts-First)**：任何跨語言介面、API 或組態結構的變更，**必須**優先在 `contracts/` 目錄下完成。
2.  **文件同步 (Docs-as-Code)**：任何影響使用者行為或系統架構的程式碼變更，都必須同步更新相關文件。
3.  **安全第一 (Security-First)**：嚴禁在版本控制中提交任何真實的密鑰或 Token。

## 2. llm.txt 強制執行機制

這是確保所有原則得以落實的核心機制。每個主要模組下都有一份 `llm.txt`，作為該模組的**專用開發檢查清單**。

* `contracts/llm.txt`
* `go-platform/llm.txt`
* `python-adk-runtime/llm.txt`

**所有貢獻者必須遵循以下三級檢查制度**：

#### 第一級：開發前強制檢查
1.  **識別模組**：明確本次開發涉及的所有模組。
2.  **熟讀檢查清單**：完整閱讀所有相關模組的 `llm.txt` 文件。
3.  **制定執行計畫**：明確哪些檢查項目適用於本次開發。

#### 第二級：開發過程中強制檢查
* **里程碑檢查**：每完成一個功能點，回顧並檢查是否符合 `llm.txt` 的要求。
* **規範遵循**：實時確認沒有違反模組的「必守規範」。

#### 第三級：提交前強制檢查
* **100% 完成度**：確保所有相關模組 `llm.txt` 的「提交前檢查清單」已 100% 完成。
* **自我報告**：在 PR 描述中，必須附上 `llm.txt` 的自我檢查報告，說明完成情況。

## 3. 開發與貢獻工作流程

### 3.1. 變更前規劃
- [ ] 識別變更類型 (SSOT 契約, 核心邏輯, 介面變更等)。
- [ ] 評估影響範圍 (UI, 系統行為, 文件, 範例)。
- [ ] 建立包含「文件更新」的 TODO 清單。

### 3.2. 實作與驗證
- [ ] 遵循 SSOT 原則，契約變更先行，並執行 `cd contracts && make gen`。
- [ ] 嚴格遵循對應模組 `llm.txt` 的所有規範。
- [ ] 編寫或更新單元測試和整合測試，確保測試覆蓋率 > 90%。

### 3.3. 文件同步更新
根據變更的影響，更新相關文件：
-   **P0 (必須更新)**：影響 CLI、環境變數、啟動流程、`config.yaml` 結構的變更。
-   **P1 (重要更新)**：影響系統架構、核心行為或錯誤處理流程的變更。
-   **P2 (建議更新)**：內部程式碼結構的重大重構。

### 3.4. 提交 Pull Request
- [ ] 在 PR 描述中清楚說明變更的摘要、風險與驗證方式。
- [ ] **附上 `llm.txt` 的自我檢查報告**。
- [ ] 確保所有 CI/CD 流程通過。

## 4. MVP 開發聚焦 (Phase 3: 事後複盤)

當前所有開發工作應聚焦於 MVP 範圍，避免添加非 Phase 3 的功能。
* **核心組件**：`postmortem_orchestrator`, `HealthAggregator`, `ReportGenerator`, `ResponseHistoryStore`。
* **目標**：實現一個可運行的、從數據查詢到報告生成的事後複盤流程。
* 詳細的 MVP 任務請參考 `TODO.md`。

---

## 5. 相關文件快速連結

* **架構**
    * [**`sre-services-map.md`**](sre-services-map.md): 架構憲法
    * [**`spec.md`**](spec.md): 平台技術規格
* **開發**
    * [**`TODO.md`**](TODO.md): 當前 MVP 任務指令
    * [**`docs/quick-reference.md`**](docs/quick-reference.md): 快速參考手冊
* **模組**
    * [**`contracts/README.md`**](contracts/README.md): SSOT 契約管理
    * [**`go-platform/README.md`**](go-platform/README.md): Go 平台開發指南
    * [**`python-adk-runtime/README.md`**](python-adk-runtime/README.md): Python ADK 開發指南