# Go Chat Application

## Description
This project is a browser-based chat application built using Go. It is designed to facilitate communication among users in a chatroom, with the ability to fetch stock quotes using a specific command format.

## Features

### Mandatory Features
- **User Authentication:** Registered users can log in to participate in the chat.
- **Real-Time Chatroom:** Users can send and receive messages in a shared chatroom.
- **Stock Quote Command:** Users can post messages in the format `/stock=stock_code` to request stock quotes.
- **Bot Integration:**
    - A decoupled bot retrieves stock information using the Stooq API.
    - Parses the CSV response and sends stock quote messages back to the chatroom using RabbitMQ.
    - Bot posts messages in the format: `"APPL.US quote is $93.42 per share"`.
- **Message Ordering:** Chat messages are displayed in chronological order.
- **Message History:** The application shows only the last 50 messages.
- **Unit Testing:** Key functionalities are tested to ensure reliability.

## Technology Stack
- **Language:** Go
- **Message Broker:** RabbitMQ
- **API Integration:** Stooq API for stock quotes
- **Frontend:** Minimal HTML/CSS for simplicity
- **Testing Framework:** Go's native `testing` package

## Installation and Setup

### Prerequisites
- [Go](https://golang.org/dl/) installed
- [Docker](https://www.docker.com/) installed
- [Make](https://www.gnu.org/software/make/manual/make.html/) installed (The README has its commands based on Make, if you don't want to install it, please refer to the Makefile to find the commands)
- [RabbitMQ](https://www.rabbitmq.com/) running (provided in the docker file)
- [PostgreSQL](https://www.postgresql.org/) running (provided in the docker file)


### Steps
Create a `.env` file in the root directory based on the provided `.env.example` file.

Run `make run-chat` to start the chat application.

Run `make run-bot` to start the bot application.

If you have issues with dependencies, run `make install` to install the dependencies.

#### chat-app
- The chat application runs on `localhost:8080`.
- Simply open the provided web interface in your browser to start chatting.
1. Register a new user
2. Log in with the registered user
3. Create or join a chatroom
4. Start chatting!

#### bot-app
- The bot application runs on `localhost:8081`.
- The bot listens to messages in the `stock_requests` queue and responds with stock quotes in the `stock_responses` queue.
- Simply send the `/stock=stock_code` command in the chatroom to receive stock quotes.
- Example: `/stock=aapl.us`


## Testing
Work in progress!


## Notes
- The `/stock` command does not persist in the database.
- The frontend is intentionally minimal, focusing on backend functionality.


## Author
- [Henrique Zecchinel](mailto:henriquezecchinel@gmail.com)
