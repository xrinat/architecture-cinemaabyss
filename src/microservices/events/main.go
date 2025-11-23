package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultPort = "8082"

// healthCheck handles /health requests
func healthCheck(c *gin.Context) {
	// In a real application, this should check the Kafka connection
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Events Service is running",
	})
}

func main() {
	// Set Gin to Release Mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Setup routes
	router.GET("/health", healthCheck)
	
	// Kafka Configuration (simplified for this example)
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Println("WARNING: KAFKA_BROKERS environment variable is not set. Assuming kafka-service:9092")
		kafkaBrokers = "kafka-service:9092"
	}
	log.Printf("Starting Kafka consumer connected to: %s", kafkaBrokers)

	// Start HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting Events Service on port %s...", port)
	s := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve error: %v", err)
	}
}