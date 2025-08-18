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
	Authenticate(ctx context.Context, credentials domain.Credentials) (string, *domain.UserResponse, error)
	GetOAuthURL(state string) string
	GetUserByID(ctx context.Context, userID string) (*domain.UserResponse, error)
}

type authService struct {
	authMethods map[string]AuthMethod
	repository  repository.AuthRepository
	jwtKey      []byte
}

func NewAuthService(authMethods map[string]AuthMethod, repo repository.AuthRepository, jwtKey []byte) AuthService {
	return &authService{
		authMethods: authMethods,
		repository:  repo,
		jwtKey:      jwtKey,
	}
}

func (s *authService) Authenticate(ctx context.Context, credentials domain.Credentials) (string, *domain.UserResponse, error) {

	method, ok := s.authMethods[credentials.Type]

	if !ok {
		return "", nil, fmt.Errorf("metodo de autenticação não suportado: %s", credentials.Type)
	}

	user, err := method.Authenticate(ctx, credentials)
	if err != nil {
		return "", nil, fmt.Errorf("falha na autenticação> %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(s.jwtKey)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao fazer parse de chave privada: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id":   user.ID,
		"user_name": user.Username,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao assinar token: %w", err)
	}
	return tokenString, user, nil
}

func (s *authService) GetOAuthURL(state string) string {
	// Verifica se existe o método "oauth" registrado
	method, ok := s.authMethods["oauth"]
	if !ok {
		return ""
	}

	// Faz type assertion para a estratégia correta
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
