package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"articleswap-scalable/services/internal/broker"
	"articleswap-scalable/services/internal/cache"
)

// HealthHandler checks connectivity to PostgreSQL, Redis, and RabbitMQ.
type HealthHandler struct {
	Pool   *pgxpool.Pool
	Cache  *cache.RedisCache
	Broker *broker.Broker
}

type healthResponse struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
	RabbitMQ string `json:"rabbitmq"`
	Time     string `json:"time"`
}

// ServeHTTP handles GET /health.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp := healthResponse{
		Status:   "ok",
		Postgres: "up",
		Redis:    "up",
		RabbitMQ: "up",
		Time:     time.Now().Format(time.RFC3339),
	}

	if err := h.Pool.Ping(ctx); err != nil {
		resp.Postgres = "down"
		resp.Status = "degraded"
	}

	if err := h.Cache.HealthCheck(ctx); err != nil {
		resp.Redis = "down"
		resp.Status = "degraded"
	}

	if err := h.Broker.HealthCheck(); err != nil {
		resp.RabbitMQ = "down"
		resp.Status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}
