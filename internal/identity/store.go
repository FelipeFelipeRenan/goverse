package identity

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRow struct {
	User
	passwordHash string
}

type newUser struct {
	Username     string
	Email        string // já normalizado (minúsculas)
	PasswordHash string
	Picture      string
	IsOAuth      bool
}

// store é a fronteira com o Postgres; existe como interface para testar os handlers sem banco.
type store interface {
	create(ctx context.Context, u newUser) (User, error)
	byID(ctx context.Context, id string) (userRow, error)
	byEmail(ctx context.Context, email string) (userRow, error)
	many(ctx context.Context, ids []string) ([]User, error)
	list(ctx context.Context) ([]User, error)
	update(ctx context.Context, id, username, picture string) (User, error)
	softDelete(ctx context.Context, id string) error
}

const userCols = `id::text, username, email, picture, is_oauth, created_at`

type pgStore struct{ pool *pgxpool.Pool }

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Picture, &u.IsOAuth, &u.CreatedAt)
	return u, mapErr(err)
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "22P02": // invalid_text_representation: id que não é UUID
			return ErrNotFound
		case "23505": // unique_violation
			return ErrEmailTaken
		}
	}
	return err
}

func (s *pgStore) create(ctx context.Context, u newUser) (User, error) {
	return scanUser(s.pool.QueryRow(ctx,
		`INSERT INTO identity.users (username, email, password_hash, picture, is_oauth)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+userCols,
		u.Username, u.Email, u.PasswordHash, u.Picture, u.IsOAuth))
}

func (s *pgStore) byID(ctx context.Context, id string) (userRow, error) {
	return s.withHash(ctx, `WHERE id = $1 AND deleted_at IS NULL`, id)
}

func (s *pgStore) byEmail(ctx context.Context, email string) (userRow, error) {
	return s.withHash(ctx, `WHERE lower(email) = lower($1) AND deleted_at IS NULL`, email)
}

func (s *pgStore) withHash(ctx context.Context, where string, arg any) (userRow, error) {
	var r userRow
	err := s.pool.QueryRow(ctx, `SELECT `+userCols+`, password_hash FROM identity.users `+where, arg).
		Scan(&r.ID, &r.Username, &r.Email, &r.Picture, &r.IsOAuth, &r.CreatedAt, &r.passwordHash)
	return r, mapErr(err)
}

func (s *pgStore) many(ctx context.Context, ids []string) ([]User, error) {
	return s.query(ctx, `SELECT `+userCols+` FROM identity.users WHERE id = ANY($1::uuid[]) AND deleted_at IS NULL`, ids)
}

func (s *pgStore) list(ctx context.Context) ([]User, error) {
	return s.query(ctx, `SELECT `+userCols+` FROM identity.users WHERE deleted_at IS NULL ORDER BY created_at`)
}

func (s *pgStore) query(ctx context.Context, sql string, args ...any) ([]User, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Picture, &u.IsOAuth, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *pgStore) update(ctx context.Context, id, username, picture string) (User, error) {
	return scanUser(s.pool.QueryRow(ctx,
		`UPDATE identity.users
		    SET username = $2, picture = COALESCE(NULLIF($3, ''), picture), updated_at = now()
		  WHERE id = $1 AND deleted_at IS NULL
		  RETURNING `+userCols, id, username, picture))
}

func (s *pgStore) softDelete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE identity.users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
