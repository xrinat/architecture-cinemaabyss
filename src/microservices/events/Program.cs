using EventsService.Kafka;
using EventsService.Controllers; // Убедимся, что контроллеры доступны

var builder = WebApplication.CreateBuilder(args);

// --- Конфигурация ---
// KAFKA_BROKERS берется из docker-compose
if (string.IsNullOrEmpty(builder.Configuration["KAFKA_BROKERS"]))
{
    Console.WriteLine("ОШИБКА КОНФИГУРАЦИИ: KAFKA_BROKERS не установлен.");
    Environment.Exit(1);
}

// --- Добавление сервисов ---
builder.Services.AddControllers();
// Регистрируем Producer как Scoped или Singleton. Singleton предпочтительнее для Producer.
builder.Services.AddSingleton<EventProducer>();
// Регистрируем фоновый сервис для Consumer
builder.Services.AddHostedService<EventConsumerService>();

var app = builder.Build();

// --- Настройка пайплайна HTTP ---
app.UseRouting();
app.MapControllers();

Console.WriteLine($"[EVENTS SERVICE] Запущен на порту {builder.Configuration["PORT"] ?? "8082"}");
Console.WriteLine($"[EVENTS SERVICE] Kafka Brokers: {builder.Configuration["KAFKA_BROKERS"]}");

app.Run();