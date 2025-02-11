package text

import (
	"chat-app/internal/storage"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type TextRoom struct {
	RoomID  string
	Content string
	Clients map[*websocket.Conn]bool
	Mutex   sync.RWMutex
}

func HandleTextRoom(w http.ResponseWriter, r *http.Request, db *storage.DB) {
	roomID := r.URL.Path[len("/room/"):]
	if roomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetTextRoom(w, r, db, roomID)
	case http.MethodPost:
		handlePostTextRoom(w, r, db, roomID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// TODO: Add Websocket support
}

func handleGetTextRoom(w http.ResponseWriter, r *http.Request, db *storage.DB, roomID string) {
	var content string
	err := db.Conn.QueryRow("SELECT content FROM text_rooms WHERE room_id = $1 ORDER BY created_at DESC LIMIT 1", roomID).Scan(&content)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": content})
}

// TODO: Add some safety rules along with limitations to prevent abuse
func handlePostTextRoom(w http.ResponseWriter, r *http.Request, db *storage.DB, roomID string) {
	var msg struct {
		NewContent string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	var oldContent string
	var oldTimestamp time.Time
	err = tx.QueryRow("SELECT content, updated_at FROM text_rooms WHERE room_id = $1", roomID).Scan(&oldContent, &oldTimestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			// Room not found, insert new record
			_, err = tx.Exec(`
				INSERT INTO text_rooms 
				    (room_id, content, content_history) 
				VALUES ($1, $2, $3)
			`, roomID, msg.NewContent, "[]")
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to create new room", http.StatusInternalServerError)
				return
			}
		} else {
			tx.Rollback()
			http.Error(w, "Failed to query room", http.StatusInternalServerError)
			return
		}
	} else {
		// Room found, update existing record
		historyEntry := map[string]interface{}{
			"timestamp": oldTimestamp,
			"content":   oldContent,
		}

		historyEntryJSON, err := json.Marshal(historyEntry)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to marshal history entry", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
			UPDATE text_rooms 
			SET 
			    content_history = jsonb_insert(content_history, '{0}', to_jsonb($1::jsonb), true), 
			    content = $2,
			    updated_at = CURRENT_TIMESTAMP
			WHERE room_id = $3
		`, string(historyEntryJSON), msg.NewContent, roomID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update message", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
