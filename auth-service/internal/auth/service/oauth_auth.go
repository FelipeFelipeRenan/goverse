package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthAuth struct {
	config     *oauth2.Config
	repository repository.AuthRepository
}

func NewOAuthAuth(repository repository.AuthRepository) *OAuthAuth {
	return &OAuthAuth{
		config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
		repository: repository,
	}
}

func (a *OAuthAuth) Authenticate(ctx context.Context, credentials domain.Credentials) (*domain.UserResponse, error) {
	if credentials.Token == "" {
		return nil, fmt.Errorf("token (authorization code) não fornecido")
	}

	// Troca o código pelo token de acesso
	token, err := a.config.Exchange(ctx, credentials.Token)
	if err != nil {
		return nil, fmt.Errorf("erro ao trocar código por token: %w", err)
	}

	// Faz a requisição para obter dados do usuário
	client := a.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar dados do usuário: %w", err)
	}
	defer resp.Body.Close()

	// Decodifica os dados do usuário
	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta do Google: %w", err)
	}

	// Tenta buscar o usuário no user-service
	user, err := a.repository.FindByEmail(ctx, userInfo.Email)
	if err == nil {
		return user, nil // usuário já existe
	}

	// Se não existir, registra o usuário
	newUser := domain.User{
		Email:    userInfo.Email,
		Username: userInfo.Name,
		Password: "-", // senha dummy, pois é OAuth
		Picture:  userInfo.Picture,
	}
	userID, err := a.repository.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar novo usuário: %w", err)
	}

	return &domain.UserResponse{
		ID:       userID,
		Email:    newUser.Email,
		Username: newUser.Username,
		Picture:  newUser.Picture,
	}, nil
}


// NOVO MÉTODO: gera a URL de login OAuth
func (a *OAuthAuth) GetAuthURL(state string) string {
	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (a *OAuthAuth) GetOAuthURL(state string) string {
	return a.config.AuthCodeURL(state)
}
