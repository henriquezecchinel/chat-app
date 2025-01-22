# Variables
DOCKER_COMPOSE = docker-compose
CHAT_SERVICE = chat-app
BOT_SERVICE = bot-app

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

# Stop running containers
stop:
	@echo "Stopping running containers..."
	$(DOCKER_COMPOSE) down

# Clean up Docker resources
clean:
	@echo "Cleaning up Docker resources..."
	docker system prune -f

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
