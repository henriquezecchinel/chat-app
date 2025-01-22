package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageRepository_AddMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(db)

	chatroomID := 1
	userID := 1
	content := "Hello, world!"

	mock.ExpectExec("INSERT INTO messages").
		WithArgs(chatroomID, userID, content).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddMessage(context.Background(), chatroomID, userID, content)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetLastMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(db)

	chatroomID := 1
	limit := 10
	timestamp := time.Now()

	rows := sqlmock.NewRows([]string{"id", "chatroom_id", "user_id", "content", "timestamp"}).
		AddRow(1, chatroomID, 1, "Hello, world!", timestamp).
		AddRow(2, chatroomID, 2, "Hi there!", timestamp)

	mock.ExpectQuery("SELECT id, chatroom_id, user_id, content, timestamp FROM messages WHERE chatroom_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2").
		WithArgs(chatroomID, limit).
		WillReturnRows(rows)

	messages, err := repo.GetLastMessages(context.Background(), chatroomID, limit)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Hello, world!", messages[0].Content)
	assert.Equal(t, "Hi there!", messages[1].Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}
