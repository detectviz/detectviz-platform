# AI協作核心規則 - SSOT

> 📌 **文檔職責**：本文檔定義AI協作的核心規則和原則，是所有AI開發工作的指導準則。

## 🎯 Agent vs Tool 黃金準則

### 核心職責分離
```
Agent (決策層)               Tool (執行層)
─────────────               ─────────────
WHY - 為什麼做              HOW - 如何做
WHAT - 做什麼              WHERE - 在哪做  
WHEN - 何時做              WITH - 用什麼做
```

### 實現原則
1. **Agent 專注決策**
   - 業務邏輯制定
   - 工作流編排
   - 知識整合
   - 策略規劃

2. **Tool 專注執行**
   - 數據查詢
   - API調用
   - 文件操作
   - 外部系統集成

3. **嚴格邊界**
   - Agent 不直接操作數據
   - Tool 不包含業務邏輯
   - 所有數據操作必須通過Tool
   - 所有決策邏輯必須在Agent

## 📋 強制執行的開發檢查

### 開發前檢查清單
- [ ] 明確模組職責邊界
- [ ] 確認Agent vs Tool分離設計
- [ ] 檢查contracts/最新狀態
- [ ] 閱讀相關模組的llm.txt

### 開發過程中檢查
- [ ] Agent不直接操作外部系統
- [ ] Tool保持無狀態和冪等性
- [ ] 跨語言通訊使用contracts定義
- [ ] 遵循模組專用開發規範

### 提交前檢查清單
- [ ] 100%通過相關測試
- [ ] 符合代碼品質標準
- [ ] 文檔已同步更新
- [ ] 無技術債務累積

## 🛠️ 技術實現規範

### SSOT契約優先
1. **contracts/是唯一事實來源**
   - 任何跨語言介面變更必須先更新contracts/
   - 禁止在下游專案手動修改生成碼
   - 使用buf工具管理proto生成

2. **配置管理統一**
   - 統一配置載入路徑和環境覆蓋
   - 使用JSON Schema驗證配置
   - 敏感信息通過環境變數管理

### 跨語言通訊標準
```python
# ✅ 正確：使用RemoteTool通過gRPC調用
remote_tool = RemoteTool("observability.health_aggregator", "0.1.0")
result = await remote_tool.invoke({"action": "query_health"})

# ❌ 錯誤：Agent直接操作外部系統  
influx_client = InfluxDBClient()  # 不應該在Agent中
```

### ADK標準遵循
```python
# ✅ 正確：使用ADK Agent團隊模式
postmortem_orchestrator = adk.Agent(
    name="postmortem_orchestrator",
    sub_agents=[data_collector, analyzer, report_writer]
)

# ❌ 錯誤：不使用ADK標準
class CustomAgent:  # 應該使用adk.Agent
    pass
```

## 🔍 品質保證機制

### 自動化驗證
```bash
# contracts驗證
make validate-with-versions

# 模組卡驗證  
make validate-cards

# proto健康檢查
make health-check-proto
```

### 技術債務預防
- 定期運行健康檢查工具
- 使用自動化修復工具
- 遵循預防性檢查清單
- 維護代碼品質指標

## 🚨 禁止的操作

### 嚴格禁止
1. **違反職責分離**
   - Agent直接查詢數據庫
   - Agent直接調用外部API
   - Tool包含業務決策邏輯

2. **繞過SSOT**
   - 修改contracts/gen/生成碼
   - 不通過contracts更新跨語言介面
   - 硬編碼配置而非使用統一載入

3. **破壞架構標準**
   - 不使用ADK標準的Agent架構
   - 不通過RemoteTool進行跨語言調用
   - 跳過模組卡驗證

## 📚 AI協作最佳實踐

### 1. 理解優先級
```
sre-services-map.md (業務邏輯) → 
spec.md (技術規範) → 
AGENT.md (協作指南) → 
模組/llm.txt (具體執行)
```

### 2. 開發流程
1. 先理解業務需求 (sre-services-map.md)
2. 確認技術實現方案 (spec.md)  
3. 遵循協作規範 (AGENT.md)
4. 執行模組專用指南 (llm.txt)

### 3. 文檔維護
- 修改代碼時同步更新文檔
- 引用而非重複SSOT內容  
- 保持模組llm.txt與通用規範一致

## 🎯 AI開發者特別注意

### 關鍵提醒
1. **永遠先檢查SSOT**：任何疑問都先查閱相應的SSOT文檔
2. **嚴格遵循職責分離**：Agent和Tool的職責邊界不可模糊
3. **使用自動化工具**：優先使用提供的自動化工具而非手動操作
4. **保持架構一致性**：所有開發都要符合混合語言架構設計

### 錯誤處理指導
```python
# ✅ 正確的錯誤處理
try:
    result = await remote_tool.invoke(params)
    if not result["ok"]:
        # 記錄錯誤但讓Agent決定如何處理
        logger.error(f"Tool execution failed: {result['error']}")
        return {"status": "tool_error", "details": result}
except Exception as e:
    # 提供足夠信息讓Agent決策
    return {"status": "exception", "error": str(e)}

# ❌ 錯誤的錯誤處理  
try:
    result = await remote_tool.invoke(params)
except Exception:
    # 直接重試或修復 - 這是Agent的決策責任
    return self.fallback_processing()  
```

---

**維護說明**：
- 更新頻率：架構原則變更時更新
- 維護責任：架構師 + AI協作負責人
- 引用方式：`{{ docs/development/AI_COLLABORATION_RULES.md#section }}`