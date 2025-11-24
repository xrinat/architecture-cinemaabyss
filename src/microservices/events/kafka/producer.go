package kafka

import (
    "context"
    "encoding/json"
    "time"

    "events-service-go/models"
    "github.com/segmentio/kafka-go"
    "github.com/sirupsen/logrus"
)

type Producer struct {
    writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
    writer := &kafka.Writer{
        Addr:         kafka.TCP(brokers...),
        RequiredAcks: kafka.RequireAll,
        Async:        false,
    }
    logrus.Infof("Kafka Producer initialized with brokers: %v", brokers)
    return &Producer{writer: writer}
}

func (p *Producer) ProduceEvent(topic string, event interface{}) (*models.EventResponse, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    value, err := json.Marshal(event)
    if err != nil {
        return nil, err
    }

    msg := kafka.Message{
        Topic: topic,
        Value: value,
    }

    err = p.writer.WriteMessages(ctx, msg)
    if err != nil {
        logrus.Errorf("Delivery failed for event to topic '%s': %v", topic, err)
        return nil, err
    }

    return &models.EventResponse{
        Status:    "success",
        Partition: 0,
        Offset:    0,
        Event:     event,
    }, nil
}

func (p *Producer) Close() error {
    return p.writer.Close()
}
