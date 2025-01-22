package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Message struct {
	ID         int       `json:"id"`
	ChatroomID int       `json:"chatroom_id"`
	UserID     int       `json:"user_id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (repo *MessageRepository) AddMessage(ctx context.Context, chatroomID, userID int, content string) error {
	_, err := repo.db.ExecContext(ctx, `
        INSERT INTO messages (chatroom_id, user_id, content)
        VALUES ($1, $2, $3)
    `, chatroomID, userID, content)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

func (repo *MessageRepository) GetLastMessages(ctx context.Context, chatroomID int, limit int) ([]Message, error) {
	rows, err := repo.db.QueryContext(ctx, `
        SELECT id, chatroom_id, user_id, content, timestamp
        FROM messages
        WHERE chatroom_id = $1
        ORDER BY timestamp DESC
        LIMIT $2
    `, chatroomID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ChatroomID, &msg.UserID, &msg.Content, &msg.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
