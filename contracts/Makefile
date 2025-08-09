.PHONY: gen lint breaking

gen:
	buf generate

lint:
	buf lint

breaking:
	buf breaking --against '.git#branch=main'

.PHONY: gen lint breaking validate validate-cards help

## Generate Go/Python stubs from proto (buf.gen.yaml)
gen:
	buf generate

## Lint proto (buf)
lint:
	buf lint

## Breaking change check against main (buf)
breaking:
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

help:
	@echo "Targets:"
	@echo "  gen              Generate stubs via buf"
	@echo "  lint             Lint proto via buf"
	@echo "  breaking         Check API breaking changes against main"
	@echo "  validate-cards   Validate all module.card.json across repo"
	@echo "  validate         Run lint + validate-cards"