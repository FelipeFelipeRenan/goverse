package service

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/repository"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Authenticate(ctx context.Context, credentials domain.Credentials) (*domain.TokenResponse, error)
}

type authService struct {
	authRepo repository.AuthRepository
	jwtKey []byte
}

func NewAuthService(authRepo repository.AuthRepository, jwtKey []byte) AuthService{
	return &authService{
		authRepo: authRepo,
		jwtKey: jwtKey,
	}
}

func (s *authService) Authenticate(ctx context.Context, credentials domain.Credentials)(*domain.TokenResponse, error){

	userResp, err := s.authRepo.ValidateCredentials(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("erro ao validar credenciais: %w", err)
	}

	// Gerar token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id" : userResp.Id,
		"username" : userResp.Name,
		"exp": time.Now().Add(time.Hour *24).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil{
		return nil, fmt.Errorf("erro ao assinar o token: %w", err)
	}
	return &domain.TokenResponse{Token: tokenString}, nil
}