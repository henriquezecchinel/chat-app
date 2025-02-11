package main

import (
	"chat-app/cmd/text"
	"flag"
	"fmt"
	"log"
	"os"

	"chat-app/cmd/bot"
	"chat-app/cmd/chat"
)

func main() {
	// Define a CLI flag to choose between chat and bot
	appType := flag.String("app", "", "Specify the application to run: 'chat' or 'bot'")
	flag.Parse()

	if *appType == "" {
		fmt.Println("Usage: go run main.go -app=<application>")
		fmt.Println("Available applications: 'chat', 'bot'")
		os.Exit(1)
	}

	// Run the selected application based on the provided flag
	switch *appType {
	case "chat":
		log.Println("Starting Chat Application...")
		if err := chat.RunChatServer(); err != nil {
			log.Fatalf("Failed to run chat server: %v", err)
		}
	case "bot":
		log.Println("Starting Bot Application...")
		if err := bot.RunBotServer(); err != nil {
			log.Fatalf("Failed to run bot: %v", err)
		}
	case "text":
		log.Println("Starting Text Application...")
		if err := text.RunTextServer(); err != nil {
			log.Fatalf("Failed to run text server: %v", err)
		}
	default:
		fmt.Printf("Unknown application '%s'. Available options: 'chat', 'bot', 'text'\n", *appType)
		os.Exit(1)
	}
}
