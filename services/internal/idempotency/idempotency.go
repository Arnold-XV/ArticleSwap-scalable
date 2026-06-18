package idempotency

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Check looks up an idempotency key in the database.
// Returns the existing article ID if found, or empty string if not.
func Check(ctx context.Context, pool *pgxpool.Pool, key string) (articleID string, exists bool, err error) {
	err = pool.QueryRow(ctx,
		`SELECT article_id FROM idempotency_keys WHERE key = $1`, key,
	).Scan(&articleID)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("idempotency check failed: %w", err)
	}
	return articleID, true, nil
}

// Store saves an idempotency key with its associated request hash and article ID.
func Store(ctx context.Context, pool *pgxpool.Pool, key, requestHash, articleID string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO idempotency_keys (key, request_hash, article_id) VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO NOTHING`,
		key, requestHash, articleID,
	)
	if err != nil {
		return fmt.Errorf("idempotency store failed: %w", err)
	}
	return nil
}
