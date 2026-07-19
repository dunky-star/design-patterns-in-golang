PACKAGES ?= ./...
TEST_FLAGS ?=

.PHONY: help run test test-race test-cover migrateup migratedown migrateforce

help: ## Show available commands
	@echo "Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Run the web application with local media directories
	go run ./cmd/web -media-input ./input -media-output ./output

test: ## Run local tests (override with TEST_FLAGS or PACKAGES)
	go test $(TEST_FLAGS) $(PACKAGES)

test-race: ## Run tests with the race detector
	go test -race -vet=off $(TEST_FLAGS) $(PACKAGES)

test-cover: ## Run tests with coverage reporting
	go test -cover $(TEST_FLAGS) $(PACKAGES)

migrateup: ## Run database migrations up (loads .env if present)
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$$DSN" ]; then echo "Error: set DSN in .env"; exit 1; fi; \
	migrate -path db/migrations -database "mysql://$$DSN" -verbose up

migratedown: ## Run database migrations down (loads .env if present)
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$$DSN" ]; then echo "Error: set DSN in .env"; exit 1; fi; \
	migrate -path db/migrations -database "mysql://$$DSN" -verbose down

migrateforce: VERSION ?= 1
migrateforce: ## Fix a dirty migration version: make migrateforce VERSION=1
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$$DSN" ]; then echo "Error: set DSN in .env"; exit 1; fi; \
	migrate -path db/migrations -database "mysql://$$DSN" force $(VERSION)
