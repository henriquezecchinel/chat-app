package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// Chatroom represents a basic structure for a chatroom with ID and Name
type Chatroom struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ChatroomRepository struct {
	db *sql.DB
}

func NewChatroomRepository(db *sql.DB) *ChatroomRepository {
	return &ChatroomRepository{db: db}
}

func (repo *ChatroomRepository) CreateChatroom(ctx context.Context, name string) (int, error) {
	var id int

	err := repo.db.QueryRowContext(ctx, `
        INSERT INTO chatrooms (name)
        VALUES ($1)
        RETURNING id
    `, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create chatroom: %w", err)
	}

	return id, nil
}

func (repo *ChatroomRepository) ListChatrooms(ctx context.Context) ([]Chatroom, error) {
	rows, err := repo.db.QueryContext(ctx, `
        SELECT id, name
        FROM chatrooms
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chatrooms: %w", err)
	}
	defer rows.Close()

	var chatrooms []Chatroom
	for rows.Next() {
		var chatroom Chatroom
		if err := rows.Scan(&chatroom.ID, &chatroom.Name); err != nil {
			return nil, fmt.Errorf("failed to scan chatroom: %w", err)
		}
		chatrooms = append(chatrooms, chatroom)
	}

	return chatrooms, nil
}
