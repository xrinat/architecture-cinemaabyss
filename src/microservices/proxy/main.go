package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultPort = "8000"

// reverseProxy creates a reverse proxy to the given target URL
func reverseProxy(target string) gin.HandlerFunc {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// healthCheck handles /health requests
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Proxy Service is running",
	})
}

func main() {
	// Service discovery targets (using Kubernetes service names)
	eventsTarget := os.Getenv("EVENTS_SERVICE_URL")
	if eventsTarget == "" {
		eventsTarget = "http://events-service:8082" 
	}
	
	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Setup routes
	router.GET("/health", healthCheck)

	// API Gateway Routing: routes all /api/* requests to events-service
	router.Group("/api").Any("/*action", reverseProxy(eventsTarget))

	// Start HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting Proxy Service on port %s...", port)
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