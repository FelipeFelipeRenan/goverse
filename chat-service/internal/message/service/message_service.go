package service

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/repository"
)

type MessageService interface {
	ProcessAndSaveMessage(ctx context.Context, msg *domain.Message) error
	GetMessagesByRoomID(ctx context.Context, roomdID string, limit, offset int) ([]domain.Message, error)
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

// GetMessagesByRoomID implements MessageService.
func (m *messageService) GetMessagesByRoomID(ctx context.Context, roomdID string, limit int, offset int) ([]domain.Message, error) {
	
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return m.repo.GetMessagesByRoomID(ctx, roomdID, limit, offset)
}