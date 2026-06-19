package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"articleswap-scalable/services/internal/broker"
	"articleswap-scalable/services/internal/cache"
	"articleswap-scalable/services/internal/config"
	"articleswap-scalable/services/internal/database"
	"articleswap-scalable/services/internal/handler"
	"articleswap-scalable/services/internal/middleware"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[api-gateway] starting ArticleSwap API Gateway")

	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL connection pool.
	pool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("[api-gateway] postgres: %v", err)
	}
	defer pool.Close()

	// Initialize Redis cache.
	rc, err := cache.NewRedisCache(ctx, cfg)
	if err != nil {
		log.Fatalf("[api-gateway] redis: %v", err)
	}
	defer rc.Close()

	// Initialize RabbitMQ broker.
	mq, err := broker.NewBroker(cfg)
	if err != nil {
		log.Fatalf("[api-gateway] rabbitmq: %v", err)
	}
	defer mq.Close()

	// Create handlers.
	healthHandler := &handler.HealthHandler{Pool: pool, Cache: rc, Broker: mq}
	usersHandler := &handler.UsersHandler{Pool: pool}
	articlesHandler := &handler.ArticlesHandler{Pool: pool, Broker: mq}
	inboxHandler := &handler.InboxHandler{Pool: pool}
	metricsHandler := &handler.MetricsHandler{Pool: pool}

	// Build a custom router using http.HandlerFunc for clean path matching.
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case path == "/health":
			healthHandler.ServeHTTP(w, r)

		case path == "/users":
			usersHandler.ServeHTTP(w, r)

		case path == "/articles":
			articlesHandler.ServeHTTP(w, r)

		case strings.HasPrefix(path, "/articles/"):
			articlesHandler.ServeHTTP(w, r)

		case strings.HasPrefix(path, "/users/") && strings.HasSuffix(path, "/inbox"):
			inboxHandler.ServeHTTP(w, r)

		case path == "/metrics/summary":
			metricsHandler.ServeHTTP(w, r)

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	})

	// Apply middleware stack: CORS → Rate Limit → Logging → Router.
	var h http.Handler = router
	h = middleware.Logging(h)
	h = middleware.RateLimit(rc, cfg.RateLimitPerMinute)(h)
	h = middleware.CORS(h)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort),
		Handler:      h,
		ReadTimeout:  cfg.APIReadTimeout,
		WriteTimeout: cfg.APIWriteTimeout,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("[api-gateway] shutting down...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("[api-gateway] shutdown error: %v", err)
		}
	}()

	log.Printf("[api-gateway] listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("[api-gateway] server error: %v", err)
	}
	log.Println("[api-gateway] stopped")
}
