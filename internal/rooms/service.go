package rooms

import (
	"context"
	"errors"
	"strings"

	"github.com/FelipeFelipeRenan/goverse/internal/identity"
)

var (
	ErrCannotRemoveOwner = errors.New("não é possível remover o dono da sala")
	ErrCannotChangeOwner = errors.New("não é possível alterar a role do dono da sala")
)

const (
	defaultMaxMembers = 10
	maxMembersLimit   = 1000
	maxNameLen        = 100
	maxDescriptionLen = 1000
)

type service struct {
	st    store
	users identity.Directory
}

type createInput struct {
	Name        string
	Description string
	IsPublic    bool
	MaxMembers  int
}

// MemberView é um membro já enriquecido com os dados do usuário.
type MemberView struct {
	Member
	User identity.User
}

func (s *service) createRoom(ctx context.Context, ownerID string, in createInput) (Room, error) {
	in.Name = strings.TrimSpace(in.Name)
	if err := validateName(in.Name); err != nil {
		return Room{}, err
	}
	if len(in.Description) > maxDescriptionLen {
		return Room{}, invalid("descrição longa demais")
	}
	if in.MaxMembers == 0 {
		in.MaxMembers = defaultMaxMembers
	}
	if err := validateMax(in.MaxMembers); err != nil {
		return Room{}, err
	}
	// Usuário removido (soft delete) ainda tem token válido por até 24h: não pode criar sala.
	if ok, err := s.users.Exists(ctx, ownerID); err != nil {
		return Room{}, err
	} else if !ok {
		return Room{}, ErrUserNotFound
	}
	return s.st.createRoom(ctx, newRoom{
		OwnerID: ownerID, Name: in.Name, Description: in.Description,
		IsPublic: in.IsPublic, MaxMembers: in.MaxMembers,
	})
}

func validateName(name string) error {
	if name == "" {
		return invalid("name é obrigatório")
	}
	if len(name) > maxNameLen {
		return invalid("name longo demais")
	}
	return nil
}

func validateMax(n int) error {
	if n < 1 || n > maxMembersLimit {
		return invalid("max_members deve estar entre 1 e 1000")
	}
	return nil
}

// visibleRoom devolve a sala se o usuário puder vê-la: pública, ou ele é membro.
// Para salas privadas de quem não é membro responde "não encontrada" (não revela que existe).
func (s *service) visibleRoom(ctx context.Context, actorID, roomID string) (Room, error) {
	r, err := s.st.room(ctx, roomID)
	if err != nil {
		return Room{}, err
	}
	if r.IsPublic {
		return r, nil
	}
	ok, err := s.IsMember(ctx, roomID, actorID)
	if err != nil {
		return Room{}, err
	}
	if !ok {
		return Room{}, ErrRoomNotFound
	}
	return r, nil
}

func (s *service) getRoom(ctx context.Context, actorID, roomID string) (Room, error) {
	return s.visibleRoom(ctx, actorID, roomID)
}

func (s *service) listPublic(ctx context.Context, limit, offset int, keyword string) ([]Room, error) {
	return s.st.listPublic(ctx, limit, offset, strings.TrimSpace(keyword))
}

func (s *service) updateRoom(ctx context.Context, actorID, roomID string, p roomPatch) (Room, error) {
	if _, err := s.st.room(ctx, roomID); err != nil {
		return Room{}, err
	}
	m, err := s.st.member(ctx, roomID, actorID)
	if errors.Is(err, ErrMemberNotFound) {
		return Room{}, ErrForbidden
	}
	if err != nil {
		return Room{}, err
	}
	if m.Role != RoleOwner && m.Role != RoleAdmin {
		return Room{}, ErrForbidden
	}
	if p.Name != nil {
		n := strings.TrimSpace(*p.Name)
		if err := validateName(n); err != nil {
			return Room{}, err
		}
		p.Name = &n
	}
	if p.Description != nil && len(*p.Description) > maxDescriptionLen {
		return Room{}, invalid("descrição longa demais")
	}
	if p.MaxMembers != nil {
		if err := validateMax(*p.MaxMembers); err != nil {
			return Room{}, err
		}
	}
	return s.st.update(ctx, roomID, p)
}

func (s *service) deleteRoom(ctx context.Context, actorID, roomID string) error {
	r, err := s.st.room(ctx, roomID)
	if err != nil {
		return err
	}
	if r.OwnerID != actorID {
		return ErrForbidden
	}
	return s.st.softDelete(ctx, roomID)
}

func (s *service) addMember(ctx context.Context, actorID, roomID, userID, email string, role Role) error {
	if role == "" {
		role = RoleMember
	}
	if !role.valid() || role == RoleOwner {
		return invalid("role inválida (use admin, moderator, member ou guest)")
	}
	r, err := s.st.room(ctx, roomID)
	if err != nil {
		return err
	}
	if r.OwnerID != actorID {
		return ErrForbidden
	}
	switch {
	case userID == "" && strings.TrimSpace(email) == "":
		return invalid("informe user_id ou email")
	case userID == "":
		u, err := s.users.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
		if errors.Is(err, identity.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return err
		}
		userID = u.ID
	default:
		if ok, err := s.users.Exists(ctx, userID); err != nil {
			return err
		} else if !ok {
			return ErrUserNotFound
		}
	}
	return s.st.addMember(ctx, roomID, userID, role)
}

func (s *service) removeMember(ctx context.Context, actorID, roomID, userID string) error {
	r, err := s.st.room(ctx, roomID)
	if err != nil {
		return err
	}
	if actorID != userID && actorID != r.OwnerID {
		return ErrForbidden
	}
	if userID == r.OwnerID {
		return ErrCannotRemoveOwner
	}
	return s.st.removeMember(ctx, roomID, userID)
}

func (s *service) setRole(ctx context.Context, actorID, roomID, userID string, role Role) error {
	r, err := s.st.room(ctx, roomID)
	if err != nil {
		return err
	}
	if actorID != r.OwnerID {
		return ErrForbidden
	}
	if userID == r.OwnerID {
		return ErrCannotChangeOwner
	}
	if !role.valid() || role == RoleOwner {
		return invalid("role inválida (use admin, moderator, member ou guest)")
	}
	return s.st.setRole(ctx, roomID, userID, role)
}

// members resolve todos os usuários numa única chamada (Directory.GetMany): sem N+1.
// Membros cujo usuário foi removido não aparecem (mesmo comportamento do serviço antigo).
func (s *service) members(ctx context.Context, actorID, roomID string) ([]MemberView, error) {
	if _, err := s.visibleRoom(ctx, actorID, roomID); err != nil {
		return nil, err
	}
	ms, err := s.st.members(ctx, roomID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(ms))
	for i, m := range ms {
		ids[i] = m.UserID
	}
	users, err := s.users.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]MemberView, 0, len(ms))
	for _, m := range ms {
		if u, ok := users[m.UserID]; ok {
			out = append(out, MemberView{Member: m, User: u})
		}
	}
	return out, nil
}

func (s *service) join(ctx context.Context, userID, roomID string) error {
	r, err := s.st.room(ctx, roomID)
	if err != nil {
		return err
	}
	if ok, err := s.IsMember(ctx, roomID, userID); err != nil {
		return err
	} else if ok {
		return nil // já é membro: sucesso silencioso (como antes)
	}
	if !r.IsPublic {
		return ErrForbidden
	}
	if ok, err := s.users.Exists(ctx, userID); err != nil {
		return err
	} else if !ok {
		return ErrUserNotFound
	}
	err = s.st.addMember(ctx, roomID, userID, RoleMember)
	if errors.Is(err, ErrMemberAlreadyExists) { // corrida entre dois joins
		return nil
	}
	return err
}

func (s *service) roomsOf(ctx context.Context, userID string) ([]Room, error) {
	return s.st.roomsOf(ctx, userID)
}

func (s *service) ownedBy(ctx context.Context, userID string) ([]Room, error) {
	return s.st.ownedBy(ctx, userID)
}

// IsMember implementa Membership.
func (s *service) IsMember(ctx context.Context, roomID, userID string) (bool, error) {
	_, err := s.st.member(ctx, roomID, userID)
	if errors.Is(err, ErrMemberNotFound) {
		return false, nil
	}
	return err == nil, err
}
