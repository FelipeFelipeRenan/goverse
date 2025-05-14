package domain

type Credentials struct {
	Type     string `json:"type"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
