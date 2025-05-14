package domain

import "time"

type Credentials struct {
	Type     string `json:"type"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Picture  string `json:"picture"`
}

type User struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"` // opcional no OAuth
	Picture   string    `json:"picture,omitempty"`  // opcional no registro comum
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
