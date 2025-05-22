# Goverse

[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml)
[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-service-ci.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/FelipeFelipeRenan/goverse)](https://pkg.go.dev/github.com/FelipeFelipeRenan/goverse)

> Goverse é uma plataforma moderna de comunicação em tempo real inspirada em soluções como Discord e Google Meet, desenvolvida com arquitetura de microsserviços em Go.

## 🧩 Funcionalidades

- Autenticação (senha, OAuth)
- Criação e gerenciamento de salas
- Chat em tempo real via WebSocket
- Vídeo via WebRTC (em desenvolvimento)
- Notificações em tempo real
- Arquitetura desacoplada com gRPC

## 🛠️ Tecnologias

- Go (Golang)
- gRPC
- WebSocket
- WebRTC
- Docker
- Kubernetes (opcional)
- PostgreSQL / Redis

## 🚀 Estrutura de Microsserviços

```
goverse/
├── auth-service/
├── user-service/
├── room-service/
├── chat-service/
├── notification-service/
└── api-gateway/
```

## 📦 Como rodar localmente

```bash
Clone o repositório

cd goverse

# Execute os serviços (exemplo com docker-compose)
docker-compose up --build
```

## 🧪 Testes

Cada serviço contém seus próprios testes. Para rodar os testes:

```bash
cd auth-service
go test ./...
```

### 🧪 Testes com curl, acessando o API Gateway

Para criar um usuário, utilize o comando: 

```bash
curl -X POST http://localhost:8088/user \
  -H "Content-Type: application/json" \
  -d '{
  "username": "usuario",
  "email": "usuario@email.com",
  "password": "senha123"
}'

```

Para retornar todos os usuários, utilize o comando:
```bash
curl http://localhost:8088/users
```

Para retornar um usuário pelo seu ID, utilize o comando: 
```bash
curl http://localhost:8088/users/{id do usuario}
```

Para realizar testes de login com senha, utilize o comando: 

```bash
 curl -X POST http://localhost:8088/login \
  -H "Content-Type: application/json" \
  -d '{
  "email": "usuario@usuario.com",
  "password": "senha123", "type":"password"
}'
```

Para testar acessando o serviço diretamente, basta mudar a porta na requisição do curl para a que os serviços foram definidos

## 📄 Licença

Distribuído sob a licença MIT. Veja `LICENSE` para mais informações.
