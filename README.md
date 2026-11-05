# Goverse

[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml)
[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-service-ci.yml)
[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/room-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/room-service-service-ci.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/FelipeFelipeRenan/goverse)](https://pkg.go.dev/github.com/FelipeFelipeRenan/goverse)

> Goverse é uma plataforma moderna de comunicação em tempo real inspirada em soluções como Discord e Google Meet, desenvolvida com arquitetura de microsserviços em Go.

## ⚡ Quickstart (monolito modular)

Só precisa de Docker. Sem `.env`, sem chaves, sem migrations manuais:

```bash
make up      # sobe app + Postgres + Grafana e espera ficar pronto (a chave JWT é gerada sozinha)
make smoke   # testa: cadastro -> login -> /auth/me -> CSRF -> editar perfil
make logs    # logs do app
make down    # para (mantém dados) | make reset: apaga banco e chave
```

- API: <http://localhost:8080> (`/healthz`, `/readyz`, `/user`, `/auth/login`, ...)
- Grafana: <http://localhost:3000> (`admin`/`admin`) -> *Explore*: traces (Tempo), logs (Loki) e métricas, todos correlacionados por `trace_id`.
- Decisões e roteiro da migração: [`docs/adr/0001-monolito-modular.md`](docs/adr/0001-monolito-modular.md).

> Os diretórios `*-service/`, `api-gateway/`, `traefik/` e `k8s/` são o desenho antigo em microsserviços
> e serão removidos conforme cada módulo for portado. As instruções abaixo se referem a ele.

## 🧩 Funcionalidades

- Autenticação (senha, OAuth com JWTs RS256)
- Criação e gerenciamento de salas
- Chat em tempo real via WebSocket (em desenvolvimento)
- Vídeo via WebRTC (em desenvolvimento)
- Notificações em tempo real
- Arquitetura desacoplada com gRPC
- Observabilidade com Prometheus e Grafana

## 🛠️ Tecnologias

- Go (Golang)
- gRPC
- WebSocket
- WebRTC
- Docker & Docker Compose
- Kubernetes
- PostgreSQL & Redis
- Traefik como API Gateway

## 🚀 Estrutura de Microsserviços

```bash
goverse/
├── auth-service/
├── user-service/
├── room-service/
├── auth-middleware/
├── chat-service/ 
├── traefik/
├── monitoring/
└── k8s/
```

Nota: O api-gateway manual foi substituído pelo Traefik + auth-middleware, uma abordagem mais robusta e de fácil manutenção para roteamento e autenticação na borda.

## 📦 Como rodar localmente

Clone o repositório:

```bash
git clone https://github.com/FelipeFelipeRenan/goverse.git
cd goverse
```

Preencha os arquivos .env em cada pasta de serviço, usando os arquivos .env.example como base.

Execute os serviços com Docker Compose:
O projeto contém dois ambientes:

Ambiente de Produção-Like (Recomendado): Usa o Traefik como gateway, forçando toda a comunicação pela borda.


```bash
docker-compose -f docker-compose-traefik.yml up --build
```
Ambiente de Desenvolvimento Rápido: Expõe as portas de todos os serviços diretamente, útil para debug.

Ambiente de Desenvolvimento Rápido: Expõe as portas de todos os serviços diretamente, útil para debug.

```bash
docker-compose up --build
```

Crie as Chaves de Autenticação (Obrigatório):
Crie uma pasta .keys na raiz do projeto e adicione-a ao seu .gitignore. Em seguida, gere as chaves:

```bash
# Gerar a chave privada
openssl genpkey -algorithm RSA -out .keys/private.pem -pkeyopt rsa_keygen_bits:2048
```

```bash
# Extrair a chave pública
openssl rsa -pubout -in .keys/private.pem -out .keys/public.pem
```

## 🧪 Acesso à documentação do Swagger

A documentação das APIs está disponível via Swagger. Após iniciar os serviços, acesse:

<http://localhost:8088/swagger/index.html>

## 🧪 Testes

Cada serviço contém seus próprios testes de unidade. Para rodar os testes de um serviço específico:

```bash
cd room-service
go test ./... -v
```

Testes com curl (Acessando o Gateway na porta 80)
Rotas Públicas
Criar um usuário:

```bash
curl -X POST http://localhost/user \
  -H "Content-Type: application/json" \
  -d '{
    "username": "novo_usuario",
    "email": "novo@email.com",
    "password": "senha123"
  }'
```

Retornar todos os usuários:

```bash
curl http://localhost/users
```

Fazer Login (para obter um token):

```bash
export CSRF_TOKEN=$(curl -s -X POST http://localhost/auth/login \
  -H "Content-Type: application/json" \
  -d '{"type": "password", "email": "admin@goverse.com", "password": "senha123"}' \
  --cookie-jar cookies.txt | jq -r '.csrf_token')
```

Rotas Protegidas
Após obter um token JWT com a rota de login, você pode usá-lo para acessar as rotas protegidas no cabeçalho Authorization.

Retornar os dados de um usuário logado

```
curl http://localhost/auth/me --cookie cookies.txt
```

Retornar um usuário pelo seu ID:

```bash
curl http://localhost/user/<id_do_usuario> \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>"
```

Criar uma sala:

```bash
curl -X POST http://localhost/rooms \
  --cookie cookies.txt \
  -H "X-CSRF-TOKEN: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sala do Admin",
    "description": "Sala criada via cURL",
    "is_public": true,
    "max_members": 20
  }'
```

Listar salas do usuário logado

```bash
curl http://localhost/rooms/mine --cookie cookies.txt
```

Listar todas as salas (com filtros):

```bash
curl "http://localhost/rooms?limit=10&offset=0&public_only=true&keyword=Estudos" \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>"
```

Atualizar informações de uma sala (requer ser dono ou admin):

```bash
curl -X PATCH http://localhost/rooms/<id-da-sala> \
  --cookie cookies.txt \
  -H "X-CSRF-TOKEN: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Novo Nome da Sala Geral"
  }'
```

Deletar uma sala (requer ser o dono):

```bash
curl -X DELETE http://localhost/rooms/<id_da_sala> \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>"
```

Mostrar todos os membros de uma sala:

```bash
curl http://localhost/rooms/<id_da_sala>/members \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>"
```

Adicionar um membro a uma sala (requer ser dono ou admin):

```bash
curl -X POST http://localhost/rooms/<id-da-sala>/members \
  --cookie cookies.txt \
  -H "X-CSRF-TOKEN: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "<id-do-usuario-adicionado>",
    "role": "member"
  }'
```

Deletar um membro da sala (requer ser dono ou admin):

```bash
curl -X DELETE http://localhost/rooms/<id_da_sala>/members/<id_do_membro> \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>"
```

Mudar a role de um membro (requer ser dono ou admin):

```bash
curl -X PUT http://localhost/rooms/<id_da_sala>/members/<id_do_membro>/role \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>" \
  -d '{ "new_role": "admin" }'
```

### Chat Service

Para se conectar a sala, basta rodar o comando:
```bash
wscat -c "ws://localhost/ws?roomId=sala-final" \
-H "Cookie: access_token=$(grep 'access_token' cookies.txt | cut -f7)"
```

## ☸️ Utilizando Kubernetes (k8s)

Para utilizar a aplicação com o Kubernetes no Minikube:

Inicie o cluster:

```bash
minikube start
```

Crie os segredos necessários (execute a partir da raiz do projeto):

```bash
# Lembre-se de criar todos os secrets para senhas de banco, etc.
# Exemplo para as chaves JWT:
kubectl create secret generic jwt-keys-secret -n goverse \
  --from-file=private.pem=./.keys/private.pem \
  --from-file=public.pem=./.keys/public.pem
```

Aplique os manifestos:

```bash
make k8s-apply
```

Exponha as portas do Traefik:


```bash
make traefik-ports
```
### Utilizando Kubernetes (k8s)

Para utilizar a aplicação com o Kubernetes, utilizando no minikube como cluster, basta utilizar os comandos:

Para iniciar o cluster:

```bash
minikube start
```

Para aplicar os manifests:
 ```bash
make k8s-apply
```

Para expor as portas:

```bash
make traefik-ports
```

### Em breve serão implementadas as features relacionadas a operações nas salas e bate papo por texto

📄 Licença
Distribuído sob a licença MIT.
