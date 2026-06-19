package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"articleswap-scalable/services/internal/config"
)

// NewPostgresPool creates a pgxpool.Pool with connection pooling parameters
// from the provided configuration. This demonstrates Bab 10 connection pooling.
func NewPostgresPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolCfg.MinConns = cfg.DBPoolMinConns
	poolCfg.MaxConns = cfg.DBPoolMaxConns
	poolCfg.MaxConnLifetime = cfg.DBPoolMaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("[postgres] pool created — min=%d max=%d lifetime=%s",
		cfg.DBPoolMinConns, cfg.DBPoolMaxConns, cfg.DBPoolMaxConnLifetime)

	return pool, nil
}

// HealthCheck pings the database to verify connectivity.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
