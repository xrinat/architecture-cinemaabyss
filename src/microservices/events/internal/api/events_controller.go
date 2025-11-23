package api

import (
	"encoding/json"
	"events-service/internal/kafka"
	"events-service/internal/models"
	"log"
	"net/http"

	"github.com/go-chi/render"
)

// EventsHandler - Контроллер, который содержит зависимости (EventProducer)
type EventsHandler struct {
	producer *kafka.EventProducer
}

// NewEventsHandler - Конструктор с инъекцией зависимостей
func NewEventsHandler(p *kafka.EventProducer) *EventsHandler {
	return &EventsHandler{producer: p}
}

// writeJSONResponse - Хелпер для отправки JSON-ответа
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// GetHealth - GET /api/events/health
func (h *EventsHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	log.Println("Health check requested.")
	writeJSONResponse(w, http.StatusOK, map[string]bool{"status": true})
}

// handleEventCreation - Универсальная функция для POST-запросов
func (h *EventsHandler) handleEventCreation(w http.ResponseWriter, r *http.Request, topic string, eventModel models.Eventer) {
	// 1. Десериализация
	if err := json.NewDecoder(r.Body).Decode(eventModel); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 2. Инициализация полей BaseEvent
	// Это критический шаг: в C# он был в конструкторе/свойствах. 
	// В Go это нужно делать явно, встраивая BaseEvent в структуру и присваивая новые значения.
	baseEvent := models.NewBaseEvent()
	// Используем reflect или интерфейс, чтобы обновить встроенные поля BaseEvent
	// Это обходной путь. Более чистый способ - создать NewXEvent() функции, 
	// которые инициализируют BaseEvent. Но для соответствия C# логике:
	switch v := eventModel.(type) {
	case *models.MovieEvent:
		v.BaseEvent = baseEvent
	case *models.UserEvent:
		v.BaseEvent = baseEvent
	case *models.PaymentEvent:
		v.BaseEvent = baseEvent
	}

	// 3. Отправка в Kafka
	result, err := h.producer.ProduceAsync(topic, eventModel)
	if err != nil {
		log.Printf("Failed to produce event to Kafka for topic %s: %v", topic, err)
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Internal Server Error during event production.",
		})
		return
	}

	// 4. Успешный ответ
	writeJSONResponse(w, http.StatusCreated, result)
}

// CreateMovieEvent - POST /api/events/movie
func (h *EventsHandler) CreateMovieEvent(w http.ResponseWriter, r *http.Request) {
	h.handleEventCreation(w, r, "movie-events", &models.MovieEvent{})
}

// CreateUserEvent - POST /api/events/user
func (h *EventsHandler) CreateUserEvent(w http.ResponseWriter, r *http.Request) {
	h.handleEventCreation(w, r, "user-events", &models.UserEvent{})
}

// CreatePaymentEvent - POST /api/events/payment
func (h *EventsHandler) CreatePaymentEvent(w http.ResponseWriter, r *http.Request) {
	h.handleEventCreation(w, r, "payment-events", &models.PaymentEvent{})
}