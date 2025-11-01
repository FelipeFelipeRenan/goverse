package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) (*domain.UserResponse, error)
	UpdateUser(ctx context.Context, id string, user domain.User) (*domain.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type userRepository struct {
	conn *pgxpool.Pool
}

func NewUserRepository(conn *pgxpool.Pool) UserRepository {
	return &userRepository{conn: conn}
}

func (r *userRepository) CreateUser(ctx context.Context, user domain.User) (*domain.UserResponse, error) {
	// Verificando se todos os campos estão preenchidos corretamente
	if user.Username == "" || user.Email == "" {
		return nil, fmt.Errorf("dados incompletos para registro")
	}

	// Salvando o usuário no banco de dados
	query := `
	INSERT INTO users (username, email, password, picture, created_at, is_oauth)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`
	var id string
	err := r.conn.QueryRow(ctx, query, user.Username, user.Email, user.Password, user.Picture, user.CreatedAt, user.IsOAuth).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("erro ao inserir usuario:  %w", err)
	}

	return &domain.UserResponse{
		ID:        id,
		Username:  user.Username,
		Email:     user.Email,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt,
		IsOAuth:   user.IsOAuth,
	}, nil
}

// UpdateUser implements UserRepository.
func (r *userRepository) UpdateUser(ctx context.Context, id string, user domain.User) (*domain.UserResponse, error) {
	query := `SET username = $1, picture = $2, updated_at = now()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING id, username, email, picture, created_at, is_oauth;`

	row := r.conn.QueryRow(ctx, query, user.Username, user.Picture, id)

	var updated domain.UserResponse
	err := row.Scan(&updated.ID, &updated.Username, &updated.Email, &updated.Picture, &updated.CreatedAt, &updated.IsOAuth)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &updated, nil

}

// DeleteUser implements UserRepository.
func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	query := `
					UPDATE users
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL;
		`

	cmdTag, err := r.conn.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
			SELECT id, username, email, password, picture, created_at, is_oauth
			FROM users
			WHERE id = $1 AND deleted_at IS NULL
		`
	row := r.conn.QueryRow(ctx, query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Picture, &user.CreatedAt, &user.IsOAuth)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("Usuario nao encontrado: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.conn.Query(ctx, `SELECT id, username, email, password, picture, created_at, is_oauth FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("erro na busca de usuarios: %w", err)
	}

	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Picture, &u.CreatedAt, &u.IsOAuth)
		if err != nil {
			return nil, fmt.Errorf("erro na busca de usuarios: %w", err)
		}
		users = append(users, u)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("erro na busca de usuarios: %w", rows.Err())
	}
	return users, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := "SELECT id, username, email, password, picture, created_at, is_oauth FROM users WHERE email = $1 AND deleted_at IS NULL"
	row := r.conn.QueryRow(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Picture, &user.CreatedAt, &user.IsOAuth)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário por e-mail: %w", err)
	}
	return &user, nil
}
