# Etapa de build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth-middleware

# Etapa final
FROM alpine:latest

LABEL Name=goverse-auth-middleware Version=0.0.1

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/auth-middleware .

# Segurança: usuário não-root
RUN adduser -D appuser
USER appuser

EXPOSE 3000
ENTRYPOINT ["./auth-middleware"]
