package chat

import (
	"chat-app/internal/chat/repository"
	"log"
)

func BroadcastMessageToChatroom(chatroomID int, message repository.Message) {
	ClientsMutex.RLock()
	defer ClientsMutex.RUnlock()

	if chatroomClients, ok := Clients[chatroomID]; ok {
		for client := range chatroomClients {
			err := client.WriteJSON(message)
			if err != nil {
				log.Println("WebSocket broadcast error:", err)
				client.Close()
				delete(chatroomClients, client)
			}
		}
	}
}
