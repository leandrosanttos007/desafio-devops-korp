# ---------- Estágio 1: build ----------
FROM golang:1.22-alpine AS build

WORKDIR /src

# Copia os manifests primeiro para aproveitar o cache de camadas do Docker.
COPY go.mod ./
COPY . .

# Resolve as dependências e gera o go.sum de forma reprodutível.
RUN go mod tidy

# Compila um binário estático (sem dependência de libc), enxuto (-s -w remove símbolos de debug).
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /http-server-projeto-korp .

# ---------- Estágio 2: execução ----------
FROM alpine:3.20

# Cria um usuário sem privilégios para não rodar como root.
RUN addgroup -S korp && adduser -S korp -G korp

COPY --from=build /http-server-projeto-korp /usr/local/bin/http-server-projeto-korp

USER korp
EXPOSE 8080

ENTRYPOINT ["http-server-projeto-korp"]
