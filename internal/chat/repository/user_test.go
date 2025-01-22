package repository

import (
	"context"
	"database/sql"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	username = "testuser"
	password = "password"
)

func TestUserRepository_Register(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	mock.ExpectExec("INSERT INTO Users").
		WithArgs(username, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Register(context.Background(), username, password)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Login(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT hashed_password FROM Users WHERE username = \\$1").
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"hashed_password"}).AddRow(string(hashedPassword)))

	err = repo.Login(context.Background(), username, password)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Login_UserDoesNotExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	username = "nonexistentuser"

	mock.ExpectQuery("SELECT hashed_password FROM Users WHERE username = \\$1").
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	err = repo.Login(context.Background(), username, password)
	assert.EqualError(t, err, "user does not exist")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Login_InvalidPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("differentpassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT hashed_password FROM Users WHERE username = \\$1").
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"hashed_password"}).AddRow(string(hashedPassword)))

	err = repo.Login(context.Background(), username, password)
	assert.EqualError(t, err, "invalid password")
	assert.NoError(t, mock.ExpectationsWereMet())
}
