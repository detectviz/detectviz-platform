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

---

## 專案結構優化操作指令（目錄調整用）

若要讓 Cursor 根據專案目錄結構自動優化架構，請提供 @tree 並使用以下 Prompt：

> 請針對 @tree 中的專案結構，依照 /docs/ENGINEERING_SPEC.md 與 Clean Architecture 原則，執行結構優化與重構。
> 
> 請執行下列動作：
> 1. 請檢查 internal/adapters/plugins 的目錄結構是否過於扁平，並依照插件類型重新分類（如 detectors/, importers/, uipages/），所有變更請 patch。
> 2. 將 dto/ 移出 internal，集中至 pkg/interfaces/dto/
> 3. 合併 internal/app/initializer 與 internal/config → internal/bootstrap/
> 4. 調整所有 import path，確保可編譯
> 5. 補上對應的 /docs/ARCHITECTURE.md 架構圖與文字說明
> 6. 將 internal/adapters/plugins 細分類為 detectors/, importers/, uipages/ 等

可依需求加上：
> 所有變更請一次 patch，並同步修改受影響的檔案
> 如有任何 doc（如 PLUGIN_GUIDE.md、ARCHITECTURE.md、interface_spec.md）需因結構變動而更新，請一併修改。