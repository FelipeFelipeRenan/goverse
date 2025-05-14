package handler

import (
	"context"
	"log"

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

func NewGRPCHandler(userService service.UserService) userpb.UserServiceServer{
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

    log.Printf("Senha enviada: %s | Senha armazenada: %s", req.Password, user.Password)

    return &userpb.UserResponse{
        Id:   user.ID,
        Name: user.Username,
    }, nil
}

func (h *GRPCHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	user := domain.User{
		Username: req.Name,
		Email:    req.Email,
		Password: req.Password,
		Picture:  req.Picture,
	}

	id, err := h.userService.Register(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao registrar usuário: %v", err)
	}

	return &userpb.RegisterResponse{Id: id}, nil
}
