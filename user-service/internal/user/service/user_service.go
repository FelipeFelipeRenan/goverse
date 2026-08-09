// Package service provides compatibility during the migration to the identity module.
package service

import (
	identityservice "github.com/FelipeFelipeRenan/goverse/internal/modules/identity/service"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
)

// UserService is retained for callers while the implementation lives in identity.
type UserService = identityservice.UserService

// NewUserService delegates to the centralized identity module.
func NewUserService(repository repository.UserRepository) UserService {
	return identityservice.NewUserService(repository)
}
