# Variables
DOCKER_COMPOSE = docker-compose
CHAT_SERVICE = chat-app
BOT_SERVICE = bot-app
TEXT_SERVICE = text-app

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	go mod tidy

# Run chat application locally in a container
run-chat:
	@echo "Starting the chat application using Docker Compose..."
	$(DOCKER_COMPOSE) up --build $(CHAT_SERVICE)

# Run bot application locally in a container
run-bot:
	@echo "Starting the bot application using Docker Compose..."
	$(DOCKER_COMPOSE) up --build $(BOT_SERVICE)

# Run text application locally in a container
run-text:
	@echo "Starting the text application using Docker Compose..."
	$(DOCKER_COMPOSE) up --build $(TEXT_SERVICE)

# Stop running containers
stop:
	@echo "Stopping running containers..."
	$(DOCKER_COMPOSE) down

# Clean up Docker resources (BE CAREFUL! IT WILL REMOVE ALL DOCKER RESOURCES!)
clean:
	@echo "Cleaning up Docker resources..."
	# Stop all running containers
	@if [ -n "$$(docker ps -q)" ]; then docker stop $$(docker ps -q); fi

	# Remove all containers
	@if [ -n "$$(docker ps -aq)" ]; then docker rm $$(docker ps -aq); fi

	# Remove all images
	@if [ -n "$$(docker images -q)" ]; then docker rmi $$(docker images -q); fi

	# Remove all volumes
	@if [ -n "$$(docker volume ls -q)" ]; then docker volume rm $$(docker volume ls -q); fi

	# Remove all user-defined networks
	@if [ -n "$$(docker network ls -q -f 'type=custom')" ]; then docker network rm $$(docker network ls -q -f 'type=custom'); fi

	# Clean up any remaining Docker resources
	docker system prune -a --volumes -f

# Run linter
lint:
	@echo "Running linter..."
	go mod verify
	golangci-lint run -c .golangci.yml

# Run linter with --fix flag
lint-fix:
	@echo "Running linter with --fix flag..."
	go mod verify
	golangci-lint run -c .golangci.yml --fix

# Format code
format:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

test:
	@echo "Running tests..."
	go test ./...