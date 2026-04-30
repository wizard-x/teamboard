.PHONY: help dev dev-backend dev-frontend build build-frontend build-backend test test-backend test-frontend migrate-up migrate-down migrate-create docker-up docker-down docker-build lint lint-backend lint-frontend clean

default: help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Development ──

dev: ## Start backend + frontend in dev mode (requires docker compose for infra)
	@echo "Starting infrastructure..."
	docker compose up -d postgres redis
	@echo "Waiting for services..."
	@sleep 3
	@echo "Running migrations..."
	cd src/backend && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "postgres://trello:trello_secret@localhost:5432/trello?sslmode=disable" up
	@echo "Starting backend and frontend..."
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start backend in dev mode (ensures frontend is built)
	@$(MAKE) --no-print-directory build-frontend-for-embed
	cd src/backend && go run ./cmd/server

build-frontend-for-embed: ## Build frontend and copy dist to backend embed directory
	@if [ ! -f src/frontend/dist/index.html ]; then \
		echo "Building frontend..."; \
		$(MAKE) --no-print-directory build-frontend; \
	fi
	@echo "Copying frontend dist to backend embed directory..."
	@rm -rf src/backend/cmd/server/static/*
	@cp -r src/frontend/dist/* src/backend/cmd/server/static/

dev-frontend: ## Start frontend in dev mode
	cd src/frontend && npm run dev

# ── Build ──

build-frontend: ## Build frontend
	cd src/frontend && npm ci && npm run build

build-backend: ## Build backend binary
	cd src/backend && go build -o ../../bin/teamboard ./cmd/server

build: build-frontend build-backend ## Build everything

# ── Tests ──

test-backend: ## Run backend tests
	cd src/backend && go test ./... -v -count=1

test-frontend: ## Run frontend tests
	cd src/frontend && npm test

test: test-backend test-frontend ## Run all tests

# ── Database ──

DSN = postgres://trello:trello_secret@localhost:5432/trello?sslmode=disable

migrate-up: ## Run database migrations
	cd src/backend && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DSN)" up

migrate-down: ## Rollback last migration
	cd src/backend && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DSN)" down

migrate-create: ## Create a new migration (usage: make migrate-create name=add_index)
	cd src/backend && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations create $(name) sql

# ── Docker ──

docker-build: ## Build Docker image
	docker build -t teamboard .

docker-up: ## Start all services with Docker Compose
	docker compose up -d --build

docker-down: ## Stop Docker Compose services
	docker compose down

# ── Lint ──

lint-backend: ## Lint backend code
	cd src/backend && go vet ./...

lint-frontend: ## Lint frontend code
	cd src/frontend && npx vue-tsc --noEmit

lint: lint-backend lint-frontend ## Lint everything

# ── Clean ──

clean: ## Clean build artifacts
	rm -rf bin/
	cd src/frontend && rm -rf dist node_modules/.vite
