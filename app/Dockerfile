# ---------- Estágio 1: build ----------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Copia apenas os manifests para aproveitar o cache de camadas do Docker
COPY go.mod go.sum ./

# Baixa as dependências e fixa essa camada em cache
RUN go mod download

# Copia o restante do código-fonte
COPY . .

# Compila um binário estático (sem libc), enxuto (-s -w remove símbolos de debug)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /http-server-projeto-korp .

# ---------- Estágio 2: execução ----------
FROM alpine:3.20

# Instala certificados SSL/TLS para conexões externas e cria usuário não-root
RUN apk add --no-cache ca-certificates && \
    addgroup -S korp && adduser -S korp -G korp

COPY --from=build /http-server-projeto-korp /usr/local/bin/http-server-projeto-korp

USER korp
EXPOSE 8080

ENTRYPOINT ["http-server-projeto-korp"]
