package service

import (
	"context"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, user domain.User) (string, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Register(ctx context.Context, user domain.User) (string, error) {
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return "", fmt.Errorf("dados incompletos para registro")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}
	user.Password = string(hashedPassword)
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAllUsers(ctx)
}
