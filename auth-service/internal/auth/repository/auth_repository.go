package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthRepository interface {
	ValidateCredentials(ctx context.Context, email, password string) (*userpb.UserResponse, error)
	FindByEmail(ctx context.Context, email string) (*domain.UserResponse, error)
	CreateUser(ctx context.Context, user domain.User) (*domain.UserResponse, error)
}

type authRepository struct {
	userClient userpb.UserServiceClient
}

func NewAuthRepository(userClient userpb.UserServiceClient) AuthRepository {
	return &authRepository{userClient: userClient}
}

func (r *authRepository) ValidateCredentials(ctx context.Context, email, password string) (*userpb.UserResponse, error) {

	req := &userpb.CredentialsRequest{
		Email:    email,
		Password: password,
	}

	return r.userClient.ValidateCredentials(ctx, req)
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	req := &userpb.EmailRequest{
		Email: email,
	}

	resp, err := r.userClient.GetUserByEmail(ctx, req)
	if err != nil {
		// Tratamento do erro NotFound
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return nil, domain.ErrUserNotFound
		}
		// Qualquer outro erro é retornado normalmente
		return nil, fmt.Errorf("erro ao buscar usuário por email: %w", err)
	}

	return &domain.UserResponse{
		ID:       resp.Id,
		Username: resp.Name,
		Email:    resp.Email,
	}, nil
}

func (r *authRepository) CreateUser(ctx context.Context, user domain.User) (*domain.UserResponse, error) {
	now := time.Now()

	req := &userpb.RegisterRequest{
		Name:      user.Username,            // Nome do usuário
		Email:     user.Email,               // E-mail do usuário
		Password:  "",                       // OAuth não tem senha
		Picture:   user.Picture,             // Foto de perfil
		CreatedAt: now.Format(time.RFC3339), // Data de criação
		IsOauth:   user.Is_OAuth,
	}

	// Chamada ao serviço gRPC
	resp, err := r.userClient.Register(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar novo usuário: %w", err)
	}

	return &domain.UserResponse{
		ID:       resp.Id,      // ID gerado no registro
		Username: resp.Name,    // Nome do usuário
		Email:    resp.Email,   // E-mail do usuário
		Picture:  resp.Picture, // Foto do usuário
	}, nil
}
