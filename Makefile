# Variables
IMAGE_NAME = my-postgres-image
CONTAINER_NAME = my-postgres-container

# Default target
.PHONY: start
start:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME) .
	@echo "Running Docker container..."
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p 5432:5432 \
		--env-file .env \
		-v my_dbdata:/var/lib/postgresql/data-$(IMAGE_NAME) \
		$(IMAGE_NAME)
.PHONY: watch
watch:
	@echo "Running Air..."
	air -c .air.toml
.PHONY: stop
stop:
	@echo "Stopping Docker container..."
	docker stop $(CONTAINER_NAME)
	@echo "Removing Docker container..."
	docker rm $(CONTAINER_NAME)
.PHONY: clean
clean:
	@echo "Removing Docker image..."
	docker rmi $(IMAGE_NAME)

# Migrations
.PHONY: create
create:
	@echo "Creating new migration..."
	migrate create -ext sql -dir ./internal/infrastructure/persistence/migrations -seq $(name)

.PHONY: migrate
migrate:
	@echo "Running migrations..."
	migrate -path ./internal/infrastructure/persistence/migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable" up

.PHONY: rollback
rollback:
	@echo "Rolling back migrations..."
	migrate -path ./internal/infrastructure/persistence/migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable" down

# Find directories with Go files containing //go:generate directives
GENERATE_DIRS := $(sort $(dir $(shell find . -name '*.go' -exec grep -l '^//go:generate' {} +)))

.PHONY: generate
generate:
	@for dir in $(GENERATE_DIRS); do \
		echo "Running go generate in $$dir"; \
		( cd $$dir && go generate ); \
	done


.PHONY: lint
lint:
	@echo "Running goimports..."
	goimports -w .
	@echo "Running gofumpt..."
	gofumpt -w -l .
	@echo "Running linter..."
	golangci-lint run