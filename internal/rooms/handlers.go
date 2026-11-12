package rooms

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/httpx"
)

type api struct {
	svc  *service
	keys *authn.Keys
}

// Rotas idênticas às do room-service. Só GET /rooms (listagem de salas públicas) é aberta.
func (a *api) routes(mux *http.ServeMux) {
	auth := func(h http.HandlerFunc) http.Handler { return a.keys.Require(h) }

	mux.HandleFunc("GET /rooms", a.list)
	mux.Handle("POST /rooms", auth(a.create))
	mux.Handle("GET /rooms/mine", auth(a.owned))
	mux.Handle("GET /user/rooms", auth(a.mine))
	mux.Handle("GET /rooms/{id}", auth(a.get))
	mux.Handle("PATCH /rooms/{id}", auth(a.update))
	mux.Handle("DELETE /rooms/{id}", auth(a.remove))

	mux.Handle("POST /rooms/{id}/join", auth(a.join))
	mux.Handle("GET /rooms/{id}/members", auth(a.listMembers))
	mux.Handle("POST /rooms/{id}/members", auth(a.addMember))
	mux.Handle("PUT /rooms/{id}/members/{memberID}/role", auth(a.setRole))
	mux.Handle("DELETE /rooms/{id}/members/{memberID}", auth(a.removeMember))
}

// O serviço antigo devolvia dois formatos de sala; os dois são mantidos para o frontend não quebrar:
//   - roomResponse (id/name/...): POST /rooms e GET /rooms/{id}
//   - roomLegacy (room_id/room_name/...): listagens e PATCH
type roomResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	IsPublic    bool   `json:"is_public"`
	MemberCount int    `json:"member_count"`
	MaxMembers  int    `json:"max_members"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

const legacyTime = "2006-01-02T15:04:05Z"

func toResponse(r Room) roomResponse {
	return roomResponse{
		ID: r.ID, Name: r.Name, Description: r.Description, OwnerID: r.OwnerID, IsPublic: r.IsPublic,
		MemberCount: r.MemberCount, MaxMembers: r.MaxMembers,
		CreatedAt: r.CreatedAt.UTC().Format(legacyTime), UpdatedAt: r.UpdatedAt.UTC().Format(legacyTime),
	}
}

type roomLegacy struct {
	ID          string     `json:"room_id"`
	Name        string     `json:"room_name"`
	Description string     `json:"room_description"`
	IsPublic    bool       `json:"is_public"`
	OwnerID     string     `json:"owner_id"`
	MemberCount int        `json:"member_count"`
	MaxMembers  int        `json:"max_members"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func toLegacy(r Room) roomLegacy {
	return roomLegacy{
		ID: r.ID, Name: r.Name, Description: r.Description, IsPublic: r.IsPublic, OwnerID: r.OwnerID,
		MemberCount: r.MemberCount, MaxMembers: r.MaxMembers, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func toLegacyList(rs []Room) []roomLegacy {
	out := make([]roomLegacy, len(rs)) // nunca nil: o cliente recebe [] e não null
	for i, r := range rs {
		out[i] = toLegacy(r)
	}
	return out
}

type userView struct {
	ID        string `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	CreatedAt string `json:"created_at"`
	IsOAuth   bool   `json:"is_oauth"`
}

type memberView struct {
	RoomID   string    `json:"room_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	User     userView  `json:"user"`
}

// fail traduz erros de domínio em status HTTP; o que não for conhecido vira 500 sem vazar detalhes.
func fail(w http.ResponseWriter, r *http.Request, err error) {
	var inv invalidError
	switch {
	case errors.As(err, &inv):
		httpx.Error(w, http.StatusBadRequest, inv.msg)
	case errors.Is(err, ErrRoomNotFound), errors.Is(err, ErrMemberNotFound), errors.Is(err, ErrUserNotFound):
		httpx.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrForbidden):
		httpx.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrMemberAlreadyExists), errors.Is(err, ErrRoomFull),
		errors.Is(err, ErrCannotRemoveOwner), errors.Is(err, ErrCannotChangeOwner):
		httpx.Error(w, http.StatusConflict, err.Error())
	default:
		slog.ErrorContext(r.Context(), "erro em rooms", "method", r.Method, "path", r.URL.Path, "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
	}
}

func actor(r *http.Request) string {
	id, _ := authn.From(r.Context()) // Require garante que existe
	return id.UserID
}

func (a *api) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
		MaxMembers  int    `json:"max_members"`
	}
	if !httpx.Decode(w, r, &in) {
		return
	}
	room, err := a.svc.createRoom(r.Context(), actor(r), createInput(in))
	if errors.Is(err, ErrUserNotFound) {
		httpx.Error(w, http.StatusUnauthorized, "usuário da sessão não existe mais")
		return
	}
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toResponse(room))
}

func (a *api) get(w http.ResponseWriter, r *http.Request) {
	room, err := a.svc.getRoom(r.Context(), actor(r), r.PathValue("id"))
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(room))
}

// list é pública e só devolve salas públicas. O parâmetro public_only deixou de existir:
// permitia listar salas privadas sem autenticação.
func (a *api) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := 10, 0
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 {
		limit = min(v, 100)
	}
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v >= 0 {
		offset = v
	}
	rooms, err := a.svc.listPublic(r.Context(), limit, offset, q.Get("keyword"))
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toLegacyList(rooms))
}

func (a *api) mine(w http.ResponseWriter, r *http.Request) {
	rooms, err := a.svc.roomsOf(r.Context(), actor(r))
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toLegacyList(rooms))
}

func (a *api) owned(w http.ResponseWriter, r *http.Request) {
	rooms, err := a.svc.ownedBy(r.Context(), actor(r))
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toLegacyList(rooms))
}

func (a *api) update(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsPublic    *bool   `json:"is_public"`
		MaxMembers  *int    `json:"max_members"`
	}
	if !httpx.Decode(w, r, &in) {
		return
	}
	room, err := a.svc.updateRoom(r.Context(), actor(r), r.PathValue("id"), roomPatch(in))
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toLegacy(room))
}

func (a *api) remove(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := a.svc.deleteRoom(r.Context(), actor(r), id); err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, id)
}

func (a *api) join(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.join(r.Context(), actor(r), r.PathValue("id")); err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Entrada bem-sucedida"})
}

func (a *api) listMembers(w http.ResponseWriter, r *http.Request) {
	ms, err := a.svc.members(r.Context(), actor(r), r.PathValue("id"))
	if err != nil {
		fail(w, r, err)
		return
	}
	// E-mail é dado pessoal: só dono e admins da sala veem o dos outros (cada um sempre vê o próprio).
	me := actor(r)
	seeEmails := false
	for _, m := range ms {
		if m.UserID == me && (m.Role == RoleOwner || m.Role == RoleAdmin) {
			seeEmails = true
		}
	}
	out := make([]memberView, len(ms))
	for i, m := range ms {
		email := ""
		if seeEmails || m.UserID == me {
			email = m.User.Email
		}
		out[i] = memberView{
			RoomID: m.RoomID, Role: string(m.Role), JoinedAt: m.JoinedAt,
			User: userView{
				ID: m.User.ID, Name: m.User.Username, Email: email, Picture: m.User.Picture,
				CreatedAt: m.User.CreatedAt.UTC().Format(time.RFC3339), IsOAuth: m.User.IsOAuth,
			},
		}
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (a *api) addMember(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"` // alternativa ao user_id: quem convida sabe o e-mail, não o UUID
		Role   Role   `json:"role"`
	}
	if !httpx.Decode(w, r, &in) {
		return
	}
	if err := a.svc.addMember(r.Context(), actor(r), r.PathValue("id"), in.UserID, in.Email, in.Role); err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, in)
}

func (a *api) removeMember(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("memberID")
	if err := a.svc.removeMember(r.Context(), actor(r), r.PathValue("id"), userID); err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, userID)
}

func (a *api) setRole(w http.ResponseWriter, r *http.Request) {
	var in struct {
		NewRole Role `json:"new_role"`
		Role    Role `json:"role"` // o frontend envia "role"; o contrato antigo era "new_role"
	}
	if !httpx.Decode(w, r, &in) {
		return
	}
	if in.NewRole == "" {
		in.NewRole = in.Role
	}
	if err := a.svc.setRole(r.Context(), actor(r), r.PathValue("id"), r.PathValue("memberID"), in.NewRole); err != nil {
		fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, in)
}
