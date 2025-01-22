package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatroomRepository_CreateChatroom(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewChatroomRepository(db)

	name := "Test Chatroom"
	chatroomID := 1

	mock.ExpectQuery("INSERT INTO chatrooms \\(name\\) VALUES \\(\\$1\\) RETURNING id").
		WithArgs(name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(chatroomID))

	id, err := repo.CreateChatroom(context.Background(), name)
	assert.NoError(t, err)
	assert.Equal(t, chatroomID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatroomRepository_ListChatrooms(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewChatroomRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Chatroom 1").
		AddRow(2, "Chatroom 2")

	mock.ExpectQuery("SELECT id, name FROM chatrooms").
		WillReturnRows(rows)

	chatrooms, err := repo.ListChatrooms(context.Background())
	assert.NoError(t, err)
	assert.Len(t, chatrooms, 2)
	assert.Equal(t, "Chatroom 1", chatrooms[0].Name)
	assert.Equal(t, "Chatroom 2", chatrooms[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
