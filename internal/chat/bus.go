package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// event é o que uma réplica avisa às outras para elas entregarem aos seus clientes locais.
type event struct {
	Origin string   `json:"o"`
	Room   string   `json:"r"`
	Except string   `json:"x,omitempty"` // não entregar a conexões deste usuário
	Only   []string `json:"y,omitempty"` // se presente, só a estes usuários
	Msg    *Message `json:"m,omitempty"`
	MsgID  string   `json:"id,omitempty"` // payload grande demais: o receptor busca a mensagem no banco
}

// bus liga as réplicas. Entrega "pelo menos uma vez por réplica conectada": quem estava
// desconectado do barramento durante uma queda perde os eventos efêmeros (digitando, presença)
// e as mensagens ao vivo; o histórico no Postgres continua certo e o front recarrega ao reconectar.
type bus interface {
	publish(ctx context.Context, ev event) error
	run(ctx context.Context, handle func(event))
}

const (
	notifyChannel    = "goverse_chat"
	maxNotifyPayload = 7500 // limite do Postgres é 8000 bytes
)

var errPayloadTooLarge = errors.New("evento grande demais para NOTIFY")

func encodeEvent(ev event) ([]byte, error) {
	b, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}
	if len(b) <= maxNotifyPayload {
		return b, nil
	}
	if ev.Msg != nil && ev.Msg.Type == TypeChat && ev.Msg.ID != "" {
		return json.Marshal(event{Origin: ev.Origin, Room: ev.Room, MsgID: ev.Msg.ID})
	}
	return nil, errPayloadTooLarge
}

// pgBus usa LISTEN/NOTIFY do Postgres. Não funciona atrás de PgBouncer em modo transaction.
type pgBus struct{ pool *pgxpool.Pool }

func (b *pgBus) publish(ctx context.Context, ev event) error {
	payload, err := encodeEvent(ev)
	if err != nil {
		return err
	}
	_, err = b.pool.Exec(ctx, "SELECT pg_notify($1, $2)", notifyChannel, string(payload))
	return err
}

func (b *pgBus) run(ctx context.Context, handle func(event)) {
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		err := b.listen(ctx, handle)
		if ctx.Err() != nil {
			return
		}
		slog.Warn("chat: barramento caiu, reconectando", "err", err, "em", backoff)
		if time.Since(start) > 30*time.Second {
			backoff = time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, 15*time.Second)
	}
}

func (b *pgBus) listen(ctx context.Context, handle func(event)) error {
	// Conexão própria (fora do pool): LISTEN é por sessão e ocuparia uma vaga do pool para sempre.
	conn, err := pgx.ConnectConfig(ctx, b.pool.Config().ConnConfig)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	if _, err := conn.Exec(ctx, "LISTEN "+notifyChannel); err != nil {
		return err
	}
	slog.Info("chat: ouvindo o barramento", "canal", notifyChannel)
	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		var ev event
		if err := json.Unmarshal([]byte(n.Payload), &ev); err != nil {
			slog.Warn("chat: evento inválido no barramento", "err", fmt.Sprint(err))
			continue
		}
		handle(ev)
	}
}
