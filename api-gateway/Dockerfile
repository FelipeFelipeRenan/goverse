# Etapa de build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Se precisar do proto futuramente, pode descomentar essa linha:
# COPY proto /proto

# Copia os arquivos de go.mod e go.sum do api-gateway
COPY api-gateway/go.mod api-gateway/go.sum ./

RUN go mod download

# Copia o restante do código do api-gateway
COPY api-gateway/. .

# Compila o binário
RUN go build -o app ./cmd/main.go

# Etapa final
FROM alpine:latest

LABEL Name=goverse-api-gateway Version=0.0.1

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/app /app/app
WORKDIR /app

RUN adduser -D appuser
USER appuser

ENTRYPOINT ["/app/app"]
EXPOSE 8080
