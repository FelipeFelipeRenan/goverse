# Goverse

[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml)
[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-service-ci.yml)
[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/room-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/room-service-service-ci.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/FelipeFelipeRenan/goverse)](https://pkg.go.dev/github.com/FelipeFelipeRenan/goverse)

> Goverse é uma plataforma moderna de comunicação em tempo real inspirada em soluções como Discord e Google Meet, desenvolvida com arquitetura de microsserviços em Go.

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
├── chat-service/     # Em breve
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
Ambiente de Desenvolvimento Rápido: Expõe as portas de todos os serviços diretamente, útil para debug.
```

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
    "username": "usuario",
    "email": "usuario@email.com",
    "password": "senha123"
  }'
```

Retornar todos os usuários:

```bash
curl http://localhost/users
```

Fazer Login (para obter um token):

```bash
curl -X POST http://localhost/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "type": "password",
    "email": "usuario@email.com",
    "password": "senha123"
  }'
```

Rotas Protegidas
Após obter um token JWT com a rota de login, você pode usá-lo para acessar as rotas protegidas no cabeçalho Authorization.

Retornar um usuário pelo seu ID:

```bash
curl http://localhost/user/<id_do_usuario> \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>"
```

Criar uma sala:

```bash
curl -X POST http://localhost/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>" \
  -d '{
    "name": "Minha Sala de Estudos",
    "description": "Sala para focar em Go e arquitetura.",
    "is_public": true,
    "max_members": 20
  }'
```

Listar todas as salas (com filtros):

```bash
curl "http://localhost/rooms?limit=10&offset=0&public_only=true&keyword=Estudos" \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>"
```

Atualizar informações de uma sala (requer ser dono ou admin):

```bash
curl -X PATCH http://localhost/rooms/<id_da_sala> \

  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>" \
  -d '{
    "name": "Novo Nome da Sala",
    "description": "Nova descrição"
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
curl -X POST http://localhost/rooms/<id_da_sala>/members \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <SEU_TOKEN_JWT>" \
  -d '{
    "user_id": "<id_do_usuario_a_ser_adicionado>",
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
