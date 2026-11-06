# Migração do chat-service → módulo `chat`

Rotas preservadas (mesmo formato JSON, inclusive `CreatedAt`):

| Antes | Agora |
|---|---|
| `GET /ws?roomId=` (cookie `access_token`) | igual; membro verificado por `rooms.Membership` (sem gRPC) |
| `GET /rooms/{id}/messages?limit&offset` | igual; não-membro recebe **403** (antes 401, que o front confundia com sessão expirada) |

## O que mudou (de propósito)

- **Sem Kafka, Zookeeper e Redis.** A mensagem é gravada no Postgres *antes* de ser distribuída, na ordem em que a sala a recebeu. Antes: Kafka → consumer → banco, e a entrega não esperava a gravação.
- **IDs e datas gerados no servidor** (UUIDv7); `user_id`, `username` e `room_id` vêm da conexão autenticada, nunca do payload.
- **Várias abas por usuário.** Antes só a última conexão contava. "entrou/saiu da sala" só dispara na 1ª/última aba.
- **Fila limitada por conexão (64).** Consumidor lento é desconectado; não trava a sala nem consome memória sem limite.
- **Origem do WebSocket** validada contra `GOVERSE_CORS_ORIGINS` (antes aceitava qualquer origem).
- `SIGNAL`/`DIRECT` só chegam a quem está conectado na mesma sala. Digitando (`TYPING_*`) não volta ao remetente. Mensagem vazia ou > 4000 bytes devolve frame `{"type":"ERROR"}`.
- Status online no Redis (`user:status:*`) foi removido: nada no front o consome.

## Limite conhecido

O hub vive na memória de **um** processo. Para várias réplicas é preciso trocar a distribuição por um barramento (Postgres `LISTEN/NOTIFY` resolve até boa escala; NATS depois). A persistência já está correta para isso.

## Dados antigos

```sql
INSERT INTO chat.messages (id, room_id, user_id, username, content, created_at)
SELECT id, room_id, user_id, username, content, created_at
FROM   <banco_antigo>.messages
WHERE  type = 'CHAT';
```
(UUIDv4 antigos ordenam mal por `id`; se o histórico antigo importar, ordene por `created_at` na cópia.)
