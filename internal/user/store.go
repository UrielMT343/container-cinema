package user

import (
	"context"
	"fmt"

	"start/internal/database"
	"start/internal/models"
)

type Store struct {
	db *database.Service
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

func (s *Store) CreateUser(user models.User) (models.User, error) {
	pool := s.db.GetDB()

	query := `
		INSERT INTO users (name, email, password_hash, is_active, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := pool.QueryRow(context.Background(), query, user.Name, user.Email, user.PasswordHash, user.IsActive, user.Role).Scan(&user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("error while creating the user: %v", err)
	}

	return user, nil
}

func (s *Store) GetUserByEmail(email string) (models.User, error) {
	pool := s.db.GetDB()

	query := `
		SELECT id, name, email, password_hash, is_active, role FROM users
		WHERE email = $1
	`
	var user models.User
	err := pool.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.IsActive, &user.Role)
	if err != nil {
		return models.User{}, fmt.Errorf("error: %v", err)
	}

	return user, nil
}
