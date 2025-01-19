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
- [RabbitMQ](https://www.rabbitmq.com/) running locally or accessible remotely

### Steps
Work in progress!

## Testing
Work in progress!

## API Usage
- Stock quotes can be fetched using the command format `/stock=stock_code`.
- Example: `/stock=aapl.us`

## Notes
- The `/stock` command does not persist in the database.
- The frontend is intentionally minimal, focusing on backend functionality.

## Author
- [Henrique Zecchinel](mailto:henriquezecchinel@gmail.com)


Work in Progress!