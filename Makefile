# Variables
DOCKER_COMPOSE = docker-compose
CHAT_SERVICE = chat-app
BOT_SERVICE = bot-app

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	docker exec $(CHAT_SERVICE) go mod tidy

# Run chat application locally in a container
run-chat:
	@echo "Starting the chat application using Docker Compose..."
	$(DOCKER_COMPOSE) up --build chat-app

# Run bot application locally in a container
run-bot:
	@echo "Starting the bot application using Docker Compose..."
	$(DOCKER_COMPOSE) up --build bot-app

# Stop running containers
stop:
	@echo "Stopping running containers..."
	$(DOCKER_COMPOSE) down

# Clean up Docker resources
clean:
	@echo "Cleaning up Docker resources..."
	docker system prune -f