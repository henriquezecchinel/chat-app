package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"chat-app/internal/auth"
	"chat-app/internal/chat"
	"chat-app/internal/storage"

	"github.com/gorilla/websocket"
)

var (
	authRepo     *auth.UserRepository
	chatroomRepo *chat.ChatroomRepository
	messageRepo  *chat.MessageRepository

	clients      = make(map[int]map[*websocket.Conn]bool) // Chatroom ID => WebSocket connections
	clientsMutex = sync.RWMutex{}
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from all origins (use only in dev!)
			return true
		},
	}
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := storage.NewDB(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	authRepo = auth.NewUserRepository(db.Conn)
	chatroomRepo = chat.NewChatroomRepository(db.Conn)
	messageRepo = chat.NewMessageRepository(db.Conn)

	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.Handle("/chatroom/create", auth.AuthMiddleware(http.HandlerFunc(handleCreateChatroom)))
	http.Handle("/chatroom/list", auth.AuthMiddleware(http.HandlerFunc(handleListChatrooms)))
	http.Handle("/chatroom/post_message", auth.AuthMiddleware(http.HandlerFunc(handlePostMessage)))
	http.Handle("/chatroom/messages", auth.AuthMiddleware(http.HandlerFunc(handleGetMessages)))

	http.HandleFunc("/ws", handleWebSocket)

	http.Handle("/", http.FileServer(http.Dir("./web/static")))

	log.Println("Chat server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// IMPORTANT: This code is for demonstration purposes only.
	// In a real-world application, we should avoid using query parameters for sensitive data such as token.
	// We can use different approaches like using HTTP Handshake or sending the first message with the Auth details.
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		log.Println("Missing token parameter")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		log.Printf("Invalid JWT token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	chatroomIDStr := r.URL.Query().Get("chatroom_id")
	if chatroomIDStr == "" {
		http.Error(w, "Missing chatroom_id", http.StatusBadRequest)
		return
	}

	chatroomID := atoi(chatroomIDStr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	addClientToChatroom(conn, chatroomID)
	defer removeClientFromChatroom(conn, chatroomID)

	// Start handling WebSocket messages
	for {
		var msg struct {
			Content string `json:"content"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		if err := messageRepo.AddMessage(r.Context(), chatroomID, userID, msg.Content); err != nil {
			log.Println("Failed to store message in the DB:", err)
			continue
		}

		broadcastMessageToChatroom(chatroomID, fmt.Sprintf("User %d: %s", userID, msg.Content))
	}
}

func addClientToChatroom(conn *websocket.Conn, chatroomID int) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if clients[chatroomID] == nil {
		clients[chatroomID] = make(map[*websocket.Conn]bool)
	}
	clients[chatroomID][conn] = true
}

func removeClientFromChatroom(conn *websocket.Conn, chatroomID int) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if _, ok := clients[chatroomID]; ok {
		delete(clients[chatroomID], conn)
		if len(clients[chatroomID]) == 0 {
			delete(clients, chatroomID)
		}
	}
}

func broadcastMessageToChatroom(chatroomID int, content string) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	if chatroomClients, ok := clients[chatroomID]; ok {
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

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = authRepo.Register(ctx, req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" || req.Password == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID, err := authRepo.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(userID, req.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func handleCreateChatroom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Name == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	id, err := chatroomRepo.CreateChatroom(ctx, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chatroom_id": id,
	})
}

func handleListChatrooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	chatrooms, err := chatroomRepo.ListChatrooms(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"chatrooms": chatrooms,
	})
}

func handlePostMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract the authenticated user ID from the context
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	var req struct {
		ChatroomID int    `json:"chatroom_id"`
		Content    string `json:"content"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ChatroomID <= 0 || req.Content == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = messageRepo.AddMessage(ctx, req.ChatroomID, userID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Message posted successfully",
	})
}

func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	chatroomID := r.URL.Query().Get("chatroom_id")
	if chatroomID == "" {
		http.Error(w, "Missing chatroom_id parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	messages, err := messageRepo.GetLastMessages(ctx, atoi(chatroomID), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
}

// atoi is a helper function to convert string to integer.
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
