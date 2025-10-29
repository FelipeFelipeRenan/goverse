package client

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger"
)

type RoomClient interface {
	IsUserMember(ctx context.Context, roomID, userID string) (bool, error)
}

type roomClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewRoomClient() RoomClient {
	baseURL := os.Getenv("ROOM_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://room-service:8080"
		logger.Warn("ROOM_SERVICE_URL não definido, usando padrão", "url", baseURL)
	}

	return &roomClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}

}

// IsUserMember implements RoomClient.
func (c *roomClient) IsUserMember(ctx context.Context, roomID string, userID string) (bool, error) {
	url := fmt.Sprintf("%s/internal/rooms/%s/members/%s", c.baseURL, roomID, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("falha ao criar requisição: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("falha ao chamar room-service: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return false, nil
	}

	return false, fmt.Errorf("room-service respondeu com status inesperado: %d", resp.StatusCode)
}
