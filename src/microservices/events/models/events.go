package models

import "time"

type BaseEvent struct {
    EventID   string    `json:"eventId"`
    Timestamp time.Time `json:"timestamp"`
}

type MovieEvent struct {
    BaseEvent
    MovieID int    `json:"movieId"`
    Title   string `json:"title"`
    Action  string `json:"action"`
    UserID  *int   `json:"userId,omitempty"`
}

type UserEvent struct {
    BaseEvent
    UserID   int     `json:"userId"`
    Action   string  `json:"action"`
    Username *string `json:"username,omitempty"`
    Email    *string `json:"email,omitempty"`
}

type PaymentEvent struct {
    BaseEvent
    PaymentID int     `json:"paymentId"`
    UserID    int     `json:"userId"`
    Amount    float64 `json:"amount"`
    Status    string  `json:"status"`
}

type EventResponse struct {
    Status    string      `json:"status"`
    Partition int         `json:"partition"`
    Offset    int64       `json:"offset"`
    Event     interface{} `json:"event"`
}
