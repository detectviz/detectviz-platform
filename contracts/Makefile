.PHONY: gen lint breaking

gen:
	buf generate

lint:
	buf lint

breaking:
	buf breaking --against '.git#branch=main'

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
	    python3 tools/validate_module_card.py "$$f"; \
	  done; \
	fi

## Full validation: proto lint + module cards
validate: lint validate-cards

## Comprehensive validation with version check
validate-with-versions: check-versions validate
	@echo "✓ All validations passed with version verification"

help:
	@echo "Detectviz Contracts Build System"
	@echo "Version Management & Code Generation Targets:"
	@echo ""
	@echo "  install-tools    Install required tools with pinned versions"
	@echo "  check-versions   Verify tool versions for consistency"
	@echo "  gen              Generate stubs via buf (with version verification)"
	@echo "  lint             Lint proto via buf"
	@echo "  breaking         Check API breaking changes against main"
	@echo "  validate-cards   Validate all module.card.json across repo"
	@echo "  validate         Run lint + validate-cards"
	@echo "  validate-with-versions  Full validation with version checks"
	@echo ""
	@echo "Pinned Versions:"
	@echo "  buf: $(BUF_VERSION)"
	@echo "  protoc-gen-go: $(PROTOC_GEN_GO_VERSION)"
	@echo "  protoc-gen-go-grpc: $(PROTOC_GEN_GO_GRPC_VERSION)"