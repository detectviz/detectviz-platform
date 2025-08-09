# Detectviz Platform

[![Status: MVP](https://img.shields.io/badge/status-MVP-brightgreen)](./spec.md)
[![SSOT: contracts](https://img.shields.io/badge/SSOT-contracts-0A84FF)](./contracts)
[![Go >= 1.21](https://img.shields.io/badge/Go-%3E%3D%201.21-00ADD8?logo=go)](#)
[![Python >= 3.11](https://img.shields.io/badge/Python-%3E%3D%203.11-3776AB?logo=python)](#)
[![Google ADK aligned](https://img.shields.io/badge/Google%20ADK-aligned-4285F4?logo=google)](https://google.github.io/adk-docs/)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-enabled-6E43E6?logo=opentelemetry)](#)
[![Grafana Alloy](https://img.shields.io/badge/Grafana-Alloy-F46800?logo=grafana)](./grafana-alloy/config.alloy)
[![License](https://img.shields.io/badge/license-MIT-blue)](./LICENSE)

以 **Google Agent Development Kit (ADK)** 為核心，結合 Go 與 Python 的混合式可觀察性與智能代理平台。此倉庫聚焦最小可運行與可擴展：以 `contracts/` 作為單一事實來源（SSOT），Go 作為平台層與 ToolBridge，Python 則承載 ADK Runtime 與多代理協作。觀測以 Grafana Alloy 為收集/轉送層，支援本地 LGTM、Grafana Cloud 與 GCP。

> 詳細規範、術語與流程請以 [`spec.md`](./spec.md) 為準。

---

## 核心概念
- **SSOT（contracts）**：所有跨語言契約、設定與分類皆以 Proto/Schema 為唯一真相，下游程式碼不得手改生成檔。
- **解耦**：Go 與 Python 僅透過 gRPC/HTTP 溝通；平台層最小化、業務邏輯下沉到 ADK Runtime。
- **可擴展**：以「模組卡（Module Card）」規範 Agent/Tool/Capability/Plugin 的 `role`/`category` 與依賴；提供樣板與驗證工具。
- **可觀測性優先**：統一 Logs/Traces/Metrics/Profiles；應用側只暴露 `pprof`，由 Alloy 以 `pyroscope.scrape` 上送。

---

## 目錄導覽
- `contracts/`：SSOT 契約與樣本
  - `proto/`：gRPC 介面（`adk_bridge.proto`）
  - `schemas/`：`config.schema.json`、`module.card.schema.json`
  - `samples/`：組態樣本（可複製到 go-platform 啟動）
  - `tools/`：合規檢查與卡片驗證
- `go-platform/`：平台核心（CLI、ToolBridge、插件、觀測初始化）
  - `cmd/detectviz/`：CLI 入口
  - `internal/configx/`：設定載入與 Schema 驗證
  - `internal/pluginhost/`：Registry 與插件目錄（telegraf-like）
  - `internal/observability/`：zap 與 OTEL 初始化（含 `pprof`）
  - `configs/`：預設 `config.yaml`
- `python-adk-runtime/`：ADK Runtime 與 RemoteTool
  - `src/detectviz_adk/`：runtime/memory/tools/capabilities（含 `remote_tool.py`）
  - `agents/`、`templates/`：範例與樣板
- `grafana-alloy/`：Alloy 管線設定（本地或雲端）

---

## 快速開始（概覽）
1. 讀取與校對規格：`spec.md`、`contracts/README.md`、`go-platform/README.md`、`python-adk-runtime/README.md`
2. 套用範例設定：
   ```bash
   cp contracts/samples/config.yaml go-platform/configs/config.yaml
   ```
3. 啟動 Alloy（請先設定 `.env` 中的雲端變數或使用本地 LGTM）：
   ```bash
   ./grafana-alloy/alloy run ./grafana-alloy/config.alloy
   ```
4. 啟動平台層與示範 HTTP（產生 traces/metrics/logs）：
   ```bash
   go run ./go-platform/cmd/detectviz plugin serve \
     --config ./go-platform/configs/config.yaml \
     --http-demo --http-demo-listen :7777
   curl -sS http://127.0.0.1:7777/hello
   ```
5. 於 Grafana 檢視：Logs/Traces/Profiles 可進行 Drilldown；Profiles 由 Alloy `pyroscope.scrape` 擷取 `pprof`。

---

## 開發流程（摘要）
- **契約先行**：任何跨語言變更請先更新 `contracts/`（proto/schema/samples），再生成並同步下游。
- **生成碼**：在 `contracts/` 執行 `buf lint && buf generate`，禁止手動編輯 `*.pb.go` 等生成檔。
- **插件/Agent 擴增**：依樣板撰寫並附 `module.card.json`，以驗證工具檢查分類與依賴。
- **觀測一致**：Go 端以 OTLP 輸出；Profiles 僅以 `pprof` 暴露，由 Alloy 上送。

---

## 參考
- [`spec.md`](./spec.md)
- [`contracts/README.md`](./contracts/README.md)
- [`go-platform/README.md`](./go-platform/README.md)
- [`python-adk-runtime/README.md`](./python-adk-runtime/README.md)
- [`grafana-alloy/config.alloy`](./grafana-alloy/config.alloy)


---

## 架構圖

```mermaid
flowchart TB
  subgraph Contracts["contracts/<br/>(SSOT)"]
    C1[proto/<br/>- adk_bridge.proto]
    C2[schemas/<br/>- config.schema.json<br/>- module.card.schema.json]
  end

  subgraph Go["go-platform<br/>(Platform Core / ToolBridge)"]
    G1[ToolBridge gRPC Server<br/>internal/pluginhost]
    G2["Plugins (telegraf-like)"]
    G3["OpenTelemetry SDK (traces/metrics)<br/>+ zap logs"]
    G4["pprof :6060"]
  end

  subgraph Py["python-adk-runtime<br/>(ADK Runtime)"]
    P1[Agents\n- coordinator\n- tool-exec]
    P2[Capabilities / Tools]
    P3["RemoteTool (gRPC client)"]
    P4[MemoryBank]
  end

  subgraph Alloy["Grafana Alloy<br/>(Collector)"]
    A1["otelcol.receiver.otlp<br/>:4317 / :4318"]
    A2["loki.source.file<br/>(file tail)"]
    A3["pyroscope.scrape<br/>(pprof)"]
    A4["batch / transform / auth"]
  end

  subgraph Destinations["Observability Backends"]
    D1["Local LGTM<br/>(Prometheus/Tempo/Loki/Pyroscope)"]
    D2["Grafana Cloud<br/>(OTLP/Loki/Pyroscope)"]
    D3["GCP<br/>(Cloud Trace/Logging/Profiler)"]
  end

  P3 -->|gRPC<br/>ToolRequest/ToolChunk| G1
  G2 --> G1
  G3 -->|OTLP gRPC/HTTP| A1
  G3 -->|logs file| A2
  G4 -->|scrape| A3

  A1 -->|traces/metrics| D2
  A2 -->|logs| D2
  A3 -->|profiles| D2

  %% 支援以 mode 切換目標（lgtm_local / grafana_cloud / gcp）
  A1 -.-> D1
  A2 -.-> D1
  A3 -.-> D1

  A1 -.-> D3
  A2 -.-> D3
  A3 -.-> D3

  Contracts -. generate .-> Go
  Contracts -. generate .-> Py
```
