using System.Net;
using System.Net.Http.Headers;
using System.Text.RegularExpressions;

var builder = WebApplication.CreateBuilder(args);

// --- Конфигурация из переменных окружения (Environment Variables) ---
var monolithUrl = builder.Configuration.GetValue<string>("MONOLITH_URL");
var moviesServiceUrl = builder.Configuration.GetValue<string>("MOVIES_SERVICE_URL");
var eventsServiceUrl = builder.Configuration.GetValue<string>("EVENTS_SERVICE_URL"); // Добавлен URL для Events Service

// Feature Flag для миграции movies
var isGradualMigrationEnabled = builder.Configuration.GetValue<bool>("GRADUAL_MIGRATION");
var migrationPercent = builder.Configuration.GetValue<int>("MOVIES_MIGRATION_PERCENT");

// Проверка обязательных конфигураций
if (string.IsNullOrEmpty(monolithUrl) || string.IsNullOrEmpty(moviesServiceUrl) || string.IsNullOrEmpty(eventsServiceUrl))
{
    Console.WriteLine("ОШИБКА КОНФИГУРАЦИИ: MONOLITH_URL, MOVIES_SERVICE_URL или EVENTS_SERVICE_URL не установлены.");
    Environment.Exit(1);
}

Console.WriteLine($"[КОНФИГУРАЦИЯ] Монолит: {monolithUrl}");
Console.WriteLine($"[КОНФИГУРАЦИЯ] Movies Microservice: {moviesServiceUrl}");
Console.WriteLine($"[КОНФИГУРАЦИЯ] Events Microservice: {eventsServiceUrl}");
Console.WriteLine($"[КОНФИГУРАЦИЯ] Постепенная миграция Movies включена: {isGradualMigrationEnabled}");
Console.WriteLine($"[КОНФИГУРАЦИЯ] Процент трафика Movies на новый сервис: {migrationPercent}%");

// Регистрируем HttpClient для безопасного и эффективного использования
builder.Services.AddHttpClient();

var app = builder.Build();

// Инициализируем Random для реализации процентного переключения трафика
var random = new Random();

// Универсальная функция проксирования
async Task ProxyRequest(HttpContext context, string targetBaseUrl, string targetServiceName, string path, IHttpClientFactory clientFactory)
{
    // Формирование полного URL для проксирования, включая оригинальный QueryString
    var targetUrl = $"{targetBaseUrl}{context.Request.Path.Value}{context.Request.QueryString}";

    // 3. Создание HTTP-запроса
    var client = clientFactory.CreateClient();
    var request = new HttpRequestMessage(new HttpMethod(context.Request.Method), targetUrl);

    // Копирование заголовков запроса
    foreach (var header in context.Request.Headers)
    {
        // Headers.TryAddWithoutValidation для заголовков запроса
        if (!request.Headers.TryAddWithoutValidation(header.Key, header.Value.ToArray()))
        {
            // Если не удалось, пробуем добавить в Content Headers (для POST/PUT/PATCH)
            request.Content?.Headers.TryAddWithoutValidation(header.Key, header.Value.ToArray());
        }
    }

    // Добавление тела запроса (для POST/PUT/PATCH)
    if (context.Request.ContentLength.HasValue && context.Request.ContentLength > 0 && context.Request.Body != null)
    {
        request.Content = new StreamContent(context.Request.Body);
        if (context.Request.ContentType != null)
        {
            // Копирование типа контента
            request.Content.Headers.ContentType = MediaTypeHeaderValue.Parse(context.Request.ContentType);
        }
    }
    
    // Добавляем специальный заголовок для отладки
    request.Headers.Add("X-Strangler-Route", targetServiceName);
    request.Headers.Add("X-Original-Path", context.Request.Path.Value);


    // 4. Отправка запроса и получение ответа
    HttpResponseMessage? response;
    try
    {
        response = await client.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, context.RequestAborted);
    }
    catch (HttpRequestException ex)
    {
        // Обработка ошибок сетевого взаимодействия (например, недоступность сервиса)
        context.Response.StatusCode = (int)HttpStatusCode.GatewayTimeout;
        await context.Response.WriteAsync($"Ошибка проксирования к {targetServiceName}: {ex.Message}");
        Console.WriteLine($"ОШИБКА ПРОКСИРОВАНИЯ к {targetServiceName}: {ex.Message}");
        return;
    }


    // 5. Копирование заголовков и статуса ответа
    context.Response.StatusCode = (int)response.StatusCode;
    
    // Копирование заголовков ответа
    foreach (var header in response.Headers)
    {
        context.Response.Headers.TryAdd(header.Key, header.Value.ToArray());
    }
    // Копирование заголовков контента
    foreach (var header in response.Content.Headers)
    {
        context.Response.Headers.TryAdd(header.Key, header.Value.ToArray());
    }

    // Удаляем потенциально конфликтующие заголовки
    context.Response.Headers.Remove("transfer-encoding");
    context.Response.Headers.Remove("Content-Length");


    // 6. Копирование тела ответа
    await response.Content.CopyToAsync(context.Response.Body);
}


// --- 1. HEALTH Check (Проверка работоспособности самого прокси) ---
app.MapGet("/health", () =>
{
    // Как указано в OpenAPI спецификации
    return Results.Text("Strangler Fig Proxy is healthy");
});


// --- 2. MOVIES Routes (С паттерном Strangler Fig) ---

// 2a. Точный путь /api/movies (обработка GET http://localhost:8000/api/movies)
app.MapMethods("/api/movies", new[] { "GET", "POST", "PUT", "DELETE" }, async (
    HttpContext context,
    IHttpClientFactory clientFactory) =>
{
    // 1. Определение целевого URL (Monolith или Microservice)
    var targetBaseUrl = monolithUrl;
    var targetServiceName = "Monolith";

    // Логика Strangler Fig с Feature Flag
    if (isGradualMigrationEnabled)
    {
        var randomNumber = random.Next(1, 101); // 1-100

        if (randomNumber <= migrationPercent)
        {
            targetBaseUrl = moviesServiceUrl;
            targetServiceName = "Movies-Service";
        }
    }
    
    Console.WriteLine($"[ЗАПРОС] URI: {context.Request.Path}. Маршрутизация: {targetServiceName} (Процент: {migrationPercent}%)");

    // Передаем пустую строку для 'path', так как маршрут не захватывает подпуть
    await ProxyRequest(context, targetBaseUrl, targetServiceName, "", clientFactory);
});


// 2b. Пути с подстрокой /api/movies/{*path}
app.MapMethods("/api/movies/{*path}", new[] { "GET", "POST", "PUT", "DELETE" }, async (
    HttpContext context,
    string path, // захватывает подпуть (например, "123" из /api/movies/123)
    IHttpClientFactory clientFactory) =>
{
    // 1. Определение целевого URL (Monolith или Microservice)
    var targetBaseUrl = monolithUrl;
    var targetServiceName = "Monolith";

    // Логика Strangler Fig с Feature Flag
    if (isGradualMigrationEnabled)
    {
        var randomNumber = random.Next(1, 101); // 1-100

        if (randomNumber <= migrationPercent)
        {
            targetBaseUrl = moviesServiceUrl;
            targetServiceName = "Movies-Service";
        }
    }
    
    // В этом случае path будет содержать подпуть, например "123"
    Console.WriteLine($"[ЗАПРОС] URI: {context.Request.Path}. Маршрутизация: {targetServiceName} (Процент: {migrationPercent}%)");

    await ProxyRequest(context, targetBaseUrl, targetServiceName, path, clientFactory);
});


// --- 3. EVENTS Routes (Полностью мигрированы) ---
// Все, что начинается с /api/events, перенаправляется в Events Microservice
app.MapMethods("/api/events/{*path}", new[] { "GET", "POST", "PUT", "DELETE" }, async (
    HttpContext context,
    string path,
    IHttpClientFactory clientFactory) =>
{
    const string targetServiceName = "Events-Service";
    Console.WriteLine($"[ЗАПРОС] URI: {context.Request.Path}. Маршрутизация: {targetServiceName} (Полностью мигрирован)");
    
    await ProxyRequest(context, eventsServiceUrl, targetServiceName, path, clientFactory);
});


// --- 4. DEFAULT Routes (Пока остаются в Монолите) ---
// Ловит все остальные запросы: users, payments, subscriptions
// Используем явные маршруты для надежного захвата базовых путей и подпутей.
const string monolithServicesRegex = "(users|payments|subscriptions)";

// 4a. Точные пути: /api/users, /api/payments, /api/subscriptions
app.MapMethods($"/api/{{service:regex({monolithServicesRegex})}}", new[] { "GET", "POST", "PUT", "DELETE" }, async (
    HttpContext context,
    string service,
    IHttpClientFactory clientFactory) =>
{
    const string targetServiceName = "Monolith";
    Console.WriteLine($"[ЗАПРОС] URI: {context.Request.Path}. Маршрутизация: {targetServiceName} (Базовый маршрут: {service})");
    
    // Используем Path.Value, чтобы сохранить полный путь, включая начальный слэш
    await ProxyRequest(context, monolithUrl, targetServiceName, service, clientFactory);
});

// 4b. Пути с подстрокой: /api/users/123, /api/payments/status
app.MapMethods($"/api/{{service:regex({monolithServicesRegex})}}/{{*path}}", new[] { "GET", "POST", "PUT", "DELETE" }, async (
    HttpContext context,
    string service,
    string path,
    IHttpClientFactory clientFactory) =>
{
    const string targetServiceName = "Monolith";
    Console.WriteLine($"[ЗАПРОС] URI: {context.Request.Path}. Маршрутизация: {targetServiceName} (Подмаршрут: {service}/{path})");
    
    // Используем Path.Value, чтобы сохранить полный путь, включая начальный слэш
    await ProxyRequest(context, monolithUrl, targetServiceName, path, clientFactory);
});

// Запросы, не пойманые ни одним из вышеперечисленных маршрутов (например, корневой путь)
app.MapGet("/", () => "Proxy Service запущен и ожидает трафика на /api/* маршрутах.");

app.Run();