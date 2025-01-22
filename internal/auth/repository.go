package auth

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Register(ctx context.Context, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = repo.db.ExecContext(ctx, `
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

	err := repo.db.QueryRowContext(ctx, `
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
