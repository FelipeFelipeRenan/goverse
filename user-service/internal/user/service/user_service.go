package service

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, user domain.User) (*domain.UserResponse, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	Authenticate(ctx context.Context, email, password string) (*domain.User, error) // 👈 adicionado

}

type userService struct {
	repo repository.UserRepository
}


func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Register(ctx context.Context, user domain.User) (*domain.UserResponse, error) {
	
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return nil, fmt.Errorf("dados incompletos para registro")
	}

	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	user.CreatedAt = time.Now()
	user.Password = string(hashedPassword)
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAllUsers(ctx)
}

// Authenticate implements UserService.
func (s *userService) Authenticate(ctx context.Context, email string, password string) (*domain.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("usuario nao encontrado: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("senha invalida")
	}
	return user, nil
}

