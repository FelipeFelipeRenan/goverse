# Migração do módulo `rooms`

O módulo cria o schema `rooms` (tabelas `rooms.rooms` e `rooms.members`) no banco do monolito, com
**chaves estrangeiras para `identity.users`** (antes `owner_id`/`user_id` eram UUIDs soltos). Por isso
`identity` precisa migrar e receber os dados **antes** de `rooms`.

## Copiar dados (banco antigo `rooms` → `goverse`)

Com o banco antigo acessível (por exemplo via `postgres_fdw` como schema `old_rooms`), depois de copiar os usuários:

```sql
INSERT INTO rooms.rooms (id, name, description, is_public, max_members, owner_id, member_count, created_at, updated_at, deleted_at)
SELECT id, name, coalesce(description, ''), is_public, max_members, owner_id, 0, created_at, created_at, deleted_at
FROM old_rooms.rooms;

INSERT INTO rooms.members (room_id, user_id, role, joined_at)
SELECT room_id, user_id, role, joined_at FROM old_rooms.room_members;

-- member_count é recalculado a partir dos membros reais (o contador antigo podia estar fora de sincronia)
UPDATE rooms.rooms r SET member_count = (SELECT count(*) FROM rooms.members m WHERE m.room_id = r.id);
```

Antes de rodar, confira se os dados antigos respeitam as novas regras (o INSERT falha se não respeitarem):
- todo `owner_id`/`user_id` existe em `identity.users`;
- cada sala tem **exatamente um** membro com role `owner`;
- nenhuma sala tem mais membros do que `max_members` (o serviço antigo não checava isso).

```sql
SELECT id FROM old_rooms.rooms r WHERE (SELECT count(*) FROM old_rooms.room_members m WHERE m.room_id = r.id AND m.role = 'owner') <> 1;
SELECT id FROM old_rooms.rooms r WHERE (SELECT count(*) FROM old_rooms.room_members m WHERE m.room_id = r.id) > r.max_members;
```

## Mudanças de comportamento (intencionais)

| Antes | Agora |
|---|---|
| `GET /rooms?public_only=false` e a busca por `keyword` listavam salas **privadas**, sem login | só salas públicas; o parâmetro `public_only` foi removido |
| `GET /rooms/{id}`, membros e demais rotas confiavam no header `X-User-ID` | exigem sessão (cookie + CSRF); só `GET /rooms` continua aberta |
| Qualquer um via sala privada por id | quem não é membro recebe **404** (não revela que existe) |
| `POST .../members` não respeitava `max_members` | respeita; sala lotada = **409** (a contagem é atômica no banco) |
| `role` livre em `POST .../members` (dava para criar vários "owner") | só `admin`, `moderator`, `member`, `guest`; 1 dono por sala (índice único) |
| `PATCH /rooms/{id}` ignorava `is_public` e `max_members`, mas a resposta fingia ter aplicado | aplica os quatro campos; `max_members` menor que o nº de membros = 400 |
| Erros de permissão e validação viravam **500** | 400 / 403 / 404 / 409 conforme o caso |
| Listas vazias podiam sair como `null` | sempre `[]` |
| `GET /rooms/{id}/members` fazia 1 chamada gRPC por membro (N+1) | 1 consulta (`identity.Directory.GetMany`) |
| Usuário removido ainda criava sala com o token válido por 24h | 401 |

Formatos de resposta preservados: `id`/`name` em `POST /rooms` e `GET /rooms/{id}`; `room_id`/`room_name` nas listagens e no `PATCH`.

## Pendente (de propósito)
- `GET /rooms/{id}/members` ainda inclui o e-mail dos membros para qualquer usuário logado que enxergue a sala (mantido por compatibilidade com o frontend). Vale restringir depois.
- A virada das rotas `/rooms*` e `/user/rooms` no Traefik para o monolito.
