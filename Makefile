NAMESPACE=goverse

# =========== DOCKER ===========

docker-bp:
	@echo "🚀 Realizando build e push para o Docker Registry"
	@echo "🛠️ Realizando build e push do User Service"
		docker build -t felipedev21/user-service:latest -f user-service/Dockerfile .
		docker push felipedev21/user-service:latest
	@echo "🛠️ Realizando build e push do Auth Service"
		docker build -t felipedev21/auth-service:latest -f auth-service/Dockerfile .
		docker push felipedev21/auth-service:latest
	@echo "🛠️ Realizando build e push do Room Service"
		docker build -t felipedev21/room-service:latest -f room-service/Dockerfile .
		docker push felipedev21/room-service:latest
	@echo "🛠️ Realizando build e push do Auth Middleware"
		docker build -t felipedev21/auth-middleware:latest ./auth-middleware
		docker push felipedev21/auth-middleware:latest

# ========== TRAEFIK ==========

# Lista todos os deployments no namespace e faz rollout restart em cada um

restart-all:
	@echo "Reiniciando todos os deployments no namespace $(NAMESPACE)..."
	@for deploy in $$(kubectl get deployments -n $(NAMESPACE) -o jsonpath='{.items[*].metadata.name}'); do \
		echo "▶ Reiniciando $$deploy..."; \
		kubectl rollout restart deployment $$deploy -n $(NAMESPACE); \
	done

traefik-apply:
	kubectl apply -f k8s/base/configmaps/traefik-configmap.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/configmaps/traefik-dynamic-configmap.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/deployments/traefik-deployment.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/services/traefik-service.yml -n $(NAMESPACE)

traefik-restart:
	kubectl rollout restart deployment traefik -n $(NAMESPACE)

traefik-port:
	@echo "👉 Acessar dashboard: http://localhost:8081/dashboard"
	@echo "👉 Acessar métricas:  http://localhost:8082/metrics"
	@echo "👉 Acessar API:       http://localhost:8080/"
	kubectl port-forward deployment/traefik -n $(NAMESPACE) 8081:8081 8082:8082 8080:80

# ========== SERVIÇOS ==========
auth-apply:
	@echo "🛠️ Aplicando configurações do cluster no namespace $(NAMESPACE)"
	kubectl apply -n $(NAMESPACE) -f k8s/ --recursive

user-apply:
	kubectl apply -f k8s/base/deployments/user-service-deployment.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/services/user-service.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/configmaps/user-service-configmap.yml -n $(NAMESPACE)

room-apply:
	kubectl apply -f k8s/base/deployments/room-service-deployment.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/services/room-service.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/configmaps/room-service-configmap.yml -n $(NAMESPACE)

auth-middleware-apply:
	kubectl apply -f k8s/base/deployments/auth-middleware-deployment.yml -n $(NAMESPACE)
	kubectl apply -f k8s/base/services/auth-middleware-service.yml -n $(NAMESPACE)

# ========== COMPLETOS ==========

services-apply: auth-apply user-apply room-apply auth-middleware-apply

k8s-apply: traefik-apply services-apply

# ========== UTILS ==========

traefik-logs:
	kubectl logs deployment/traefik -n $(NAMESPACE) -f

describe-traefik:
	kubectl describe deployment traefik -n $(NAMESPACE)

# ========== HELP ==========

help:
	@echo "🧪 Goverse Kubernetes Makefile:"
	@echo "  make docker-bp               # Build e push dos serviços para Docker Registry"
	@echo "  make k8s-apply           	# Aplica tudo: traefik + serviços"
	@echo "  make traefik-apply       	# Aplica apenas configs/deploy de traefik"
	@echo "  make traefik-restart     	# Reinicia o traefik"
	@echo "  make traefik-port        	# Faz port-forward para dashboard e métricas"
	@echo "  make services-apply      	# Aplica todos os serviços (auth, user, room)"
	@echo "  make auth-apply          	# Aplica auth-service"
	@echo "  make user-apply          	# Aplica user-service"
	@echo "  make room-apply          	# Aplica room-service"
	@echo "  make auth-middleware-apply    # Aplica auth-middleware"
	@echo "  make traefik-logs        	# Mostra logs do traefik"
	@echo "  make restart-all        	# reinicia todos os pods"

# ========== MONOLITO (desenvolvimento) ==========

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
