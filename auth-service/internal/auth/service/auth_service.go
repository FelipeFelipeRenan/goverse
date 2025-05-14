package service

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Authenticate(ctx context.Context, credentials domain.Credentials) (*domain.TokenResponse, error)
}

type authService struct {
	authMethods map[string]AuthMethod
	jwtKey []byte
}

func NewAuthService(authMethods map[string]AuthMethod, jwtKey []byte) AuthService{
	return &authService{
		authMethods: authMethods,
		jwtKey: jwtKey,
	}
}

func (s *authService) Authenticate(ctx context.Context, credentials domain.Credentials)(*domain.TokenResponse, error){

	method, ok := s.authMethods[credentials.Type]

	if !ok {
		return nil, fmt.Errorf("metodo de autenticação não suportado: %s",credentials.Type)
	}

	user, err := method.Authenticate(ctx, credentials)
	if err != nil {
		return nil, fmt.Errorf("falha na autenticação> %w", err)
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id" : user.ID,
		"user_name": user.Username,
		"exp" : time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar token: %w", err)
	}
	return &domain.TokenResponse{Token:tokenString}, nil
}