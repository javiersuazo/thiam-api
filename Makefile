ifneq ($(wildcard .env),)
include .env
export
else
$(warning WARNING: .env file not found! Using .env.example)
include .env.example
export
endif

BASE_STACK = docker compose -f docker-compose.yml
INTEGRATION_TEST_STACK = $(BASE_STACK) -f docker-compose-integration-test.yml
ALL_STACK = $(INTEGRATION_TEST_STACK)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Run docker compose (without backend and reverse proxy)
	$(BASE_STACK) up --build -d db rabbitmq nats && docker compose logs -f
.PHONY: compose-up

compose-up-all: ### Run docker compose (with backend and reverse proxy)
	$(BASE_STACK) up --build -d
.PHONY: compose-up-all

compose-up-integration-test: ### Run docker compose with integration test
	$(INTEGRATION_TEST_STACK) up --build --abort-on-container-exit --exit-code-from integration-test
.PHONY: compose-up-integration-test

compose-down: ### Down docker compose
	$(ALL_STACK) down --remove-orphans
.PHONY: compose-down

swag-v1: ### swag init
	swag init -g internal/controller/http/router.go
.PHONY: swag-v1

proto-v1: ### generate source files from proto
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		docs/proto/v1/*.proto
.PHONY: proto-v1

deps: ### deps tidy + verify
	go mod tidy && go mod verify
.PHONY: deps

deps-audit: ### check dependencies vulnerabilities
	govulncheck ./...
.PHONY: deps-audit

format: ### Run code formatter
	gofumpt -l -w .
	gci write . --skip-generated -s standard -s default
.PHONY: format

run: deps swag-v1 proto-v1 ### swag run for API v1
	go mod download && \
	CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run

docker-rm-volume: ### remove docker volume
	docker volume rm go-clean-template_pg-data
.PHONY: docker-rm-volume

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

linter-hadolint: ### check by hadolint linter
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
.PHONY: linter-hadolint

linter-dotenv: ### check by dotenv linter
	dotenv-linter
.PHONY: linter-dotenv

test: ### run test
	go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/...
.PHONY: test

integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...
.PHONY: integration-test

mock: ### run mockgen
	mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
	mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go
.PHONY: mock

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations '$(word 2,$(MAKECMDGOALS))'
.PHONY: migrate-create

migrate-up: ### migration up
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' up
.PHONY: migrate-up

bin-deps: ### install tools
	go install tool
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
.PHONY: bin-deps

pre-commit: swag-v1 proto-v1 mock format linter-golangci test ### run pre-commit
.PHONY: pre-commit

# ==============================================================================
# Developer Experience
# ==============================================================================

setup: ### one-command dev environment setup
	@echo "Installing Go tools..."
	@go install tool
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
	@go install github.com/air-verse/air@latest
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 && pre-commit install || echo "pre-commit not found. Install with: brew install pre-commit"
	@echo "Copying environment file..."
	@test -f .env || cp .env.example .env
	@echo "Setup complete! Run 'make dev' to start developing."
.PHONY: setup

dev: ### run app with hot-reload
	@command -v air >/dev/null 2>&1 || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air
.PHONY: dev

doctor: ### diagnose development environment
	@echo "Checking development environment..."
	@echo ""
	@echo "Go:"
	@go version || echo "  ERROR: Go not found"
	@echo ""
	@echo "Docker:"
	@docker --version || echo "  ERROR: Docker not found"
	@docker compose version || echo "  ERROR: Docker Compose not found"
	@echo ""
	@echo "Go Tools:"
	@command -v golangci-lint >/dev/null 2>&1 && echo "  golangci-lint: OK" || echo "  golangci-lint: MISSING (run: make bin-deps)"
	@command -v air >/dev/null 2>&1 && echo "  air: OK" || echo "  air: MISSING (run: go install github.com/air-verse/air@latest)"
	@command -v swag >/dev/null 2>&1 && echo "  swag: OK" || echo "  swag: MISSING (run: make bin-deps)"
	@command -v mockgen >/dev/null 2>&1 && echo "  mockgen: OK" || echo "  mockgen: MISSING (run: make bin-deps)"
	@command -v migrate >/dev/null 2>&1 && echo "  migrate: OK" || echo "  migrate: MISSING (run: make bin-deps)"
	@command -v protoc >/dev/null 2>&1 && echo "  protoc: OK" || echo "  protoc: MISSING (install protobuf compiler)"
	@echo ""
	@echo "Optional Tools:"
	@command -v pre-commit >/dev/null 2>&1 && echo "  pre-commit: OK" || echo "  pre-commit: MISSING (run: brew install pre-commit)"
	@command -v direnv >/dev/null 2>&1 && echo "  direnv: OK" || echo "  direnv: MISSING (run: brew install direnv)"
	@echo ""
	@echo "Environment:"
	@test -f .env && echo "  .env file: OK" || echo "  .env file: MISSING (run: cp .env.example .env)"
.PHONY: doctor

clean: ### clean generated files and caches
	@echo "Cleaning generated files..."
	@rm -rf tmp/
	@rm -rf coverage.txt
	@rm -rf build-errors.log
	@go clean -cache -testcache
	@echo "Clean complete."
.PHONY: clean

reset-db: ### reset database to clean state
	@echo "Stopping containers..."
	@$(BASE_STACK) down db -v 2>/dev/null || true
	@echo "Starting fresh database..."
	@$(BASE_STACK) up -d db
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Running migrations..."
	@$(MAKE) migrate-up
	@echo "Database reset complete."
.PHONY: reset-db

logs: ### tail application logs
	@$(BASE_STACK) logs -f app
.PHONY: logs

test-watch: ### run tests in watch mode
	@command -v watchexec >/dev/null 2>&1 || (echo "Installing watchexec..." && brew install watchexec)
	watchexec -e go -r -- go test -v -race ./internal/...
.PHONY: test-watch

env-check: ### validate required environment variables
	@echo "Checking required environment variables..."
	@test -n "$$PG_URL" || (echo "ERROR: PG_URL is not set" && exit 1)
	@test -n "$$HTTP_PORT" || (echo "ERROR: HTTP_PORT is not set" && exit 1)
	@echo "All required environment variables are set."
.PHONY: env-check

security: ### run security checks
	@echo "Running security checks..."
	@govulncheck ./...
	@echo "Security checks complete."
.PHONY: security
