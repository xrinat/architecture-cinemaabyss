# --- Build Stage ---
FROM golang:1.23 AS builder

WORKDIR /app

# Копируем файлы модулей для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Компилируем статический бинарник для Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o proxy_service main.go

# --- Run Stage ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из build stage
COPY --from=builder /app/proxy_service .

# Порт, который слушает Go-сервис
EXPOSE 8000

# Запуск приложения
CMD ["./proxy_service"]
