.PHONY: help run test test-race coverage vet build fmt lint docker-up docker-down

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "%-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the API
	go run ./cmd/api

test: ## Run tests
	go test ./...

test-race: ## Run tests with the race detector
	go test -race ./...

coverage: ## Run tests and write coverage.out
	go test -v -race -coverprofile=coverage.out ./...

vet: ## Run go vet
	go vet ./...

build: ## Build the API
	go build ./cmd/api

fmt: ## Format Go files
	gofmt -w ./cmd ./internal

lint: ## Run golangci-lint when installed
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "golangci-lint is not installed; skipping lint"; fi

docker-up: ## Start the API with Docker Compose
	docker compose up --build

docker-down: ## Stop Docker Compose services
	docker compose down