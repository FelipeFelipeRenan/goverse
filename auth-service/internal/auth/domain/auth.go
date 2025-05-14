package domain

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
}

type TokenResponse struct {
	Token string `json:"token"`
}
