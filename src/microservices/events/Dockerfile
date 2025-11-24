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

# Сервис запускается на порту 8082, как указано в docker-compose
ENV ASPNETCORE_URLS=http://+:8082

# Определение точки входа
ENTRYPOINT ["dotnet", "EventsService.dll"]
=======
# --- Этап сборки ---
FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o events-service

# --- Этап запуска ---
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/events-service .

ENV PORT=8082
EXPOSE 8082

ENTRYPOINT ["/app/events-service"]
>>>>>>> cinema
