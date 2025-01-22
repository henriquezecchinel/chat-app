package auth

import (
	"context"
	_ "database/sql"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Authenticate(ctx context.Context, username, password string) (int, error) {
	args := m.Called(ctx, username, password)
	return args.Int(0), args.Error(1)
}

func TestGenerateJWT(t *testing.T) {
	userID := 1
	username := "testuser"

	token, err := GenerateJWT(userID, username)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

func TestValidateJWT(t *testing.T) {
	userID := 1
	username := "testuser"

	token, err := GenerateJWT(userID, username)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidateJWT(token)

	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
}

func TestAuthenticate(t *testing.T) {
	repo := new(MockUserRepository)
	ctx := context.Background()

	t.Run("user exists and password matches", func(t *testing.T) {
		repo.On("Authenticate", ctx, "testuser", "password").Return(1, nil)

		userID, err := repo.Authenticate(ctx, "testuser", "password")

		assert.NoError(t, err)
		assert.Equal(t, 1, userID)
		repo.AssertExpectations(t)
	})

	t.Run("user does not exist", func(t *testing.T) {
		repo.On("Authenticate", ctx, "unknownuser", "password").Return(0, errors.New("user does not exist"))

		userID, err := repo.Authenticate(ctx, "unknownuser", "password")

		assert.EqualError(t, err, "user does not exist")
		assert.Equal(t, 0, userID)
		repo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		repo.On("Authenticate", ctx, "testuser", "wrongpassword").Return(0, errors.New("invalid password"))

		userID, err := repo.Authenticate(ctx, "testuser", "wrongpassword")

		assert.EqualError(t, err, "invalid password")
		assert.Equal(t, 0, userID)
		repo.AssertExpectations(t)
	})
}
