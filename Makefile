.PHONY: *
.DEFAULT_GOAL := help

APP_NAME := telegram-owl
MAIN_PACKAGE := ./cmd/telegram-owl
BUILD_DIR := dist
GOEXE := $(shell go env GOEXE)
BUILD_OUTPUT := $(BUILD_DIR)/$(APP_NAME)$(GOEXE)

help: ## Show this help
	@printf "\n\033[37m%s\033[0m\n" 'Usage: make [target]'
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: export CGO_ENABLED := 0
build: ## Build local production-style binary
	@mkdir -p $(BUILD_DIR)
	go build -trimpath -ldflags "-s -w" -o $(BUILD_OUTPUT) $(MAIN_PACKAGE)

release-local: ## Build local snapshot release without publishing
	GORELEASER_LOCAL=1 goreleaser release --snapshot --clean --skip=before

mod: ## go mod tidy
	go mod tidy

fmt: ## Format code
	@go tool gofumpt -l -w .

test: ## Run tests
	go test ./...

lint: ## golangci-lint
	go tool golangci-lint run --fix
