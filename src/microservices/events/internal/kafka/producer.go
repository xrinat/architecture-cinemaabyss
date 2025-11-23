package kafka

import (
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"events-service/internal/models"
)

// EventProducer - Эквивалент C# EventProducer
type EventProducer struct {
	p *kafka.Producer
}

// NewEventProducer создает и инициализирует новый Kafka Producer
func NewEventProducer(brokers string) (*EventProducer, error) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"acks":              "all", // Ждем подтверждения от всех реплик
	}

	p, err := kafka.NewProducer(configMap)
	if err != nil {
		return nil, err
	}

	log.Printf("Kafka Producer initialized with brokers: %s", brokers)
	return &EventProducer{p: p}, nil
}

// ProduceAsync - Универсальный метод для отправки события
func (ep *EventProducer) ProduceAsync(topic string, eventData models.Eventer) (models.EventResponse, error) {
	// 1. Инициализация BaseEvent полей (EventID, Timestamp)
	// В Go это делается вручную, так как нет автоматических конструкторов/инициализаторов.
	// При этом предполагаем, что поля BaseEvent встроены в структуры событий.
	// Для простоты, этот шаг лучше сделать в EventModels.go или в контроллере. 
	// Здесь мы просто сериализуем то, что пришло.

	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return models.EventResponse{}, err
	}

	// Канал для доставки
	deliveryChan := make(chan kafka.Event)

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          eventJSON,
		Key:            nil, // Null ключ
	}

	err = ep.p.Produce(message, deliveryChan)
	if err != nil {
		return models.EventResponse{}, err
	}

	// Ждем результата доставки
	e := <-deliveryChan
	m := e.(*kafka.Message)
	close(deliveryChan)

	if m.TopicPartition.Error != nil {
		log.Printf("Delivery failed for event to topic '%s': %v", topic, m.TopicPartition.Error)
		return models.EventResponse{}, m.TopicPartition.Error
	}

	// Успешная доставка
	response := models.EventResponse{
		Status:    "success",
		Partition: m.TopicPartition.Partition.Int32(),
		Offset:    m.TopicPartition.Offset.Int64(),
		Event:     eventData,
	}

	log.Printf("Event produced successfully to topic '%s' | Type: %s | Partition: %d | Offset: %d",
		topic, eventData.GetEventType(), response.Partition, response.Offset)

	return response, nil
}

// Close - Эквивалент Dispose
func (ep *EventProducer) Close() {
	// Ждем завершения всех асинхронных операций
	ep.p.Flush(10000) // 10000 мс = 10 секунд
	ep.p.Close()
	log.Println("Kafka Producer closed.")
}