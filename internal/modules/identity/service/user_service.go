package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/modules/identity/domain"
	"github.com/FelipeFelipeRenan/goverse/internal/modules/identity/port"
	"golang.org/x/crypto/bcrypt"
)

// UserService contains the user-management use cases shared by identity.
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
	repository port.UserRepository
}

// NewUserService creates the user-management application service.
func NewUserService(repository port.UserRepository) UserService {
	return &userService{repository: repository}
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

	return s.repository.CreateUser(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repository.DeleteUser(ctx, id)
}

func (s *userService) UpdateUser(ctx context.Context, id string, user domain.User) (*domain.UserResponse, error) {
	if user.Username == "" {
		return nil, errors.New("username não pode ser vazio")
	}

	return s.repository.UpdateUser(ctx, id, user)
}

func (s *userService) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repository.GetUserByID(ctx, id)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.repository.GetAllUsers(ctx)
}

func (s *userService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("usuario nao encontrado: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("senha invalida")
	}

	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	user, err := s.repository.GetUserByEmail(ctx, email)
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
	user, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}

		return false, err
	}

	return user != nil, nil
}

func generateRandomPassword(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for index := range bytes {
		bytes[index] = letters[bytes[index]%byte(len(letters))]
	}

	return string(bytes), nil
}
