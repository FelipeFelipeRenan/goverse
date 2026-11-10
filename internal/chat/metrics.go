package chat

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Métricas de negócio do chat (OTLP -> Prometheus: chat_connections, chat_messages_total,
// chat_spool_pending, chat_bus_events_total; dead letters = chat_messages_total{result="dead_letter"}).
type metrics struct {
	conns metric.Int64UpDownCounter
	msgs  metric.Int64Counter
	bus   metric.Int64Counter
}

func newMetrics(h *hub) *metrics {
	m := otel.Meter("github.com/FelipeFelipeRenan/goverse/internal/chat")
	out := &metrics{}
	out.conns, _ = m.Int64UpDownCounter("chat.connections",
		metric.WithUnit("{connection}"), metric.WithDescription("conexões WebSocket abertas nesta réplica"))
	out.msgs, _ = m.Int64Counter("chat.messages",
		metric.WithUnit("{message}"), metric.WithDescription("mensagens de chat por resultado (ok, spooled, rejected, failed)"))
	out.bus, _ = m.Int64Counter("chat.bus.events",
		metric.WithUnit("{event}"), metric.WithDescription("eventos entre réplicas (direction=in|out, outcome)"))
	_, _ = m.Int64ObservableGauge("chat.spool.pending",
		metric.WithUnit("{message}"), metric.WithDescription("mensagens aguardando gravação no Postgres"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			if h.spool != nil {
				o.Observe(h.spool.pending())
			} else {
				o.Observe(0)
			}
			return nil
		}))
	// Séries nascem em 0: sem isso o painel mostra "sem dados" até o primeiro evento (e não dá
	// para distinguir "nada aconteceu" de "a métrica não existe").
	ctx := context.Background()
	out.conns.Add(ctx, 0)
	for _, r := range []string{"ok", "spooled", "rejected", "failed", "dead_letter"} {
		out.msgs.Add(ctx, 0, metric.WithAttributes(attribute.String("result", r)))
	}
	for _, d := range []string{"in", "out"} {
		for _, o := range []string{"ok", "error", "dropped"} {
			out.bus.Add(ctx, 0, metric.WithAttributes(attribute.String("direction", d), attribute.String("outcome", o)))
		}
	}
	return out
}

func (m *metrics) message(result string) {
	m.msgs.Add(context.Background(), 1, metric.WithAttributes(attribute.String("result", result)))
}

func (m *metrics) busEvent(direction, outcome string) {
	m.bus.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("direction", direction), attribute.String("outcome", outcome)))
}
