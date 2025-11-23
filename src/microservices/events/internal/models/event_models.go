package models

import (
	"time"

	"github.com/google/uuid"
)

// BaseEvent - Базовый класс для всех событий
type BaseEvent struct {
	EventID   uuid.UUID `json:"eventId"`
	Timestamp time.Time `json:"timestamp"`
}

type Eventer interface {
	GetEventType() string
}

// MovieEvent - Событие фильма
type MovieEvent struct {
	BaseEvent
	MovieID int    `json:"movieId"`
	Title   string `json:"title"`
	Action  string `json:"action"` // e.g., "viewed", "updated"
	UserID  *int   `json:"userId,omitempty"` // Nullable int
}

func (e *MovieEvent) GetEventType() string {
	return "Movie"
}

// UserEvent - Событие пользователя
type UserEvent struct {
	BaseEvent
	UserID   int     `json:"userId"`
	Action   string  `json:"action"` // e.g., "registered", "logged_in"
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
}

func (e *UserEvent) GetEventType() string {
	return "User"
}

// PaymentEvent - Событие платежа
type PaymentEvent struct {
	BaseEvent
	PaymentID int     `json:"paymentId"`
	UserID    int     `json:"userId"`
	Amount    float64 `json:"amount"` // Используем float64 для decimal
	Status    string  `json:"status"` // e.g., "completed", "failed"
}

func (e *PaymentEvent) GetEventType() string {
	return "Payment"
}

// EventResponse - Ответ API
type EventResponse struct {
	Status    string      `json:"status"`
	Partition int32       `json:"partition"`
	Offset    int64       `json:"offset"`
	Event     interface{} `json:"event"` // Храним общее событие
}

// NewBaseEvent инициализирует базовые поля
func NewBaseEvent() BaseEvent {
	return BaseEvent{
		EventID:   uuid.New(),
		Timestamp: time.Now().UTC(),
	}
}