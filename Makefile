.PHONY: help run-backend test-backend fmt lint docker-up docker-down docker-logs docker-ps health ready

# Default goal
.DEFAULT_GOAL := help

help: ## Display available make targets
	@echo "Recess Development Makefile"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run-backend: ## Run Go backend locally on host
	@echo "Starting Recess backend..."
	cd backend && go run ./cmd/server

test-backend: ## Run all Go backend unit and integration tests
	@echo "Running Go tests..."
	cd backend && go test -v -race -cover ./...

fmt: ## Format Go source code
	@echo "Formatting Go files..."
	cd backend && gofmt -w -s .

lint: ## Run Go vet
	@echo "Running go vet..."
	cd backend && go vet ./...

docker-up: ## Start all infrastructure services with Docker Compose
	@echo "Starting Docker Compose services..."
	docker compose up -d --build

docker-down: ## Stop all Docker Compose services and remove containers
	@echo "Stopping Docker Compose services..."
	docker compose down

docker-logs: ## Follow Docker Compose logs
	docker compose logs -f

docker-ps: ## Check status and health of all Docker Compose containers
	docker compose ps

health: ## Check application liveness
	@curl -s http://localhost:8080/health | jq . || curl -s http://localhost:8080/health

ready: ## Check application and infrastructure readiness (Postgres & Redis)
	@curl -s http://localhost:8080/ready | jq . || curl -s http://localhost:8080/ready
