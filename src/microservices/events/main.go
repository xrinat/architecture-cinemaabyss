package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"events-service/internal/api"
	"events-service/internal/kafka"
	"events-service/internal/models"
)

func main() {
	// --- Конфигурация ---
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Fatal("ОШИБКА КОНФИГУРАЦИИ: KAFKA_BROKERS не установлен.")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Значение по умолчанию
	}

	log.Printf("[EVENTS SERVICE] Kafka Brokers: %s", kafkaBrokers)

	// --- Инициализация сервисов ---
	producer, err := kafka.NewEventProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("Ошибка инициализации Kafka Producer: %v", err)
	}
	defer producer.Close()

	// Consumer запускается как фоновый сервис в отдельной горутине
	consumerService := kafka.NewEventConsumerService(kafkaBrokers, []string{"movie-events", "user-events", "payment-events"})
	ctx, cancel := context.WithCancel(context.Background())
	go consumerService.Execute(ctx)
	
	// --- Настройка HTTP пайплайна и контроллеров ---
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Инжекция зависимостей в контроллер (эквивалент EventsController.cs)
	eventHandler := api.NewEventsHandler(producer)
	
	r.Route("/api/events", func(r chi.Router) {
		r.Get("/health", eventHandler.GetHealth)
		r.Post("/movie", eventHandler.CreateMovieEvent)
		r.Post("/user", eventHandler.CreateUserEvent)
		r.Post("/payment", eventHandler.CreatePaymentEvent)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// --- Запуск и Graceful Shutdown ---
	log.Printf("[EVENTS SERVICE] Запущен на порту %s", port)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Не удалось запустить HTTP-сервер: %v", err)
		}
	}()

	// Ждем сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Получен сигнал завершения. Запускается graceful shutdown...")

	// Завершение работы Consumer Service
	cancel()

	// Остановка HTTP-сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP-сервер принудительно завершил работу: %v", err)
	}

	log.Println("Сервис завершил работу.")
}