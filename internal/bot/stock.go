package bot

import (
	"chat-app/internal/chat"
	"chat-app/internal/chat/repository"
	"chat-app/internal/messaging"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

func FetchStockData(stockCode string) string {
	stockCode = strings.ToUpper(strings.TrimSpace(stockCode))

	if !IsValidStockCode(stockCode) {
		return fmt.Sprintf("Invalid stock code: %s", stockCode)
	}

	// TODO: Move the URL to a config file
	// TODO: Should we process JSON responses instead of CSV?
	urlStr := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Println("Failed to parse URL:", err)
		return fmt.Sprintf("Error fetching stock data for %s", stockCode)
	}

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		log.Println("Failed to fetch stock data:", err)
		return fmt.Sprintf("Error fetching stock data for %s", stockCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return fmt.Sprintf("Error reading stock data for %s", stockCode)
	}

	// Parse the CSV response, valid stock_code response example below:
	// Symbol,Date,Time,Open,High,Low,Close,Volume
	// AAPL.US,2025-01-22,16:15:22,219.79,223.3528,219.79,222.4683,8385754
	lines := strings.Split(string(body), "\n")
	if len(lines) > 1 {
		data := strings.Split(lines[1], ",")
		if len(data) >= 6 {
			stockSymbol := strings.ToUpper(stockCode)
			price := data[3] // The stock open price in the CSV file

			// Example of invalid stock_code response:
			// Symbol,Date,Time,Open,High,Low,Close,Volume
			// BAD.US,N/D,N/D,N/D,N/D,N/D,N/D,N/D
			if price == "N/D" {
				return fmt.Sprintf("No data available for stock code %s", stockSymbol)
			}

			return fmt.Sprintf("%s quote is $%s per share", stockSymbol, price)
		}
	}

	log.Println("Unexpected data format received from stock API")
	return fmt.Sprintf("No data available for stock code %s", strings.ToUpper(stockCode))
}

func ConsumeStockResponses(rabbitMQ *messaging.RabbitMQ) {
	msgs, err := rabbitMQ.Channel.Consume(
		"stock_responses",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to start consuming stock_responses:", err)
	}

	for msg := range msgs {
		var response struct {
			ChatroomID int    `json:"chatroom_id"`
			Content    string `json:"content"`
		}

		if err := json.Unmarshal(msg.Body, &response); err != nil {
			log.Println("Failed to unmarshal stock response:", err)
			continue
		}

		msgToSend := repository.Message{
			ChatroomID: response.ChatroomID,
			UserID:     0,
			Content:    response.Content,
			Timestamp:  time.Now(),
		}

		chat.BroadcastMessageToChatroom(response.ChatroomID, msgToSend)
	}
}

func SendStockRequestToQueue(rabbitMQ *messaging.RabbitMQ, request map[string]interface{}) {
	body, err := json.Marshal(request)
	if err != nil {
		log.Println("Failed to marshal stock request:", err)
		return
	}

	err = rabbitMQ.Channel.Publish(
		"",
		"stock_requests",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Println("Failed to publish stock request:", err)
	}
}

func ConsumeStockRequests(rabbitMQ *messaging.RabbitMQ) {
	msgs, err := rabbitMQ.Channel.Consume(
		"stock_requests",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to start consuming stock_requests:", err)
	}

	log.Println("Bot is ready to receive messages...")

	for msg := range msgs {
		log.Println("Processing stock request:", string(msg.Body))
		var request struct {
			ChatroomID int    `json:"chatroom_id"`
			StockCode  string `json:"stock_code"`
		}

		if err := json.Unmarshal(msg.Body, &request); err != nil {
			log.Println("Failed to unmarshal stock request:", err)
			continue
		}

		stockResponse := FetchStockData(request.StockCode)

		response := map[string]interface{}{
			"chatroom_id": request.ChatroomID,
			"content":     stockResponse,
		}

		PublishStockResponse(rabbitMQ, response)
	}
}

func PublishStockResponse(rabbitMQ *messaging.RabbitMQ, response map[string]interface{}) {
	body, err := json.Marshal(response)
	if err != nil {
		log.Println("Failed to marshal stock response:", err)
		return
	}

	err = rabbitMQ.Channel.Publish(
		"",
		"stock_responses",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Println("Failed to publish stock response:", err)
	}
}

func IsValidStockCode(code string) bool {
	if code == "" {
		return false
	}

	if len(code) < 2 || len(code) > 12 { // Arbitrary max length
		return false
	}

	// Allowed characters: 'A-Z', '0-9', '.', and '-'
	validPattern := `^[A-Z0-9.-]+$`
	matched, err := regexp.MatchString(validPattern, code)
	if err != nil {
		log.Println("Regex validation failed:", err)
		return false
	}

	return matched
}
