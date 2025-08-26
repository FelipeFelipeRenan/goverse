package service

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/repository"
)

type MessageService interface {
	ProcessAndSaveMessage(ctx context.Context, msg *domain.Message) error
}

type messageService struct {
	repo repository.MessageRepository
}

func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{repo: repo}
}

// ProcessAndSaveMessage implements MessageService.
func (m *messageService) ProcessAndSaveMessage(ctx context.Context, msg *domain.Message) error {
	return m.repo.SaveMessage(ctx, msg)
}
