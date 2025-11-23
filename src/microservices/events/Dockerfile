# --- Этап сборки (Build Stage) ---
# Используем базовый образ Golang с Alpine для небольшого размера
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем приложение в статический бинарник
# CGO_ENABLED=0 и флаги -ldflags '-s -w' уменьшают размер бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-s -w' -o events_service .


# --- Этап запуска (Run Stage) ---
# Используем минимальный образ Alpine Linux
FROM alpine:latest
# Устанавливаем сертификаты CA для HTTPS-запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник из этапа сборки
COPY --from=builder /app/events_service .

# Открываем порт, который используется в Go-коде (8082)
EXPOSE 8082

# Команда запуска
CMD ["./events_service"]