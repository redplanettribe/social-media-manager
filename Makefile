# Variables
IMAGE_NAME = my-postgres-image
CONTAINER_NAME = my-postgres-container

# Default target
.PHONY: start
start:
	@echo "Composing local containers..."
	docker compose up -d
	@echo "Running backend with Air..."
	make watch

.PHONY: watch
watch:
	@echo "Running Air..."
	air -c .air.toml
.PHONY: stop
stop:
	@echo "Stopping local containers..."
	docker compose down



# Migrations
.PHONY: create-migration
create-migration:
	@echo "Creating new migration..."
	migrate create -ext sql -dir ./internal/infrastructure/persistence/migrations -seq $(name)

.PHONY: migrate-up
migrate-up:
	@echo "Running migrations..."
	migrate -path ./internal/infrastructure/persistence/migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable" up $(step)

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back migrations..."
	migrate -path ./internal/infrastructure/persistence/migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable" down $(step)

# Find directories with Go files containing //go:generate directives
GENERATE_DIRS := $(sort $(dir $(shell find . -name '*.go' -exec grep -l '^//go:generate' {} +)))

.PHONY: generate-mocks
generate-mocks:
	@for dir in $(GENERATE_DIRS); do \
		echo "Running go generate in $$dir"; \
		( cd $$dir && go generate ); \
	done

.PHONY: generate-api-docs
generate-api-docs:
	@echo "Generating API documentation..."
	swag init -g cmd/server/main.go


.PHONY: lint
lint:
	@echo "Running goimports..."
	goimports -w .
	@echo "Running gofumpt..."
	gofumpt -w -l .
	@echo "Running linter..."
	golangci-lint run