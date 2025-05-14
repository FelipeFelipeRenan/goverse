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
	// trocar o código por um access token
	token, err := a.config.Exchange(ctx, credentials.Token)
	if err != nil {
		return nil, fmt.Errorf("eror ao trocar código por token: %w", err)
	}

	client := a.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar dados do usuário: %w", err)
	}
	defer resp.Body.Close()
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

	// Validação real com o user-service:
	user, err := a.repository.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	return user, nil

}
