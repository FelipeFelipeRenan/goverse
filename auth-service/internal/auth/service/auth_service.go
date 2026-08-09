package service

import (
	"context"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/repository"
	"github.com/FelipeFelipeRenan/goverse/internal/modules/identity/token"
)

type AuthService interface {
	Authenticate(ctx context.Context, credentials domain.Credentials) (string, *domain.UserResponse, error)
	GetOAuthURL(state string) string
	GetUserByID(ctx context.Context, userID string) (*domain.UserResponse, error)
}

type authService struct {
	authMethods map[string]AuthMethod
	repository  repository.AuthRepository
	tokenIssuer *token.Issuer
}

func NewAuthService(authMethods map[string]AuthMethod, repository repository.AuthRepository, jwtKey []byte) AuthService {
	return &authService{
		authMethods: map[string]AuthMethod{},
		repository:  repository,
		tokenIssuer: token.NewIssuer(jwtKey),
	}
}

func (s *authService) Authenticate(ctx context.Context, credentials domain.Credentials) (string, *domain.UserResponse, error) {
	method, ok := s.authMethods[credentials.Type]
	if !ok {
		return "", nil, fmt.Errorf("metodo de autenticação não suportado: %s", credentials.Type)
	}

	user, err := method.Authenticate(ctx, credentials)
	if err != nil {
		return "", nil, fmt.Errorf("falha na autenticação: %w", err)
	}

	tokenString, err := s.tokenIssuer.Issue(user.ID, user.Username)
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

func (s *authService) GetOAuthURL(state string) string {
	method, ok := s.authMethods["oauth"]
	if !ok {
		return ""
	}

	oauthMethod, ok := method.(interface {
		GetOAuthURL(state string) string
	})
	if !ok {
		return ""
	}

	return oauthMethod.GetOAuthURL(state)
}

func (s *authService) GetUserByID(ctx context.Context, userID string) (*domain.UserResponse, error) {
	return s.repository.FindByID(ctx, userID)
}
