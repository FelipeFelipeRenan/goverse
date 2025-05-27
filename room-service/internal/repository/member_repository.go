package repository

import (
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

type RoomMemberRepository interface {
	AddMember(member *domain.RoomMember) error
	RemoveMember(roomID, userID string) error
	GetMembers(roomID string) ([]*domain.RoomMember, error)
	GetUserRole(roomID, userID string) (domain.Role, error)
	IsMember(roomID, userID string) (bool, error)
}

type roomMemberRepository struct {
	db *pgx.Conn
}

func NewRoomMemberRepository(db *pgx.Conn) RoomMemberRepository {
	return &roomMemberRepository{db: db}
}

// AddMember implements RoomMemberRepository.
func (r *roomMemberRepository) AddMember(member *domain.RoomMember) error {
	panic("unimplemented")
}

// GetMembers implements RoomMemberRepository.
func (r *roomMemberRepository) GetMembers(roomID string) ([]*domain.RoomMember, error) {
	panic("unimplemented")
}

// GetUserRole implements RoomMemberRepository.
func (r *roomMemberRepository) GetUserRole(roomID string, userID string) (domain.Role, error) {
	panic("unimplemented")
}

// IsMember implements RoomMemberRepository.
func (r *roomMemberRepository) IsMember(roomID string, userID string) (bool, error) {
	panic("unimplemented")
}

// RemoveMember implements RoomMemberRepository.
func (r *roomMemberRepository) RemoveMember(roomID string, userID string) error {
	panic("unimplemented")
}
