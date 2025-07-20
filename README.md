# Goverse

[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/auth-service-ci.yml)
[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/user-service-service-ci.yml)
[![Build](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/room-service-ci.yml/badge.svg)](https://github.com/FelipeFelipeRenan/goverse/actions/workflows/room-service-service-ci.yml)
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
- Kubernetes
- PostgreSQL / Redis
- Traefik for API Gateway

## 🚀 Estrutura de Microsserviços

```bash
goverse/
├── auth-service/
├── user-service/
├── room-service/
├── chat-service/
├── notification-service/
├── traefik/
├── monitoring/
├── k8s/
└── api-gateway/
```

Retirei o uso do api gateway por enquanto, substituindo pelo Traefik,
pois estava se tornando dificil de realizar manutenção nas rotas dos microsseviços

## 📦 Como rodar localmente

```bash
Clone o repositório

cd goverse

preencha os arquivos .env baseados nos .env.example

# Execute os serviços (exemplo com docker-compose)
docker-compose up --build
```

## 🧪 Acesso à documentação do Swagger

Os endpoints para os serviços estão disponíveis na interface do Swagger, ao acessar o link abaixo:

<http://localhost:8088/swagger/index.html>

## 🧪 Testes

Cada serviço contém seus próprios testes. Para rodar os testes:

```bash
cd auth-service
go test ./...
```

### 🧪 Testes com curl, acessando o API Gateway

Para criar um usuário, utilize o comando:

```bash
curl -X POST http://localhost/user \
  -H "Content-Type: application/json" \
  -d '{
  "username": "usuario",
  "email": "usuario@email.com",
  "password": "senha123"
}'

```

Para retornar todos os usuários, utilize o comando:

```bash
curl http://localhost/users
```

Para retornar um usuário pelo seu ID, utilize o comando:

```bash
curl http://localhost/users/<id do usuario>
```

Para realizar testes de login com senha, utilize o comando:

```bash
 curl -X POST http://localhost/auth/login \
  -H "Content-Type: application/json" \
  -d '{
  "email": "usuario@usuario.com",
  "password": "senha123", "type":"password"
}'
```

Para realizar o acesso à rotas protegidas, utilize o comando:

```bash
  curl -X GET http://localhost/user/<id do usuario> \
  -H "Authorization: Bearer <TOKEN>"    
```

Para testar acessando o serviço diretamente, basta mudar a porta na requisição do curl para a que os serviços foram definidos

Para criação de salas, utilize o comando:

```bash
curl -X POST http://localhost/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \

  -d '{
    "owner_id": "<id do dono>",
    "name": "<nome da sala>",
    "description": "<descrição da sala>",
    "is_public": <boleano indicando se a sala é publica ou não>
  }'
```

Para verificar retornar uma sala por ID, utilize o comando:

```bash
curl -X GET http://localhost/rooms/<id da sala>
```

Para listar todas as salas, junto com ulitização de filtros, utilize o comando:

```bash
 curl "http://localhost/rooms?limit=<numero de salas>&offset=<numero de salas puladas >&public_only=<true ou false>&keyword=<palavra chave da sala>"
```

Caso deseje, basta omitir o filtro

Para atualizar informações de uma sala, utilize o comando:

```bash
curl -X PATCH http://localhost/rooms/3 
  -H "Content-Type: application/json" 
  -H "X-User-ID: 1"
    -d '{
      "name": "<novo nome>",
      "description": "<nova descrição>"
       }'
```

Podem ser adicionados outros campos para ser modificado, como o is_public, ou omitido os que desejar não atualizar

Para deletar uma sala, utilize o comando

```bash
  curl -X DELETE http://localhost/rooms/<id da sala> \
  -H "X-User-ID: <id do dono da sala>"

```

Para mostrar todos os membros de uma sala, utilize o comando:

```bash
curl -X GET http://localhost/rooms/<id da sala>/members \
```

Para adicionar um membro a sala, utilize o comando:

```bash
  curl -X POST http://localhost/rooms/<id da sala>/members \
    -H "X-User-ID: <id do dono da sala>" \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "<id do usuario>",
        "role": "<role>"
      }'
```

Para deletar um membro da sala, utilize o comando:

```bash
curl -X DELETE http://localhost/rooms/<id da sala>/members/<id do membro> \
  -H "X-User-ID: <id do dono da sala>"
```

Para mudar a role de um membro, utilize o comando:

```bash
 curl -X PUT http://localhost/rooms/<id da sala>/members/<id do membro>/role \
  -H "X-User-ID: <id do dono da sala>" \
  -H "Content-Type: application/json" \
  -d '{ "new_role": "<role>" }'

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

## 📄 Licença

Distribuído sob a licença MIT. Veja `LICENSE` para mais informações.
