package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) (string, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)

}

type userRepository struct {
	conn *pgx.Conn
}

func NewUserRepository(conn *pgx.Conn) UserRepository {
	return &userRepository{conn: conn}
}

func (r *userRepository) CreateUser(ctx context.Context, user domain.User) (string, error) {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id;
	
	`
	var id string
	err := r.conn.QueryRow(ctx, query, user.Username, user.Email, user.Password).Scan(&id)
	if err != nil {
		return "0", fmt.Errorf("Erro ao inserir usuario:  %w", err)
	}

	return id, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, created_at
		FROM users
		WHERE id = $1
	`

	row := r.conn.QueryRow(ctx, query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("Usuario nao encontrado: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]domain.User, error){
	rows, err := r.conn.Query(ctx, `SELECT id, username, email FROM users`)
	if err != nil {
		return nil, fmt.Errorf("erro na busca de usuarios: %w", err)
	}

	defer rows.Close()

	var users []domain.User
	for rows.Next(){
		var u domain.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email)
		if err != nil {
			return nil, fmt.Errorf("erro na busca de usuarios: %w", err)
		}
		users = append(users, u)
	}

	if rows.Err() != nil{
		return nil, fmt.Errorf("erro na busca de usuarios: %w", rows.Err())
	}
	return users, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, created_at
		FROM users
		WHERE email = $1
	`

	row := r.conn.QueryRow(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("usuário não encontrado com e-mail %s: %w", email, err)
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário por e-mail: %w", err)
	}
	return &user, nil
}

