.PHONY: *
.DEFAULT_GOAL := help

help: ## Show this help
	@printf "\n\033[37m%s\033[0m\n" 'Usage: make [target]'
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build binary
	go build -ldflags "-s -w"

mod: ## go mod tidy
	go mod tidy

fmt: ## Format code
	@go tool gofumpt -l -w .

lint: ## golangci-lint
	go tool golangci-lint run --out-format tab --sort-results --fix
