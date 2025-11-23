package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// --- Конфигурация из переменных окружения ---
	monolithURL := os.Getenv("MONOLITH_URL")
	moviesServiceURL := os.Getenv("MOVIES_SERVICE_URL")
	eventsServiceURL := os.Getenv("EVENTS_SERVICE_URL")

	if monolithURL == "" || moviesServiceURL == "" || eventsServiceURL == "" {
		log.Fatal("ОШИБКА КОНФИГУРАЦИИ: MONOLITH_URL, MOVIES_SERVICE_URL или EVENTS_SERVICE_URL не установлены")
	}

	gradualMigration := false
	if val := os.Getenv("GRADUAL_MIGRATION"); val == "true" {
		gradualMigration = true
	}

	migrationPercent := 0
	if val := os.Getenv("MOVIES_MIGRATION_PERCENT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			migrationPercent = p
		}
	}

	fmt.Printf("[КОНФИГУРАЦИЯ] Монолит: %s\n", monolithURL)
	fmt.Printf("[КОНФИГУРАЦИЯ] Movies Microservice: %s\n", moviesServiceURL)
	fmt.Printf("[КОНФИГУРАЦИЯ] Events Microservice: %s\n", eventsServiceURL)
	fmt.Printf("[КОНФИГУРАЦИЯ] Постепенная миграция Movies включена: %v\n", gradualMigration)
	fmt.Printf("[КОНФИГУРАЦИЯ] Процент трафика Movies на новый сервис: %d%%\n", migrationPercent)

	rand.Seed(time.Now().UnixNano())

	r := mux.NewRouter()

	// --- 1. HEALTH Check ---
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Strangler Fig Proxy is healthy"))
	})

	// --- 2. MOVIES Routes ---
	handleMovies := func(w http.ResponseWriter, r *http.Request) {
		targetURL := monolithURL
		targetService := "Monolith"

		if gradualMigration && rand.Intn(100)+1 <= migrationPercent {
			targetURL = moviesServiceURL
			targetService = "Movies-Service"
		}

		fmt.Printf("[ЗАПРОС] URI: %s. Маршрутизация: %s (Процент: %d%%)\n", r.RequestURI, targetService, migrationPercent)
		proxyRequest(w, r, targetURL, targetService)
	}

	r.HandleFunc("/api/movies", handleMovies).Methods("GET", "POST", "PUT", "DELETE")
	r.HandleFunc("/api/movies/{rest:.*}", handleMovies).Methods("GET", "POST", "PUT", "DELETE")

	// --- 3. EVENTS Routes ---
	r.HandleFunc("/api/events/{rest:.*}", func(w http.ResponseWriter, r *http.Request) {
		targetService := "Events-Service"
		fmt.Printf("[ЗАПРОС] URI: %s. Маршрутизация: %s (Полностью мигрирован)\n", r.RequestURI, targetService)
		proxyRequest(w, r, eventsServiceURL, targetService)
	}).Methods("GET", "POST", "PUT", "DELETE")

	// --- 4. DEFAULT Routes (Monolith) ---
	monolithServices := []string{"users", "payments", "subscriptions"}
	for _, svc := range monolithServices {
		r.HandleFunc("/api/"+svc, func(s string) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				targetService := "Monolith"
				fmt.Printf("[ЗАПРОС] URI: %s. Маршрутизация: %s (Базовый маршрут: %s)\n", r.RequestURI, targetService, s)
				proxyRequest(w, r, monolithURL, targetService)
			}
		}(svc)).Methods("GET", "POST", "PUT", "DELETE")

		r.HandleFunc("/api/"+svc+"/{rest:.*}", func(s string) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				targetService := "Monolith"
				fmt.Printf("[ЗАПРОС] URI: %s. Маршрутизация: %s (Подмаршрут: %s)\n", r.RequestURI, targetService, s)
				proxyRequest(w, r, monolithURL, targetService)
			}
		}(svc)).Methods("GET", "POST", "PUT", "DELETE")
	}

	// --- 5. ROOT ---
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Proxy Service запущен и ожидает трафика на /api/* маршрутах."))
	})

	fmt.Println("Proxy Service запущен на :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

// proxyRequest проксирует запрос к целевому сервису
func proxyRequest(w http.ResponseWriter, r *http.Request, targetBaseURL, targetService string) {
	// Парсим target URL
	targetParsed, err := url.Parse(targetBaseURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid target URL: %v", err), http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetParsed)

	// Добавляем заголовки для трассировки
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Strangler-Route", targetService)
		req.Header.Set("X-Original-Path", r.URL.Path)
	}

	// Убираем конфликтующие заголовки
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Content-Length")
		resp.Header.Del("Transfer-Encoding")
		return nil
	}

	// Обработка ошибок проксирования
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		http.Error(rw, fmt.Sprintf("Ошибка проксирования к %s: %v", targetService, err), http.StatusGatewayTimeout)
		log.Printf("ОШИБКА ПРОКСИРОВАНИЯ к %s: %v\n", targetService, err)
	}

	proxy.ServeHTTP(w, r)
}
