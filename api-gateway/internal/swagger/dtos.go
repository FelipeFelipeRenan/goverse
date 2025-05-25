package swagger

type LoginRequest struct {
	Email    string `json:"email" example:"joao@email.com"`
	Password string `json:"password" example:"senha123"`
	Type     string `json:"type" example:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"joaogate"`
	Email    string `json:"email" example:"joao@email.com"`
}
