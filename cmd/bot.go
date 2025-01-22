package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	log.Println("Stock bot is starting...")

	// This is just a placeholder for now.
	// TODO: Connect to RabbitMQ for real functionality
	stockCode := "aapl.us"
	handleStockRequest(stockCode)

	log.Println("Placeholder execution is done...")
}

func handleStockRequest(stockCode string) {
	// Example API URL: https://stooq.com/q/l/?s=aapl.us&f=sd2t2ohlcv&h&e=csv
	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to fetch stock data:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: received HTTP status %d from stock API", resp.StatusCode)
	}

	// Parse stock data (CSV response)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response body:", err)
	}

	lines := strings.Split(string(body), "\n")
	if len(lines) > 1 {
		// Assuming the second line in the response contains the data
		data := strings.Split(lines[1], ",")
		if len(data) >= 6 {
			stockSymbol := strings.ToUpper(stockCode)
			price := data[3]
			fmt.Printf("%s quote is $%s per share\n", stockSymbol, price)
		} else {
			log.Println("Unexpected data format received from stock API")
		}
	} else {
		log.Println("No data received from stock API")
	}
}
