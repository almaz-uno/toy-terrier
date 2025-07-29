.PHONY: build run test clean lint docker-build docker-run help generate build-full version app-help test-coverage deps env-setup migrate-up migrate-down migrate-create docker-stop docker-compose-up docker-compose-down docker-compose-logs dev install-tools release git-hooks

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=toy-terrier-bot
BINARY_PATH=./cmd/toy-terrier-bot

# Configuration
CONFIG ?= config.yaml
ENV_FILE ?= .env

# Docker parameters
DOCKER_IMAGE=toy-terrier-bot
DOCKER_TAG=latest

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@egrep '^(.+)\:\ ##\ (.+)' $(MAKEFILE_LIST) | column -t -c 2 -s ':#'

$(BINARY_NAME): ## Build the application
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH)

build: $(BINARY_NAME) ## Build the binary (alias for $(BINARY_NAME))

generate: ## Generate sqlc code
	sqlc generate

build-full: generate build ## Generate code and build

run: build ## Run the application (usage: make run [CONFIG=path/to/config.yaml] [ENV_FILE=path/to/.env])
	@if [ -f "$(ENV_FILE)" ]; then \
		echo "Loading environment variables from $(ENV_FILE)"; \
		export $$(grep -v '^#' $(ENV_FILE) | xargs) && ./$(BINARY_NAME) --config=$(CONFIG); \
	else \
		echo "Environment file $(ENV_FILE) not found, running without env file"; \
		./$(BINARY_NAME) --config=$(CONFIG); \
	fi

version: ## Show version
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH)
	@if [ -f "$(ENV_FILE)" ]; then \
		export $$(grep -v '^#' $(ENV_FILE) | xargs) && ./$(BINARY_NAME) --version; \
	else \
		./$(BINARY_NAME) --version; \
	fi

app-help: ## Show application help
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH)
	@if [ -f "$(ENV_FILE)" ]; then \
		export $$(grep -v '^#' $(ENV_FILE) | xargs) && ./$(BINARY_NAME) --help; \
	else \
		./$(BINARY_NAME) --help; \
	fi

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

lint: ## Run linter
	golangci-lint run

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

env-setup: ## Create .env file from .env.example
	@if [ ! -f ".env" ]; then \
		if [ -f ".env.example" ]; then \
			cp .env.example .env; \
			echo "Created .env file from .env.example"; \
			echo "Please edit .env file with your actual values"; \
		else \
			echo ".env.example file not found"; \
			exit 1; \
		fi; \
	else \
		echo ".env file already exists"; \
	fi

# Database operations
migrate-up: ## Run database migrations up
	migrate -path ./internal/database/migrations -database "postgres://postgres:password@localhost:5432/toy_terrier?sslmode=disable" up

migrate-down: ## Run database migrations down
	migrate -path ./internal/database/migrations -database "postgres://postgres:password@localhost:5432/toy_terrier?sslmode=disable" down

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	migrate create -ext sql -dir ./internal/database/migrations -seq $(NAME)

# Docker operations
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	docker run -d --name toy-terrier \
		-p 8080:8080 \
		-v $(PWD)/config.yaml:/app/config.yaml:ro \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop: ## Stop Docker container
	docker stop toy-terrier
	docker rm toy-terrier

docker-compose-up: ## Start services with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	docker-compose down

docker-compose-logs: ## Show docker-compose logs
	docker-compose logs -f

# Development operations
dev: run ## Run in development mode

install-tools: ## Install development tools
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Release operations
release: clean lint test build ## Prepare release build

# Git operations
git-hooks: ## Install git hooks
	cp scripts/pre-commit .git/hooks/
	chmod +x .git/hooks/pre-commit
