.PHONY: *
.DEFAULT_GOAL := help

help: ## Show this help
	@printf "\n\033[37m%s\033[0m\n" 'Usage: make [target]'
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-gofumpt: ## Install gofumpt
	@go install mvdan.cc/gofumpt@latest

fmt: ## Format code
	@gofumpt -l -w .

build: ## Format code
	go build -ldflags "-s -w"

mod: ## go mod tidy
	go mod tidy

lint: ## golangci-lint
	golangci-lint run --out-format tab --sort-results --fix
