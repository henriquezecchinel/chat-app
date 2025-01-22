package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"chat-app/internal/auth"
	"chat-app/internal/chat"
	"chat-app/internal/storage"
)

var (
	authRepo     *auth.UserRepository
	chatroomRepo *chat.ChatroomRepository
	messageRepo  *chat.MessageRepository
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

	// TODO: Add WebSocket handlers for real-time chat functionality.

	log.Println("Chat server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start server:", err)
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
