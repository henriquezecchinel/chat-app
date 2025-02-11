package chat

import (
	"chat-app/internal/bot"
	"chat-app/internal/chat"
	"chat-app/internal/chat/repository"
	"chat-app/internal/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"chat-app/internal/auth"
	"chat-app/internal/messaging"
	"chat-app/internal/storage"

	"github.com/gorilla/websocket"
)

var (
	userRepo     *repository.UserRepository
	chatroomRepo *repository.ChatroomRepository
	messageRepo  *repository.MessageRepository

	upgrader = websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			// Allow connections from all origins (use only in dev!)
			return true
		},
	}

	chatRabbitMQ *messaging.RabbitMQ
)

func RunChatServer() error {
	var err error

	chatRabbitMQ, err = messaging.SetupRabbitMQ("stock_requests", "stock_responses")
	if err != nil {
		log.Fatal("chatRabbitMQ setup failed:", err)
	}
	defer chatRabbitMQ.Close()

	go bot.ConsumeStockResponses(chatRabbitMQ)

	db, err := storage.SetupDatabaseConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	userRepo = repository.NewUserRepository(db.Conn)
	chatroomRepo = repository.NewChatroomRepository(db.Conn)
	messageRepo = repository.NewMessageRepository(db.Conn)

	// TODO: Migrate the 'handle' functions to separate files
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.Handle("/chatroom/create", auth.Middleware(http.HandlerFunc(handleCreateChatroom)))
	http.Handle("/chatroom/list", auth.Middleware(http.HandlerFunc(handleListChatrooms)))
	http.Handle("/chatroom/post_message", auth.Middleware(http.HandlerFunc(handlePostMessage)))
	http.Handle("/chatroom/messages", auth.Middleware(http.HandlerFunc(handleGetMessages)))

	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir("./web/static")))

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
		// TODO: REMOVE THE CORS MIDDLEWARE FOR PROD ENVIRONMENT
		Handler: utils.CorsMiddleware(http.DefaultServeMux),
	}

	log.Println("Chat server is running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}

	return nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// IMPORTANT: This code is for demonstration purposes only.
	// In a real-world application, we should avoid using query parameters for sensitive data such as token.
	// TODO: Implement a more secure way to authenticate WebSocket connections [ephemeral access tokens, cookies, etc.]
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

	chatroomID, err := utils.Atoi(chatroomIDStr)
	if err != nil {
		http.Error(w, "Invalid chatroom_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	chat.AddClientToChatroom(conn, chatroomID)
	defer chat.RemoveClientFromChatroom(conn, chatroomID)

	// Start handling WebSocket messages
	for {
		var msg struct {
			Content string `json:"content"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error for chatroom: %v", err)
			break
		}

		if strings.HasPrefix(msg.Content, "/stock=") {
			stockCode := strings.TrimPrefix(msg.Content, "/stock=")
			stockRequest := map[string]interface{}{
				"chatroom_id": chatroomID,
				"stock_code":  stockCode,
			}

			bot.SendStockRequestToQueue(chatRabbitMQ, stockRequest)
		} else {
			if err := messageRepo.AddMessage(r.Context(), chatroomID, userID, msg.Content); err != nil {
				log.Println("Failed to store message in the DB:", err)
				continue
			}

			msgToSend := repository.Message{
				ChatroomID: chatroomID,
				UserID:     userID,
				Content:    msg.Content,
				Timestamp:  time.Now(),
			}
			chat.BroadcastMessageToChatroom(chatroomID, msgToSend)
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
	err = userRepo.Register(ctx, req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement storing the JWT token in an HttpOnly cookie
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
	userID, err := userRepo.Authenticate(ctx, req.Username, req.Password)
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
	err = json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"chatroom_id": id,
	})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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

	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"chatrooms": chatrooms,
	})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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
	err = json.NewEncoder(w).Encode(map[string]string{
		"message": "Message posted successfully",
	})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	chatroomIDStr := r.URL.Query().Get("chatroom_id")
	if chatroomIDStr == "" {
		http.Error(w, "Missing chatroom_id", http.StatusBadRequest)
		return
	}

	chatroomID, err := utils.Atoi(chatroomIDStr)
	if err != nil {
		http.Error(w, "Invalid chatroom_id", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	messages, err := messageRepo.GetLastMessages(ctx, chatroomID, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
