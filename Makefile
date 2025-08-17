# Detectviz Platform - 根目錄 Makefile
# 整合各模組的自動化工具，提供統一的開發者介面

.PHONY: help validate-with-versions validate-cards health-check-proto maintain-proto
.PHONY: generate-module-card fix-module-cards lint gen breaking check-versions
.PHONY: setup-development validate-implementation fix-common-issues

## 顯示所有可用命令
help:
	@echo "Detectviz Platform - 統一開發介面"
	@echo ""
	@echo "🔧 主要開發命令："
	@echo "  validate-implementation  完整實作驗證 (推薦)"
	@echo "  setup-development       設置開發環境"
	@echo "  fix-common-issues       修復常見問題"
	@echo ""
	@echo "📋 Contract 管理："
	@echo "  validate-with-versions  完整驗證 (含版本檢查)" 
	@echo "  validate-cards          驗證所有模組卡"
	@echo "  generate-module-card    生成新模組卡"
	@echo "  fix-module-cards        修復模組卡問題"
	@echo ""
	@echo "🛠️ Proto 工具："
	@echo "  health-check-proto      Proto 健康檢查"
	@echo "  maintain-proto          Proto 維護工作流"
	@echo "  lint                    Proto 語法檢查"
	@echo "  gen                     生成 Proto 程式碼"
	@echo ""
	@echo "Usage: make generate-module-card NAME=<name> ROLE=<role> CATEGORY=<category> DESC='<description>'"

## 完整實作驗證 - AI 協作者主要使用
validate-implementation:
	@echo "🚀 執行完整實作驗證..."
	@cd contracts && $(MAKE) validate-with-versions
	@echo "✅ 實作驗證完成"

## 設置開發環境
setup-development:
	@echo "🔧 設置開發環境..."
	@cd contracts && $(MAKE) install-tools
	@echo "✅ 開發環境設置完成"

## 修復常見問題
fix-common-issues:
	@echo "🔧 修復常見問題..."
	@cd contracts && $(MAKE) fix-module-cards
	@cd contracts && $(MAKE) fix-proto
	@echo "✅ 常見問題修復完成"

## Contract 相關命令 - 委派到 contracts/Makefile
validate-with-versions:
	@cd contracts && $(MAKE) validate-with-versions

validate-cards:
	@cd contracts && $(MAKE) validate-cards

generate-module-card:
	@cd contracts && $(MAKE) generate-module-card NAME="$(NAME)" ROLE="$(ROLE)" CATEGORY="$(CATEGORY)" DESC="$(DESC)" OUTPUT="$(OUTPUT)"

fix-module-cards:
	@cd contracts && $(MAKE) fix-module-cards

health-check-proto:
	@cd contracts && $(MAKE) health-check-proto

maintain-proto:
	@cd contracts && $(MAKE) maintain-proto

lint:
	@cd contracts && $(MAKE) lint

gen:
	@cd contracts && $(MAKE) gen

breaking:
	@cd contracts && $(MAKE) breaking

check-versions:
	@cd contracts && $(MAKE) check-versions