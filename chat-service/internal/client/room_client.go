package client

import (
	"context"
	"fmt"

	roompb "github.com/FelipeFelipeRenan/goverse/proto/room"
	"google.golang.org/grpc"
)

type RoomClient interface {
	IsUserMember(ctx context.Context, roomID, userID string) (bool, error)
}

type roomClient struct {
	grpcClient roompb.RoomServiceClient
}

func NewRoomClient(coon *grpc.ClientConn) RoomClient {
	return &roomClient{
		grpcClient: roompb.NewRoomServiceClient(coon),
	}

}

// IsUserMember implements RoomClient.
func (c *roomClient) IsUserMember(ctx context.Context, roomID string, userID string) (bool, error) {

	req := &roompb.IsMemberRequest{
		RoomId: roomID,
		UserId: userID,
	}

	resp, err := c.grpcClient.IsMember(ctx, req)
	if err != nil {
		return false, fmt.Errorf("falha ao chamar room-service via gRPC: %w", err)
	}

	return resp.IsMember, nil
}
