package chat

import "log"

func BroadcastMessageToChatroom(chatroomID int, content string) {
	ClientsMutex.RLock()
	defer ClientsMutex.RUnlock()

	if chatroomClients, ok := Clients[chatroomID]; ok {
		for client := range chatroomClients {
			err := client.WriteJSON(map[string]string{
				"message": content,
			})
			if err != nil {
				log.Println("WebSocket broadcast error:", err)
				client.Close()
				delete(chatroomClients, client)
			}
		}
	}
}
