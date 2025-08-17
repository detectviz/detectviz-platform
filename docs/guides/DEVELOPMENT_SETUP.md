# 開發環境設置指南

> 📌 **文檔職責**：本文檔提供完整的開發環境設置指南，確保開發者能快速搭建可用的開發環境。

## 🎯 快速開始

### 系統需求
- **作業系統**：Linux, macOS, Windows (WSL2)
- **Go**：1.24+ (必須)
- **Python**：3.11+ (必須)
- **Docker**：20.10+ (推薦)
- **Git**：2.30+ (必須)

### 一鍵設置腳本
```bash
# 下載設置腳本
curl -sSL https://raw.githubusercontent.com/detectviz/platform/main/tools/setup-dev.sh | bash

# 或者手動克隆後執行
git clone <repository-url>
cd detectviz-platform
./tools/setup-dev.sh
```

## 🔧 詳細設置步驟

### 1. 基礎環境準備

#### Go 環境設置
```bash
# 下載並安裝 Go 1.24
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz

# 配置環境變數 (添加到 ~/.bashrc 或 ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin

# 驗證安裝
go version  # 應顯示 go version go1.24 ...
```

#### Python 環境設置
```bash
# 使用 pyenv 管理 Python 版本 (推薦)
curl https://pyenv.run | bash

# 安裝 Python 3.11
pyenv install 3.11.0
pyenv global 3.11.0

# 或使用系統包管理器
# Ubuntu/Debian
sudo apt update && sudo apt install python3.11 python3.11-venv python3.11-dev

# macOS
brew install python@3.11

# 驗證安裝
python3 --version  # 應顯示 Python 3.11.x
```

#### Docker 環境設置
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# macOS
brew install docker docker-compose

# 或下載 Docker Desktop
# https://www.docker.com/products/docker-desktop

# 驗證安裝
docker --version
docker-compose --version
```

### 2. 專案依賴安裝

#### 克隆專案
```bash
git clone <repository-url>
cd detectviz-platform

# 檢查專案結構
ls -la
# 應該看到：contracts/, go-platform/, python-adk-runtime/, docs/ 等目錄
```

#### 安裝 Buf (Protocol Buffers 工具)
```bash
# 安裝 buf
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# 驗證安裝
buf --version
```

#### Go 模組初始化
```bash
cd go-platform

# 下載依賴
go mod download

# 驗證編譯
go build -o detectviz cmd/detectviz/main.go

# 運行測試
go test ./...
```

#### Python 環境設置
```bash
cd python-adk-runtime

# 創建虛擬環境
python3 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 安裝依賴
pip install -r requirements.txt

# 安裝開發依賴
pip install -r requirements-dev.txt

# 以開發模式安裝本地包
pip install -e .
```

### 3. 生成跨語言契約

```bash
cd contracts

# 生成 Protocol Buffers 代碼
buf lint        # 檢查 proto 語法
buf generate    # 生成 Go 和 Python 代碼

# 或使用 Makefile
make gen

# 驗證生成結果
ls -la gen/go/detectviz/contracts/v1/
ls -la gen/python/detectviz/contracts/v1/
```

### 4. 配置管理

#### 創建本地配置
```bash
# 複製範例配置
cp contracts/samples/config.yaml ./config.yaml
cp contracts/samples/.env.template ./.env

# 編輯配置文件
vim config.yaml
```

#### 範例開發配置
```yaml
# config.yaml - 開發環境配置
env: dev

grpc:
  listen: ":5002"
  max_recv_bytes: 4194304
  max_send_bytes: 4194304

observability:
  mode: lgtm_local
  otlp:
    protocol: grpc
    endpoint: "127.0.0.1:4317"
    insecure: true
  logs:
    mode: stdout
  profiling:
    enabled: true
    pprof_address: "127.0.0.1:6060"
    application_name: "detectviz-dev"

plugin:
  paths: 
    - "./go-platform/internal/pluginhost/plugins"
  registry: file

memory:
  backend: inmem
  default_ttl_seconds: 3600
```

#### 環境變數設置
```bash
# .env 文件
export DETECTVIZ_CONFIG_FILE=./config.yaml
export LOG_LEVEL=debug
export DETECTVIZ_ENV=development

# 如果使用外部服務
export INFLUXDB_URL=http://localhost:8086
export INFLUXDB_TOKEN=your-token
export REDIS_URL=redis://localhost:6379
```

### 5. 外部服務設置 (可選)

#### 啟動 LGTM Stack (本地可觀察性)
```bash
# 使用 Docker Compose 啟動
docker-compose -f docker/lgtm-stack.yml up -d

# 檢查服務狀態
docker-compose -f docker/lgtm-stack.yml ps

# 服務端點
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
# Loki: http://localhost:3100
# Tempo: http://localhost:3200
```

#### 啟動 InfluxDB (如需要)
```bash
# 啟動 InfluxDB
docker run -d \
  --name detectviz-influxdb \
  -p 8086:8086 \
  -v influxdb-data:/var/lib/influxdb2 \
  influxdb:2.7

# 初始化設置
docker exec -it detectviz-influxdb influx setup \
  --bucket detectviz \
  --org detectviz \
  --password secretpassword \
  --username admin \
  --force
```

#### 啟動 Redis (如需要)
```bash
# 啟動 Redis
docker run -d \
  --name detectviz-redis \
  -p 6379:6379 \
  redis:7-alpine

# 測試連接
redis-cli ping  # 應回傳 PONG
```

## 🧪 開發環境驗證

### 1. 驗證 Go 平台
```bash
cd go-platform

# 編譯專案
go build -o detectviz cmd/detectviz/main.go

# 驗證配置
./detectviz config validate -f ../config.yaml

# 啟動服務 (背景執行)
./detectviz plugin serve --config ../config.yaml &

# 檢查健康狀態
curl http://localhost:8081/healthz
curl http://localhost:8081/readyz

# 停止服務
pkill detectviz
```

### 2. 驗證 Python Runtime
```bash
cd python-adk-runtime
source venv/bin/activate

# 運行基本測試
python -m pytest test_simple_adk.py -v

# 測試 gRPC 連接
python test_adk_integration.py

# 啟動 Web 服務器 (測試用)
python web_server.py &

# 測試 HTTP 端點
curl http://localhost:8080/health
curl -X POST http://localhost:8080/postmortem \
  -H "Content-Type: application/json" \
  -d '{"incident_id": "test-001", "service": "test-service"}'

# 停止服務
pkill -f web_server.py
```

### 3. 端到端通訊測試
```bash
# 啟動 Go 平台
cd go-platform
./detectviz plugin serve --config ../config.yaml &

# 等待服務啟動
sleep 5

# 啟動 Python Runtime
cd ../python-adk-runtime
python example_usage.py

# 檢查日誌
tail -f ../var/log/detectviz/detectviz.log
```

## 🔧 開發工具設置

### IDE 配置

#### VS Code 設置
```json
// .vscode/settings.json
{
  "go.gopath": "${workspaceFolder}/go-platform",
  "go.goroot": "/usr/local/go",
  "python.defaultInterpreterPath": "./python-adk-runtime/venv/bin/python",
  "python.terminal.activateEnvironment": true,
  "files.associations": {
    "*.proto": "proto"
  },
  "proto.protoc_path": "/usr/local/bin/protoc"
}
```

#### 推薦擴展
```json
// .vscode/extensions.json
{
  "recommendations": [
    "golang.go",
    "ms-python.python",
    "zxh404.vscode-proto3",
    "ms-vscode.makefile-tools",
    "ms-azuretools.vscode-docker"
  ]
}
```

### Git 配置
```bash
# 設置 Git hooks
cp tools/git-hooks/* .git/hooks/
chmod +x .git/hooks/*

# 配置 Git 忽略
cat >> .gitignore << EOF
# 開發環境特定
.env
*.local
debug.log

# IDE 設置
.idea/
*.swp
*.swo

# 編譯產物
go-platform/detectviz
python-adk-runtime/dist/
EOF
```

### 調試設置

#### Go 調試配置
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Go Platform",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/go-platform/cmd/detectviz",
      "args": ["plugin", "serve", "--config", "../config.yaml"],
      "cwd": "${workspaceFolder}/go-platform"
    }
  ]
}
```

#### Python 調試配置
```json
// .vscode/launch.json (添加到 configurations 陣列)
{
  "name": "Launch Python Runtime",
  "type": "python",
  "request": "launch",
  "program": "${workspaceFolder}/python-adk-runtime/example_usage.py",
  "cwd": "${workspaceFolder}/python-adk-runtime",
  "env": {
    "PYTHONPATH": "${workspaceFolder}/python-adk-runtime/src"
  }
}
```

## 📋 開發工作流

### 日常開發流程
```bash
# 1. 同步最新代碼
git pull origin main

# 2. 重新生成契約 (如有 proto 變更)
cd contracts && make gen

# 3. 更新依賴
cd ../go-platform && go mod tidy
cd ../python-adk-runtime && pip install -r requirements.txt

# 4. 運行測試
make test-all

# 5. 啟動開發服務
make dev-start

# 6. 進行開發工作...

# 7. 提交前檢查
make pre-commit-check
```

### 測試命令
```bash
# 所有測試
make test-all

# 僅 Go 測試
make test-go

# 僅 Python 測試  
make test-python

# 整合測試
make test-integration

# 覆蓋率報告
make coverage
```

### 代碼格式化
```bash
# Go 代碼格式化
gofmt -w go-platform/
go mod tidy

# Python 代碼格式化
cd python-adk-runtime
black .
isort .
flake8 .
```

## 🚨 常見問題排除

### Go 相關問題

#### 問題：go mod tidy 失敗
```bash
# 解決方案：清理模組緩存
go clean -modcache
go mod download
go mod tidy
```

#### 問題：編譯錯誤 "package not found"
```bash
# 檢查 GOPATH 和 GOROOT
echo $GOPATH
echo $GOROOT

# 重新設置 Go 環境
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
```

### Python 相關問題

#### 問題：ImportError: No module named 'detectviz_adk'
```bash
# 解決方案：重新安裝本地包
cd python-adk-runtime
pip install -e .

# 檢查 Python 路徑
python -c "import sys; print(sys.path)"
```

#### 問題：gRPC 連接失敗
```bash
# 檢查 Go 服務是否運行
curl http://localhost:8081/healthz

# 檢查埠口占用
netstat -tlpn | grep 5002

# 檢查防火牆設置
sudo ufw status
```

### Docker 相關問題

#### 問題：Docker 權限錯誤
```bash
# 將用戶添加到 docker 組
sudo usermod -aG docker $USER
newgrp docker

# 或使用 sudo 運行 docker 命令
sudo docker ps
```

#### 問題：容器啟動失敗
```bash
# 檢查容器日誌
docker logs detectviz-influxdb

# 檢查埠口衝突
docker ps -a
netstat -tlpn | grep 8086
```

### 協議緩衝區相關問題

#### 問題：buf generate 失敗
```bash
# 檢查 buf 安裝
buf --version

# 檢查 proto 語法
buf lint contracts/proto

# 重新安裝 buf
rm /usr/local/bin/buf
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
```

## 🎯 生產環境準備

### 環境變數清單
```bash
# 必需環境變數
export DETECTVIZ_CONFIG_FILE=/path/to/config.yaml
export LOG_LEVEL=info

# 可選環境變數 (依據配置需要)
export INFLUXDB_URL=https://your-influxdb.com
export INFLUXDB_TOKEN=your-secret-token
export REDIS_URL=redis://your-redis:6379
export GRAFANA_URL=https://your-grafana.com
export GRAFANA_API_KEY=your-api-key
```

### 配置檢查清單
- [ ] 生產配置文件已創建並驗證
- [ ] 所有敏感信息已移至環境變數
- [ ] 日誌級別設置為 info 或 warn
- [ ] 可觀察性端點配置正確
- [ ] 資源限制已配置
- [ ] 安全設置已啟用 (TLS、認證等)

---

**維護說明**：
- 更新頻率：依賴或設置步驟變更時更新
- 維護責任：開發環境維護團隊
- 引用方式：`{{ docs/guides/DEVELOPMENT_SETUP.md#section }}`