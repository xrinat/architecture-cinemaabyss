package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"events-service/internal/models"
)

// EventConsumerService - Эквивалент C# EventConsumerService
type EventConsumerService struct {
	brokers string
	topics  []string
}

func NewEventConsumerService(brokers string, topics []string) *EventConsumerService {
	return &EventConsumerService{
		brokers: brokers,
		topics:  topics,
	}
}

// Execute - Запускает логику потребителя. Эквивалент Protected override Task ExecuteAsync
func (ecs *EventConsumerService) Execute(ctx context.Context) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers":  ecs.brokers,
		"group.id":           "events-service-group-1",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
		"client.id":          "events-consumer-1",
	}

	c, err := kafka.NewConsumer(configMap)
	if err != nil {
		log.Printf("Не удалось создать Kafka Consumer: %v", err)
		return
	}
	defer c.Close()

	if err := c.SubscribeTopics(ecs.topics, nil); err != nil {
		log.Printf("Не удалось подписаться на топики: %v", err)
		return
	}
	log.Printf("Kafka Consumer subscribed to topics: %s", ecs.topics)

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка Kafka Consumer Service...")
			return
		default:
			// Блокирующая операция потребления с таймаутом (эквивалент Consume(stoppingToken))
			msg, err := c.ReadMessage(time.Second) 
			
			if err == nil {
				// Успешно получено сообщение
				topic := *msg.TopicPartition.Topic
				message := string(msg.Value)
				
				log.Printf("--- Event Consumed --- | Topic: %s | Partition: %d | Offset: %d", 
					topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)

				// Простая имитация обработки: десериализация и логирование
				var eventObject interface{}
				var jsonErr error

				switch topic {
				case "movie-events":
					var event models.MovieEvent
					jsonErr = json.Unmarshal(msg.Value, &event)
					eventObject = event
				case "user-events":
					var event models.UserEvent
					jsonErr = json.Unmarshal(msg.Value, &event)
					eventObject = event
				case "payment-events":
					var event models.PaymentEvent
					jsonErr = json.Unmarshal(msg.Value, &event)
					eventObject = event
				default:
					log.Printf("Received message from unknown topic %s. Message: %s", topic, message)
				}

				if jsonErr != nil {
					log.Printf("Failed to deserialize event from topic %s. Message: %s. Error: %v", topic, message, jsonErr)
				} else {
					log.Printf("Successfully processed %T event: %s", eventObject, message)
				}
				
			} else if !err.(kafka.Error).IsTimeout() {
				// Ошибка потребления, не таймаут
				log.Printf("Error consuming message: %v", err)
			}
		}
	}
}