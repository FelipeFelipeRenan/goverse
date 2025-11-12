package httpa

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/client"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/service"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger"
)

type MessageHandler struct {
	svc        service.MessageService
	roomClient client.RoomClient
}

func NewMessageHandler(svc service.MessageService, roomClient client.RoomClient) *MessageHandler {
	return &MessageHandler{
		svc:        svc,
		roomClient: roomClient,
	}
}

func (h *MessageHandler) GetMessagesByRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Não autorizado: X-User-ID não encontrado", http.StatusUnauthorized)
		return
	}

	roomID := r.PathValue("roomId")
	if roomID == "" {
		http.Error(w, "ID da sala é obrigatório", http.StatusBadRequest)
		return
	}

	isMember, err := h.roomClient.IsUserMember(ctx, roomID, userID)
	if err != nil {
		logger.Error("Erro ao verificar membro no room-service", "err", err, "userID", userID, "roomID", roomID)
		http.Error(w, "Erro interno ao verificar permissão", http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, "Acesso negado: Você não é membro desta sala", http.StatusUnauthorized)
		return
	}

	// parametros de paginação
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// buscar mensagens
	messages, err := h.svc.GetMessagesByRoomID(ctx, roomID, limit, offset)
	if err != nil {
		http.Error(w, "Erro ao buscar mensagens", http.StatusInternalServerError)
		return
	}

	// retornar JSON
	w.Header().Set("Content-Type", "application/json")
	if messages == nil {
		// retorna lista vazia ao invés de null
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(messages)

}
