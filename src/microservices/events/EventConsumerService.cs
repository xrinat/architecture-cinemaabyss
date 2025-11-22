using System.Text.Json;
using Confluent.Kafka;
using EventsService.Models;

namespace EventsService.Kafka
{
    public class EventConsumerService : BackgroundService
    {
        private readonly ILogger<EventConsumerService> _logger;
        private readonly IConfiguration _configuration;
        private readonly string[] _topics = new[] { "movie-events", "user-events", "payment-events" };

        public EventConsumerService(ILogger<EventConsumerService> logger, IConfiguration configuration)
        {
            _logger = logger;
            _configuration = configuration;
        }

        protected override Task ExecuteAsync(CancellationToken stoppingToken)
        {
            // Выполняем логику Consumer'а в отдельной задаче, чтобы не блокировать BackgroundService
            return Task.Run(() => ConsumeEvents(stoppingToken), stoppingToken);
        }

        private void ConsumeEvents(CancellationToken stoppingToken)
        {
            var config = new ConsumerConfig
            {
                BootstrapServers = _configuration["KAFKA_BROKERS"],
                GroupId = "events-service-group-1", // Уникальная группа для нашего сервиса
                AutoOffsetReset = AutoOffsetReset.Earliest, // Начинаем чтение с начала, если нет сохраненного смещения
                EnableAutoCommit = true, // Автоматически фиксируем смещение
                ClientId = "events-consumer-1"
            };

            try
            {
                using var consumer = new ConsumerBuilder<Ignore, string>(config)
                    .SetErrorHandler((_, e) => _logger.LogError("Kafka Error: {Reason}", e.Reason))
                    .SetLogHandler((_, log) => _logger.LogInformation("Kafka Log: {Facility} - {Message}", log.Facility, log.Message))
                    .Build();

                consumer.Subscribe(_topics);
                _logger.LogInformation("Kafka Consumer subscribed to topics: {Topics}", string.Join(", ", _topics));

                while (!stoppingToken.IsCancellationRequested)
                {
                    try
                    {
                        var consumeResult = consumer.Consume(stoppingToken);
                        
                        // Получено сообщение
                        var topic = consumeResult.Topic;
                        var message = consumeResult.Message.Value;
                        
                        _logger.LogWarning(
                            "--- Event Consumed --- | Topic: {Topic} | Partition: {Partition} | Offset: {Offset}", 
                            topic, consumeResult.Partition.Value, consumeResult.Offset.Value);

                        // Простая имитация обработки: десериализация и логирование
                        // На реальном проекте здесь была бы сложная бизнес-логика (например, обновление БД, отправка уведомлений)
                        
                        // Пытаемся десериализовать в общий тип для логирования
                        try
                        {
                            // В реальном проекте необходимо более сложное определение типа по топику
                            object? eventObject = null;
                            if (topic == "movie-events")
                            {
                                eventObject = JsonSerializer.Deserialize<MovieEvent>(message);
                            }
                            else if (topic == "user-events")
                            {
                                eventObject = JsonSerializer.Deserialize<UserEvent>(message);
                            }
                            else if (topic == "payment-events")
                            {
                                eventObject = JsonSerializer.Deserialize<PaymentEvent>(message);
                            }
                            
                            _logger.LogInformation("Successfully processed {Type} event: {Message}", 
                                eventObject?.GetType().Name ?? "Unknown", message);
                        }
                        catch (JsonException ex)
                        {
                            _logger.LogError(ex, "Failed to deserialize event from topic {Topic}. Message: {Message}", topic, message);
                        }
                        
                    }
                    catch (ConsumeException e)
                    {
                        _logger.LogError(e, "Error consuming message: {Reason}", e.Error.Reason);
                    }
                    catch (OperationCanceledException)
                    {
                        // Нормальное завершение при остановке сервиса
                        break;
                    }
                    // Короткая пауза для снижения нагрузки, если нет сообщений (хотя Consume блокирующий)
                    // Task.Delay(100, stoppingToken).Wait(stoppingToken);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Unexpected error in Kafka consumer service.");
            }
        }
    }
}