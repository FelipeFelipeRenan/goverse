package domain

import (
	"errors"
	"time"
)

type User struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	Picture   string     `json:"picture"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	IsOAuth   bool       `json:"is_oauth"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	IsOAuth   bool      `json:"is_oauth"`
}

var ErrUserNotFound = errors.New("usuário não encontrado")
