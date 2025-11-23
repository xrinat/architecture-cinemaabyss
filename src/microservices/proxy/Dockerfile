FROM --platform=linux/amd64 mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /app

# Копирование файлов проекта и восстановление зависимостей
COPY *.csproj ./.
RUN dotnet restore

COPY . .
RUN dotnet publish -c Release -o /app/publish /p:UseAppHost=false
FROM --platform=linux/amd64 mcr.microsoft.com/dotnet/aspnet:8.0 AS final
WORKDIR /app
COPY --from=build /app/publish .


ENV ASPNETCORE_URLS=http://+:8000
EXPOSE 8000


ENTRYPOINT ["dotnet", "ProxyService.dll"]