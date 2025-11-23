FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /app

# Копирование файлов проекта и восстановление зависимостей
COPY *.csproj ./.
RUN dotnet restore

# Копирование всего исходного кода и публикация
COPY . .
# Добавил /p:UseAppHost=false для повышения совместимости в Docker, хотя это необязательно для ASP.NET Core
RUN dotnet publish -c Release -o /app/publish /p:UseAppHost=false

# --- Этап запуска (Final Stage) ---
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS final
WORKDIR /app
COPY --from=build /app/publish .

# Сервис запускается на порту 8082
ENV ASPNETCORE_URLS=http://+:8082
EXPOSE 8082

# Определение точки входа (должно быть правильным, если имя DLL совпадает с именем проекта)
ENTRYPOINT ["dotnet", "EventsService.dll"]