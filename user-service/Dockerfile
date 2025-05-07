# Etapa de build
FROM golang:1.24-alpine AS builder

# Instala dependências necessárias
RUN apk add --no-cache git gcc musl-dev

# Define diretório de trabalho
WORKDIR /app

# Copia os arquivos de dependências primeiro para melhor aproveitamento do cache
COPY go.mod go.sum ./

# Baixa as dependências sem copiar todo o código-fonte ainda
RUN go mod download

# Copia o restante do código
COPY . .

# Compila o binário da aplicação
RUN go build -o app ./cmd/main.go

# Etapa final (imagem mais leve)
FROM alpine:latest

LABEL Name=goverse-user-service Version=0.0.1

# Apenas adiciona certificados SSL, evitando pacotes desnecessários
RUN apk --no-cache add ca-certificates

# Copia apenas o binário, reduzindo o tamanho da imagem final
COPY --from=builder /app/app /app/app
WORKDIR /app

# Define usuário não-root para melhor segurança
RUN adduser -D appuser
USER appuser

# Define o ponto de entrada
ENTRYPOINT ["/app/app"]

# Expõe a porta padrão da aplicação
EXPOSE 8080