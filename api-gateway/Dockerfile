# Etapa de build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copia os arquivos de módulo
COPY api-gateway/go.mod api-gateway/go.sum ./

# Copia todo o código-fonte (que agora é só o cmd/, docs/, pkg/)
COPY api-gateway/. .

# Roda o tidy para baixar dependências (só o swagger e o logger)
RUN go mod tidy

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

EXPOSE 8090