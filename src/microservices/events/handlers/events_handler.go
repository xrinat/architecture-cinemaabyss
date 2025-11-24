package handlers

import (
    "events-service-go/kafka"
    "events-service-go/models"
    "github.com/gin-gonic/gin"
    "net/http"
)

type Handler struct {
    Producer *kafka.Producer
}

func NewHandler(producer *kafka.Producer) *Handler {
    return &Handler{Producer: producer}
}

func (h *Handler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": true})
}

func (h *Handler) CreateMovieEvent(c *gin.Context) {
    var ev models.MovieEvent
    if err := c.BindJSON(&ev); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    resp, err := h.Producer.ProduceEvent("movie-events", ev)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreateUserEvent(c *gin.Context) {
    var ev models.UserEvent
    if err := c.BindJSON(&ev); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    resp, err := h.Producer.ProduceEvent("user-events", ev)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreatePaymentEvent(c *gin.Context) {
    var ev models.PaymentEvent
    if err := c.BindJSON(&ev); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    resp, err := h.Producer.ProduceEvent("payment-events", ev)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, resp)
}
