.PHONY: gen lint breaking

gen:
	buf generate

lint:
	buf lint

breaking:
	buf breaking --against '.git#branch=main'
