# Version control and consistency settings
BUF_VERSION := v1.47.2
PROTOC_GEN_GO_VERSION := v1.34.2
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1
PROTOC_GEN_PYTHON_VERSION := v5.28.3

.PHONY: gen lint breaking validate validate-cards help check-versions install-tools

## Check tool versions for consistency
check-versions:
	@echo "Checking tool versions..."
	@command -v buf >/dev/null 2>&1 || { echo "buf not found. Run 'make install-tools' first."; exit 1; }
	@BUF_ACTUAL=$$(buf --version 2>&1 | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1); \
	if [ "$$BUF_ACTUAL" != "$(BUF_VERSION)" ]; then \
		echo "WARNING: buf version mismatch. Expected: $(BUF_VERSION), Found: $$BUF_ACTUAL"; \
		echo "Run 'make install-tools' to install correct versions."; \
	else \
		echo "✓ buf version: $$BUF_ACTUAL"; \
	fi

## Install required tools with pinned versions
install-tools:
	@echo "Installing required tools with pinned versions..."
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Please install Go first."; exit 1; }
	curl -sSL https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/buf-$$(uname -s)-$$(uname -m) -o /tmp/buf && \
		sudo mv /tmp/buf /usr/local/bin/buf && \
		sudo chmod +x /usr/local/bin/buf
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@echo "✓ Tools installed with pinned versions"

## Generate Go/Python stubs from proto (buf.gen.yaml) with version verification
gen: check-versions
	@echo "Generating proto stubs..."
	buf generate
	@echo "Generating version metadata..."
	@mkdir -p gen/metadata
	@echo '{"buf_version": "$(BUF_VERSION)", "generated_at": "'$$(date -Iseconds)'", "proto_hash": "'$$(find proto -name "*.proto" -exec sha256sum {} \; | sha256sum | cut -d' ' -f1)'"}' > gen/metadata/version.json
	@echo "✓ Generated stubs with version metadata"

## Lint proto (buf)
lint: check-versions
	buf lint

## Breaking change check against main (buf)
breaking: check-versions
	buf breaking --against '.git#branch=main'

## Generate a new module card with validation
generate-module-card:
	@echo "Generating new module card..."
	@if [ -z "$(NAME)" ] || [ -z "$(ROLE)" ] || [ -z "$(CATEGORY)" ] || [ -z "$(DESC)" ]; then \
		echo "Usage: make generate-module-card NAME=<name> ROLE=<role> CATEGORY=<category> DESC='<description>' [OUTPUT=<path>]"; \
		echo "Valid roles: agent.coordinator, agent.tool_exec, tool, capability, plugin.gateway, memory.backend, security.module, observability.module, storage.module"; \
		echo "Valid categories: collector.input, transform.processor, aggregate.aggregator, sink.output, gateway, llm, retriever, workflow, a2a, capability, memory.backend, security.authn, security.authz, observability.exporter, observability.processor, storage.blob, storage.kv, storage.vector"; \
		exit 1; \
	fi
	@OUTPUT_PATH=$${OUTPUT:-$(NAME)_module.card.json}; \
	python3 tools/generate_module_card.py "$(NAME)" "$(ROLE)" "$(CATEGORY)" "$(DESC)" "$$OUTPUT_PATH"; \
	python3 tools/validate_module_card.py "$$OUTPUT_PATH"

## Validate all module.card.json across the repo using contracts/tools/validate_module_card.py
validate-cards:
	@echo "Validating Module Cards..."
	@set -e; \
	FILES=$$(find .. -type f -name 'module.card.json'); \
	if [ -z "$$FILES" ]; then \
	  echo "No module.card.json found."; \
	else \
	  for f in $$FILES; do \
	    echo " - $$f"; \
	    python3 tools/validate_module_card.py "$$f" || exit 1; \
	  done; \
	fi

## Fix common module card issues automatically
fix-module-cards:
	@echo "Fixing common module card issues..."
	@FILES=$$(find .. -type f -name 'module.card.json'); \
	for f in $$FILES; do \
		echo "Checking $$f..."; \
		python3 tools/fix_module_card.py "$$f"; \
	done

## Full validation: proto lint + module cards
validate: lint validate-cards

## Comprehensive validation with version check
validate-with-versions: check-versions validate
	@echo "✓ All validations passed with version verification"

## Health check for proto files
health-check-proto:
	@echo "Checking proto files health..."
	@find proto -name "*.proto" -exec python3 tools/proto_health_check.py {} \; | grep -E "(❌|✅)" || true

## Auto-fix common proto issues
fix-proto:
	@echo "Auto-fixing proto issues..."
	@echo "⚠️  This is a future enhancement - currently only validates"
	@$(MAKE) health-check-proto

## Full proto maintenance workflow
maintain-proto: health-check-proto lint
	@echo "✅ Proto maintenance completed"

help:
	@echo "Detectviz Contracts Build System"
	@echo "Version Management & Code Generation Targets:"
	@echo ""
	@echo "  install-tools       Install required tools with pinned versions"
	@echo "  check-versions      Verify tool versions for consistency"
	@echo "  gen                 Generate stubs via buf (with version verification)"
	@echo "  lint                Lint proto via buf"
	@echo "  breaking            Check API breaking changes against main"
	@echo "  validate-cards      Validate all module.card.json across repo"
	@echo "  generate-module-card Generate new module card with validation"
	@echo "  fix-module-cards    Fix common module card issues automatically"
	@echo "  health-check-proto  Check proto files for common issues"
	@echo "  fix-proto           Auto-fix common proto issues"
	@echo "  maintain-proto      Full proto maintenance workflow"
	@echo "  validate            Run lint + validate-cards"
	@echo "  validate-with-versions Full validation with version checks"
	@echo ""
	@echo "Pinned Versions:"
	@echo "  buf: $(BUF_VERSION)"
	@echo "  protoc-gen-go: $(PROTOC_GEN_GO_VERSION)"
	@echo "  protoc-gen-go-grpc: $(PROTOC_GEN_GO_GRPC_VERSION)"