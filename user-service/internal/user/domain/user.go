package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	IsOAuth   bool      `json:"is_oauth"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	IsOAuth   bool      `json:"is_oauth"`
}
