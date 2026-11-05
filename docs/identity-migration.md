# Migração do módulo `identity`

O módulo cria `identity.users` no banco do monolito (migration embutida, aplicada no boot).
Os dados do `user-service` antigo precisam ser copiados uma vez.

## 1. Copiar usuários (banco antigo `users` → `goverse`)

Com o banco antigo acessível (por exemplo via `postgres_fdw` como schema `old_users`):

```sql
INSERT INTO identity.users (id, username, email, password_hash, picture, is_oauth, created_at, updated_at, deleted_at)
SELECT id, username, lower(email), password, coalesce(picture, ''), coalesce(is_oauth, false),
       created_at, coalesce(updated_at, created_at), deleted_at
FROM old_users.users;
```

Atenção:
- Os hashes bcrypt continuam válidos; ninguém precisa redefinir senha.
- O índice único é por `lower(email)`: se houver e-mails que só diferem em maiúsculas, o INSERT falha e é preciso resolver à mão.
- As colunas antigas eram `timestamp` sem fuso; elas serão interpretadas no fuso da sessão (use `SET TIME ZONE 'UTC'` se foram gravadas em UTC).
- A migration `000002_seed_users` do serviço antigo criava `admin@goverse.com` e `usuario@goverse.com` com a senha `senha123`. **Remova ou troque a senha dessas contas em qualquer ambiente exposto.**

## 2. Virada (quando o monolito atender produção)
Aponte as rotas `/auth/*`, `/oauth/*`, `/user*` e `/users` do Traefik para o monolito e remova o
`forwardAuth`. Rotas, cookies (`access_token`, `csrf_token`) e o formato do JWT (`user_id`,
`username`, RS256) foram mantidos, então tokens emitidos pelo serviço antigo continuam válidos
enquanto a chave privada for a mesma (`GOVERSE_JWT_PRIVATE_KEY_PATH`).

## Mudanças de comportamento (intencionais)
| Antes | Agora |
|---|---|
| `PUT /user/me` falhava (SQL sem `UPDATE`) | funciona |
| `GET /users` público e `GET /user/{id}` devolviam o hash da senha | exigem sessão e nunca devolvem senha |
| Login diferenciava "usuário não encontrado" de "senha inválida" | 401 `credenciais inválidas` em ambos |
| Login Google não emitia `csrf_token` (POST/PUT/DELETE falhavam) | emite |
| Google sem checar `verified_email` | exige e-mail verificado |
| Duplicidade de e-mail virava 500 | 409 |
| Senha sem regra | 8 a 72 caracteres |
| CSRF inválido respondia 401 | 403 |
| `POST /auth/login` com `type: "oauth"` | removido; use `/oauth/google/login` |
