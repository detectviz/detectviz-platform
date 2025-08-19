# Detectviz Platform Makefile

.PHONY: help setup start stop restart status logs clean test build run-go run-python

# 預設目標
.DEFAULT_GOAL := help

# 顏色定義
YELLOW := \033[1;33m
GREEN := \033[0;32m
RED := \033[0;31m
NC := \033[0m # No Color

# 變數
<<<<<<< HEAD
DOCKER_COMPOSE := docker-compose
=======
DOCKER_COMPOSE := docker-compose --env-file ../.env
>>>>>>> 08fa581 (update)
GO_CMD := go
PYTHON_CMD := python3
PROJECT_NAME := detectviz

# Help 目標
help: ## 顯示此幫助信息
	@echo "$(GREEN)Detectviz Platform - Makefile 命令$(NC)"
	@echo "======================================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# ============================================
# 環境設置
# ============================================

setup: ## 初始化開發環境（第一次運行）
	@echo "$(YELLOW)🚀 初始化開發環境...$(NC)"
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

env: ## 檢查環境變數
	@echo "$(YELLOW)📋 環境變數檢查...$(NC)"
<<<<<<< HEAD
	@test -f .env || (echo "$(RED)❌ .env 文件不存在，請先運行 make setup$(NC)" && exit 1)
	@echo "$(GREEN)✅ .env 文件存在$(NC)"
	@grep -E "^[A-Z_]+=" .env | head -10
=======
	@test -f ../.env || (echo "$(RED)❌ .env 文件不存在，請先運行 make setup$(NC)" && exit 1)
	@echo "$(GREEN)✅ .env 文件存在$(NC)"
	@grep -E "^[A-Z_]+=" ../.env | head -10
>>>>>>> 08fa581 (update)

# ============================================
# Docker 服務管理
# ============================================

start: ## 啟動所有服務
	@echo "$(YELLOW)🚀 啟動服務...$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✅ 服務啟動完成$(NC)"
	@make status

stop: ## 停止所有服務
	@echo "$(YELLOW)⏹️  停止服務...$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✅ 服務已停止$(NC)"

restart: stop start ## 重啟所有服務

status: ## 檢查服務狀態
	@echo "$(YELLOW)📊 服務狀態：$(NC)"
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "$(GREEN)服務健康檢查：$(NC)"
	@curl -s http://localhost:9090/-/healthy > /dev/null 2>&1 && echo "✅ Prometheus: 健康" || echo "❌ Prometheus: 不健康"
	@curl -s http://localhost:3000/api/health > /dev/null 2>&1 && echo "✅ Grafana: 健康" || echo "❌ Grafana: 不健康"
	@curl -s http://localhost:3100/ready > /dev/null 2>&1 && echo "✅ Loki: 健康" || echo "❌ Loki: 不健康"
	@docker exec detectviz-postgres pg_isready > /dev/null 2>&1 && echo "✅ PostgreSQL: 健康" || echo "❌ PostgreSQL: 不健康"
	@docker exec detectviz-redis redis-cli ping > /dev/null 2>&1 && echo "✅ Redis: 健康" || echo "❌ Redis: 不健康"

logs: ## 查看所有服務日誌
	@$(DOCKER_COMPOSE) logs -f --tail=100

logs-prometheus: ## 查看 Prometheus 日誌
	@$(DOCKER_COMPOSE) logs -f prometheus --tail=100

logs-grafana: ## 查看 Grafana 日誌
	@$(DOCKER_COMPOSE) logs -f grafana --tail=100

clean: ## 清理所有數據（危險！）
	@echo "$(RED)⚠️  警告：這將刪除所有數據！$(NC)"
	@read -p "確定要繼續嗎？(yes/no): " confirm; \
	if [ "$confirm" = "yes" ]; then \
		$(DOCKER_COMPOSE) down -v; \
		rm -rf data/*; \
		echo "$(GREEN)✅ 數據已清理$(NC)"; \
	fi

# ============================================
# Go 開發
# ============================================

build-go: ## 構建 Go 平台
	@echo "$(YELLOW)🔨 構建 Go 平台...$(NC)"
	@cd go-platform && $(GO_CMD) build -o bin/detectviz cmd/detectviz/main.go
	@echo "$(GREEN)✅ Go 平台構建完成$(NC)"

run-go: ## 運行 Go 平台
	@echo "$(YELLOW)🚀 啟動 Go 平台...$(NC)"
	@cd go-platform && $(GO_CMD) run cmd/detectviz/main.go plugin serve --config ../config.yaml

test-go: ## 運行 Go 測試
	@echo "$(YELLOW)🧪 運行 Go 測試...$(NC)"
	@cd go-platform && $(GO_CMD) test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✅ Go 測試完成$(NC)"

test-go-metrics: ## 測試 MetricsProvider
	@echo "$(YELLOW)🧪 測試 MetricsProvider...$(NC)"
	@cd go-platform && $(GO_CMD) test -v ./internal/metrics/... -run TestProvider

test-go-health: ## 測試 HealthAggregator
	@echo "$(YELLOW)🧪 測試 HealthAggregator...$(NC)"
	@cd go-platform && $(GO_CMD) test -v ./internal/pluginhost/plugins/observability/health_aggregator/...

lint-go: ## Go 代碼檢查
	@echo "$(YELLOW)🔍 Go 代碼檢查...$(NC)"
	@cd go-platform && golangci-lint run
	@echo "$(GREEN)✅ Go 代碼檢查完成$(NC)"

# ============================================
# Python 開發
# ============================================

build-python: ## 構建 Python ADK
	@echo "$(YELLOW)🔨 構建 Python ADK...$(NC)"
	@cd python-adk-runtime && pip install -e .
	@echo "$(GREEN)✅ Python ADK 構建完成$(NC)"

run-python: ## 運行 Python ADK
	@echo "$(YELLOW)🚀 啟動 Python ADK...$(NC)"
	@cd python-adk-runtime && $(PYTHON_CMD) -m detectviz_adk.main

test-python: ## 運行 Python 測試
	@echo "$(YELLOW)🧪 運行 Python 測試...$(NC)"
	@cd python-adk-runtime && pytest -v --cov=detectviz_adk tests/
	@echo "$(GREEN)✅ Python 測試完成$(NC)"

lint-python: ## Python 代碼檢查
	@echo "$(YELLOW)🔍 Python 代碼檢查...$(NC)"
	@cd python-adk-runtime && black --check . && flake8 . && mypy .
	@echo "$(GREEN)✅ Python 代碼檢查完成$(NC)"

# ============================================
# 整合測試
# ============================================

test-e2e: ## 運行端到端測試
	@echo "$(YELLOW)🧪 運行端到端測試...$(NC)"
	@make start
	@sleep 10  # 等待服務啟動
	@cd tests/e2e && $(PYTHON_CMD) test_postmortem_flow.py
	@echo "$(GREEN)✅ 端到端測試完成$(NC)"

test-integration: ## 運行整合測試
	@echo "$(YELLOW)🧪 運行整合測試...$(NC)"
	@make test-prometheus-integration
	@make test-postgres-integration
	@make test-redis-integration
	@echo "$(GREEN)✅ 整合測試完成$(NC)"

test-prometheus-integration: ## 測試 Prometheus 整合
	@echo "$(YELLOW)測試 Prometheus 連接...$(NC)"
	@curl -s http://localhost:9090/api/v1/query?query=up | jq '.status' | grep -q success && \
		echo "$(GREEN)✅ Prometheus API 正常$(NC)" || \
		echo "$(RED)❌ Prometheus API 異常$(NC)"

test-postgres-integration: ## 測試 PostgreSQL 整合
	@echo "$(YELLOW)測試 PostgreSQL 連接...$(NC)"
	@docker exec detectviz-postgres psql -U detectviz -d detectviz -c "SELECT 1" > /dev/null 2>&1 && \
		echo "$(GREEN)✅ PostgreSQL 連接正常$(NC)" || \
		echo "$(RED)❌ PostgreSQL 連接異常$(NC)"

test-redis-integration: ## 測試 Redis 整合
	@echo "$(YELLOW)測試 Redis 連接...$(NC)"
	@docker exec detectviz-redis redis-cli SET test_key test_value > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Redis 連接正常$(NC)" || \
		echo "$(RED)❌ Redis 連接異常$(NC)"

# ============================================
# 開發工具
# ============================================

dev: ## 開發模式（啟動所有服務並監控日誌）
	@make start
	@make logs

dev-reload: ## 熱重載開發（Go + Python）
	@echo "$(YELLOW)🔄 啟動熱重載開發模式...$(NC)"
	@tmux new-session -d -s detectviz-dev
	@tmux split-window -h
	@tmux send-keys -t detectviz-dev:0.0 'cd go-platform && air' C-m
	@tmux send-keys -t detectviz-dev:0.1 'cd python-adk-runtime && watchmedo auto-restart -d src -p "*.py" -- python -m detectviz_adk.main' C-m
	@tmux attach -t detectviz-dev

port-forward: ## 端口轉發（用於 K8s 環境）
	@echo "$(YELLOW)🔗 設置端口轉發...$(NC)"
	@kubectl port-forward svc/prometheus 9090:9090 &
	@kubectl port-forward svc/grafana 3000:3000 &
	@echo "$(GREEN)✅ 端口轉發已設置$(NC)"

# ============================================
# 監控和調試
# ============================================

metrics: ## 查看當前指標
	@echo "$(YELLOW)📊 當前指標：$(NC)"
	@curl -s http://localhost:9090/api/v1/query?query=up | jq '.data.result[] | {job: .metric.job, instance: .metric.instance, value: .value[1]}'

traces: ## 查看追蹤數據
	@echo "$(YELLOW)🔍 開啟 Tempo UI...$(NC)"
	@open http://localhost:3200 || xdg-open http://localhost:3200

profiles: ## 查看性能分析
	@echo "$(YELLOW)🔥 開啟 Pyroscope UI...$(NC)"
	@open http://localhost:4040 || xdg-open http://localhost:4040

dashboard: ## 開啟 Grafana Dashboard
	@echo "$(YELLOW)📊 開啟 Grafana...$(NC)"
	@open http://localhost:3000 || xdg-open http://localhost:3000

# ============================================
# 文檔生成
# ============================================

docs: ## 生成文檔
	@echo "$(YELLOW)📚 生成文檔...$(NC)"
	@cd go-platform && godoc -http=:6060 &
	@cd python-adk-runtime && sphinx-build -b html docs/ docs/_build/
	@echo "$(GREEN)✅ 文檔生成完成$(NC)"
	@echo "Go 文檔: http://localhost:6060"
	@echo "Python 文檔: file://$(PWD)/python-adk-runtime/docs/_build/index.html"

# ============================================
# 發布和部署
# ============================================

build-images: ## 構建 Docker 映像
	@echo "$(YELLOW)🐳 構建 Docker 映像...$(NC)"
	@docker build -t detectviz/go-platform:latest -f go-platform/Dockerfile go-platform/
	@docker build -t detectviz/python-adk:latest -f python-adk-runtime/Dockerfile python-adk-runtime/
	@echo "$(GREEN)✅ Docker 映像構建完成$(NC)"

push-images: ## 推送 Docker 映像
	@echo "$(YELLOW)📤 推送 Docker 映像...$(NC)"
	@docker push detectviz/go-platform:latest
	@docker push detectviz/python-adk:latest
	@echo "$(GREEN)✅ Docker 映像推送完成$(NC)"

deploy-k8s: ## 部署到 Kubernetes
	@echo "$(YELLOW)☸️  部署到 Kubernetes...$(NC)"
	@kubectl apply -f k8s/
	@echo "$(GREEN)✅ Kubernetes 部署完成$(NC)"

# ============================================
# 效能測試
# ============================================

benchmark: ## 運行效能測試
	@echo "$(YELLOW)⚡ 運行效能測試...$(NC)"
	@cd go-platform && $(GO_CMD) test -bench=. -benchmem ./...
	@echo "$(GREEN)✅ 效能測試完成$(NC)"

load-test: ## 運行負載測試
	@echo "$(YELLOW)🔨 運行負載測試...$(NC)"
	@vegeta attack -duration=30s -rate=100 -targets=tests/load/targets.txt | vegeta report
	@echo "$(GREEN)✅ 負載測試完成$(NC)"

# ============================================
# 工具安裝
# ============================================

install-tools: ## 安裝開發工具
	@echo "$(YELLOW)🔧 安裝開發工具...$(NC)"
	# Go 工具
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	# Python 工具
	@pip install black flake8 mypy pytest pytest-cov watchdog
	# 其他工具
	@brew install jq tmux vegeta || apt-get install jq tmux vegeta
	@echo "$(GREEN)✅ 開發工具安裝完成$(NC)"

# ============================================
# 快捷命令
# ============================================

up: start ## 啟動服務（別名）
down: stop ## 停止服務（別名）
ps: status ## 查看狀態（別名）
build: build-go build-python ## 構建所有組件
test: test-go test-python ## 運行所有測試
lint: lint-go lint-python ## 運行所有代碼檢查