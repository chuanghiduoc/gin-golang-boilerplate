.PHONY: build run test clean docker-build docker-up docker-down migrate-up migrate-down migrate-create sqlc swagger lint help dev dev-setup dev-docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Binary names
API_BINARY=bin/api
MIGRATE_BINARY=bin/migrate

# Default target
.DEFAULT_GOAL := help

## Build commands
build: ## Build the API binary
	$(GOBUILD) -o $(API_BINARY) ./cmd/api
	$(GOBUILD) -o $(MIGRATE_BINARY) ./cmd/migrate

run: ## Run the API server locally
	$(GOCMD) run ./cmd/api

## Test commands
test: ## Run all tests
	$(GOTEST) -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Clean commands
clean: ## Remove build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

## Docker commands
docker-build: ## Build Docker image
	docker build -t backend-gin:latest .

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker container logs
	docker-compose logs -f api

docker-restart: ## Restart Docker containers
	docker-compose restart

## Database commands
migrate-up: ## Run all up migrations
	$(GOCMD) run ./cmd/migrate -direction up

migrate-down: ## Run all down migrations
	$(GOCMD) run ./cmd/migrate -direction down

migrate-down-one: ## Run one down migration
	$(GOCMD) run ./cmd/migrate -direction down -steps 1

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	migrate create -ext sql -dir db/migrations -seq $(name)

## Code generation
sqlc: ## Generate SQLC code
	sqlc generate

swagger: ## Generate Swagger documentation
	swag init -g cmd/api/main.go -o docs

## Code quality
lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	$(GOFMT) ./...

## Dependencies
deps: ## Download dependencies
	$(GOMOD) download

tidy: ## Tidy dependencies
	$(GOMOD) tidy

## Development
dev: ## Run with hot reload (requires air: go install github.com/air-verse/air@latest)
	air

dev-setup: ## Install development tools (air, sqlc, swag, migrate, golangci-lint)
	go install github.com/air-verse/air@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ All dev tools installed!"

dev-docker: ## Run development with Docker (PostgreSQL + hot reload)
	docker-compose -f docker-compose.dev.yml up

## Help
help: ## Display this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
