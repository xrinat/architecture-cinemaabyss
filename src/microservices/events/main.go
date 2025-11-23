package main

import (
    "context"
    "events-service-go/handlers"
    "events-service-go/kafka"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8082"
    }

    brokers := os.Getenv("KAFKA_BROKERS")
    if brokers == "" {
        logrus.Fatal("KAFKA_BROKERS not set")
    }

    producer := kafka.NewProducer([]string{brokers})
    defer producer.Close()

    consumer := kafka.NewConsumer([]string{brokers}, "events-service-group-1", []string{"movie-events", "user-events", "payment-events"})
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    consumer.StartConsuming(ctx)
    defer consumer.Close()

    h := handlers.NewHandler(producer)
    r := gin.Default()

    api := r.Group("/api/events")
    {
        api.GET("/health", h.Health)
        api.POST("/movie", h.CreateMovieEvent)
        api.POST("/user", h.CreateUserEvent)
        api.POST("/payment", h.CreatePaymentEvent)
    }

    go func() {
        if err := r.Run(":" + port); err != nil {
            logrus.Fatalf("Failed to run server: %v", err)
        }
    }()

    logrus.Infof("[EVENTS SERVICE] Running on port %s | Kafka Brokers: %s", port, brokers)

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logrus.Info("Shutting down server...")
    cancel()
    time.Sleep(1 * time.Second)
}
