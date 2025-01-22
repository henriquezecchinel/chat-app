package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecret string

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET not set in .env file")
	}
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 1-hour expiration
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid claims")
}

func (repo *UserRepository) Authenticate(ctx context.Context, username, password string) (int, error) {
	var hashedPassword string
	var userID int

	err := repo.db.QueryRowContext(ctx, `
        SELECT id, hashed_password
        FROM users
        WHERE username = $1
    `, username).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("user does not exist")
	} else if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return 0, fmt.Errorf("invalid password")
	}

	return userID, nil
}
