package bot

import (
	"chat-app/internal/bot"
	"chat-app/internal/messaging"
	"log"
)

var botRabbitMQ *messaging.RabbitMQ

func RunBot() error {
	var err error

	botRabbitMQ, err = messaging.SetupRabbitMQ("stock_requests", "stock_responses")
	if err != nil {
		log.Fatal("RabbitMQ setup failed:", err)
	}
	defer botRabbitMQ.Close()

	bot.ConsumeStockRequests(botRabbitMQ)

	return nil
}
