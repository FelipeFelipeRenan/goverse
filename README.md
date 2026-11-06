# Goverse

[![CI](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/ci.yml)

> Plataforma de comunicação em tempo real (salas, chat, auth) em Go, como **monolito modular**: um binário, um Postgres, observabilidade com OpenTelemetry.

## ⚡ Quickstart (monolito modular)

Só precisa de Docker. Sem `.env`, sem chaves, sem migrations manuais:

```bash
make up      # sobe app + Postgres + Grafana e espera ficar pronto (a chave JWT é gerada sozinha)
make smoke   # testa: cadastro -> login -> /auth/me -> CSRF -> editar perfil
make logs    # logs do app
make down    # para (mantém dados) | make reset: apaga banco e chave
```

- API: <http://localhost> (porta 80, onde o frontend a espera; `/healthz`, `/readyz`, `/user`, `/auth/login`, ...). Se a 80 estiver ocupada (por exemplo pelo Traefik antigo), pare-o ou use `GOVERSE_PUBLIC_PORT=8080 make up`.
- Grafana: <http://localhost:3000> (`admin`/`admin`) -> *Explore*: traces (Tempo), logs (Loki) e métricas, todos correlacionados por `trace_id`.
- **Login com Google (opcional):** ponha `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` e `GOOGLE_REDIRECT_URI` (`http://localhost/oauth/google/callback`) no `.env` da raiz e rode `make up` de novo. Sem isso, `/oauth/google/login` responde 501.
- **Frontend:** `ng serve` em <http://localhost:4200> já é aceito pelo CORS (`GOVERSE_CORS_ORIGINS` muda a lista).
- Chat: WebSocket em `/ws?roomId=` e histórico em `/rooms/{id}/messages` ([`docs/chat-migration.md`](docs/chat-migration.md)).
- Decisões e roteiro da migração: [`docs/adr/0001-monolito-modular.md`](docs/adr/0001-monolito-modular.md).

## 🧩 Estrutura

```
cmd/goverse/          composição: config, telemetria, DB, módulos, servidor HTTP
internal/identity/    usuários, login, OAuth Google
internal/rooms/       salas e participantes
internal/chat/        WebSocket + histórico
internal/platform/    config, authn (JWT RS256 + CSRF), migrate, httpx, telemetria
deploy/               Dockerfile, compose de desenvolvimento, smoke test
docs/                 ADR e guias de migração por módulo
```

Regras: módulos só se falam por interfaces pequenas na raiz do pacote (`identity.Directory`, `rooms.Membership`); nada de SQL entre módulos; um schema Postgres por módulo.

## 🧪 Testes

```bash
make test    # go test -race ./...
make smoke   # fluxo HTTP completo contra a stack no ar
```

O CI (`.github/workflows/ci.yml`) roda gofmt, vet, testes com `-race`, govulncheck, build da imagem e o smoke contra Postgres real.

## 🚢 Deploy

Uma imagem (`deploy/Dockerfile`, distroless) e variáveis `GOVERSE_*` (ver `.env.example` e `internal/platform/config`). `GOVERSE_ENV=prod` ativa cookies `Secure`/`SameSite=None`. O desenho antigo (7 microsserviços, Kafka, gRPC, Traefik, k8s) está no histórico do git, antes da remoção.
