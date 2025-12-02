# Etapa de build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copia proto antes de rodar o go mod download (necessário por causa do replace ../proto)
COPY proto /proto

COPY common /common

# Copia os arquivos de go.mod e go.sum do user-service
COPY auth-service/go.mod auth-service/go.sum ./
# Copia o restante do código do user-service
COPY auth-service/. .

# Rodar o download agora funciona porque o /proto existe
RUN go mod tidy

# Compila o binário
RUN go build -o app ./cmd/main.go

# Etapa final
FROM alpine:latest

LABEL Name=goverse-auth-service Version=0.0.1

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/app /app/app
WORKDIR /app

RUN adduser -D appuser
USER appuser

ENTRYPOINT ["/app/app"]
EXPOSE 8080
