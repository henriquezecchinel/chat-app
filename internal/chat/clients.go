package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

var (
	Clients      = make(map[int]map[*websocket.Conn]bool) // Chatroom ID => WebSocket connections
	ClientsMutex = sync.RWMutex{}
)

func AddClientToChatroom(conn *websocket.Conn, chatroomID int) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	if Clients[chatroomID] == nil {
		Clients[chatroomID] = make(map[*websocket.Conn]bool)
	}
	Clients[chatroomID][conn] = true
}

func RemoveClientFromChatroom(conn *websocket.Conn, chatroomID int) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	if _, ok := Clients[chatroomID]; ok {
		delete(Clients[chatroomID], conn)
		if len(Clients[chatroomID]) == 0 {
			delete(Clients, chatroomID)
		}
	}
}
