# Etapa de build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copia proto antes de rodar o go mod download (necessário por causa do replace ../proto)
COPY proto /proto

# Copia os arquivos de go.mod e go.sum do user-service
COPY room-service/go.mod room-service/go.sum ./

# Rodar o download agora funciona porque o /proto existe
RUN go mod download

# Copia o restante do código do user-service
COPY room-service/. .

# Compila o binário
RUN go build -o app ./cmd/main.go

# Etapa final
FROM alpine:latest

LABEL Name=goverse-room-service Version=0.0.1

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/app /app/app
WORKDIR /app

RUN adduser -D appuser
USER appuser

ENTRYPOINT ["/app/app"]
EXPOSE 8083

