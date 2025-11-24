<<<<<<< HEAD
# Этап сборки (Build Stage)
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /app

# Копирование файлов проекта и восстановление зависимостей
COPY *.csproj ./
RUN dotnet restore

# Копирование всего исходного кода и публикация
COPY . .
RUN dotnet publish -c Release -o /app/publish

# Этап запуска (Final Stage)
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS final
WORKDIR /app
COPY --from=build /app/publish .

# Сервис запускается на порту 8000, как указано в docker-compose
ENV ASPNETCORE_URLS=http://+:8000

# Определение точки входа
ENTRYPOINT ["dotnet", "ProxyService.dll"]
=======
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
>>>>>>> cinema
