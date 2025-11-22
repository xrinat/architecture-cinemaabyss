using System.Text.Json;
using Confluent.Kafka;
using EventsService.Models;

namespace EventsService.Kafka
{
    public class EventProducer : IDisposable
    {
        private readonly IProducer<Null, string> _producer;
        private readonly ILogger<EventProducer> _logger;

        public EventProducer(IConfiguration configuration, ILogger<EventProducer> logger)
        {
            _logger = logger;
            // KAFKA_BROKERS: kafka:9092 - берется из docker-compose
            var config = new ProducerConfig
            {
                BootstrapServers = configuration["KAFKA_BROKERS"],
                Acks = Acks.All // Ждем подтверждения от всех реплик
            };

            // Создаем Producer с Null ключом и строковым значением (JSON)
            _producer = new ProducerBuilder<Null, string>(config).Build();
            _logger.LogInformation("Kafka Producer initialized with brokers: {Brokers}", config.BootstrapServers);
        }

        // Универсальный метод для отправки любого события в указанный топик
        public async Task<EventResponse> ProduceAsync<T>(string topic, T eventData) where T : BaseEvent
        {
            try
            {
                var eventJson = JsonSerializer.Serialize(eventData);
                
                // Создание Kafka сообщения
                var message = new Message<Null, string> { Value = eventJson };

                // Отправка сообщения
                var deliveryResult = await _producer.ProduceAsync(topic, message);
                
                _logger.LogInformation(
                    "Event produced successfully to topic '{Topic}' | Type: {EventType} | Partition: {Partition} | Offset: {Offset}",
                    topic, eventData.EventType, deliveryResult.Partition.Value, deliveryResult.Offset.Value);

                return new EventResponse
                {
                    Status = "success",
                    Partition = deliveryResult.Partition.Value,
                    Offset = deliveryResult.Offset.Value,
                    Event = eventData
                };
            }
            catch (ProduceException<Null, string> e)
            {
                _logger.LogError(e, "Delivery failed for event to topic '{Topic}': {Reason}", topic, e.Error.Reason);
                throw;
            }
        }

        public void Dispose()
        {
            // Ждем завершения всех асинхронных операций
            _producer.Flush(TimeSpan.FromSeconds(10));
            _producer.Dispose();
        }
    }
}