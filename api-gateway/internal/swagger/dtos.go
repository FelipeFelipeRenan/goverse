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
	Picture  string `json:"picture"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Picture  string `json:"picture"`
}

type RoomResponse struct {
	ID          int    `json:"id" example:"1"`
	Name        string `json:"name" example:"Sala de Estudos"`
	Description string `json:"description" example:"Sala para estudo de algoritmos"`
	OwnerID     int    `json:"owner_id" example:"1"`
	CreatedAt   string `json:"created_at" example:"2025-06-06T18:30:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2025-06-06T19:00:00Z"`
}

type CreateRoomRequest struct {
	Name        string `json:"name" example:"Sala de Estudos" binding:"required"`
	Description string `json:"description" example:"Sala para estudo de algoritmos"`
}

type UpdateRoomRequest struct {
	Name        *string `json:"name" example:"Sala Atualizada"`
	Description *string `json:"description" example:"Descrição atualizada"`
}

type MemberResponse struct {
	ID       int    `json:"id" example:"5"`
	UserID   int    `json:"user_id" example:"2"`
	RoomID   int    `json:"room_id" example:"1"`
	Username string `json:"username" example:"joaogate"`
	Role     string `json:"role" example:"admin"`
	JoinedAt string `json:"joined_at" example:"2025-06-05T20:00:00Z"`
}

type AddMemberRequest struct {
	UserID int    `json:"user_id" example:"3" binding:"required"`
	Role   string `json:"role" example:"member"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" example:"admin" binding:"required"`
}
