package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// API Gateway
	APIHost             string
	APIPort             string
	APIReadTimeout      time.Duration
	APIWriteTimeout     time.Duration
	RateLimitPerMinute  int

	// PostgreSQL
	DatabaseURL            string
	DBPoolMinConns         int32
	DBPoolMaxConns         int32
	DBPoolMaxConnLifetime  time.Duration

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTL      time.Duration

	// RabbitMQ
	RabbitMQURL      string
	RabbitMQExchange string

	// Queue names
	QueueSubmitted  string
	QueueStemming   string
	QueueWordcloud  string
	QueueAggregator string
	QueueFailed     string

	// Worker
	WorkerConcurrency      int
	WorkerMaxRetries       int
	WorkerBackoffBaseMs    int
	WorkerBackoffMaxMs     int
	CircuitFailureThreshold int
	CircuitOpenSeconds     int
}

// Load reads environment variables and returns a Config with defaults applied.
func Load() *Config {
	return &Config{
		APIHost:             envOrDefault("API_HOST", "0.0.0.0"),
		APIPort:             envOrDefault("API_PORT", "8080"),
		APIReadTimeout:      time.Duration(envIntOrDefault("API_READ_TIMEOUT_SECONDS", 10)) * time.Second,
		APIWriteTimeout:     time.Duration(envIntOrDefault("API_WRITE_TIMEOUT_SECONDS", 10)) * time.Second,
		RateLimitPerMinute:  envIntOrDefault("API_RATE_LIMIT_PER_MINUTE", 60),

		DatabaseURL:           envOrDefault("DATABASE_URL", "postgres://articleswap:articleswap@postgres:5432/articleswap?sslmode=disable"),
		DBPoolMinConns:        int32(envIntOrDefault("DB_POOL_MIN_CONNS", 2)),
		DBPoolMaxConns:        int32(envIntOrDefault("DB_POOL_MAX_CONNS", 20)),
		DBPoolMaxConnLifetime: time.Duration(envIntOrDefault("DB_POOL_MAX_CONN_LIFETIME_MINUTES", 30)) * time.Minute,

		RedisAddr:     envOrDefault("REDIS_ADDR", "redis:6379"),
		RedisPassword: envOrDefault("REDIS_PASSWORD", ""),
		RedisDB:       envIntOrDefault("REDIS_DB", 0),
		CacheTTL:      time.Duration(envIntOrDefault("REDIS_CACHE_TTL_SECONDS", 86400)) * time.Second,

		RabbitMQURL:      envOrDefault("RABBITMQ_URL", "amqp://articleswap:articleswap@rabbitmq:5672/"),
		RabbitMQExchange: envOrDefault("RABBITMQ_EXCHANGE", "articles.exchange"),

		QueueSubmitted:  envOrDefault("QUEUE_ARTICLE_SUBMITTED", "articles.submitted"),
		QueueStemming:   envOrDefault("QUEUE_ARTICLE_STEMMING", "articles.stemming"),
		QueueWordcloud:  envOrDefault("QUEUE_ARTICLE_WORDCLOUD", "articles.wordcloud"),
		QueueAggregator: envOrDefault("QUEUE_ARTICLE_AGGREGATOR", "articles.aggregator"),
		QueueFailed:     envOrDefault("QUEUE_ARTICLE_FAILED", "articles.failed"),

		WorkerConcurrency:      envIntOrDefault("WORKER_CONCURRENCY", 4),
		WorkerMaxRetries:       envIntOrDefault("WORKER_MAX_RETRIES", 3),
		WorkerBackoffBaseMs:    envIntOrDefault("WORKER_BACKOFF_BASE_MS", 500),
		WorkerBackoffMaxMs:     envIntOrDefault("WORKER_BACKOFF_MAX_MS", 8000),
		CircuitFailureThreshold: envIntOrDefault("CIRCUIT_FAILURE_THRESHOLD", 5),
		CircuitOpenSeconds:     envIntOrDefault("CIRCUIT_OPEN_SECONDS", 30),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
