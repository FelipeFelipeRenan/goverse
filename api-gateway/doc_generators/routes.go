package doc_generators

import "net/http"

// --- AUTH ---

// SwaggerLoginPlaceholder godoc
// @Summary Login (Obter Cookies)
// @Description Realiza login com email e senha. Salva os cookies 'access_token' (HttpOnly) e 'csrf_token' (JS-readable) no navegador.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciais de login"
// @Success 200 {object} LoginResponse "Retorna o usuário e o token CSRF"
// @Failure 401 {string} string "Credenciais inválidas"
// @Router /auth/login [post]
func SwaggerLoginPlaceholder(w http.ResponseWriter, r *http.Request) {}

// SwaggerGetMe godoc
// @Summary Obter dados do usuário logado
// @Description Retorna os dados do usuário associado ao cookie 'access_token'.
// @Tags Auth
// @Produce json
// @Security CookieAuth
// @Success 200 {object} UserResponse
// @Failure 401 {string} string "Não autorizado"
// @Router /auth/me [get]
func SwaggerGetMe(w http.ResponseWriter, r *http.Request) {}

// SwaggerLogout godoc
// @Summary Logout (Limpar Cookies)
// @Description Limpa os cookies 'access_token' and 'csrf_token' do navegador.
// @Tags Auth
// @Produce json
// @Security CookieAuth, CsrfAuth
// @Success 200 {string} string "Logout bem-sucedido"
// @Router /auth/logout [post]
func SwaggerLogout(w http.ResponseWriter, r *http.Request) {}

// --- USER ---

// SwaggerCreateUser godoc
// @Summary Criar novo usuário (Público)
// @Description Rota pública para registrar um novo usuário no sistema.
// @Tags User
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "Dados do novo usuário"
// @Success 201 {object} UserResponse
// @Failure 400 {string} string "Dados inválidos"
// @Router /user [post]
func SwaggerCreateUser(w http.ResponseWriter, r *http.Request) {}

// SwaggerListUsers godoc
// @Summary Listar todos os usuários (Público)
// @Description Rota pública que retorna uma lista de todos os usuários.
// @Tags User
// @Produce json
// @Success 200 {array} UserResponse
// @Router /users [get]
func SwaggerListUsers(w http.ResponseWriter, r *http.Request) {}

// SwaggerUpdateUser godoc
// @Summary Atualizar o usuário logado
// @Description Atualiza o 'username' ou 'picture' do usuário autenticado.
// @Tags User
// @Accept json
// @Produce json
// @Security CookieAuth, CsrfAuth
// @Param input body UpdateUserRequest true "Dados do usuário para atualizar"
// @Success 200 {object} UserResponse
// @Failure 400 {string} string "Requisição inválida"
// @Failure 401 {string} string "Não autorizado"
// @Router /user/me [put]
func SwaggerUpdateUser(w http.ResponseWriter, r *http.Request) {}

// SwaggerDeleteUser godoc
// @Summary Deletar o usuário logado
// @Description Realiza um 'soft delete' no usuário autenticado.
// @Tags User
// @Security CookieAuth, CsrfAuth
// @Success 204 {string} string "Usuário deletado com sucesso"
// @Failure 401 {string} string "Não autorizado"
// @Router /user/me [delete]
func SwaggerDeleteUser(w http.ResponseWriter, r *http.Request) {}

// --- ROOM ---

// SwaggerCreateRoom godoc
// @Summary Criar nova sala
// @Description Cria uma nova sala. O usuário logado será o 'owner'.
// @Tags Room
// @Security CookieAuth, CsrfAuth
// @Accept json
// @Produce json
// @Param room body CreateRoomRequest true "Dados da nova sala"
// @Success 201 {object} RoomResponse
// @Failure 400 {string} string "Dados inválidos"
// @Router /rooms [post]
func SwaggerCreateRoom(w http.ResponseWriter, r *http.Request) {}

// SwaggerGetRoomByID godoc
// @Summary Buscar sala por ID
// @Description Retorna detalhes de uma sala específica. Requer autenticação.
// @Tags Room
// @Security CookieAuth
// @Produce json
// @Param id path string true "ID da sala (UUID)" format(uuid)
// @Success 200 {object} RoomResponse
// @Failure 404 {string} string "Sala não encontrada"
// @Router /rooms/{id} [get]
func SwaggerGetRoomByID(w http.ResponseWriter, r *http.Request) {}

// SwaggerListRooms godoc
// @Summary Listar salas (Público)
// @Description Retorna uma lista de salas. Por padrão, apenas salas públicas.
// @Tags Room
// @Produce json
// @Success 200 {array} RoomResponse
// @Router /rooms [get]
func SwaggerListRooms(w http.ResponseWriter, r *http.Request) {}

// SwaggerGetOwnedRooms godoc
// @Summary Listar minhas salas (criadas por mim)
// @Description Retorna todas as salas onde o usuário logado é o proprietário.
// @Tags Room
// @Security CookieAuth
// @Produce json
// @Success 200 {array} RoomResponse "Lista de salas"
// @Failure 401 {string} string "Não autorizado"
// @Router /rooms/mine [get]
func SwaggerGetOwnedRooms(w http.ResponseWriter, r *http.Request) {}

// SwaggerGetRoomsByUserID godoc
// @Summary Listar salas que eu participo
// @Description Retorna todas as salas das quais o usuário logado é membro.
// @Tags Room
// @Security CookieAuth
// @Produce json
// @Success 200 {array} RoomResponse
// @Failure 401 {string} string "Não autorizado"
// @Router /user/rooms [get]
func SwaggerGetRoomsByUserID(w http.ResponseWriter, r *http.Request) {}

// SwaggerUpdateRoom godoc
// @Summary Atualizar sala
// @Description Atualiza os dados de uma sala. Requer ser 'owner' ou 'admin' da sala.
// @Tags Room
// @Security CookieAuth, CsrfAuth
// @Accept json
// @Produce json
// @Param id path string true "ID da sala (UUID)" format(uuid)
// @Param room body UpdateRoomRequest true "Dados atualizados da sala"
// @Success 200 {object} RoomResponse
// @Failure 400 {string} string "Dados inválidos"
// @Router /rooms/{id} [patch]
func SwaggerUpdateRoom(w http.ResponseWriter, r *http.Request) {}

// SwaggerDeleteRoom godoc
// @Summary Excluir sala
// @Description Exclui uma sala. Requer ser o 'owner' da sala.
// @Tags Room
// @Security CookieAuth, CsrfAuth
// @Param id path string true "ID da sala (UUID)" format(uuid)
// @Success 204 {string} string "Sala excluída"
// @Failure 404 {string} string "Sala não encontrada"
// @Router /rooms/{id} [delete]
func SwaggerDeleteRoom(w http.ResponseWriter, r *http.Request) {}

// --- MEMBER ---

// SwaggerJoinRoom godoc
// @Summary Entrar em uma sala
// @Description Entra em uma sala pública.
// @Tags Member
// @Security CookieAuth, CsrfAuth
// @Param roomID path string true "ID da sala (UUID)" format(uuid)
// @Success 200 {string} string "Entrada bem-sucedida"
// @Failure 403 {string} string "Acesso negado (ex: sala privada)"
// @Router /rooms/{roomID}/join [post]
func SwaggerJoinRoom(w http.ResponseWriter, r *http.Request) {}

// SwaggerListMembers godoc
// @Summary Listar membros da sala
// @Description Retorna a lista de usuários em uma sala específica.
// @Tags Member
// @Security CookieAuth
// @Param roomID path string true "ID da sala (UUID)" format(uuid)
// @Success 200 {array} MemberWithUser
// @Router /rooms/{roomID}/members [get]
func SwaggerListMembers(w http.ResponseWriter, r *http.Request) {}

// SwaggerAddMember godoc
// @Summary Adicionar membro à sala
// @Description Adiciona um usuário a uma sala. Requer ser 'owner' ou 'admin' da sala.
// @Tags Member
// @Security CookieAuth, CsrfAuth
// @Accept json
// @Produce json
// @Param roomID path string true "ID da sala (UUID)" format(uuid)
// @Param member body AddMemberRequest true "Dados do membro"
// @Success 201 {object} MemberWithUser
// @Failure 400 {string} string "Dados inválidos"
// @Router /rooms/{roomID}/members [post]
func SwaggerAddMember(w http.ResponseWriter, r *http.Request) {}

// SwaggerUpdateMemberRole godoc
// @Summary Atualizar papel do membro
// @Description Atualiza o 'role' de um membro na sala. Requer ser 'owner' ou 'admin'.
// @Tags Member
// @Security CookieAuth, CsrfAuth
// @Accept json
// @Param roomID path string true "ID da sala (UUID)" format(uuid)
// @Param memberID path string true "ID do usuário (UUID)" format(uuid)
// @Param request body UpdateRoleRequest true "Novo papel"
// @Success 200 {object} MemberWithUser
// @Router /rooms/{roomID}/members/{memberID}/role [put]
func SwaggerUpdateMemberRole(w http.ResponseWriter, r *http.Request) {}

// SwaggerRemoveMember godoc
// @Summary Remover membro da sala
// @Description Remove um usuário da sala. Requer ser 'owner'/'admin', ou o próprio usuário.
// @Tags Member
// @Security CookieAuth, CsrfAuth
// @Param roomID path string true "ID da sala (UUID)" format(uuid)
// @Param memberID path string true "ID do usuário (UUID)" format(uuid)
// @Success 204 {string} string "Removido com sucesso"
// @Router /rooms/{roomID}/members/{memberID} [delete]
func SwaggerRemoveMember(w http.ResponseWriter, r *http.Request) {}