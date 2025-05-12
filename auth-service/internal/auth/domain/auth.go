package domain

type Credentials struct {
	Email    string `json:"email"`
	Password string `json: "password`
}

type TokenResponse struct {
	Token string `json:"token"`
}
