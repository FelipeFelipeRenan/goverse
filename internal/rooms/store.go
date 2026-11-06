package rooms

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type newRoom struct {
	OwnerID     string
	Name        string
	Description string
	IsPublic    bool
	MaxMembers  int
}

// roomPatch: campos nil não mudam.
type roomPatch struct {
	Name        *string
	Description *string
	IsPublic    *bool
	MaxMembers  *int
}

// store é a fronteira com o Postgres. As operações que mexem em member_count
// (createRoom, addMember, removeMember) são transacionais e atômicas aqui dentro.
type store interface {
	createRoom(ctx context.Context, r newRoom) (Room, error)
	room(ctx context.Context, id string) (Room, error)
	listPublic(ctx context.Context, limit, offset int, keyword string) ([]Room, error)
	roomsOf(ctx context.Context, userID string) ([]Room, error)
	ownedBy(ctx context.Context, userID string) ([]Room, error)
	update(ctx context.Context, id string, p roomPatch) (Room, error)
	softDelete(ctx context.Context, id string) error
	member(ctx context.Context, roomID, userID string) (Member, error)
	members(ctx context.Context, roomID string) ([]Member, error)
	addMember(ctx context.Context, roomID, userID string, role Role) error
	removeMember(ctx context.Context, roomID, userID string) error
	setRole(ctx context.Context, roomID, userID string, role Role) error
}

type pgStore struct{ pool *pgxpool.Pool }

type scanner interface{ Scan(...any) error }

func roomCols(a string) string {
	return a + `.id::text, ` + a + `.name, ` + a + `.description, ` + a + `.is_public, ` +
		a + `.owner_id::text, ` + a + `.member_count, ` + a + `.max_members, ` + a + `.created_at, ` + a + `.updated_at`
}

func scanRoom(s scanner) (Room, error) {
	var r Room
	err := s.Scan(&r.ID, &r.Name, &r.Description, &r.IsPublic, &r.OwnerID,
		&r.MemberCount, &r.MaxMembers, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "22P02": // id que não é UUID
			return ErrRoomNotFound
		case "23505": // unique_violation: (room_id,user_id) ou segundo owner
			return ErrMemberAlreadyExists
		case "23503": // foreign_key_violation: usuário inexistente
			return ErrUserNotFound
		case "23514": // check_violation: max_members < member_count etc.
			return invalid("valor inválido: max_members não pode ser menor que o número de membros")
		}
	}
	return err
}

func (s *pgStore) createRoom(ctx context.Context, n newRoom) (Room, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Room{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	r, err := scanRoom(tx.QueryRow(ctx,
		`INSERT INTO rooms.rooms AS r (name, description, is_public, max_members, owner_id, member_count)
		 VALUES ($1, $2, $3, $4, $5, 1) RETURNING `+roomCols("r"),
		n.Name, n.Description, n.IsPublic, n.MaxMembers, n.OwnerID))
	if err != nil {
		return Room{}, mapErr(err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO rooms.members (room_id, user_id, role) VALUES ($1, $2, 'owner')`, r.ID, n.OwnerID); err != nil {
		return Room{}, mapErr(err)
	}
	return r, tx.Commit(ctx)
}

func (s *pgStore) room(ctx context.Context, id string) (Room, error) {
	r, err := scanRoom(s.pool.QueryRow(ctx,
		`SELECT `+roomCols("r")+` FROM rooms.rooms r WHERE r.id = $1 AND r.deleted_at IS NULL`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Room{}, ErrRoomNotFound
	}
	return r, mapErr(err)
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (s *pgStore) listPublic(ctx context.Context, limit, offset int, keyword string) ([]Room, error) {
	return s.queryRooms(ctx,
		`SELECT `+roomCols("r")+` FROM rooms.rooms r
		  WHERE r.is_public AND r.deleted_at IS NULL
		    AND ($3::text = '' OR r.name ILIKE '%' || $3::text || '%')
		  ORDER BY r.created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset, likeEscaper.Replace(keyword))
}

func (s *pgStore) roomsOf(ctx context.Context, userID string) ([]Room, error) {
	return s.queryRooms(ctx,
		`SELECT `+roomCols("r")+` FROM rooms.rooms r
		   JOIN rooms.members m ON m.room_id = r.id
		  WHERE m.user_id = $1 AND r.deleted_at IS NULL ORDER BY r.created_at DESC`, userID)
}

func (s *pgStore) ownedBy(ctx context.Context, userID string) ([]Room, error) {
	return s.queryRooms(ctx,
		`SELECT `+roomCols("r")+` FROM rooms.rooms r
		  WHERE r.owner_id = $1 AND r.deleted_at IS NULL ORDER BY r.created_at DESC`, userID)
}

func (s *pgStore) queryRooms(ctx context.Context, sql string, args ...any) ([]Room, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []Room{}
	for rows.Next() {
		r, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *pgStore) update(ctx context.Context, id string, p roomPatch) (Room, error) {
	r, err := scanRoom(s.pool.QueryRow(ctx,
		`UPDATE rooms.rooms r
		    SET name = COALESCE($2, r.name), description = COALESCE($3, r.description),
		        is_public = COALESCE($4, r.is_public), max_members = COALESCE($5, r.max_members),
		        updated_at = now()
		  WHERE r.id = $1 AND r.deleted_at IS NULL
		  RETURNING `+roomCols("r"), id, p.Name, p.Description, p.IsPublic, p.MaxMembers))
	if errors.Is(err, pgx.ErrNoRows) {
		return Room{}, ErrRoomNotFound
	}
	return r, mapErr(err)
}

func (s *pgStore) softDelete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE rooms.rooms SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRoomNotFound
	}
	return nil
}

func (s *pgStore) member(ctx context.Context, roomID, userID string) (Member, error) {
	var m Member
	err := s.pool.QueryRow(ctx,
		`SELECT m.room_id::text, m.user_id::text, m.role, m.joined_at
		   FROM rooms.members m JOIN rooms.rooms r ON r.id = m.room_id AND r.deleted_at IS NULL
		  WHERE m.room_id = $1 AND m.user_id = $2`, roomID, userID).
		Scan(&m.RoomID, &m.UserID, &m.Role, &m.JoinedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}
	if err != nil {
		if mapped := mapErr(err); errors.Is(mapped, ErrRoomNotFound) {
			return Member{}, ErrMemberNotFound
		}
		return Member{}, mapErr(err)
	}
	return m, nil
}

func (s *pgStore) members(ctx context.Context, roomID string) ([]Member, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT room_id::text, user_id::text, role, joined_at FROM rooms.members
		  WHERE room_id = $1 ORDER BY joined_at`, roomID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.RoomID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// addMember reserva a vaga de forma atômica (member_count < max_members) e insere o membro.
func (s *pgStore) addMember(ctx context.Context, roomID, userID string, role Role) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE rooms.rooms SET member_count = member_count + 1, updated_at = now()
		  WHERE id = $1 AND deleted_at IS NULL AND member_count < max_members`, roomID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM rooms.rooms WHERE id = $1 AND deleted_at IS NULL)`, roomID).Scan(&exists); err != nil {
			return mapErr(err)
		}
		if !exists {
			return ErrRoomNotFound
		}
		return ErrRoomFull
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO rooms.members (room_id, user_id, role) VALUES ($1, $2, $3)`, roomID, userID, string(role)); err != nil {
		return mapErr(err)
	}
	return tx.Commit(ctx)
}

func (s *pgStore) removeMember(ctx context.Context, roomID, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `DELETE FROM rooms.members WHERE room_id = $1 AND user_id = $2`, roomID, userID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	if _, err := tx.Exec(ctx,
		`UPDATE rooms.rooms SET member_count = member_count - 1, updated_at = now() WHERE id = $1`, roomID); err != nil {
		return mapErr(err)
	}
	return tx.Commit(ctx)
}

func (s *pgStore) setRole(ctx context.Context, roomID, userID string, role Role) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE rooms.members SET role = $3 WHERE room_id = $1 AND user_id = $2`, roomID, userID, string(role))
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}
