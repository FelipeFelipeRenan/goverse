package handler

import (
	"context"

	roompb "github.com/FelipeFelipeRenan/goverse/proto/room"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	roompb.UnimplementedRoomServiceServer
	memberService service.MemberService
}

func NewGRPCHandler(ms service.MemberService) *GRPCHandler {
	return &GRPCHandler{memberService: ms}
}

func (h *GRPCHandler) IsMember(ctx context.Context, req *roompb.IsMemberRequest) (*roompb.IsMemberResponse, error) {
	isMember, err := h.memberService.IsMember(ctx, req.RoomId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao verificar membro: %v", err)
	}

	return &roompb.IsMemberResponse{IsMember: isMember}, nil
}
