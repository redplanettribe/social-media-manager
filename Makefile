# Variables
IMAGE_NAME = my-postgres-image
CONTAINER_NAME = my-postgres-container

# Default target
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
stop:
	@echo "Stopping Docker container..."
	docker stop $(CONTAINER_NAME)
	@echo "Removing Docker container..."
	docker rm $(CONTAINER_NAME)

clean:
	@echo "Removing Docker image..."
	docker rmi $(IMAGE_NAME)
