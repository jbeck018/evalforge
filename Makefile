# EvalForge Development Makefile
# Outstanding Developer Experience - One command to rule them all

.PHONY: help dev dev-up dev-down dev-reset dev-seed dev-clean test test-unit test-integration \
        build build-backend build-frontend build-mock-llm \
        fmt lint lint-go lint-js \
        profile profile-cpu profile-mem profile-trace \
        logs logs-api logs-worker logs-clickhouse logs-postgres \
        db-reset db-migrate db-seed \
        deps deps-go deps-js \
        clean docker-clean \
        status health check

# Colors for pretty output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[1;37m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

# Configuration
COMPOSE_FILE := docker-compose.yml
COMPOSE_CMD := docker-compose -f $(COMPOSE_FILE)
BACKEND_DIR := backend
FRONTEND_DIR := frontend
DEV_DIR := dev

# Development environment variables
export EVALFORGE_ENV := development
export EVALFORGE_LOG_LEVEL := debug
export EVALFORGE_MOCK_LLMS := true

## Help
help: ## Show this help message
	@echo "$(CYAN)EvalForge Development Commands$(NC)"
	@echo "$(BLUE)=============================$(NC)"
	@echo ""
	@echo "$(GREEN)🚀 Quick Start:$(NC)"
	@echo "  $(WHITE)make dev$(NC)          - Start complete development environment"
	@echo "  $(WHITE)make dev-reset$(NC)    - Reset and restart with fresh data"
	@echo "  $(WHITE)make status$(NC)       - Check status of all services"
	@echo ""
	@echo "$(GREEN)📋 Available Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(WHITE)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(GREEN)🔧 Service URLs:$(NC)"
	@echo "  $(WHITE)Frontend:$(NC)       http://localhost:3000"
	@echo "  $(WHITE)API Server:$(NC)     http://localhost:8000"
	@echo "  $(WHITE)API Docs:$(NC)       http://localhost:8089"
	@echo "  $(WHITE)Mock OpenAI:$(NC)    http://localhost:8080"
	@echo "  $(WHITE)Mock Anthropic:$(NC) http://localhost:8081"
	@echo "  $(WHITE)ClickHouse:$(NC)     http://localhost:8123"
	@echo "  $(WHITE)PostgreSQL:$(NC)     localhost:5432"
	@echo "  $(WHITE)Grafana:$(NC)        http://localhost:3001"
	@echo "  $(WHITE)Prometheus:$(NC)     http://localhost:9090"
	@echo "  $(WHITE)Jaeger:$(NC)         http://localhost:16686"
	@echo "  $(WHITE)MinIO Console:$(NC)  http://localhost:9001"
	@echo "  $(WHITE)MailHog:$(NC)        http://localhost:8025"

## Development Environment
dev: ## Start complete development environment (one command to rule them all)
	@echo "$(CYAN)🚀 Starting EvalForge development environment...$(NC)"
	@$(MAKE) -s deps-check
	@$(MAKE) -s dev-up
	@$(MAKE) -s wait-for-services
	@$(MAKE) -s dev-seed
	@echo ""
	@echo "$(GREEN)✅ EvalForge is ready for development!$(NC)"
	@echo ""
	@echo "$(BLUE)📱 Frontend:$(NC)      http://localhost:3000"
	@echo "$(BLUE)🔧 API Docs:$(NC)      http://localhost:8089"
	@echo "$(BLUE)📊 Grafana:$(NC)       http://localhost:3001 (admin/evalforge_dev)"
	@echo "$(BLUE)📈 Prometheus:$(NC)    http://localhost:9090"
	@echo "$(BLUE)🔍 Jaeger:$(NC)        http://localhost:16686"
	@echo ""
	@echo "$(YELLOW)💡 Tip: Run 'make logs' to see service logs$(NC)"
	@echo "$(YELLOW)💡 Tip: Run 'make status' to check service health$(NC)"

dev-up: ## Start all services in background
	@echo "$(BLUE)🐳 Starting Docker services...$(NC)"
	@$(COMPOSE_CMD) up -d
	@echo "$(GREEN)✅ Services started$(NC)"

dev-down: ## Stop all services
	@echo "$(BLUE)🛑 Stopping all services...$(NC)"
	@$(COMPOSE_CMD) down
	@echo "$(GREEN)✅ Services stopped$(NC)"

dev-reset: ## Reset environment with fresh data
	@echo "$(YELLOW)🔄 Resetting development environment...$(NC)"
	@$(COMPOSE_CMD) down -v
	@docker system prune -f --volumes
	@$(MAKE) -s dev-up
	@$(MAKE) -s wait-for-services
	@$(MAKE) -s dev-seed
	@echo "$(GREEN)✅ Environment reset complete!$(NC)"

dev-clean: ## Clean up development environment completely
	@echo "$(RED)🧹 Cleaning up development environment...$(NC)"
	@$(COMPOSE_CMD) down -v --remove-orphans
	@docker system prune -af --volumes
	@docker volume prune -f
	@echo "$(GREEN)✅ Cleanup complete$(NC)"

dev-seed: ## Seed databases with development data
	@echo "$(BLUE)🌱 Seeding databases with development data...$(NC)"
	@echo "$(CYAN)  - PostgreSQL already seeded via init scripts$(NC)"
	@echo "$(CYAN)  - Generating ClickHouse events...$(NC)"
	@cd $(DEV_DIR)/data-generator && go run . --events-per-day=1000 --days-back=7 --verbose
	@echo "$(GREEN)✅ Database seeding complete$(NC)"

## Service Management
status: ## Check status of all services
	@echo "$(CYAN)📊 Service Status$(NC)"
	@echo "$(BLUE)===============$(NC)"
	@$(COMPOSE_CMD) ps
	@echo ""
	@$(MAKE) -s health

health: ## Check health of all services
	@echo "$(CYAN)🏥 Health Check$(NC)"
	@echo "$(BLUE)===============$(NC)"
	@echo -n "$(WHITE)PostgreSQL:$(NC)     "
	@curl -s -f http://localhost:5432 > /dev/null 2>&1 && echo "$(GREEN)✅ Healthy$(NC)" || echo "$(RED)❌ Unhealthy$(NC)"
	@echo -n "$(WHITE)ClickHouse:$(NC)     "
	@curl -s -f http://localhost:8123/ping > /dev/null 2>&1 && echo "$(GREEN)✅ Healthy$(NC)" || echo "$(RED)❌ Unhealthy$(NC)"
	@echo -n "$(WHITE)Redis:$(NC)          "
	@docker exec evalforge_redis redis-cli ping > /dev/null 2>&1 && echo "$(GREEN)✅ Healthy$(NC)" || echo "$(RED)❌ Unhealthy$(NC)"
	@echo -n "$(WHITE)Mock LLM:$(NC)       "
	@curl -s -f http://localhost:8080/health > /dev/null 2>&1 && echo "$(GREEN)✅ Healthy$(NC)" || echo "$(RED)❌ Unhealthy$(NC)"
	@echo -n "$(WHITE)MinIO:$(NC)          "
	@curl -s -f http://localhost:9000/minio/health/live > /dev/null 2>&1 && echo "$(GREEN)✅ Healthy$(NC)" || echo "$(RED)❌ Unhealthy$(NC)"
	@echo -n "$(WHITE)Grafana:$(NC)        "
	@curl -s -f http://localhost:3001/api/health > /dev/null 2>&1 && echo "$(GREEN)✅ Healthy$(NC)" || echo "$(RED)❌ Unhealthy$(NC)"

wait-for-services: ## Wait for all services to be healthy
	@echo "$(BLUE)⏳ Waiting for services to be ready...$(NC)"
	@echo -n "$(WHITE)PostgreSQL$(NC)"; while ! docker exec evalforge_postgres pg_isready -U evalforge > /dev/null 2>&1; do echo -n "."; sleep 1; done; echo " $(GREEN)✅$(NC)"
	@echo -n "$(WHITE)ClickHouse$(NC)"; while ! curl -s -f http://localhost:8123/ping > /dev/null 2>&1; do echo -n "."; sleep 1; done; echo " $(GREEN)✅$(NC)"
	@echo -n "$(WHITE)Redis$(NC)"; while ! docker exec evalforge_redis redis-cli ping > /dev/null 2>&1; do echo -n "."; sleep 1; done; echo " $(GREEN)✅$(NC)"
	@echo -n "$(WHITE)Mock LLM$(NC)"; while ! curl -s -f http://localhost:8080/health > /dev/null 2>&1; do echo -n "."; sleep 2; done; echo " $(GREEN)✅$(NC)"
	@echo "$(GREEN)🎉 All services are ready!$(NC)"

## Logs
logs: ## Show logs from all services
	@$(COMPOSE_CMD) logs -f --tail=100

logs-api: ## Show API server logs
	@echo "$(BLUE)📋 API Server Logs$(NC)"
	@echo "$(BLUE)==================$(NC)"
	@$(COMPOSE_CMD) logs -f api

logs-worker: ## Show background worker logs
	@echo "$(BLUE)📋 Worker Logs$(NC)"
	@echo "$(BLUE)==============$(NC)"
	@$(COMPOSE_CMD) logs -f worker

logs-clickhouse: ## Show ClickHouse logs
	@echo "$(BLUE)📋 ClickHouse Logs$(NC)"
	@echo "$(BLUE)==================$(NC)"
	@$(COMPOSE_CMD) logs -f clickhouse

logs-postgres: ## Show PostgreSQL logs
	@echo "$(BLUE)📋 PostgreSQL Logs$(NC)"
	@echo "$(BLUE)==================$(NC)"
	@$(COMPOSE_CMD) logs -f postgres

## Database Management
db-reset: ## Reset all databases
	@echo "$(YELLOW)🔄 Resetting databases...$(NC)"
	@$(COMPOSE_CMD) exec postgres psql -U evalforge -d evalforge -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(COMPOSE_CMD) exec clickhouse clickhouse-client --query="DROP DATABASE IF EXISTS evalforge; CREATE DATABASE evalforge;"
	@$(MAKE) -s db-migrate
	@$(MAKE) -s db-seed
	@echo "$(GREEN)✅ Databases reset$(NC)"

db-migrate: ## Run database migrations
	@echo "$(BLUE)🔄 Running database migrations...$(NC)"
	@echo "$(CYAN)  - PostgreSQL migrations via init scripts$(NC)"
	@echo "$(CYAN)  - ClickHouse migrations via init scripts$(NC)"
	@echo "$(GREEN)✅ Migrations complete$(NC)"

db-seed: ## Seed databases with test data
	@$(MAKE) -s dev-seed

## Testing
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "$(BLUE)🧪 Running unit tests...$(NC)"
	@cd $(BACKEND_DIR) && go test -short -v ./...
	@cd $(FRONTEND_DIR) && npm test -- --watchAll=false
	@echo "$(GREEN)✅ Unit tests passed$(NC)"

test-integration: ## Run integration tests
	@echo "$(BLUE)🔬 Running integration tests...$(NC)"
	@cd $(BACKEND_DIR) && go test -v ./tests/integration/...
	@echo "$(GREEN)✅ Integration tests passed$(NC)"

test-load: ## Run load tests
	@echo "$(BLUE)⚡ Running load tests...$(NC)"
	@cd $(DEV_DIR)/scripts && ./load-test.sh
	@echo "$(GREEN)✅ Load tests complete$(NC)"

## Building
build: build-backend build-frontend build-mock-llm ## Build all components

build-backend: ## Build backend services
	@echo "$(BLUE)🔨 Building backend services...$(NC)"
	@cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api
	@cd $(BACKEND_DIR) && go build -o bin/worker ./cmd/worker
	@echo "$(GREEN)✅ Backend built$(NC)"

build-frontend: ## Build frontend application
	@echo "$(BLUE)🔨 Building frontend application...$(NC)"
	@cd $(FRONTEND_DIR) && npm run build
	@echo "$(GREEN)✅ Frontend built$(NC)"

build-mock-llm: ## Build mock LLM service
	@echo "$(BLUE)🔨 Building mock LLM service...$(NC)"
	@cd $(DEV_DIR)/mock-llm && go build -o mock-llm .
	@echo "$(GREEN)✅ Mock LLM built$(NC)"

## Code Quality
fmt: fmt-go fmt-js ## Format all code

fmt-go: ## Format Go code
	@echo "$(BLUE)🎨 Formatting Go code...$(NC)"
	@cd $(BACKEND_DIR) && gofmt -w .
	@cd $(BACKEND_DIR) && goimports -w .
	@echo "$(GREEN)✅ Go code formatted$(NC)"

fmt-js: ## Format JavaScript/TypeScript code
	@echo "$(BLUE)🎨 Formatting JS/TS code...$(NC)"
	@cd $(FRONTEND_DIR) && npm run format
	@echo "$(GREEN)✅ JS/TS code formatted$(NC)"

lint: lint-go lint-js ## Lint all code

lint-go: ## Lint Go code
	@echo "$(BLUE)🔍 Linting Go code...$(NC)"
	@cd $(BACKEND_DIR) && golangci-lint run
	@echo "$(GREEN)✅ Go linting passed$(NC)"

lint-js: ## Lint JavaScript/TypeScript code
	@echo "$(BLUE)🔍 Linting JS/TS code...$(NC)"
	@cd $(FRONTEND_DIR) && npm run lint
	@echo "$(GREEN)✅ JS/TS linting passed$(NC)"

## Performance Profiling
profile: ## Start profiling session
	@echo "$(BLUE)📊 Starting performance profiling...$(NC)"
	@echo "$(CYAN)Available profiles:$(NC)"
	@echo "  $(WHITE)make profile-cpu$(NC)    - CPU profiling"
	@echo "  $(WHITE)make profile-mem$(NC)    - Memory profiling"
	@echo "  $(WHITE)make profile-trace$(NC)  - Execution tracing"

profile-cpu: ## Run CPU profiling
	@echo "$(BLUE)📊 Starting CPU profiling...$(NC)"
	@cd $(BACKEND_DIR) && go test -cpuprofile=cpu.prof -bench=. ./internal/...
	@cd $(BACKEND_DIR) && go tool pprof -http=:8091 cpu.prof &
	@echo "$(GREEN)🔍 CPU profile available at http://localhost:8091$(NC)"

profile-mem: ## Run memory profiling
	@echo "$(BLUE)📊 Starting memory profiling...$(NC)"
	@cd $(BACKEND_DIR) && go test -memprofile=mem.prof -bench=. ./internal/...
	@cd $(BACKEND_DIR) && go tool pprof -http=:8092 mem.prof &
	@echo "$(GREEN)🔍 Memory profile available at http://localhost:8092$(NC)"

profile-trace: ## Run execution tracing
	@echo "$(BLUE)📊 Starting execution tracing...$(NC)"
	@cd $(BACKEND_DIR) && go test -trace=trace.out -bench=. ./internal/...
	@cd $(BACKEND_DIR) && go tool trace trace.out &
	@echo "$(GREEN)🔍 Trace viewer will open in browser$(NC)"

## Dependencies
deps: deps-go deps-js deps-tools ## Install all dependencies

deps-go: ## Install Go dependencies
	@echo "$(BLUE)📦 Installing Go dependencies...$(NC)"
	@cd $(BACKEND_DIR) && go mod download
	@cd $(BACKEND_DIR) && go mod tidy
	@cd $(DEV_DIR)/mock-llm && go mod download
	@cd $(DEV_DIR)/data-generator && go mod download
	@echo "$(GREEN)✅ Go dependencies installed$(NC)"

deps-js: ## Install JavaScript dependencies
	@echo "$(BLUE)📦 Installing JavaScript dependencies...$(NC)"
	@cd $(FRONTEND_DIR) && npm ci
	@echo "$(GREEN)✅ JavaScript dependencies installed$(NC)"

deps-tools: ## Install development tools
	@echo "$(BLUE)🔧 Installing development tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)✅ Development tools installed$(NC)"

deps-check: ## Check if required dependencies are installed
	@echo "$(BLUE)🔍 Checking dependencies...$(NC)"
	@command -v docker >/dev/null 2>&1 || { echo "$RED❌ Docker is required but not installed$(NC)"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "$(RED)❌ Docker Compose is required but not installed$(NC)"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "$(RED)❌ Go 1.21+ is required but not installed$(NC)"; exit 1; }
	@command -v node >/dev/null 2>&1 || { echo "$(RED)❌ Node.js 18+ is required but not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✅ All dependencies are available$(NC)"

## Utilities
clean: ## Clean build artifacts
	@echo "$(BLUE)🧹 Cleaning build artifacts...$(NC)"
	@rm -rf $(BACKEND_DIR)/bin/
	@rm -rf $(FRONTEND_DIR)/dist/
	@rm -rf $(FRONTEND_DIR)/build/
	@cd $(BACKEND_DIR) && go clean -cache -testcache -modcache
	@echo "$(GREEN)✅ Build artifacts cleaned$(NC)"

docker-clean: ## Clean Docker resources
	@echo "$(BLUE)🧹 Cleaning Docker resources...$(NC)"
	@docker system prune -f
	@docker volume prune -f
	@echo "$(GREEN)✅ Docker resources cleaned$(NC)"

check: ## Run all checks (lint, test, format)
	@echo "$(CYAN)🔍 Running all quality checks...$(NC)"
	@$(MAKE) -s fmt
	@$(MAKE) -s lint
	@$(MAKE) -s test-unit
	@echo "$(GREEN)✅ All checks passed!$(NC)"

## Quick Actions
quick-test: ## Quick test of ingestion pipeline
	@echo "$(BLUE)⚡ Quick ingestion test...$(NC)"
	@curl -X POST http://localhost:8000/api/v1/events \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ef_dev_test_key" \
		-d '{"trace_id":"test_trace","span_id":"test_span","project_id":"p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1","model":"gpt-4","tokens":150,"latency_ms":523,"cost_cents":45}' \
		&& echo "$(GREEN)✅ Ingestion test passed$(NC)" || echo "$(RED)❌ Ingestion test failed$(NC)"

perf-test: ## Run performance test
	@echo "$(BLUE)⚡ Running performance test...$(NC)"
	@cd $(DEV_DIR)/scripts && ./perf-test.sh

## Documentation
docs: ## Generate API documentation
	@echo "$(BLUE)📚 Generating API documentation...$(NC)"
	@cd $(BACKEND_DIR) && swag init -g cmd/api/main.go -o docs/
	@echo "$(GREEN)✅ Documentation generated$(NC)"
	@echo "$(CYAN)📖 Available at http://localhost:8089$(NC)"

## Troubleshooting
troubleshoot: ## Run troubleshooting diagnostics
	@echo "$(CYAN)🔧 EvalForge Troubleshooting$(NC)"
	@echo "$(BLUE)============================$(NC)"
	@echo ""
	@echo "$(WHITE)1. System Information:$(NC)"
	@echo "   Docker version: $$(docker --version)"
	@echo "   Docker Compose version: $$(docker-compose --version)"
	@echo "   Go version: $$(go version 2>/dev/null || echo 'Not installed')"
	@echo "   Node version: $$(node --version 2>/dev/null || echo 'Not installed')"
	@echo ""
	@echo "$(WHITE)2. Service Status:$(NC)"
	@$(MAKE) -s status
	@echo ""
	@echo "$(WHITE)3. Port Availability:$(NC)"
	@for port in 3000 5432 6379 8000 8080 8081 8123 9000; do \
		if lsof -Pi :$$port -sTCP:LISTEN -t >/dev/null 2>&1; then \
			echo "   Port $$port: $(GREEN)✅ In use$(NC)"; \
		else \
			echo "   Port $$port: $(YELLOW)⚠️  Available$(NC)"; \
		fi; \
	done
	@echo ""
	@echo "$(WHITE)4. Common Issues:$(NC)"
	@echo "   $(CYAN)• Port conflicts:$(NC) Stop other services using ports 3000, 5432, 6379, 8000, 8080, 8081, 8123, 9000"
	@echo "   $(CYAN)• Docker issues:$(NC) Try 'make dev-clean' then 'make dev'"
	@echo "   $(CYAN)• Permission issues:$(NC) Ensure Docker daemon is running and you have permissions"
	@echo ""
	@echo "$(GREEN)💡 Need help? Check the logs with 'make logs'$(NC)"

# Include optional local overrides
-include Makefile.local