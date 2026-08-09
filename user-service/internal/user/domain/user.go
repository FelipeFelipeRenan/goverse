package domain

import identitydomain "github.com/FelipeFelipeRenan/goverse/internal/modules/identity/domain"

// Deprecated: use identitydomain.User.
type User = identitydomain.User

// Deprecated: use identitydomain.UserResponse.
type UserResponse = identitydomain.UserResponse

var ErrUserNotFound = identitydomain.ErrUserNotFound
