package service

import "context"

type UserValidator interface {
	IsUserValid(ctx context.Context, userID string) (bool, error)
}
