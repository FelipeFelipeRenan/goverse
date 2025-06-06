package swagger

import "net/http"

// Login godoc
// @Summary Login
// @Description Realiza login com email e senha
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciais de login"
// @Success 200 {object} LoginResponse
// @Failure 401 {string} string "Credenciais inválidas"
// @Router /login [post]
func SwaggerLoginPlaceholder(w http.ResponseWriter, r *http.Request) {}

// GetUserByID godoc
// @Summary Obter usuário por ID
// @Tags User
// @Security ApiKeyAuth
// @Param   Authorization   header  string  true  "Token de autenticação (Bearer token)"
// @Produce json
// @Param id path int true "ID do usuário"
// @Success 200 {object} UserResponse
// @Failure 401 {string} string "Não autorizado"
// @Router /user/{id} [get]
func SwaggerGetUserPlaceholder(w http.ResponseWriter, r *http.Request) {}

// CreateUser godoc
// @Summary Criar novo usuário
// @Tags User
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "Dados do novo usuário"
// @Success 201 {object} UserResponse
// @Failure 400 {string} string "Dados inválidos"
// @Router /user [post]
func SwaggerCreateUser(w http.ResponseWriter, r *http.Request) {}

// ListUsers godoc
// @Summary Listar todos os usuários
// @Tags User
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} UserResponse
// @Router /users [get]
func SwaggerListUsers(w http.ResponseWriter, r *http.Request) {}

// CreateRoom godoc
// @Summary Criar nova sala
// @Tags Room
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param room body CreateRoomRequest true "Dados da nova sala"
// @Success 201 {object} RoomResponse
// @Failure 400 {string} string "Dados inválidos"
// @Router /rooms [post]
func SwaggerCreateRoom(w http.ResponseWriter, r *http.Request) {}

// GetRoomByID godoc
// @Summary Buscar sala por ID
// @Tags Room
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "ID da sala"
// @Success 200 {object} RoomResponse
// @Failure 404 {string} string "Sala não encontrada"
// @Router /rooms/{id} [get]
func SwaggerGetRoomByID(w http.ResponseWriter, r *http.Request) {}

// ListRooms godoc
// @Summary Listar salas
// @Tags Room
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} RoomResponse
// @Router /rooms [get]
func SwaggerListRooms(w http.ResponseWriter, r *http.Request) {}

// UpdateRoom godoc
// @Summary Atualizar sala
// @Tags Room
// @Security ApiKeyAuth
// @Param   Authorization   header  string  true  "Token de autenticação (Bearer token)"
// @Accept json
// @Produce json
// @Param id path int true "ID da sala"
// @Param room body UpdateRoomRequest true "Dados atualizados da sala"
// @Success 200 {object} RoomResponse
// @Failure 400 {string} string "Dados inválidos"
// @Router /rooms/{id} [patch]
func SwaggerUpdateRoom(w http.ResponseWriter, r *http.Request) {}

// DeleteRoom godoc
// @Summary Excluir sala
// @Tags Room
// @Security ApiKeyAuth
// @Param id path int true "ID da sala"
// @Success 204 {string} string "Sala excluída"
// @Failure 404 {string} string "Sala não encontrada"
// @Router /rooms/{id} [delete]
func SwaggerDeleteRoom(w http.ResponseWriter, r *http.Request) {}

// JoinRoom godoc
// @Summary Entrar em uma sala
// @Tags Member
// @Security ApiKeyAuth
// @Param   Authorization   header  string  true  "Token de autenticação (Bearer token)"
// @Param roomID path int true "ID da sala"
// @Success 200 {object} MemberResponse
// @Failure 403 {string} string "Acesso negado"
// @Router /rooms/{roomID}/join [post]
func SwaggerJoinRoom(w http.ResponseWriter, r *http.Request) {}

// ListMembers godoc
// @Summary Listar membros da sala
// @Tags Member
// @Security ApiKeyAuth
// @Param roomID path int true "ID da sala"
// @Success 200 {array} MemberResponse
// @Router /rooms/{roomID}/members [get]
func SwaggerListMembers(w http.ResponseWriter, r *http.Request) {}

// AddMember godoc
// @Summary Adicionar membro à sala
// @Tags Member
// @Security ApiKeyAuth
// @Param   Authorization   header  string  true  "Token de autenticação (Bearer token)"
// @Accept json
// @Produce json
// @Param roomID path int true "ID da sala"
// @Param member body AddMemberRequest true "Dados do membro"
// @Success 201 {object} MemberResponse
// @Failure 400 {string} string "Dados inválidos"
// @Router /rooms/{roomID}/members [post]
func SwaggerAddMember(w http.ResponseWriter, r *http.Request) {}

// UpdateMemberRole godoc
// @Summary Atualizar papel do membro
// @Tags Member
// @Security ApiKeyAuth
// @Param   Authorization   header  string  true  "Token de autenticação (Bearer token)"
// @Accept json
// @Param roomID path int true "ID da sala"
// @Param memberID path int true "ID do membro"
// @Param request body UpdateRoleRequest true "Novo papel"
// @Success 200 {object} MemberResponse
// @Router /rooms/{roomID}/members/{memberID}/role [put]
func SwaggerUpdateMemberRole(w http.ResponseWriter, r *http.Request) {}

// RemoveMember godoc
// @Summary Remover membro da sala
// @Tags Member
// @Security ApiKeyAuth
// @Param   Authorization   header  string  true  "Token de autenticação (Bearer token)"
// @Param roomID path int true "ID da sala"
// @Param memberID path int true "ID do membro"
// @Success 204 {string} string "Removido com sucesso"
// @Router /rooms/{roomID}/members/{memberID} [delete]
func SwaggerRemoveMember(w http.ResponseWriter, r *http.Request) {}
