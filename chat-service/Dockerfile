# --- Estágio de Build ---
# Usamos uma imagem oficial do Go com Alpine Linux para compilar nossa aplicação.
FROM golang:1.24-alpine AS builder

# Define o diretório de trabalho dentro do contêiner.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências.
COPY chat-service/go.mod chat-service/go.sum ./

# Baixa as dependências. Este passo é cacheado pelo Docker se os arquivos não mudarem.
RUN go mod download

# Copia todo o código fonte do chat-service para o contêiner.
COPY chat-service/ .

# Compila a aplicação.
# CGO_ENABLED=0 cria um binário estático, sem depender de bibliotecas C.
# -o chat-service define o nome do arquivo de saída.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chat-service ./cmd/main.go

# --- Estágio Final ---
# Começamos com uma imagem Alpine vazia, que é extremamente leve.
FROM alpine:latest

# Define o diretório de trabalho para a aplicação final.
WORKDIR /root/

# Copia apenas o binário compilado do estágio de build.
COPY --from=builder /app/chat-service .

# Expõe a porta que o nosso servidor de chat usa.
EXPOSE 8084

# O comando que será executado quando o contêiner iniciar.
CMD ["./chat-service"]