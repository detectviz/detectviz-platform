# Detectviz 平台技術規格

> **文檔職責**：定義技術實作規格、API 文檔、配置範例和實作指南，為開發者提供具體的技術實現方案

## 文檔定位

- **目標受眾**：開發者、實作者、系統整合工程師
- **更新頻率**：每週更新，隨功能開發同步
- **版本**：1.0.0
- **最後更新**：2025-08-18

## 文檔關係

```bash
README.md → AGENT.md → ARCHITECTURE.md → [SPEC.md] → TASKS.md
```

**閱讀路徑**：
- **前置閱讀**：[ARCHITECTURE.md - 系統架構設計](ARCHITECTURE.md#系統架構設計)
- **後續閱讀**：[TASKS.md - 開發任務進度](TASKS.md#實作任務清單)
- **相關參考**：[AGENT.md - AI協作指南](AGENT.md#ai協作原則)

> **架構決策與設計原則**：參見 [ARCHITECTURE.md - 系統架構設計](ARCHITECTURE.md#系統架構設計)

## 技術棧與依賴

> **完整技術棧概覽**：參見 [README.md - 技術棧](README.md#技術棧)  
> **系統需求詳情**：參見 [README.md - 系統需求](README.md#系統需求)

### 技術實作重點

本文檔專注於技術實作層面的詳細規格，包括：
- **API 規格與介面定義**：gRPC 服務、REST API、Protocol Buffers 結構
- **目錄結構與配置**：專案組織、配置檔案、部署腳本
- **實作指南與範例**：程式碼範例、最佳實踐、故障排除

> **術語索引**：參見 [完整術語對照表](.kiro/specs/documentation-normalization/terminology-index.md)

* * *

## 專案目錄結構

### 契約與配置目錄（contracts/）

```bash
contracts/                              
├── proto/detectviz/contracts/v1/        # Protocol Buffers 定義
│   ├── adk_bridge.proto                # ToolBridge / Health APIs
│   └── postmortem.proto                # MVP: 事後複盤服務定義
├── schemas/                            # JSON Schema 驗證規範
│   ├── config.schema.json              # 平台組態規範
│   ├── module.card.schema.json         # 模組卡規範
│   ├── plugin.schema.json              # 插件規範  
│   └── postmortem-request.schema.json  # MVP: 事後複盤請求規範
├── gen/                                # 自動生成的程式碼
│   ├── go/detectviz/contracts/v1/      # Go 生成碼
│   ├── python/detectviz/contracts/v1/  # Python 生成碼
│   └── metadata/version.json           # 版本元數據
├── samples/                            # 配置範例檔案
│   ├── .env.template                   # 環境變數範例
│   ├── config.yaml                     # 完整配置範例
│   ├── module.card.json                # 模組卡範例
│   └── plugin.yaml                     # 插件配置範例
├── tools/                              # 驗證與開發工具
│   ├── validate_config.py              # 配置驗證工具
│   ├── validate_module_card.py         # 模組卡驗證工具
│   └── generate_module_card.py         # 模組卡生成工具
├── buf.yaml                            # Buf 工作區配置
├── buf.gen.yaml                        # 程式碼生成配置
└── Makefile                            # 建置與版本控制
```

### 支援目錄結構

```bash
├── docs/                           # 技術文檔目錄
│   ├── development/                # 開發指南
│   │   ├── AGENT_DEVELOPMENT_GUIDE.md
│   │   ├── TOOL_DEVELOPMENT_GUIDE.md
│   │   └── TESTING_GUIDELINES.md
│   ├── guides/                     # 使用指南
│   │   ├── DEVELOPMENT_SETUP.md
│   │   └── TESTING_GUIDE.md
│   └── reference/                  # 參考文檔
│       └── QUICK_REFERENCE.md
├── .github/workflows/              # CI/CD 自動化
│   ├── contracts-validation.yml    # 契約驗證流程
│   ├── go-tests.yml                # Go 測試流程
│   └── python-tests.yml            # Python 測試流程
├── depoly/alloy/                   # 監控配置
│   └── config.alloy                # Alloy 收集器配置
└── assets/                         # 靜態資源
    ├── diagrams/                   # 架構圖表
    └── screenshots/                # 介面截圖
```

### Go 平台核心（go-platform/）

```bash
go-platform/                           # Go 高性能執行平台
├── cmd/detectviz/                     # 主程式入口
│   └── main.go                        # CLI 應用程式
├── tools/                             # 開發工具
│   └── scaffold.go                    # 插件腳手架生成器
├── configs/                           # 配置檔案
│   ├── config.yaml                    # 預設配置
│   └── config-dev.yaml               # 開發環境配置
└── internal/                          # 內部套件
    ├── configx/                       # 配置管理
    │   ├── loader.go                  # 配置載入器
    │   └── validator.go               # 配置驗證器
    ├── contracts/                     # 契約驗證
    │   └── version_check.go           # 版本一致性檢查
    ├── health/                        # 健康檢查服務
    │   └── server.go                  # 健康檢查 HTTP 服務
    ├── metrics/                       # 指標管理
    │   ├── factory.go                 # 指標工廠
    │   └── provider.go                # 指標提供者介面
    ├── observability/                 # 可觀測性
    │   ├── logging.go                 # 結構化日誌
    │   └── otel_init.go              # OpenTelemetry 初始化
    └── pluginhost/                    # 插件託管系統
        ├── bridge_server.go           # gRPC ToolBridge 服務
        ├── interceptors.go            # gRPC 攔截器
        ├── monitored_handler.go       # 監控包裝器
        ├── registry.go                # 插件註冊表
        ├── resource_monitor.go        # 資源監控器
        ├── runtime.go                 # 插件運行時
        ├── security.go                # 安全邊界控制
        └── plugins/                   # 插件實作
            ├── gateway/               # 閘道器插件
            │   └── http_request/      # HTTP 請求工具
            └── observability/         # 監控插件
                └── health_aggregator/ # 健康數據聚合器
```

### Python ADK 運行時（python-adk-runtime/）

```bash
python-adk-runtime/                    # Python AI Agent 運行時
├── README.md                          # 模組說明文檔
├── requirements.txt                   # Python 依賴清單
├── example_usage.py                   # 使用範例
├── web_server.py                      # 開發用 Web 服務
├── test_adk_integration.py           # 整合測試
├── test_simple_adk.py                # 單元測試
├── agents/                           # Agent 實作範例
│   └── postmortem/                   # MVP: 事後複盤 Agent
│       ├── __init__.py
│       └── agent.py                  # 複盤協調 Agent
└── src/detectviz_adk/               # 核心 ADK 套件
    ├── __init__.py
    ├── config/                       # 配置管理
    │   ├── __init__.py
    │   └── loader.py                 # 配置載入器
    ├── memory/                       # 記憶體與知識管理
    │   ├── __init__.py
    │   └── stores/                   # 存儲後端
    │       ├── __init__.py
    │       └── response_history_store.py
    ├── agents/                       # Agent 基礎類別
    │   ├── __init__.py
    │   └── postmortem/              # 事後複盤 Agent 團隊
    │       ├── __init__.py
    │       ├── orchestrator.py      # 協調器 Agent
    │       ├── data_collector.py    # 數據收集 Agent
    │       ├── analyzer.py          # 分析 Agent
    │       └── report_writer.py     # 報告生成 Agent
    ├── tools/                       # 工具抽象層
    │   ├── __init__.py
    │   ├── adk_tools.py            # ADK 工具包裝器
    │   ├── memory_tools.py         # 記憶體工具
    │   └── remote_tool.py          # 遠端工具客戶端
    ├── runners/                     # 執行器
    │   ├── __init__.py
    │   └── postmortem_runner.py    # 複盤執行器
    └── sessions/                    # 會話管理
        ├── __init__.py
        └── session_manager.py      # 會話管理器
```

* * *

## API 規格與介面定義

### gRPC 服務定義

#### HealthService API

```protobuf
// 健康檢查與系統資訊服務
service HealthService {
  // 基本健康檢查
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
  
  // 詳細健康狀態
  rpc GetStatus(StatusRequest) returns (StatusResponse);
  
  // 系統版本資訊
  rpc GetVersion(VersionRequest) returns (VersionResponse);
  
  // 可用能力列表
  rpc ListCapabilities(CapabilitiesRequest) returns (CapabilitiesResponse);
}

// 健康檢查請求
message HealthCheckRequest {
  string service = 1;  // 可選：特定服務名稱
}

// 健康檢查回應
message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
    SERVICE_UNKNOWN = 3;
  }
  ServingStatus status = 1;
  string message = 2;
  int64 timestamp = 3;
}
```

#### ToolBridgeService API

```protobuf
// 工具橋接服務 - 跨語言工具調用
service ToolBridgeService {
  // 執行工具（支援串流回應）
  rpc ExecuteTool(ToolRequest) returns (stream ToolChunk);
  
  // 批量執行工具
  rpc BatchExecute(BatchToolRequest) returns (stream BatchToolResponse);
  
  // 開啟會話
  rpc OpenSession(SessionRequest) returns (SessionResponse);
  
  // 關閉會話
  rpc CloseSession(CloseSessionRequest) returns (CloseSessionResponse);
  
  // 列出可用工具
  rpc ListTools(ListToolsRequest) returns (ListToolsResponse);
}

// 工具執行請求
message ToolRequest {
  string name = 1;           // 工具名稱
  string version = 2;        // 工具版本
  map<string, string> args = 3;  // 執行參數
  map<string, string> metadata = 4;  // 元數據
  string trace_id = 5;       // 追蹤 ID
  string session_id = 6;     // 會話 ID（可選）
  int32 timeout_seconds = 7; // 超時設定
}

// 工具執行回應塊
message ToolChunk {
  enum ChunkType {
    DATA = 0;      // 數據塊
    STATUS = 1;    // 狀態更新
    PROGRESS = 2;  // 進度更新
    LOG = 3;       // 日誌訊息
    ERROR = 4;     // 錯誤訊息
    COMPLETE = 5;  // 執行完成
  }
  
  ChunkType type = 1;
  bytes data = 2;
  string message = 3;
  int32 progress_percent = 4;
  string log_level = 5;
  int64 timestamp = 6;
}
```

#### PostmortemService API（MVP 專用）

```protobuf
// 事後複盤服務
service PostmortemService {
  // 啟動複盤分析
  rpc StartAnalysis(PostmortemRequest) returns (stream PostmortemResponse);
  
  // 獲取複盤報告
  rpc GetReport(GetReportRequest) returns (PostmortemReport);
  
  // 列出歷史複盤
  rpc ListPostmortems(ListPostmortemsRequest) returns (ListPostmortemsResponse);
}

// 複盤請求
message PostmortemRequest {
  string incident_id = 1;        // 事件 ID
  int64 start_time = 2;         // 開始時間（Unix 時間戳）
  int64 end_time = 3;           // 結束時間
  repeated string services = 4;  // 相關服務列表
  string severity = 5;          // 嚴重程度
  map<string, string> metadata = 6;  // 額外元數據
}

// 複盤回應
message PostmortemResponse {
  enum Stage {
    COLLECTING = 0;  // 數據收集中
    ANALYZING = 1;   // 分析中
    REPORTING = 2;   // 生成報告中
    COMPLETED = 3;   // 完成
    FAILED = 4;      // 失敗
  }
  
  Stage stage = 1;
  string message = 2;
  int32 progress_percent = 3;
  string report_url = 4;  // 報告 URL（完成時）
}
```

### HTTP REST API

#### 健康檢查端點

```http
GET /health
GET /health/ready
GET /health/live
GET /version
```

**回應範例**：
```json
{
  "status": "healthy",
  "timestamp": "2025-08-17T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "components": {
    "database": "healthy",
    "cache": "healthy",
    "external_api": "degraded"
  }
}
```

#### 工具管理端點

```http
GET /api/v1/tools                    # 列出所有工具
GET /api/v1/tools/{name}            # 獲取特定工具資訊
POST /api/v1/tools/{name}/execute   # 執行工具
GET /api/v1/sessions                # 列出活動會話
POST /api/v1/sessions               # 創建新會話
DELETE /api/v1/sessions/{id}        # 關閉會話
```

### 程式碼生成配置

#### buf.gen.yaml 配置

```yaml
version: v1
managed:
  enabled: true
  go_package_prefix:
    default: github.com/detectviz/platform/contracts/gen/go
plugins:
  # Go 程式碼生成
  - plugin: buf.build/protocolbuffers/go
    out: gen/go
    opt:
      - paths=source_relative
      - module=github.com/detectviz/platform/contracts
  
  # Go gRPC 程式碼生成
  - plugin: buf.build/grpc/go
    out: gen/go
    opt:
      - paths=source_relative
      - require_unimplemented_servers=false
  
  # Python 程式碼生成
  - plugin: buf.build/protocolbuffers/python
    out: gen/python
    opt:
      - pyi_out=gen/python
  
  # Python gRPC 程式碼生成
  - plugin: buf.build/grpc/python
    out: gen/python
```

#### 版本管理

生成的程式碼包含版本資訊：
- **Go 套件路徑**：`github.com/detectviz/platform/contracts/gen/go/detectviz/contracts/v1`
- **Python 套件路徑**：`detectviz.contracts.v1`
- **版本元數據**：`contracts/gen/metadata/version.json`

```json
{
  "version": "1.0.0",
  "generated_at": "2025-08-17T10:30:00Z",
  "buf_version": "1.28.1",
  "proto_files": [
    "detectviz/contracts/v1/adk_bridge.proto",
    "detectviz/contracts/v1/postmortem.proto"
  ],
  "checksums": {
    "adk_bridge.proto": "sha256:abc123...",
    "postmortem.proto": "sha256:def456..."
  }
}
```

## 配置管理與模組規範

### 模組卡規範（module.card.json）

每個模組都必須包含一個 `module.card.json` 檔案，定義模組的基本資訊和依賴關係。

#### 模組卡結構

```json
{
  "$schema": "https://detectviz.dev/schemas/module.card.schema.json",
  "name": "observability.health_aggregator",
  "version": "1.0.0",
  "description": "健康數據聚合器插件",
  "language": "go",
  "entrypoint": "plugin.go",
  
  "classification": {
    "role": "plugin.gateway",
    "category": "observability",
    "subcategory": "metrics"
  },
  
  "dependencies": {
    "requires": [
      {
        "name": "detectviz.contracts",
        "version": ">=1.0.0",
        "type": "proto"
      }
    ],
    "contracts": {
      "min_proto_version": "1.0.0"
    }
  },
  
  "runtime": {
    "resources": {
      "min_memory_mb": 64,
      "min_cpu_cores": 0.1,
      "max_memory_mb": 512,
      "max_cpu_cores": 1.0
    },
    "config_schema": "schemas/health_aggregator.schema.json",
    "timeout_seconds": 30
  },
  
  "observability": {
    "metrics_enabled": true,
    "tracing_enabled": true,
    "logging_level": "info",
    "tags": {
      "component": "health_aggregator",
      "layer": "plugin"
    }
  },
  
  "security": {
    "permissions": [
      "metrics:read",
      "network:http_client"
    ],
    "sandbox": true
  }
}
```

#### 角色分類定義

| 角色 | 描述 | 範例 |
|------|------|------|
| `agent.coordinator` | 協調型 Agent，負責決策和任務分派 | postmortem_orchestrator |
| `agent.specialist` | 專家型 Agent，負責特定領域分析 | data_collector, analyzer |
| `tool` | 工具，提供具體功能實作 | http_client, database_query |
| `plugin.gateway` | 閘道器插件，處理外部整合 | prometheus_client, grafana_api |
| `capability` | 能力模組，提供可重用的功能 | nlp_processor, chart_generator |
| `memory.backend` | 記憶體後端，提供存儲能力 | redis_store, file_store |
| `observability.module` | 可觀測性模組 | metrics_collector, tracer |

### 平台配置規範（config.yaml）

#### 完整配置範例

```yaml
# 基本環境配置
env: "development"  # development, staging, production

# gRPC 服務配置
grpc:
  listen: ":5002"
  max_recv_bytes: 4194304    # 4MB
  max_send_bytes: 4194304    # 4MB
  timeout_seconds: 30
  
  # TLS 配置（可選）
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
    ca_file: ""

# HTTP 服務配置
http:
  enabled: true
  listen: ":8080"
  cors:
    enabled: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]

# 可觀測性配置
observability:
  # 模式選擇：lgtm_local, grafana_cloud, gcp
  mode: "lgtm_local"
  
  # OpenTelemetry 配置
  otlp:
    protocol: "grpc"          # grpc 或 http
    endpoint: "127.0.0.1:4317"
    insecure: true
    headers: {}
    
  # 日誌配置
  logs:
    mode: "file"              # file, stdout, off
    level: "info"             # debug, info, warn, error
    file:
      path: "./var/log/detectviz/detectviz.log"
      max_size_mb: 100
      max_backups: 5
      max_age_days: 30
      compress: true
      
  # 效能分析配置
  profiling:
    enabled: true
    pprof_address: "127.0.0.1:6060"
    application_name: "detectviz-platform"
    tags:
      service.name: "detectviz-platform"
      deployment.environment: "development"

# 插件系統配置
plugin:
  paths:
    - "./go-platform/internal/pluginhost/plugins"
  registry: "file"            # file, database, remote
  
  # 插件安全配置
  security:
    sandbox_enabled: true
    allowed_networks: ["127.0.0.1", "localhost"]
    max_execution_time: 300
    max_memory_mb: 1024

# 記憶體系統配置
memory:
  backend: "inmem"            # inmem, redis, postgresql
  default_ttl_seconds: 3600
  
  # Redis 配置（當 backend = redis）
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10
    
  # 檔案存儲配置（當 backend = file）
  file:
    path: "./data/memory.db"
    sync_interval: 60

# MVP 專用：事後複盤配置
postmortem:
  enabled: true
  max_analysis_days: 7
  report_formats: ["markdown", "json", "html"]
  
  # 數據源配置
  data_sources:
    prometheus:
      enabled: true
      url: "${PROMETHEUS_URL:-http://localhost:9090}"
      timeout_seconds: 30
      
    grafana:
      enabled: true
      url: "${GRAFANA_URL:-http://localhost:3000}"
      api_key: "${GRAFANA_API_KEY}"
      timeout_seconds: 30
      
  # 報告配置
  reports:
    template_dir: "./templates/postmortem"
    output_dir: "./reports"
    include_charts: true
    auto_archive: true
    retention_days: 365

# 外部服務整合
integrations:
  slack:
    enabled: false
    webhook_url: "${SLACK_WEBHOOK_URL}"
    
  email:
    enabled: false
    smtp_host: "${SMTP_HOST}"
    smtp_port: 587
    username: "${SMTP_USERNAME}"
    password: "${SMTP_PASSWORD}"
```

#### 環境變數範例（.env.template）

```bash
# 基本配置
DETECTVIZ_ENV=development
DETECTVIZ_LOG_LEVEL=info

# 數據源配置
PROMETHEUS_URL=http://localhost:9090
GRAFANA_URL=http://localhost:3000
GRAFANA_API_KEY=your_grafana_api_key_here

# Grafana Cloud 配置（如果使用）
GF_CLOUD_OTEL_ID=your_otel_id
GF_CLOUD_PROFILES_ID=your_profiles_id
GCLOUD_RW_API_KEY=your_api_key

# 通知配置
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
SMTP_HOST=smtp.gmail.com
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password

# 安全配置
JWT_SECRET=your_jwt_secret_here
ENCRYPTION_KEY=your_32_char_encryption_key_here
```

### 配置驗證工具

#### 配置驗證指令

```bash
# 驗證主配置檔案
detectviz config validate -f config.yaml

# 驗證模組卡
detectviz module validate -f module.card.json

# 驗證所有配置
detectviz validate --all
```

#### 驗證工具實作範例

```python
#!/usr/bin/env python3
"""配置驗證工具"""

import json
import yaml
import jsonschema
from pathlib import Path

def validate_config(config_path: str, schema_path: str) -> bool:
    """驗證配置檔案"""
    try:
        # 載入配置
        with open(config_path, 'r', encoding='utf-8') as f:
            if config_path.endswith('.yaml') or config_path.endswith('.yml'):
                config = yaml.safe_load(f)
            else:
                config = json.load(f)
        
        # 載入 schema
        with open(schema_path, 'r', encoding='utf-8') as f:
            schema = json.load(f)
        
        # 執行驗證
        jsonschema.validate(config, schema)
        print(f"✅ {config_path} 驗證通過")
        return True
        
    except jsonschema.ValidationError as e:
        print(f"❌ {config_path} 驗證失敗: {e.message}")
        return False
    except Exception as e:
        print(f"❌ 驗證過程發生錯誤: {e}")
        return False

if __name__ == "__main__":
    import sys
    if len(sys.argv) != 3:
        print("使用方式: python validate_config.py <config_file> <schema_file>")
        sys.exit(1)
    
    success = validate_config(sys.argv[1], sys.argv[2])
    sys.exit(0 if success else 1)
```

* * *

## MVP 實作指南（事後複盤系統）

> **架構設計與決策原則**：參見 [ARCHITECTURE.md](ARCHITECTURE.md#sre-三階段生命週期設計)  
> **當前實作進度**：參見 [TASKS.md - MVP 核心交付目標](TASKS.md#mvp-核心交付目標)

### 核心組件實作

#### 1. 事後複盤協調器（Python ADK Agent）

**檔案位置**：`python-adk-runtime/src/detectviz_adk/agents/postmortem/orchestrator.py`

```python
"""事後複盤協調器 - ADK Root Agent 實作"""

from google import adk
from typing import Dict, Any, List
import asyncio
import logging

class PostmortemOrchestrator:
    """事後複盤協調器 - 負責整個複盤流程的決策和協調"""
    
    def __init__(self):
        self.logger = logging.getLogger(__name__)
        self.sub_agents = {
            'data_collector': DataCollectorAgent(),
            'analyzer': RootCauseAnalyzer(), 
            'report_writer': ReportWriterAgent()
        }
    
    async def execute_postmortem(self, incident_request: Dict[str, Any]) -> Dict[str, Any]:
        """執行完整的事後複盤流程"""
        try:
            # 階段 1: 數據收集決策
            collection_strategy = await self._decide_collection_strategy(incident_request)
            
            # 階段 2: 委派數據收集
            collected_data = await self.sub_agents['data_collector'].collect_data(
                collection_strategy
            )
            
            # 階段 3: 分析決策
            analysis_approach = await self._decide_analysis_approach(
                incident_request, collected_data
            )
            
            # 階段 4: 委派根因分析
            analysis_results = await self.sub_agents['analyzer'].analyze(
                collected_data, analysis_approach
            )
            
            # 階段 5: 報告決策
            report_format = await self._decide_report_format(
                incident_request, analysis_results
            )
            
            # 階段 6: 委派報告生成
            final_report = await self.sub_agents['report_writer'].generate_report(
                analysis_results, report_format
            )
            
            return {
                'status': 'completed',
                'report_url': final_report['url'],
                'summary': final_report['summary'],
                'recommendations': final_report['recommendations']
            }
            
        except Exception as e:
            self.logger.error(f"複盤執行失敗: {e}")
            return {
                'status': 'failed',
                'error': str(e)
            }
    
    async def _decide_collection_strategy(self, incident: Dict[str, Any]) -> Dict[str, Any]:
        """決策數據收集策略"""
        severity = incident.get('severity', 'medium')
        duration = incident.get('duration_minutes', 60)
        affected_services = incident.get('services', [])
        
        # 根據嚴重程度決定收集範圍
        if severity == 'critical':
            time_window = duration * 2  # 收集事件前後各一倍時間
            detail_level = 'high'
        elif severity == 'high':
            time_window = duration * 1.5
            detail_level = 'medium'
        else:
            time_window = duration
            detail_level = 'basic'
        
        return {
            'time_window_minutes': time_window,
            'detail_level': detail_level,
            'services': affected_services,
            'include_dependencies': severity in ['critical', 'high']
        }
```

#### 2. 健康數據聚合器（Go Plugin）

**檔案位置**：`go-platform/internal/pluginhost/plugins/observability/health_aggregator/plugin.go`

```go
package health_aggregator

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "go.uber.org/zap"
    
    pb "github.com/detectviz/platform/contracts/gen/go/detectviz/contracts/v1"
)

// HealthAggregatorPlugin 健康數據聚合器插件
type HealthAggregatorPlugin struct {
    promClient v1.API
    logger     *zap.Logger
    config     *Config
}

// Config 插件配置
type Config struct {
    PrometheusURL    string        `json:"prometheus_url"`
    QueryTimeout     time.Duration `json:"query_timeout"`
    MaxDataPoints    int           `json:"max_data_points"`
    DefaultTimeRange time.Duration `json:"default_time_range"`
}

// NewHealthAggregatorPlugin 創建新的健康聚合器插件
func NewHealthAggregatorPlugin(configData []byte) (*HealthAggregatorPlugin, error) {
    var config Config
    if err := json.Unmarshal(configData, &config); err != nil {
        return nil, fmt.Errorf("解析配置失敗: %w", err)
    }
    
    // 設定預設值
    if config.QueryTimeout == 0 {
        config.QueryTimeout = 30 * time.Second
    }
    if config.MaxDataPoints == 0 {
        config.MaxDataPoints = 1000
    }
    
    // 創建 Prometheus 客戶端
    client, err := api.NewClient(api.Config{
        Address: config.PrometheusURL,
    })
    if err != nil {
        return nil, fmt.Errorf("創建 Prometheus 客戶端失敗: %w", err)
    }
    
    return &HealthAggregatorPlugin{
        promClient: v1.NewAPI(client),
        logger:     zap.L().Named("health_aggregator"),
        config:     &config,
    }, nil
}

// Execute 執行健康數據聚合
func (h *HealthAggregatorPlugin) Execute(ctx context.Context, req *pb.ToolRequest) (*pb.ToolResponse, error) {
    h.logger.Info("開始執行健康數據聚合", 
        zap.String("trace_id", req.TraceId),
        zap.Any("args", req.Args))
    
    // 解析查詢參數
    params, err := h.parseQueryParams(req.Args)
    if err != nil {
        return nil, fmt.Errorf("解析查詢參數失敗: %w", err)
    }
    
    // 執行數據查詢
    results, err := h.queryHealthMetrics(ctx, params)
    if err != nil {
        return nil, fmt.Errorf("查詢健康指標失敗: %w", err)
    }
    
    // 聚合數據
    aggregated := h.aggregateResults(results)
    
    // 序列化結果
    responseData, err := json.Marshal(aggregated)
    if err != nil {
        return nil, fmt.Errorf("序列化結果失敗: %w", err)
    }
    
    return &pb.ToolResponse{
        Data:      responseData,
        Status:    "success",
        Timestamp: time.Now().Unix(),
    }, nil
}

// QueryParams 查詢參數結構
type QueryParams struct {
    Services    []string  `json:"services"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    MetricTypes []string  `json:"metric_types"`
    Resolution  string    `json:"resolution"`
}

// parseQueryParams 解析查詢參數
func (h *HealthAggregatorPlugin) parseQueryParams(args map[string]string) (*QueryParams, error) {
    params := &QueryParams{
        MetricTypes: []string{"cpu", "memory", "disk", "network", "error_rate", "response_time"},
        Resolution:  "1m",
    }
    
    // 解析服務列表
    if servicesStr, ok := args["services"]; ok {
        if err := json.Unmarshal([]byte(servicesStr), &params.Services); err != nil {
            return nil, fmt.Errorf("解析服務列表失敗: %w", err)
        }
    }
    
    // 解析時間範圍
    if startTimeStr, ok := args["start_time"]; ok {
        startTime, err := time.Parse(time.RFC3339, startTimeStr)
        if err != nil {
            return nil, fmt.Errorf("解析開始時間失敗: %w", err)
        }
        params.StartTime = startTime
    }
    
    if endTimeStr, ok := args["end_time"]; ok {
        endTime, err := time.Parse(time.RFC3339, endTimeStr)
        if err != nil {
            return nil, fmt.Errorf("解析結束時間失敗: %w", err)
        }
        params.EndTime = endTime
    }
    
    // 設定預設時間範圍
    if params.StartTime.IsZero() || params.EndTime.IsZero() {
        now := time.Now()
        params.EndTime = now
        params.StartTime = now.Add(-h.config.DefaultTimeRange)
    }
    
    return params, nil
}

// queryHealthMetrics 查詢健康指標
func (h *HealthAggregatorPlugin) queryHealthMetrics(ctx context.Context, params *QueryParams) (map[string]interface{}, error) {
    results := make(map[string]interface{})
    
    for _, service := range params.Services {
        serviceMetrics := make(map[string]interface{})
        
        for _, metricType := range params.MetricTypes {
            query := h.buildPrometheusQuery(service, metricType)
            
            // 執行範圍查詢
            result, warnings, err := h.promClient.QueryRange(ctx, query, v1.Range{
                Start: params.StartTime,
                End:   params.EndTime,
                Step:  h.parseResolution(params.Resolution),
            })
            
            if err != nil {
                h.logger.Warn("查詢指標失敗", 
                    zap.String("service", service),
                    zap.String("metric", metricType),
                    zap.Error(err))
                continue
            }
            
            if len(warnings) > 0 {
                h.logger.Warn("查詢產生警告", zap.Strings("warnings", warnings))
            }
            
            serviceMetrics[metricType] = result
        }
        
        results[service] = serviceMetrics
    }
    
    return results, nil
}

// buildPrometheusQuery 建構 Prometheus 查詢語句
func (h *HealthAggregatorPlugin) buildPrometheusQuery(service, metricType string) string {
    queries := map[string]string{
        "cpu":           fmt.Sprintf(`rate(cpu_usage_total{service="%s"}[5m])`, service),
        "memory":        fmt.Sprintf(`memory_usage_bytes{service="%s"}`, service),
        "disk":          fmt.Sprintf(`disk_usage_percent{service="%s"}`, service),
        "network":       fmt.Sprintf(`rate(network_bytes_total{service="%s"}[5m])`, service),
        "error_rate":    fmt.Sprintf(`rate(http_requests_total{service="%s",status=~"5.."}[5m])`, service),
        "response_time": fmt.Sprintf(`histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{service="%s"}[5m]))`, service),
    }
    
    if query, exists := queries[metricType]; exists {
        return query
    }
    
    // 預設查詢
    return fmt.Sprintf(`up{service="%s"}`, service)
}

// aggregateResults 聚合查詢結果
func (h *HealthAggregatorPlugin) aggregateResults(results map[string]interface{}) map[string]interface{} {
    aggregated := map[string]interface{}{
        "summary": map[string]interface{}{
            "total_services": len(results),
            "healthy_services": 0,
            "degraded_services": 0,
            "unhealthy_services": 0,
        },
        "services": results,
        "recommendations": []string{},
    }
    
    // 計算服務健康狀態
    for service, metrics := range results {
        status := h.calculateServiceHealth(metrics)
        
        switch status {
        case "healthy":
            aggregated["summary"].(map[string]interface{})["healthy_services"] = 
                aggregated["summary"].(map[string]interface{})["healthy_services"].(int) + 1
        case "degraded":
            aggregated["summary"].(map[string]interface{})["degraded_services"] = 
                aggregated["summary"].(map[string]interface{})["degraded_services"].(int) + 1
        case "unhealthy":
            aggregated["summary"].(map[string]interface{})["unhealthy_services"] = 
                aggregated["summary"].(map[string]interface{})["unhealthy_services"].(int) + 1
        }
        
        // 添加服務狀態
        if serviceMap, ok := aggregated["services"].(map[string]interface{})[service].(map[string]interface{}); ok {
            serviceMap["health_status"] = status
        }
    }
    
    return aggregated
}

// calculateServiceHealth 計算服務健康狀態
func (h *HealthAggregatorPlugin) calculateServiceHealth(metrics interface{}) string {
    // 簡化的健康狀態計算邏輯
    // 實際實作中會根據具體指標閾值進行判斷
    return "healthy"
}

// parseResolution 解析查詢解析度
func (h *HealthAggregatorPlugin) parseResolution(resolution string) time.Duration {
    switch resolution {
    case "15s":
        return 15 * time.Second
    case "30s":
        return 30 * time.Second
    case "1m":
        return time.Minute
    case "5m":
        return 5 * time.Minute
    case "15m":
        return 15 * time.Minute
    case "1h":
        return time.Hour
    default:
        return time.Minute
    }
}

// Close 清理資源
func (h *HealthAggregatorPlugin) Close() error {
    h.logger.Info("關閉健康聚合器插件")
    return nil
}
```

### 實作最佳實踐

#### 1. 錯誤處理策略

```go
// 統一錯誤處理
type PluginError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *PluginError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 錯誤分類
const (
    ErrCodeInvalidConfig   = "INVALID_CONFIG"
    ErrCodeQueryFailed     = "QUERY_FAILED"
    ErrCodeTimeout         = "TIMEOUT"
    ErrCodeInternalError   = "INTERNAL_ERROR"
)
```

#### 2. 效能優化

```go
// 並行查詢優化
func (h *HealthAggregatorPlugin) queryMetricsParallel(ctx context.Context, queries []string) ([]interface{}, error) {
    results := make([]interface{}, len(queries))
    errors := make([]error, len(queries))
    
    var wg sync.WaitGroup
    for i, query := range queries {
        wg.Add(1)
        go func(index int, q string) {
            defer wg.Done()
            
            result, _, err := h.promClient.Query(ctx, q, time.Now())
            if err != nil {
                errors[index] = err
                return
            }
            results[index] = result
        }(i, query)
    }
    
    wg.Wait()
    
    // 檢查錯誤
    for _, err := range errors {
        if err != nil {
            return nil, err
        }
    }
    
    return results, nil
}
```

#### 3. 快取策略

```go
// 結果快取
type CacheEntry struct {
    Data      interface{}
    Timestamp time.Time
    TTL       time.Duration
}

func (h *HealthAggregatorPlugin) getCachedResult(key string) (interface{}, bool) {
    if entry, exists := h.cache[key]; exists {
        if time.Since(entry.Timestamp) < entry.TTL {
            return entry.Data, true
        }
        delete(h.cache, key)
    }
    return nil, false
}
```

* * *

## CLI 工具與指令

### 基本指令

```bash
# 啟動平台服務
detectviz serve [選項]

# 插件管理
detectviz plugin serve [選項]           # 啟動插件服務
detectviz plugin new <category>/<name>  # 生成插件腳手架
detectviz plugin list                   # 列出已安裝插件
detectviz plugin validate <path>        # 驗證插件配置

# 配置管理
detectviz config validate -f <config.yaml>  # 驗證配置檔案
detectviz config generate                    # 生成配置範本
detectviz config show                        # 顯示當前配置

# 開發工具
detectviz dev generate-contracts             # 重新生成契約程式碼
detectviz dev test-connection               # 測試外部服務連線
detectviz dev benchmark                     # 執行效能基準測試
```

### 服務啟動選項

```bash
detectviz plugin serve \
  --config ./config.yaml \              # 配置檔案路徑
  --listen :5002 \                      # gRPC 監聽位址
  --http-listen :8080 \                 # HTTP 監聽位址
  --log-level info \                    # 日誌等級
  --enable-pprof \                      # 啟用效能分析
  --mtls-cert ./certs/server.crt \      # mTLS 憑證（可選）
  --mtls-key ./certs/server.key \       # mTLS 私鑰（可選）
  --mtls-ca ./certs/ca.crt              # mTLS CA 憑證（可選）
```

### 啟動流程優化

**模組化啟動序列**：
1. **參數解析** - 驗證 CLI 參數和環境變數
2. **配置載入** - 載入並驗證配置檔案
3. **契約檢查** - 驗證 Protocol Buffer 版本一致性
4. **日誌初始化** - 設定結構化日誌系統
5. **可觀測性** - 初始化 OpenTelemetry 和指標收集
6. **插件系統** - 載入和註冊插件
7. **服務啟動** - 啟動 gRPC 和 HTTP 服務
8. **健康檢查** - 啟動健康檢查端點

**錯誤處理機制**：
```go
// 啟動錯誤處理範例
func (s *Server) Start(ctx context.Context) error {
    // 設定 panic 恢復
    defer func() {
        if r := recover(); r != nil {
            s.logger.Error("服務啟動 panic", zap.Any("panic", r))
            s.Shutdown(ctx)
        }
    }()
    
    // 階段性啟動
    stages := []struct {
        name string
        fn   func() error
    }{
        {"配置驗證", s.validateConfig},
        {"契約檢查", s.checkContracts},
        {"插件載入", s.loadPlugins},
        {"服務啟動", s.startServices},
    }
    
    for _, stage := range stages {
        s.logger.Info("執行啟動階段", zap.String("stage", stage.name))
        if err := stage.fn(); err != nil {
            return fmt.Errorf("啟動階段 %s 失敗: %w", stage.name, err)
        }
    }
    
    return nil
}
```

## 部署與運維指南

### 本地開發環境設定

#### 1. 環境準備

```bash
# 安裝依賴
go version  # 需要 Go 1.21+
python --version  # 需要 Python 3.11+

# 安裝 Protocol Buffers 工具
# macOS
brew install protobuf buf

# Ubuntu/Debian
apt-get install -y protobuf-compiler
curl -sSL https://github.com/bufbuild/buf/releases/latest/download/buf-Linux-x86_64.tar.gz | tar -xz -C /usr/local/bin

# 安裝 Python 依賴
cd python-adk-runtime
pip install -r requirements.txt
```

#### 2. 建置與啟動

```bash
# 生成契約程式碼
cd contracts
make generate

# 建置 Go 平台
cd go-platform
go build -o bin/detectviz ./cmd/detectviz

# 啟動服務（開發模式）
./bin/detectviz plugin serve --config configs/config-dev.yaml --log-level debug

# 啟動 Python ADK 運行時
cd python-adk-runtime
python web_server.py
```

#### 3. 驗證部署

```bash
# 檢查服務健康狀態
curl http://localhost:8080/health

# 檢查 gRPC 服務
grpcurl -plaintext localhost:5002 detectviz.contracts.v1.HealthService/Check

# 執行整合測試
cd python-adk-runtime
python test_adk_integration.py
```

### 生產環境部署

#### Docker 容器化部署

**Dockerfile 範例**：
```dockerfile
# Go 平台容器
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go-platform/ ./go-platform/
COPY contracts/ ./contracts/

RUN cd go-platform && go build -o bin/detectviz ./cmd/detectviz

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/go-platform/bin/detectviz .
COPY --from=builder /app/go-platform/configs/ ./configs/

EXPOSE 5002 8080
CMD ["./detectviz", "plugin", "serve", "--config", "configs/config.yaml"]
```

**Docker Compose 配置**：
```yaml
version: '3.8'

services:
  detectviz-platform:
    build: .
    ports:
      - "5002:5002"  # gRPC
      - "8080:8080"  # HTTP
    environment:
      - DETECTVIZ_ENV=production
      - PROMETHEUS_URL=http://prometheus:9090
      - GRAFANA_URL=http://grafana:3000
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    depends_on:
      - prometheus
      - grafana
    restart: unless-stopped

  detectviz-adk:
    build: ./python-adk-runtime
    ports:
      - "8000:8000"
    environment:
      - DETECTVIZ_PLATFORM_URL=http://detectviz-platform:5002
    depends_on:
      - detectviz-platform
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
    restart: unless-stopped

volumes:
  grafana-storage:
```

#### Kubernetes 部署

**部署清單範例**：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: detectviz-platform
  labels:
    app: detectviz-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app: detectviz-platform
  template:
    metadata:
      labels:
        app: detectviz-platform
    spec:
      containers:
      - name: detectviz-platform
        image: detectviz/platform:latest
        ports:
        - containerPort: 5002
          name: grpc
        - containerPort: 8080
          name: http
        env:
        - name: DETECTVIZ_ENV
          value: "production"
        - name: PROMETHEUS_URL
          value: "http://prometheus:9090"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

```yaml
apiVersion: v1
kind: Service
metadata:
  name: detectviz-platform-service
spec:
  selector:
    app: detectviz-platform
  ports:
  - name: grpc
    port: 5002
    targetPort: 5002
  - name: http
    port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 監控與可觀測性設定

#### Grafana Alloy 配置

**檔案位置**：`depoly/alloy/config.alloy`

```hcl
// 日誌收集
loki.source.file "detectviz_logs" {
  targets = [
    {__path__ = "/app/logs/*.log"},
  ]
  forward_to = [loki.write.grafana_cloud.receiver]
}

loki.write "grafana_cloud" {
  endpoint {
    url = env("LOKI_ENDPOINT")
    basic_auth {
      username = env("LOKI_USERNAME")
      password = env("LOKI_PASSWORD")
    }
  }
}

// 指標收集
otelcol.receiver.otlp "detectviz" {
  grpc {
    endpoint = "0.0.0.0:4317"
  }
  http {
    endpoint = "0.0.0.0:4318"
  }
  
  output {
    metrics = [otelcol.exporter.otlphttp.grafana_cloud.input]
    traces  = [otelcol.exporter.otlphttp.grafana_cloud.input]
  }
}

otelcol.exporter.otlphttp "grafana_cloud" {
  client {
    endpoint = env("OTLP_ENDPOINT")
    auth     = otelcol.auth.basic.grafana_cloud.handler
  }
}

otelcol.auth.basic "grafana_cloud" {
  username = env("OTLP_USERNAME")
  password = env("OTLP_PASSWORD")
}

// 效能分析收集
pyroscope.scrape "detectviz" {
  targets = [
    {"__address__" = "detectviz-platform:6060"},
  ]
  forward_to = [pyroscope.write.grafana_cloud.receiver]
}

pyroscope.write "grafana_cloud" {
  endpoint {
    url = env("PYROSCOPE_ENDPOINT")
    basic_auth {
      username = env("PYROSCOPE_USERNAME")
      password = env("PYROSCOPE_PASSWORD")
    }
  }
}
```

## 故障排除與除錯指南

### 常見問題診斷

#### 1. 服務啟動失敗

**問題症狀**：
```
ERROR: 服務啟動失敗: 契約版本檢查失敗
```

**診斷步驟**：
```bash
# 檢查契約版本
cat contracts/gen/metadata/version.json

# 重新生成契約程式碼
cd contracts
make clean
make generate

# 驗證生成結果
ls -la gen/go/detectviz/contracts/v1/
ls -la gen/python/detectviz/contracts/v1/
```

#### 2. gRPC 連線問題

**問題症狀**：
```
ERROR: rpc error: code = Unavailable desc = connection error
```

**診斷步驟**：
```bash
# 檢查服務是否運行
netstat -tlnp | grep 5002

# 測試 gRPC 連線
grpcurl -plaintext localhost:5002 list

# 檢查防火牆設定
sudo ufw status
```

#### 3. 插件載入失敗

**問題症狀**：
```
WARN: 插件載入失敗: observability.health_aggregator
```

**診斷步驟**：
```bash
# 檢查插件配置
detectviz plugin validate go-platform/internal/pluginhost/plugins/observability/health_aggregator/

# 檢查模組卡
cat go-platform/internal/pluginhost/plugins/observability/health_aggregator/module.card.json

# 檢查依賴
go mod tidy
go mod verify
```

#### 4. 記憶體使用過高

**診斷工具**：
```bash
# 檢查記憶體使用
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 檢查 goroutine 洩漏
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof

# 即時監控
watch -n 1 'curl -s http://localhost:8080/health | jq .memory'
```

### 效能調優指南

#### 1. gRPC 連線池優化

```go
// 客戶端連線池配置
conn, err := grpc.Dial(address,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    }),
    grpc.WithDefaultCallOptions(
        grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
        grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
    ),
)
```

#### 2. 查詢效能優化

```go
// 批量查詢優化
func (h *HealthAggregatorPlugin) optimizedBatchQuery(ctx context.Context, queries []string) error {
    // 使用 worker pool 限制並發數
    const maxWorkers = 10
    semaphore := make(chan struct{}, maxWorkers)
    
    var wg sync.WaitGroup
    for _, query := range queries {
        wg.Add(1)
        go func(q string) {
            defer wg.Done()
            semaphore <- struct{}{}        // 獲取 worker
            defer func() { <-semaphore }() // 釋放 worker
            
            // 執行查詢
            h.executeQuery(ctx, q)
        }(query)
    }
    
    wg.Wait()
    return nil
}
```

#### 3. 快取策略

```go
// LRU 快取實作
type LRUCache struct {
    capacity int
    cache    map[string]*list.Element
    list     *list.List
    mutex    sync.RWMutex
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    if elem, exists := c.cache[key]; exists {
        c.list.MoveToFront(elem)
        return elem.Value.(*CacheEntry).Value, true
    }
    return nil, false
}
```

### 日誌分析

#### 重要日誌模式

```bash
# 查看錯誤日誌
grep "ERROR" /app/logs/detectviz.log | tail -20

# 查看效能警告
grep "WARN.*timeout\|WARN.*slow" /app/logs/detectviz.log

# 查看 gRPC 呼叫統計
grep "grpc" /app/logs/detectviz.log | grep -E "duration|latency" | tail -10

# 查看記憶體使用趨勢
grep "memory_usage" /app/logs/detectviz.log | awk '{print $1, $NF}' | tail -20
```

#### 結構化日誌查詢

```bash
# 使用 jq 分析 JSON 日誌
cat /app/logs/detectviz.log | jq 'select(.level == "ERROR")' | head -5

# 統計錯誤類型
cat /app/logs/detectviz.log | jq -r 'select(.level == "ERROR") | .error_code' | sort | uniq -c

# 查看特定 trace 的日誌
cat /app/logs/detectviz.log | jq 'select(.trace_id == "abc123")'
```

* * *

## Python ADK 整合實作

> **Agent 協作模式與設計原則**：參見 [ARCHITECTURE.md](ARCHITECTURE.md#agent-協作模式)

### RemoteTool 實作

**檔案位置**：`python-adk-runtime/src/detectviz_adk/tools/remote_tool.py`

```python
"""遠端工具客戶端 - 連接 Go 平台的 gRPC 服務"""

import grpc
import json
import asyncio
from typing import Dict, Any, Optional, AsyncIterator
from google.protobuf.json_format import MessageToDict

from detectviz.contracts.v1 import adk_bridge_pb2
from detectviz.contracts.v1 import adk_bridge_pb2_grpc

class RemoteTool:
    """遠端工具客戶端"""
    
    def __init__(self, platform_url: str = "localhost:5002"):
        self.platform_url = platform_url
        self.channel = None
        self.stub = None
    
    async def __aenter__(self):
        """異步上下文管理器入口"""
        self.channel = grpc.aio.insecure_channel(self.platform_url)
        self.stub = adk_bridge_pb2_grpc.ToolBridgeServiceStub(self.channel)
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """異步上下文管理器出口"""
        if self.channel:
            await self.channel.close()
    
    async def execute_tool(
        self,
        tool_name: str,
        tool_version: str = "latest",
        args: Dict[str, Any] = None,
        timeout_seconds: int = 30,
        trace_id: Optional[str] = None
    ) -> AsyncIterator[Dict[str, Any]]:
        """執行遠端工具"""
        
        request = adk_bridge_pb2.ToolRequest(
            name=tool_name,
            version=tool_version,
            args=args or {},
            timeout_seconds=timeout_seconds,
            trace_id=trace_id or self._generate_trace_id()
        )
        
        try:
            async for chunk in self.stub.ExecuteTool(request):
                yield {
                    'type': chunk.type,
                    'data': chunk.data.decode('utf-8') if chunk.data else None,
                    'message': chunk.message,
                    'progress_percent': chunk.progress_percent,
                    'timestamp': chunk.timestamp
                }
        except grpc.RpcError as e:
            raise RemoteToolError(f"工具執行失敗: {e.details()}")
    
    async def execute_tool_sync(
        self,
        tool_name: str,
        tool_version: str = "latest",
        args: Dict[str, Any] = None,
        timeout_seconds: int = 30,
        trace_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """同步執行遠端工具（等待完成）"""
        
        results = []
        final_result = None
        
        async for chunk in self.execute_tool(
            tool_name, tool_version, args, timeout_seconds, trace_id
        ):
            if chunk['type'] == 'COMPLETE':
                final_result = chunk
                break
            elif chunk['type'] == 'ERROR':
                raise RemoteToolError(f"工具執行錯誤: {chunk['message']}")
            else:
                results.append(chunk)
        
        return {
            'status': 'completed',
            'data': final_result['data'] if final_result else None,
            'chunks': results
        }
    
    def _generate_trace_id(self) -> str:
        """生成追蹤 ID"""
        import uuid
        return str(uuid.uuid4())

class RemoteToolError(Exception):
    """遠端工具錯誤"""
    pass

# ADK 工具包裝器
class ADKRemoteTool:
    """ADK 相容的遠端工具包裝器"""
    
    def __init__(self, tool_name: str, platform_url: str = "localhost:5002"):
        self.tool_name = tool_name
        self.platform_url = platform_url
    
    async def __call__(self, **kwargs) -> Dict[str, Any]:
        """ADK 工具調用介面"""
        async with RemoteTool(self.platform_url) as remote_tool:
            return await remote_tool.execute_tool_sync(
                tool_name=self.tool_name,
                args=kwargs
            )

# 工具工廠
def create_remote_tool(tool_name: str, platform_url: str = "localhost:5002") -> ADKRemoteTool:
    """創建遠端工具實例"""
    return ADKRemoteTool(tool_name, platform_url)
```

### Agent 實作範例

**檔案位置**：`python-adk-runtime/src/detectviz_adk/agents/postmortem/data_collector.py`

```python
"""數據收集 Agent - 負責收集事故相關數據"""

from google import adk
from typing import Dict, Any, List
import asyncio
import logging

from detectviz_adk.tools.remote_tool import create_remote_tool

class DataCollectorAgent:
    """數據收集專家 Agent"""
    
    def __init__(self):
        self.logger = logging.getLogger(__name__)
        
        # 初始化遠端工具
        self.health_aggregator = create_remote_tool("observability.health_aggregator")
        self.metrics_query = create_remote_tool("observability.metrics_query")
        self.log_analyzer = create_remote_tool("observability.log_analyzer")
    
    async def collect_data(self, strategy: Dict[str, Any]) -> Dict[str, Any]:
        """根據策略收集數據"""
        
        self.logger.info("開始數據收集", extra={"strategy": strategy})
        
        # 並行收集不同類型的數據
        tasks = []
        
        # 收集健康指標
        if strategy.get('include_health_metrics', True):
            tasks.append(self._collect_health_metrics(strategy))
        
        # 收集應用指標
        if strategy.get('include_app_metrics', True):
            tasks.append(self._collect_application_metrics(strategy))
        
        # 收集日誌
        if strategy.get('include_logs', True):
            tasks.append(self._collect_logs(strategy))
        
        # 等待所有收集任務完成
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # 整合收集結果
        collected_data = {
            'health_metrics': None,
            'app_metrics': None,
            'logs': None,
            'collection_summary': {
                'total_tasks': len(tasks),
                'successful_tasks': 0,
                'failed_tasks': 0
            }
        }
        
        # 處理結果
        for i, result in enumerate(results):
            if isinstance(result, Exception):
                self.logger.error(f"數據收集任務 {i} 失敗", exc_info=result)
                collected_data['collection_summary']['failed_tasks'] += 1
            else:
                collected_data['collection_summary']['successful_tasks'] += 1
                
                # 根據任務類型分配結果
                if i == 0:  # 健康指標
                    collected_data['health_metrics'] = result
                elif i == 1:  # 應用指標
                    collected_data['app_metrics'] = result
                elif i == 2:  # 日誌
                    collected_data['logs'] = result
        
        self.logger.info("數據收集完成", extra=collected_data['collection_summary'])
        return collected_data
    
    async def _collect_health_metrics(self, strategy: Dict[str, Any]) -> Dict[str, Any]:
        """收集健康指標"""
        
        services = strategy.get('services', [])
        time_window = strategy.get('time_window_minutes', 60)
        
        # 計算時間範圍
        import datetime
        end_time = datetime.datetime.now()
        start_time = end_time - datetime.timedelta(minutes=time_window)
        
        # 調用健康聚合器
        result = await self.health_aggregator(
            services=services,
            start_time=start_time.isoformat(),
            end_time=end_time.isoformat(),
            metric_types=['cpu', 'memory', 'disk', 'network', 'error_rate']
        )
        
        return result
    
    async def _collect_application_metrics(self, strategy: Dict[str, Any]) -> Dict[str, Any]:
        """收集應用指標"""
        
        services = strategy.get('services', [])
        detail_level = strategy.get('detail_level', 'medium')
        
        # 根據詳細程度決定查詢範圍
        metric_queries = self._build_metric_queries(services, detail_level)
        
        results = {}
        for service, queries in metric_queries.items():
            service_metrics = {}
            for metric_name, query in queries.items():
                try:
                    result = await self.metrics_query(
                        query=query,
                        service=service
                    )
                    service_metrics[metric_name] = result
                except Exception as e:
                    self.logger.warning(f"查詢 {service}.{metric_name} 失敗", exc_info=e)
                    service_metrics[metric_name] = None
            
            results[service] = service_metrics
        
        return results
    
    async def _collect_logs(self, strategy: Dict[str, Any]) -> Dict[str, Any]:
        """收集日誌數據"""
        
        services = strategy.get('services', [])
        
        # 調用日誌分析器
        result = await self.log_analyzer(
            services=services,
            time_window_minutes=time_window,
            log_levels=['ERROR', 'WARN'],
            include_stack_traces=True
        )
        
        return result
    
    def _build_metric_queries(self, services: List[str], detail_level: str) -> Dict[str, Dict[str, str]]:
        """建構指標查詢"""
        
        base_queries = {
            'request_rate': 'rate(http_requests_total[5m])',
            'error_rate': 'rate(http_requests_total{status=~"5.."}[5m])',
            'response_time': 'histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))'
        }
        
        if detail_level == 'high':
            base_queries.update({
                'cpu_usage': 'rate(cpu_usage_total[5m])',
                'memory_usage': 'memory_usage_bytes',
                'disk_io': 'rate(disk_io_bytes_total[5m])',
                'network_io': 'rate(network_io_bytes_total[5m])'
            })
        
        # 為每個服務生成查詢
        queries = {}
        for service in services:
            service_queries = {}
            for metric_name, base_query in base_queries.items():
                service_queries[metric_name] = f'{base_query}{{service="{service}"}}'
            queries[service] = service_queries
        
        return queries
```

### 記憶體管理實作

**檔案位置**：`python-adk-runtime/src/detectviz_adk/memory/stores/response_history_store.py`

```python
# 注意：以下為程式碼片段，可能需要在完整上下文中使用
# 注意：以下為程式碼片段，可能需要在完整上下文中使用
"""響應歷史存儲 - 知識庫實作"""

import json
import sqlite3
import hashlib
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta
import asyncio
import aiosqlite

class ResponseHistoryStore:
    """響應歷史存儲"""
    
    def __init__(self, db_path: str = "./data/response_history.db"):
        self.db_path = db_path
        self._init_db()
    
    def _init_db(self):
        """初始化資料庫"""
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS response_history (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    incident_id TEXT NOT NULL,
                    agent_name TEXT NOT NULL,
                    request_hash TEXT NOT NULL,
                    request_data TEXT NOT NULL,
                    response_data TEXT NOT NULL,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    expires_at TIMESTAMP,
                    tags TEXT,
                    UNIQUE(incident_id, agent_name, request_hash)
                )
            """)
            
            conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_incident_id 
                ON response_history(incident_id)
            """)
            
            conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_agent_name 
                ON response_history(agent_name)
            """)
            
            conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_created_at 
                ON response_history(created_at)
            """)
    
    async def store_response(
        self,
        incident_id: str,
        agent_name: str,
        request_data: Dict[str, Any],
        response_data: Dict[str, Any],
        ttl_hours: int = 24 * 7,  # 預設保存 7 天
        tags: List[str] = None
    ) -> str:
        """存儲響應數據"""
        
        # 計算請求雜湊
        request_str = json.dumps(request_data, sort_keys=True)
        request_hash = hashlib.sha256(request_str.encode()).hexdigest()
        
        # 計算過期時間
        expires_at = datetime.now() + timedelta(hours=ttl_hours)
        
        async with aiosqlite.connect(self.db_path) as conn:
            await conn.execute("""
                INSERT OR REPLACE INTO response_history 
                (incident_id, agent_name, request_hash, request_data, response_data, expires_at, tags)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            """, (
                incident_id,
                agent_name,
                request_hash,
                json.dumps(request_data),
                json.dumps(response_data),
                expires_at,
                json.dumps(tags or [])
            ))
            await conn.commit()
        
        return request_hash
    
    async def get_response(
        self,
        incident_id: str,
        agent_name: str,
        request_data: Dict[str, Any]
    ) -> Optional[Dict[str, Any]]:
        """獲取快取的響應"""
        
        # 計算請求雜湊
        
            async with conn.execute("""
                SELECT response_data, created_at, expires_at 
                FROM response_history 
                WHERE incident_id = ? AND agent_name = ? AND request_hash = ?
                AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
            """, (incident_id, agent_name, request_hash)) as cursor:
                
                row = await cursor.fetchone()
                if row:
                    return {
                        'data': json.loads(row[0]),
                        'created_at': row[1],
                        'expires_at': row[2]
                    }
        
        return None
    
    async def search_similar_incidents(
        self,
        services: List[str],
        error_patterns: List[str] = None,
        limit: int = 10
    ) -> List[Dict[str, Any]]:
        """搜尋相似事件"""
        
        # 建構搜尋條件
        conditions = []
        params = []
        
        if services:
            service_conditions = []
            for service in services:
                service_conditions.append("request_data LIKE ?")
                params.append(f'%"{service}"%')
            conditions.append(f"({' OR '.join(service_conditions)})")
        
        if error_patterns:
            pattern_conditions = []
            for pattern in error_patterns:
                pattern_conditions.append("response_data LIKE ?")
                params.append(f'%{pattern}%')
            conditions.append(f"({' OR '.join(pattern_conditions)})")
        
        where_clause = " AND ".join(conditions) if conditions else "1=1"
        params.append(limit)
        
            async with conn.execute(f"""
                SELECT incident_id, agent_name, request_data, response_data, created_at, tags
                FROM response_history 
                WHERE {where_clause}
                ORDER BY created_at DESC
                LIMIT ?
            """, params) as cursor:
                
                results = []
                async for row in cursor:
                    results.append({
                        'incident_id': row[0],
                        'agent_name': row[1],
                        'request_data': json.loads(row[2]),
                        'response_data': json.loads(row[3]),
                        'created_at': row[4],
                        'tags': json.loads(row[5])
                    })
                
                return results
    
    async def cleanup_expired(self) -> int:
        """清理過期數據"""
        
            cursor = await conn.execute("""
                DELETE FROM response_history 
                WHERE expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP
            """)
            await conn.commit()
            return cursor.rowcount
```

* * *

## 開發擴展指南

### 新增 Go 插件

#### 1. 使用腳手架生成插件

```bash
# 生成新插件骨架
detectviz plugin new observability/custom_metrics

# 生成的目錄結構
go-platform/internal/pluginhost/plugins/observability/custom_metrics/
├── plugin.go           # 主要實作檔案
├── module.card.json    # 模組卡配置
├── config.go          # 配置結構
└── README.md          # 插件說明
```

#### 2. 實作插件介面

```go
package custom_metrics

import (
    "context"
    "encoding/json"
    
)

type CustomMetricsPlugin struct {
    config *Config
    logger *zap.Logger
}

// Execute 實作插件執行介面
func (p *CustomMetricsPlugin) Execute(ctx context.Context, req *pb.ToolRequest) (*pb.ToolResponse, error) {
    // 解析請求參數
    var params struct {
        MetricName string `json:"metric_name"`
        TimeRange  string `json:"time_range"`
    }
    
    if err := json.Unmarshal([]byte(req.Args["params"]), &params); err != nil {
        return nil, fmt.Errorf("解析參數失敗: %w", err)
    }
    
    // 執行業務邏輯
    result, err := p.processMetrics(ctx, params.MetricName, params.TimeRange)
    if err != nil {
        return nil, err
    }
    
    // 返回結果
    responseData, _ := json.Marshal(result)
    return &pb.ToolResponse{
        Data:      responseData,
        Status:    "success",
        Timestamp: time.Now().Unix(),
    }, nil
}

// Close 清理資源
func (p *CustomMetricsPlugin) Close() error {
    return nil
}
```

#### 3. 註冊插件

```go
// 在 go-platform/internal/pluginhost/registry.go 中註冊
func (r *Registry) registerBuiltinPlugins() {
    // 註冊自訂指標插件
    r.Register("observability.custom_metrics", func(config []byte) (Plugin, error) {
        return NewCustomMetricsPlugin(config)
    })
}
```

### 新增 Python Agent

#### 1. 創建 Agent 結構

```python
# python-adk-runtime/src/detectviz_adk/agents/custom/my_agent.py

from google import adk
from typing import Dict, Any
import asyncio

class MyCustomAgent:
    """自訂 Agent 實作"""
    
    def __init__(self):
        self.name = "my_custom_agent"
        self.tools = {
            'remote_tool': create_remote_tool("observability.custom_metrics")
        }
    
    async def process_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """處理請求的主要邏輯"""
        
        # 決策邏輯
        strategy = await self._make_decision(request)
        
        # 執行工具調用
        result = await self.tools['remote_tool'](
            metric_name=strategy['metric_name'],
            time_range=strategy['time_range']
        )
        
        # 處理結果
        processed_result = await self._process_result(result)
        
        return processed_result
    
    async def _make_decision(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """決策邏輯 - Agent 的核心職責"""
        # 實作決策邏輯
        pass
    
    async def _process_result(self, result: Dict[str, Any]) -> Dict[str, Any]:
        """結果處理邏輯"""
        # 實作結果處理
        pass
```

#### 2. 創建模組卡


### 測試與驗證

#### 單元測試範例

```python
# tests/test_my_agent.py

import pytest
import asyncio
from unittest.mock import AsyncMock, patch

from detectviz_adk.agents.custom.my_agent import MyCustomAgent

class TestMyCustomAgent:
    
    @pytest.fixture
    def agent(self):
        return MyCustomAgent()
    
    @pytest.mark.asyncio
    async def test_process_request_success(self, agent):
        """測試成功處理請求"""
        
        # 模擬工具調用
        with patch.object(agent.tools['remote_tool'], '__call__', new_callable=AsyncMock) as mock_tool:
            mock_tool.return_value = {'status': 'success', 'data': {'value': 100}}
            
            request = {
                'type': 'analysis',
                'target': 'service_a'
            }
            
            result = await agent.process_request(request)
            
            assert result['status'] == 'completed'
            mock_tool.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_decision_making(self, agent):
        """測試決策邏輯"""
        
        request = {
            'severity': 'high',
            'services': ['service_a', 'service_b']
        }
        
        strategy = await agent._make_decision(request)
        
        assert 'metric_name' in strategy
        assert 'time_range' in strategy
```

#### 整合測試

```python
# tests/integration/test_agent_integration.py

import pytest
import asyncio
from detectviz_adk.tools.remote_tool import RemoteTool

class TestAgentIntegration:
    
    @pytest.mark.asyncio
    async def test_end_to_end_workflow(self):
        """端到端工作流程測試"""
        
        agent = MyCustomAgent()
        
        # 真實的請求
        request = {
            'incident_id': 'test_incident_001',
            'services': ['web_service', 'api_service'],
            'time_range': '1h'
        }
        
        # 執行完整流程
        result = await agent.process_request(request)
        
        # 驗證結果
        assert result['status'] in ['completed', 'partial']
        assert 'data' in result
        assert 'recommendations' in result
```

## 效能考量與最佳化

### 查詢效能最佳化

#### 1. 批量查詢策略

```go
// 批量查詢實作
func (p *MetricsPlugin) BatchQuery(ctx context.Context, queries []string) ([]QueryResult, error) {
    const maxConcurrency = 10
    semaphore := make(chan struct{}, maxConcurrency)
    
    results := make([]QueryResult, len(queries))
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for i, query := range queries {
        wg.Add(1)
        go func(index int, q string) {
            defer wg.Done()
            
            semaphore <- struct{}{}        // 獲取信號量
            defer func() { <-semaphore }() // 釋放信號量
            
            result, err := p.executeQuery(ctx, q)
            
            mu.Lock()
            results[index] = QueryResult{Data: result, Error: err}
            mu.Unlock()
        }(i, query)
    }
    
    wg.Wait()
    return results, nil
}
```

#### 2. 快取策略

```go
// 多層快取實作
type CacheManager struct {
    l1Cache *sync.Map           // 記憶體快取
    l2Cache redis.Cmdable       // Redis 快取
    ttl     time.Duration
}

func (c *CacheManager) Get(ctx context.Context, key string) (interface{}, bool) {
    // L1 快取檢查
    if value, ok := c.l1Cache.Load(key); ok {
        return value, true
    }
    
    // L2 快取檢查
    if c.l2Cache != nil {
        value, err := c.l2Cache.Get(ctx, key).Result()
        if err == nil {
            // 回填 L1 快取
            c.l1Cache.Store(key, value)
            return value, true
        }
    }
    
    return nil, false
}
```

#### 3. 連線池管理

```go
// gRPC 連線池
type ConnectionPool struct {
    connections chan *grpc.ClientConn
    factory     func() (*grpc.ClientConn, error)
    mu          sync.RWMutex
    closed      bool
}

func NewConnectionPool(size int, factory func() (*grpc.ClientConn, error)) *ConnectionPool {
    pool := &ConnectionPool{
        connections: make(chan *grpc.ClientConn, size),
        factory:     factory,
    }
    
    // 預先建立連線
    for i := 0; i < size; i++ {
        conn, err := factory()
        if err == nil {
            pool.connections <- conn
        }
    }
    
    return pool
}

func (p *ConnectionPool) Get() (*grpc.ClientConn, error) {
    select {
    case conn := <-p.connections:
        return conn, nil
    default:
        return p.factory()
    }
}

func (p *ConnectionPool) Put(conn *grpc.ClientConn) {
    if p.closed {
        conn.Close()
        return
    }
    
    select {
    case p.connections <- conn:
    default:
        conn.Close()
    }
}
```

### 記憶體最佳化

#### 1. 物件池

```go
// 結果物件池
var resultPool = sync.Pool{
    New: func() interface{} {
        return &QueryResult{
            Data: make(map[string]interface{}),
        }
    },
}

func getResult() *QueryResult {
    return resultPool.Get().(*QueryResult)
}

func putResult(result *QueryResult) {
    // 清理資料
    for k := range result.Data {
        delete(result.Data, k)
    }
    result.Error = nil
    result.Timestamp = 0
    
    resultPool.Put(result)
}
```

#### 2. 串流處理

```go
// 串流處理大量數據
func (p *Plugin) StreamProcess(ctx context.Context, req *pb.ToolRequest) (<-chan *pb.ToolChunk, error) {
    resultChan := make(chan *pb.ToolChunk, 100)
    
    go func() {
        defer close(resultChan)
        
        // 分批處理數據
        batchSize := 1000
        for offset := 0; ; offset += batchSize {
            batch, err := p.fetchBatch(ctx, offset, batchSize)
            if err != nil {
                resultChan <- &pb.ToolChunk{
                    Type:    pb.ToolChunk_ERROR,
                    Message: err.Error(),
                }
                return
            }
            
            if len(batch) == 0 {
                break
            }
            
            // 處理批次數據
            processed := p.processBatch(batch)
            
            // 發送結果
            data, _ := json.Marshal(processed)
            resultChan <- &pb.ToolChunk{
                Type: pb.ToolChunk_DATA,
                Data: data,
            }
        }
        
        // 發送完成信號
        resultChan <- &pb.ToolChunk{
            Type: pb.ToolChunk_COMPLETE,
        }
    }()
    
    return resultChan, nil
}
```

---

*本技術規格文檔提供了 Detectviz Platform 的完整實作指南，包含 API 定義、配置管理、部署指南和最佳實踐。*

*最後更新：2025-08-17*  
*版本：1.0.0*  
*維護者：開發團隊*