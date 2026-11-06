.PHONY: up down reset logs smoke test

COMPOSE=docker compose $(if $(wildcard .env),--env-file .env) -f deploy/compose.yml

up: ## sobe tudo (app + Postgres + Grafana) e espera ficar pronto
	$(COMPOSE) up -d --build
	@./deploy/smoke.sh wait
	@echo ""
	@echo "API:     http://localhost        (teste: make smoke)"
	@echo "Grafana: http://localhost:3000   (admin/admin -> Explore: traces, logs, métricas)"

down: ## para tudo (mantém os dados)
	$(COMPOSE) down

reset: ## para tudo e APAGA banco e chave
	$(COMPOSE) down -v

logs: ## logs do app
	$(COMPOSE) logs -f goverse

smoke: ## roda o fluxo cadastro -> login -> /auth/me -> editar perfil
	@./deploy/smoke.sh

test: ## testes Go com detector de race
	go test -race ./...
