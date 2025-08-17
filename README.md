# Detectviz Platform

[![Current Focus: Phase 3](https://img.shields.io/badge/focus-Phase%203%20Postmortem-blue)](./sre-services-map.md)
[![SSOT: contracts](https://img.shields.io/badge/SSOT-contracts-0A84FF)](./contracts)
[![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](#)
[![Python >= 3.11](https://img.shields.io/badge/Python-%3E%3D%203.11-3776AB?logo=python)](#)
[![Google ADK aligned](https://img.shields.io/badge/Google%20ADK-aligned-4285F4?logo=google)](https://google.github.io/adk-docs/)
[![License](https://img.shields.io/badge/license-MIT-blue)](./LICENSE)

基於 Google Agent Development Kit (ADK) 的混合語言智能代理平台，結合 Go 與 Python 實現可觀察性與 SRE 自動化。

> 專案狀態：參見 [專案狀態文檔](docs/status/PROJECT_STATUS.md)  
> 開發指南：參見 [開發者文檔](AGENT.md)  
> 技術規範：參見 [技術規格文檔](spec.md)

## 核心架構

本平台採用混合語言架構設計：
- **contracts/** - 單一事實來源 (SSOT)，定義跨語言契約
- **go-platform/** - 高效能平台核心與 ToolBridge
- **python-adk-runtime/** - ADK Runtime 與多代理協作
- **grafana-alloy/** - 可觀察性收集與轉送

> 詳細架構說明：參見 [技術架構文檔](docs/technical/ARCHITECTURE_OVERVIEW.md)

## 快速開始

### 前置需求
- Go 1.24+
- Python 3.11+
- Docker & Docker Compose

### 安裝步驟
```bash
# 1. 克隆專案
git clone <repository-url>
cd detectviz-platform

# 2. 生成跨語言契約
cd contracts
make gen

# 3. 啟動平台核心
cd ../go-platform
go run main.go

# 4. 啟動 ADK Runtime
cd ../python-adk-runtime
python main.py
```

> 完整設置指南：參見 [開發環境設置](docs/guides/DEVELOPMENT_SETUP.md)

## 專案結構

```
detectviz-platform/
├── contracts/           # SSOT - 跨語言契約定義
├── go-platform/        # Go 平台核心
├── python-adk-runtime/ # Python ADK Runtime
├── grafana-alloy/      # 可觀察性配置
├── docs/               # 文檔體系
└── AGENT.md           # AI 協作指南
```

## 開發文檔

- [專案狀態](docs/status/PROJECT_STATUS.md) - 當前進度與里程碑
- [SRE 服務地圖](sre-services-map.md) - 業務架構與職責分工
- [技術規格](spec.md) - 詳細技術實現規範
- [AI 協作指南](AGENT.md) - 開發守則與最佳實踐
- [契約文檔](contracts/README.md) - SSOT 使用說明

## 核心設計原則

- **SSOT 契約優先** - 所有跨語言介面變更必須先更新 contracts/
- **職責分離** - Agent 負責決策，Tool 負責執行
- **可觀察性優先** - 統一 Logs/Traces/Metrics/Profiles 收集
- **模組化設計** - 以模組卡規範組件分類與依賴關係

## 授權

MIT License - 詳見 [LICENSE](./LICENSE) 檔案