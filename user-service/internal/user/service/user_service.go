package service

import (
	"context"
	"errors"
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
	Authenticate(ctx context.Context, email, password string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error)
	ExistsByID(ctx context.Context, id string) (bool, error)

	UpdateUser(ctx context.Context, id string, user domain.User) (*domain.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
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

// DeleteUser implements UserService.
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

// UpdateUser implements UserService.
func (s *userService) UpdateUser(ctx context.Context, id string, user domain.User) (*domain.UserResponse, error) {
	if user.Username == "" {
		return nil, errors.New("username não pode ser vazio")
	}
	return s.repo.UpdateUser(ctx, id, user)
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

func (s *userService) GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário por email: %w", err)
	}
	return &domain.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Picture:  user.Picture,
	}, nil
}

func (s *userService) ExistsByID(ctx context.Context, id string) (bool, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}
	return user != nil, nil
}
