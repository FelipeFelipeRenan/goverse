# Migração do chat-service → módulo `chat`

Rotas preservadas (mesmo formato JSON, inclusive `CreatedAt`):

| Antes | Agora |
|---|---|
| `GET /ws?roomId=` (cookie `access_token`) | igual; membro verificado por `rooms.Membership` (sem gRPC) |
| `GET /rooms/{id}/messages?limit&offset` | igual; não-membro recebe **403** (antes 401, que o front confundia com sessão expirada) |

## O que mudou (de propósito)

- **Sem Kafka, Zookeeper e Redis.** A mensagem é gravada no Postgres *antes* de ser distribuída, na ordem em que a sala a recebeu. Antes: Kafka → consumer → banco, e a entrega não esperava a gravação.
- **Resiliência a queda do banco:** a gravação tenta 4 vezes (esperas de 100/300/900 ms). Se ainda falhar, ninguém recebe a mensagem e o remetente recebe `{"type":"ERROR"}` para poder reenviar. O desenho antigo entregava ao vivo e gravava depois via Kafka; a DLQ antiga era só de escrita (nada a reprocessava).
- **IDs e datas gerados no servidor** (UUIDv7); `user_id`, `username` e `room_id` vêm da conexão autenticada, nunca do payload.
- **Várias abas por usuário.** Antes só a última conexão contava. "entrou/saiu da sala" só dispara na 1ª/última aba.
- **Fila limitada por conexão (64).** Consumidor lento é desconectado; não trava a sala nem consome memória sem limite.
- **Origem do WebSocket** validada contra `GOVERSE_CORS_ORIGINS` (antes aceitava qualquer origem).
- `SIGNAL`/`DIRECT` só chegam a quem está conectado na mesma sala. Digitando (`TYPING_*`) não volta ao remetente. Mensagem vazia ou > 4000 bytes devolve frame `{"type":"ERROR"}`.
- Status online no Redis (`user:status:*`) foi removido: nada no front o consome.

## Falha do banco: spool

`GOVERSE_CHAT_SPOOL_DIR` liga um spool durável (arquivo JSON por linha, com fsync). Se o Postgres falha mesmo após as tentativas:

1. a mensagem (que já tem ID UUIDv7 e horário) é gravada no spool e **entregue ao vivo**, como no desenho antigo com Kafka;
2. um disjuntor manda as próximas direto ao spool por 5 s, para não segurar a sala em timeouts;
3. a cada 5 s o spool é reprocessado **em ordem**; o `INSERT ... ON CONFLICT DO NOTHING` torna o replay idempotente (inclusive após crash no meio);
4. erro **permanente** (violação de integridade, ex.: sala apagada) vai para `chat.dead_letters` e é registrado em log; erro permanente numa gravação ao vivo é recusado com `{"type":"ERROR"}`, sem spool.

Limites honestos: o spool sobrevive a restart do processo, **não** à perda do disco (em k8s, `emptyDir` não sobrevive a reagendamento do pod); até o reprocessamento a mensagem não aparece no histórico nem para outras réplicas (que só a veem se o barramento estiver de pé, e ele depende do mesmo Postgres); a métrica `chat_spool_pending` deve ser 0. Sem `GOVERSE_CHAT_SPOOL_DIR`, vale o comportamento anterior: falha vira erro para o remetente.

## Várias réplicas

`GOVERSE_CHAT_BUS=postgres` (padrão) liga as réplicas por LISTEN/NOTIFY. Cada réplica entrega aos seus clientes e publica o evento; as outras entregam aos delas. Mensagens grandes viajam só como ID (o limite do NOTIFY é 8000 bytes) e o receptor lê do banco. Sinais WebRTC acima de ~7,5 KB não atravessam réplicas.

- Entrega "melhor esforço" entre réplicas: se uma réplica ficou desconectada do barramento, perde mensagens ao vivo, presença e digitação do período; o histórico está certo e o front recarrega ao reconectar.
- Presença é por réplica: o mesmo usuário em duas réplicas pode gerar "entrou/saiu" duplicados.
- Não funciona atrás de PgBouncer em modo *transaction* (LISTEN é por sessão).
- Escala: cada mensagem vira um NOTIFY lido por todas as réplicas. Serve para poucas réplicas e estudo; acima disso, NATS/Redis Streams.

## Dados antigos

```sql
INSERT INTO chat.messages (id, room_id, user_id, username, content, created_at)
SELECT id, room_id, user_id, username, content, created_at
FROM   <banco_antigo>.messages
WHERE  type = 'CHAT';
```
(UUIDv4 antigos ordenam mal por `id`; se o histórico antigo importar, ordene por `created_at` na cópia.)
