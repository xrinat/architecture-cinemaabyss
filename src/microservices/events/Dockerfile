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