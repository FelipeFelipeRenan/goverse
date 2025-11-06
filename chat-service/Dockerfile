# --- Estágio de Build ---
# Usamos uma imagem oficial do Go com Alpine Linux para compilar nossa aplicação.
FROM golang:1.24-alpine AS builder

# Define o diretório de trabalho dentro do contêiner.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências.
COPY chat-service/go.mod chat-service/go.sum ./

# Copia todo o código fonte do chat-service para o contêiner.
COPY chat-service/ .
# Baixa a CLI do migrate
RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Baixa as dependências. Este passo é cacheado pelo Docker se os arquivos não mudarem.
RUN go mod tidy


# Compila a aplicação.
# CGO_ENABLED=0 cria um binário estático, sem depender de bibliotecas C.
# -o chat-service define o nome do arquivo de saída.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/main.go

# --- Estágio Final ---
# Começamos com uma imagem Alpine vazia, que é extremamente leve.
FROM alpine:latest

# Define o diretório de trabalho para a aplicação final.
WORKDIR /app

# Copia apenas o binário compilado do estágio de build.
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/app /app/app
COPY --from=builder /app/migrations /app/migrations


RUN adduser -D appuser
USER appuser
# Expõe a porta que o nosso servidor de chat usa.
EXPOSE 8084

# O comando que será executado quando o contêiner iniciar.
# ENTRYPOINT ["/app/app"]