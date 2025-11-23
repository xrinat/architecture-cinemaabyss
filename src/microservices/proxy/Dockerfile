# --- Этап сборки (Build Stage) ---
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем приложение в статический бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-s -w' -o proxy_service .


# --- Этап запуска (Run Stage) ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник из этапа сборки
COPY --from=builder /app/proxy_service .

# Открываем порт, который используется в Go-коде (8000)
EXPOSE 8000

# Команда запуска
CMD ["./proxy_service"]