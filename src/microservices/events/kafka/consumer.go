package kafka

import (
    "context"
    "encoding/json"
    "events-service-go/models"
    "github.com/segmentio/kafka-go"
    "github.com/sirupsen/logrus"
)

type Consumer struct {
    reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID string, topics []string) *Consumer {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: brokers,
        GroupID: groupID,
        Topic:   topics[0], // Для нескольких топиков нужно отдельный reader или динамическая подписка
    })
    logrus.Infof("Kafka Consumer initialized for topics: %v", topics)
    return &Consumer{reader: reader}
}

func (c *Consumer) StartConsuming(ctx context.Context) {
    go func() {
        for {
            m, err := c.reader.ReadMessage(ctx)
            if err != nil {
                if ctx.Err() != nil {
                    break
                }
                logrus.Errorf("Error consuming message: %v", err)
                continue
            }

            logrus.Infof("--- Event Consumed --- | Topic: %s | Partition: %d | Offset: %d", m.Topic, m.Partition, m.Offset)

            // Простейшая десериализация
            switch m.Topic {
            case "movie-events":
                var ev models.MovieEvent
                if err := json.Unmarshal(m.Value, &ev); err != nil {
                    logrus.Errorf("Failed to deserialize MovieEvent: %v", err)
                } else {
                    logrus.Infof("Processed MovieEvent: %+v", ev)
                }
            case "user-events":
                var ev models.UserEvent
                if err := json.Unmarshal(m.Value, &ev); err != nil {
                    logrus.Errorf("Failed to deserialize UserEvent: %v", err)
                } else {
                    logrus.Infof("Processed UserEvent: %+v", ev)
                }
            case "payment-events":
                var ev models.PaymentEvent
                if err := json.Unmarshal(m.Value, &ev); err != nil {
                    logrus.Errorf("Failed to deserialize PaymentEvent: %v", err)
                } else {
                    logrus.Infof("Processed PaymentEvent: %+v", ev)
                }
            }
        }
    }()
}

func (c *Consumer) Close() error {
    return c.reader.Close()
}
