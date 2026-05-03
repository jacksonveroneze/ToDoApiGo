# Etapa de build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copia arquivos de dependência primeiro para aproveitar cache
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do projeto
COPY . .

# Compila o binário
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Etapa final
FROM alpine:3.20

WORKDIR /app

# Certificados para chamadas HTTPS, caso a API use integrações externas
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/api .

EXPOSE 8080

CMD ["./api"]