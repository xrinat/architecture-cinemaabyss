# Этап сборки (Builder Stage)
# Используем образ Go 1.23
FROM --platform=linux/amd64 golang:1.23-alpine AS build
WORKDIR /app

# Копируем go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем приложение
# CGO_ENABLED=0 и static линковка для создания полностью статического бинарника
# (это не строго обязательно, но хорошая практика для Alpine)
RUN CGO_ENABLED=0 go build -o /app/events-service ./main.go

# --- Этап запуска (Final Stage) ---
# Используем минимальный образ Alpine
FROM --platform=linux/amd64 alpine:latest AS final
WORKDIR /app

# Копируем собранный бинарник из этапа сборки
COPY --from=build /app/events-service .

# Сервис запускается на порту 8082
ENV ASPNETCORE_URLS=http://+:8082
ENV PORT=8082
EXPOSE 8082

# Определение точки входа
ENTRYPOINT ["/app/events-service"]