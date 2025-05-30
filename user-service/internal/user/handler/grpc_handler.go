package handler

import (
	"context"
	"errors"
	"time"

	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	userpb.UnimplementedUserServiceServer
	userService service.UserService
}

func NewGRPCHandler(userService service.UserService) userpb.UserServiceServer {
	return &GRPCHandler{userService: userService}
}

func (h *GRPCHandler) ValidateCredentials(ctx context.Context, req *userpb.CredentialsRequest) (*userpb.UserResponse, error) {
	user, err := h.userService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	// Validando se a senha fornecida corresponde ao hash armazenado
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unknown, "senha invalida!")
	}

	return &userpb.UserResponse{
		Id:        user.ID,
		Name:      user.Username,
		Email:     user.Email,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		IsOauth:   user.IsOAuth,
	}, nil
}

func (h *GRPCHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	// Validando o formato da data

	createdAt := time.Now()
	if req.CreatedAt != "" {
		t, err := time.Parse(time.RFC3339, req.CreatedAt)
		if err == nil {
			createdAt = t
		}
	}
	// Criando o usuário com os dados fornecidos
	user := domain.User{
		Username:  req.Name,
		Email:     req.Email,
		Password:  req.Password, // Aqui pode ser uma senha vazia, caso OAuth
		Picture:   req.Picture,  // Foto de perfil
		CreatedAt: createdAt,    // Data de criação
		IsOAuth:   req.IsOauth,
	}
	if user.Password == "" {
		user.Password, _ = generateRandomPassword(16)
		//	if err != nil {
		//		return nil, status.Errorf(codes.Internal, "erro ao gerar senha: %v", err)
		//	}
	}

	// Registrando no banco de dados
	id, err := h.userService.Register(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao registrar usuário: %v", err)
	}

	// Retornando a resposta do registro
	return &userpb.RegisterResponse{
		Id:        id.ID,
		Name:      user.Username,
		Email:     user.Email,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		IsOauth:   true,
	}, nil

}

func (s *GRPCHandler) GetUserByEmail(ctx context.Context, req *userpb.EmailRequest) (*userpb.UserResponse, error) {
	user, err := s.userService.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "usuário não encontrado")
		}
		return nil, status.Errorf(codes.Internal, "erro ao buscar usuário: %v", err)
	}

	return &userpb.UserResponse{
		Id:      user.ID,
		Email:   user.Email,
		Name:    user.Username,
		Picture: user.Picture,
	}, nil
}
