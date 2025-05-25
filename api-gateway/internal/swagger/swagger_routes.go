package swagger

import "net/http"

// @Summary Login
// @Description Realiza login com email e senha
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciais de login"
// @Success 200 {object} LoginResponse
// @Failure 401 {string} string "Credenciais inválidas"
// @Router /login [post]
func SwaggerLoginPlaceholder(w http.ResponseWriter, r *http.Request){}


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