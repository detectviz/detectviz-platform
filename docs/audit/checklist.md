


# AI Scaffold Code Review Checklist

本檔案由 `todo.md` 中未完成任務自動產出，供 AI Agent 在 scaffold 或重構過程中進行自我檢查與修正回饋使用。請搭配 `/docs/ai_scaffold/scaffold_prerequisites.md` 一併使用。

---

## 通用檢查項目（所有 scaffold 任務適用）

- [x] 是否已更新 `/todo.md` 並標記為已完成？
- [x] 是否已補上 plugin 對應的 JSON Schema 至 `/schemas/plugins/`？ (password_hasher.json)
- [x] 是否已撰寫 `/docs/plugins/plugin-xxx.md` 說明文件，包含用途與範例？ (plugin-password-hasher.md)
- [x] 是否於 plugin 或 interface 補上 AI_PLUGIN_TYPE、AI_IMPL_PACKAGE 等 scaffold 標註？
- [x] 是否已在 `/docs/architecture/interface_spec.md` 補上新增 interface？ (PasswordHasher)
- [x] 是否補上至少一組 `plugin_test.go`，測試 `Init`, `Start`, 錯誤流程？ (hasher_bcrypt_test.go)
- [x] 是否已修復運行時錯誤，確保應用程式可正常啟動？ (修復 nil pointer 問題)

---

## Plugin Scaffold 類型檢查

### ImporterPlugin
- [ ] 是否實作 `ImportData()`，並支援 config 傳入來源欄位？
- [ ] 是否含資料批次處理流程？
- [ ] 是否對應 plugin schema 中所有欄位正確解析？
- [ ] 是否處理錯誤列、空值等特殊輸入情境？

### DetectorPlugin
- [ ] 是否實作 `RunDetection()` 並回傳 `AnalysisResult`？
- [ ] 是否處理輸入資料解析、條件比對、觸發紀錄？
- [ ] 是否支援由 config 傳入門檻與欄位 mapping？
- [ ] 是否紀錄觸發次數 / 分數 / 類型？

### UIPagePlugin
- [ ] 是否實作 `RegisterRoutes()` 並綁定 `/ui/xxx` endpoint？
- [ ] 是否提供 title、description 等 metadata？
- [ ] 是否支援 iframe URL 動態設定？

### CLIPlugin
- [ ] 是否有 `RegisterCLICmds()` 並註冊至少一個 command？
- [ ] 是否正確列印輸出與支援 CLI flags？
- [ ] 是否提供 scaffold 指令說明文字？

---

## 安全與結構檢查

- [x] `User.Password` 是否已重構為 `PasswordHash`？ (pkg/domain/entities/user.go)
- [x] 是否導入 `PasswordHasher` interface 並實作 bcrypt？ (internal/auth/hasher/)
- [ ] 是否使用 EmailVO / IDVO 等 Value Object 封裝基本欄位？
- [x] 是否已將明文 log/output 濾除？ (PasswordHash 有 json:"-" 標籤)
- [ ] 是否已補上 AuthProvider/SessionStore interface 定義與 mock？

---

## RAG 與知識庫可用性檢查

- [x] 是否於 plugin/interface 補上 AI scaffold 標籤？ (PasswordHasher interface)
- [x] 是否對應 plugin 建立了 JSON schema（含範例）？ (schemas/plugins/password_hasher.json)
- [x] 是否建立 plugin 對應的 scaffold doc（可讀性良好）？ (docs/plugins/plugin-password-hasher.md)
- [ ] 是否於 `/docs/ai_scaffold/rag_ingest_plan.md` 登記此 plugin 為知識來源？

---

## LLM Plugin 擴展性檢查（若為 AI Plugin）

- [ ] 是否支援 context 向量查詢？
- [ ] 是否處理 prompt struct 並可動態替換？
- [ ] 是否將分析記錄餵入向量資料庫？
- [ ] 是否支援查詢出處與回推回答來源？

---

## 文件同步與一致性

- [ ] `/README.md` 是否已更新 plugin 清單或 scaffold 狀態？
- [x] `/todo.md` 是否標記已完成？是否刪除重複？ (已標記完成 8 個任務)
- [x] 是否在 `interface_spec.md` 中補上完整定義與註解？ (PasswordHasher interface)
- [x] 是否同步更新 plugin scaffold metadata？ (進度統計已更新)
