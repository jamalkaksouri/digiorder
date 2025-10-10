# ===============================
# Makefile Bash-friendly
# Auto-load .env.example
# ===============================

# Load environment variables from .env.example
ifneq (,$(wildcard .env.example))
	include .env.example
	export $(shell sed 's/=.*//' .env.example)
endif

# ===============================
# Phony targets
# ===============================
.PHONY: help build run test clean migrate-up migrate-down sqlc docker-up docker-down install-tools mod-tidy lint fmt

# -------------------------------
# Help
# -------------------------------
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# -------------------------------
# Build & Run
# -------------------------------
build: ## Build the application
	go build -o bin/digiorder cmd/main.go

run: ## Run the application
	go run cmd/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/

# -------------------------------
# Database Migrations
# -------------------------------
migrate-up: ## Run database migrations
	@echo "Running migrations using database URL:"
	@echo "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down: ## Rollback database migrations
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

# -------------------------------
# Docker Setup
# -------------------------------
docker-up: ## Start PostgreSQL in Docker
	docker run -d \
		--name digiorder-postgres \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		postgres:15-alpine

docker-down: ## Stop PostgreSQL container
	docker stop digiorder-postgres || true
	docker rm digiorder-postgres || true

# -------------------------------
# Development Tools
# -------------------------------
install-tools: ## Install development tools
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc: ## Generate SQL code
	sqlc generate

mod-tidy: ## Tidy and vendor Go modules
	go mod tidy
	go mod vendor

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	gofumpt -w .
