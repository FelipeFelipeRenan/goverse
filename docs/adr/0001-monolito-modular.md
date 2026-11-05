# ADR 0001 — Migrar de microsserviços para monolito modular

**Status:** em andamento · **Data:** 2026-11-04

## Contexto
Sete serviços Go (~5 mil linhas) em 8 módulos `go.mod`, 14 containers no compose, um único
Postgres compartilhado e uma equipe de uma pessoa. O custo de instrumentar (traces entre
Traefik → forwardAuth → gRPC → Postgres → Kafka) e de fazer deploy superou qualquer benefício.

## Decisão
Um binário (`cmd/goverse`), um `go.mod`, módulos de domínio em `internal/` e infraestrutura
transversal em `internal/platform/`. Migração incremental (strangler): os serviços antigos
continuam funcionando até cada módulo ser portado.

| Módulo | Substitui |
|---|---|
| `identity` | user-service, auth-service, auth-middleware |
| `rooms` | room-service |
| `chat` | chat-service (hub, mensagens, presença) |

## Regras de modularidade
1. Um módulo expõe uma interface pequena na raiz do seu pacote; o resto fica em `internal/`.
2. Nenhum módulo executa SQL em tabelas de outro; um schema Postgres por módulo.
3. Contratos HTTP em OpenAPI; sem dependência de gRPC entre módulos.
4. Observabilidade via OpenTelemetry; correlação por `trace_id` nos logs (`slog`).
5. `GOVERSE_ROLE` = `all` | `api` | `realtime` permite separar o tráfego WebSocket depois,
   sem fork de código.

## Roteiro
0. Estancar: CI e vazamento de salas privadas (PR anterior).
1. **Esqueleto + observabilidade** (este PR).
2. `identity` → 3. `rooms` → 4. `chat` (sem Kafka; persistir antes do fan-out) → 5. remover legado.

## Consequências
Domínio de falha compartilhado e deploy do realtime derrubando conexões abertas; mitigados por
`GOVERSE_ROLE` e pelos limites de módulo. Mantemos paths HTTP e cookies estáveis para o
`goverse-frontend` não perceber a migração.
