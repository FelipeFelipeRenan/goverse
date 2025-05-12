package repository

import (
	"context"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/jackc/pgx/v5"
)


type AuthRepository interface {
	ValidateCredentials(ctx context.Context, email, password string)(*domain.Credentials, error )
}


type authRepository struct {
  conn *pgx.Conn
}

func NewAuthRepository(conn *pgx.Conn) AuthRepository{
	return &authRepository{conn:conn}
}

func (r *authRepository) ValidateCredentials(ctx context.Context, email, password string)(*domain.Credentials, error){
	query := `SELECT email, password FROM users WHERE email = $1`

	var stored domain.Credentials
	err := r.conn.QueryRow(ctx, query, email).Scan(&stored.Email, &stored.Password)
	if err != nil{
		if err == pgx.ErrNoRows{
			return nil, fmt.Errorf("usuario nao encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar usuario: %w", err)
	}
	return &stored, nil
}