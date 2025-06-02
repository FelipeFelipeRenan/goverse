package domain

import "errors"

var (
	ErrUnauthorized          = errors.New("ação não autorizada")
	ErrUserNotFound          = errors.New("usuário não encontrado")
	ErrMemberAlreadyExists   = errors.New("usuário já é membro da sala")
	ErrCannotRemoveOwner     = errors.New("não é possível remover o dono da sala")
	ErrCannotUpdateOwnerRole = errors.New("não é possível alterar a role do dono da sala")
	ErrMemberNotFound        = errors.New("membro não encontrado")
)
