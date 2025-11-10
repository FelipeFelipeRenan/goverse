package doc_generators

import "time"

// (Structs de Autenticação)
type LoginRequest struct {
	Email    string `json:"email" example:"admin@goverse.com"`
	Password string `json:"password" example:"senha123"`
	Type     string `json:"type" example:"password"`
}

type LoginResponse struct {
	User      UserResponse `json:"user"`
	CsrfToken string       `json:"csrf_token" example:"a1b2c3d4-..."`
}

// (Structs de Usuário)
type UserResponse struct {
	ID        string    `json:"id" example:"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"`
	Username  string    `json:"username" example:"Admin Goverse"`
	Email     string    `json:"email" example:"admin@goverse.com"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	IsOAuth   bool      `json:"is_oauth"`
}

type CreateUserRequest struct {
	Username string `json:"username" example:"novo_usuario"`
	Email    string `json:"email" example:"novo@email.com"`
	Password string `json:"password" example:"senha123"`
}

type UpdateUserRequest struct {
	Username string `json:"username" example:"Novo Nome"`
	Picture  string `json:"picture" example:"http://.../img.png"`
}

// (Structs de Sala)
type RoomResponse struct {
	ID          string    `json:"id" example:"c0eebc99-9c0b-4ef8-bb6d-6bb9bd380b11"`
	Name        string    `json:"name" example:"Sala Geral"`
	Description string    `json:"description" example:"Sala pública para todos."`
	OwnerID     string    `json:"owner_id" example:"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"`
	IsPublic    bool      `json:"is_public" example:"true"`
	MemberCount int       `json:"member_count" example:"1"`
	MaxMembers  int       `json:"max_members" example:"100"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRoomRequest struct {
	Name        string `json:"name" example:"Minha Sala de Jogos" binding:"required"`
	Description string `json:"description" example:"Sala para jogar."`
	IsPublic    bool   `json:"is_public" example:"true"`
	MaxMembers  int    `json:"max_members" example:"50"`
}

type UpdateRoomRequest struct {
	Name        *string `json:"name,omitempty" example:"Sala de Estudos"`
	Description *string `json:"description,omitempty" example:"Novo foco."`
	IsPublic    *bool   `json:"is_public,omitempty" example:"false"`
	MaxMembers  *int    `json:"max_members,omitempty" example:"10"`
}

// (Structs de Membros)
type MemberWithUser struct {
	RoomID   string       `json:"room_id" example:"c0eebc99-9c0b-4ef8-bb6d-6bb9bd380b11"`
	Role     string       `json:"role" example:"owner"`
	JoinedAt time.Time    `json:"joined_at"`
	User     UserResponse `json:"user"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" example:"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12" binding:"required"`
	Role   string `json:"role" example:"member"`
}

type UpdateRoleRequest struct {
	NewRole string `json:"new_role" example:"admin" binding:"required"`
}
