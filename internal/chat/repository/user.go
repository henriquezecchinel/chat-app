package repository

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (repo *UserRepository) Register(ctx context.Context, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = repo.DB.ExecContext(ctx, `
        INSERT INTO Users (username, hashed_password)
        VALUES ($1, $2)
    `, username, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	return nil
}

func (repo *UserRepository) Login(ctx context.Context, username, password string) error {
	var hashedPassword string

	err := repo.DB.QueryRowContext(ctx, `
        SELECT hashed_password
        FROM Users
        WHERE username = $1
    `, username).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		return fmt.Errorf("user does not exist")
	} else if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}

func (repo *UserRepository) Authenticate(ctx context.Context, username, password string) (int, error) {
	var hashedPassword string
	var userID int

	err := repo.DB.QueryRowContext(ctx, `
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
