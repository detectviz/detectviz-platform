# 測試指導原則 - SSOT

> 📌 **文檔職責**：本文檔定義整個平台的測試策略、標準和執行流程，是所有測試工作的權威指南。

## 🎯 測試策略概覽

### 測試金字塔
```
        /\
       /E2E\     端到端測試 (少量，關鍵流程)
      /______\
     /        \
    / 整合測試  \   跨模組、跨語言通訊驗證
   /____________\
  /              \
 /    單元測試    \  大量，快速反饋
/________________\
```

### 測試覆蓋率要求
- **單元測試**：≥ 90%
- **整合測試**：≥ 80% 關鍵路徑覆蓋
- **端到端測試**：100% 主要業務流程覆蓋

## 📋 測試層級詳解

### 1. 單元測試 (Unit Tests)
**目標**：驗證模組內部邏輯的正確性

#### Go 平台單元測試
```go
// go-platform/internal/pluginhost/observability_test.go
func TestStartSpan(t *testing.T) {
    tracer := NewTracer("test-service")
    
    span := tracer.StartSpan("test-operation")
    assert.NotNil(t, span)
    
    span.End()
}
```

#### Python ADK 單元測試
```python
# python-adk-runtime/tests/test_remote_tool.py
@pytest.mark.asyncio
async def test_remote_tool_invoke():
    tool = RemoteTool("test_tool", "1.0.0")
    result = await tool.invoke({"test": "data"})
    assert result["status"] == "success"
```

### 2. 整合測試 (Integration Tests)
**目標**：驗證模組間通訊和跨語言橋接

#### gRPC 通訊整合測試
```python
# tests/integration/test_grpc_bridge.py
@pytest.mark.integration
async def test_go_python_communication():
    # 啟動 Go ToolBridge
    go_bridge = start_test_toolbridge()
    
    # Python RemoteTool 調用
    tool = RemoteTool("capability.gateway.http_request", "1.0.0")
    result = await tool.invoke({"url": "https://example.com"})
    
    assert result["ok"] is True
    go_bridge.stop()
```

#### Contract 驗證測試
```python
# tests/integration/test_contracts.py
def test_proto_message_compatibility():
    # 驗證 Go 和 Python 生成的 proto 訊息相容性
    go_request = create_go_invoke_request()
    python_request = create_python_invoke_request()
    
    assert proto_compatible(go_request, python_request)
```

### 3. 端到端測試 (E2E Tests)
**目標**：驗證完整業務流程和用戶場景

#### Agent 工作流測試
```python
# tests/e2e/test_postmortem_workflow.py
@pytest.mark.e2e
async def test_complete_postmortem_workflow():
    # 完整的事後複盤流程測試
    orchestrator = get_postmortem_orchestrator()
    
    result = await orchestrator.run(
        "請為事件 INC-2025-001 進行事後複盤，受影響服務為 api-gateway，"
        "時間範圍是 2025-08-16 14:00:00 到 2025-08-16 16:00:00"
    )
    
    # 驗證完整流程
    assert "事後檢討報告" in result
    assert "根本原因分析" in result
    assert "改善建議" in result
```

## 🛠️ 測試工具和框架

### Go 測試工具
```go
// 使用標準庫和關鍵工具
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "go.uber.org/goleak"  // goroutine 洩漏檢測
)
```

### Python 測試工具
```python
# requirements-test.txt
pytest>=7.0.0
pytest-asyncio>=0.21.0
pytest-mock>=3.10.0
coverage>=7.0.0
```

### Mock 和 Stub 策略
```python
# 使用 pytest-mock 進行模擬
@pytest.fixture
def mock_go_toolbridge(mocker):
    mock_tool = mocker.patch('detectviz_adk.tools.RemoteTool')
    mock_tool.invoke.return_value = {"status": "success", "data": "mock"}
    return mock_tool
```

## 📊 測試執行和 CI/CD

### 本地測試執行
```bash
# Go 測試
cd go-platform
go test ./... -v -race -cover

# Python 測試
cd python-adk-runtime
pytest tests/ -v --cov=src --cov-report=html

# 整合測試
pytest tests/integration/ -v --integration

# 端到端測試
pytest tests/e2e/ -v --e2e
```

### CI/CD Pipeline 測試階段
```yaml
# .github/workflows/test.yml
stages:
  - unit-tests:
      parallel:
        - go-unit-tests
        - python-unit-tests
  
  - integration-tests:
      depends_on: [unit-tests]
      script: pytest tests/integration/
  
  - e2e-tests:
      depends_on: [integration-tests]
      script: pytest tests/e2e/
```

## 🎯 測試品質標準

### 測試命名約定
```python
# 測試函數命名格式
def test_[功能]_[條件]_[預期結果]():
    pass

# 範例
def test_remote_tool_invoke_with_valid_params_returns_success():
    pass

def test_agent_coordination_when_sub_agent_fails_handles_gracefully():
    pass
```

### 測試結構模式
```python
# AAA 模式：Arrange, Act, Assert
@pytest.mark.asyncio
async def test_agent_processes_request():
    # Arrange - 準備測試數據
    agent = create_test_agent()
    input_data = "test request"
    
    # Act - 執行被測試的行為
    result = await agent.run(input_data)
    
    # Assert - 驗證結果
    assert result is not None
    assert "response" in result
```

## 🔍 測試驗證檢查清單

### 提交前測試檢查
- [ ] 所有單元測試通過 (`pytest tests/unit/`)
- [ ] 整合測試通過 (`pytest tests/integration/`)
- [ ] 測試覆蓋率達到要求 (`coverage report`)
- [ ] 無測試洩漏 (goroutine, memory)
- [ ] 測試執行時間合理 (< 5分鐘總計)

### 程式碼審查測試檢查
- [ ] 新功能有對應測試
- [ ] 錯誤情況有測試覆蓋
- [ ] Mock 使用恰當，不過度模擬
- [ ] 測試可讀性和可維護性良好
- [ ] 測試資料隔離，無副作用

### 性能和壓力測試
```python
# 性能基準測試
@pytest.mark.benchmark
def test_agent_response_time_benchmark(benchmark):
    agent = create_production_agent()
    
    result = benchmark(agent.run, "standard request")
    
    # 驗證響應時間 < 2秒
    assert benchmark.stats['mean'] < 2.0
```

## 🚨 常見測試問題和解決方案

### 1. 跨語言測試問題
**問題**：Go 和 Python 組件測試隔離困難
**解決**：使用 Docker Compose 建立隔離測試環境

### 2. 異步測試問題
**問題**：Python async/await 測試不穩定
**解決**：使用 pytest-asyncio 和適當的事件循環管理

### 3. Mock 過度使用
**問題**：過度 Mock 導致測試脫離實際
**解決**：優先使用 Fake 實現，僅在必要時使用 Mock

## 📚 測試文檔和報告

### 測試報告生成
```bash
# 生成詳細測試報告
pytest --html=reports/test_report.html --cov-report=html:reports/coverage

# 整合到 CI/CD
make test-with-reports
```

### 測試指標追蹤
- 測試通過率
- 測試覆蓋率變化
- 測試執行時間趨勢
- 缺陷發現率

---

**維護說明**：
- 更新頻率：測試策略變更時更新
- 維護責任：測試工程師 + 各模組負責人
- 引用方式：`{{ docs/development/TESTING_GUIDELINES.md#section }}`