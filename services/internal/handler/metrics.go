package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MetricsHandler returns a summary of system metrics from PostgreSQL.
// Useful for poster data and observability.
type MetricsHandler struct {
	Pool *pgxpool.Pool
}

type metricsSummary struct {
	TotalUsers    int            `json:"total_users"`
	TotalArticles int            `json:"total_articles"`
	ArticlesByStatus map[string]int `json:"articles_by_status"`
	PipelineEvents int           `json:"pipeline_events"`
	IdempotencyKeys int          `json:"idempotency_keys"`
	ProcessingSummary processingStats `json:"processing_summary"`
}

type processingStats struct {
	StemmingDone   int `json:"stemming_done"`
	StemmingQueued int `json:"stemming_queued"`
	StemmingFailed int `json:"stemming_failed"`
	WordcloudDone   int `json:"wordcloud_done"`
	WordcloudQueued int `json:"wordcloud_queued"`
	WordcloudFailed int `json:"wordcloud_failed"`
}

// ServeHTTP handles GET /metrics/summary.
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	var s metricsSummary
	s.ArticlesByStatus = make(map[string]int)

	// Total users.
	_ = h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&s.TotalUsers)

	// Total articles.
	_ = h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM articles`).Scan(&s.TotalArticles)

	// Articles by status.
	rows, err := h.Pool.Query(ctx,
		`SELECT status, COUNT(*) FROM articles GROUP BY status`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if rows.Scan(&status, &count) == nil {
				s.ArticlesByStatus[status] = count
			}
		}
	}

	// Pipeline events count.
	_ = h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM pipeline_events`).Scan(&s.PipelineEvents)

	// Idempotency keys count.
	_ = h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM idempotency_keys`).Scan(&s.IdempotencyKeys)

	// Processing results breakdown.
	_ = h.Pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN stemming_status = 'done' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN stemming_status = 'queued' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN stemming_status = 'failed' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN wordcloud_status = 'done' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN wordcloud_status = 'queued' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN wordcloud_status = 'failed' THEN 1 ELSE 0 END), 0)
		 FROM article_processing_results`,
	).Scan(
		&s.ProcessingSummary.StemmingDone,
		&s.ProcessingSummary.StemmingQueued,
		&s.ProcessingSummary.StemmingFailed,
		&s.ProcessingSummary.WordcloudDone,
		&s.ProcessingSummary.WordcloudQueued,
		&s.ProcessingSummary.WordcloudFailed,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
