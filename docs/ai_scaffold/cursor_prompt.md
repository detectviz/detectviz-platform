請讀取 @todo.md ，並依照出現順序逐條處理每個 - [ ] 項目。請依照 @scaffold_prerequisites.md  中規則執行：

- 依照 <!-- SCAFFOLD_TYPE --> 與 <!-- TARGET --> 判斷該如何 scaffold
- 每完成一項任務請：
  - 勾選 /todo.md 的項目為 [x]
  - 更新 @interface_spec.md （若有新增 interface）
  - 更新 @ROADMAP.md （若有達成里程碑）
  - 新增 plugin schema + plugin doc（若是 plugin 類型）
  - 補上測試檔 plugin_test.go（若是 plugin）

---

Checklist 最終產出要求

- 每次 Scaffold 任務執行後，請根據已完成的 @todo.md  項目，自動更新 @checklist.md 。
- checklist.md 為平台 Scaffold 驗證與品質追蹤標準的最終產出，必須與 todo.md 保持同步。
- 所有完成項目應在 checklist 中標示 [x] 或附上完成證據（檔案路徑、function 名稱等）。