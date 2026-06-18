package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"articleswap-scalable/services/internal/broker"
	"articleswap-scalable/services/internal/idempotency"
	"articleswap-scalable/services/internal/models"
)

// ArticlesHandler handles POST /articles and GET /articles/:id.
type ArticlesHandler struct {
	Pool   *pgxpool.Pool
	Broker *broker.Broker
}

// ServeHTTP routes to the appropriate handler method.
func (h *ArticlesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route: POST /articles
	if r.URL.Path == "/articles" && r.Method == http.MethodPost {
		h.createArticle(w, r)
		return
	}

	// Route: GET /articles/:id
	if strings.HasPrefix(r.URL.Path, "/articles/") && r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/articles/")
		if id == "" {
			http.Error(w, `{"error":"article id is required"}`, http.StatusBadRequest)
			return
		}
		h.getArticle(w, r, id)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

// createArticle handles POST /articles.
func (h *ArticlesHandler) createArticle(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields.
	if req.SenderID == "" || req.ReceiverID == "" || req.Title == "" || req.Content == "" {
		http.Error(w, `{"error":"sender_id, receiver_id, title, and content are required"}`, http.StatusBadRequest)
		return
	}

	if req.IdempotencyKey == "" {
		http.Error(w, `{"error":"idempotency_key is required"}`, http.StatusBadRequest)
		return
	}

	// Check idempotency key — if this key was already used, return the existing article.
	existingArticleID, exists, err := idempotency.Check(r.Context(), h.Pool, req.IdempotencyKey)
	if err != nil {
		log.Printf("[articles] idempotency check error: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	if exists {
		// Return the existing article (idempotent response).
		article, err := h.queryArticleByID(r, existingArticleID)
		if err != nil {
			http.Error(w, `{"error":"failed to fetch existing article"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(article)
		return
	}

	// Compute content hash for caching.
	hash := sha256.Sum256([]byte(req.Content))
	contentHash := hex.EncodeToString(hash[:])

	// Begin transaction: insert article + processing_results + idempotency key.
	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		log.Printf("[articles] begin tx error: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var articleID string
	err = tx.QueryRow(r.Context(),
		`INSERT INTO articles (sender_id, receiver_id, title, content, content_hash, status)
		 VALUES ($1, $2, $3, $4, $5, 'queued')
		 RETURNING id`,
		req.SenderID, req.ReceiverID, req.Title, req.Content, contentHash,
	).Scan(&articleID)
	if err != nil {
		log.Printf("[articles] insert article error: %v", err)
		http.Error(w, `{"error":"failed to create article"}`, http.StatusInternalServerError)
		return
	}

	// Create processing results row.
	_, err = tx.Exec(r.Context(),
		`INSERT INTO article_processing_results (article_id) VALUES ($1)`, articleID)
	if err != nil {
		log.Printf("[articles] insert processing results error: %v", err)
		http.Error(w, `{"error":"failed to create processing results"}`, http.StatusInternalServerError)
		return
	}

	// Store idempotency key.
	requestHash := fmt.Sprintf("%s:%s:%s:%s", req.SenderID, req.ReceiverID, req.Title, contentHash)
	_, err = tx.Exec(r.Context(),
		`INSERT INTO idempotency_keys (key, request_hash, article_id) VALUES ($1, $2, $3)`,
		req.IdempotencyKey, requestHash, articleID)
	if err != nil {
		log.Printf("[articles] insert idempotency key error: %v", err)
		http.Error(w, `{"error":"failed to store idempotency key"}`, http.StatusInternalServerError)
		return
	}

	// Log pipeline event.
	_, err = tx.Exec(r.Context(),
		`INSERT INTO pipeline_events (article_id, service_name, event_type, message)
		 VALUES ($1, 'api-gateway', 'article_submitted', 'Article queued for processing')`,
		articleID)
	if err != nil {
		log.Printf("[articles] insert pipeline event error: %v", err)
		// Non-critical, continue.
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("[articles] commit error: %v", err)
		http.Error(w, `{"error":"failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	// Publish jobs to stemming and wordcloud queues (fan-out).
	job := models.ArticleJob{
		ArticleID:   articleID,
		ContentHash: contentHash,
		Content:     req.Content,
		RetryCount:  0,
	}

	if err := h.Broker.Publish(r.Context(), "article.stemming", job); err != nil {
		log.Printf("[articles] publish stemming error: %v", err)
	}
	if err := h.Broker.Publish(r.Context(), "article.wordcloud", job); err != nil {
		log.Printf("[articles] publish wordcloud error: %v", err)
	}

	log.Printf("[articles] created article=%s sender=%s receiver=%s hash=%s",
		articleID, req.SenderID, req.ReceiverID, contentHash[:12])

	// Return the created article.
	article, err := h.queryArticleByID(r, articleID)
	if err != nil {
		// Article was created but fetch failed — still return 202 with minimal info.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"id": articleID, "status": "queued"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(article)
}

// getArticle handles GET /articles/:id.
func (h *ArticlesHandler) getArticle(w http.ResponseWriter, r *http.Request, id string) {
	article, err := h.queryArticleByID(r, id)
	if err != nil {
		http.Error(w, `{"error":"article not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

// queryArticleByID fetches an article with its processing results.
func (h *ArticlesHandler) queryArticleByID(r *http.Request, id string) (*models.Article, error) {
	var a models.Article
	var p models.ArticleProcessingResult

	err := h.Pool.QueryRow(r.Context(),
		`SELECT
			a.id, a.sender_id, a.receiver_id, a.title, a.content, a.content_hash,
			a.status, a.created_at, a.updated_at,
			p.stemmed_content, p.word_frequencies_json, p.stemming_status, p.wordcloud_status,
			p.processing_started_at, p.processing_finished_at
		 FROM articles a
		 LEFT JOIN article_processing_results p ON a.id = p.article_id
		 WHERE a.id = $1`, id,
	).Scan(
		&a.ID, &a.SenderID, &a.ReceiverID, &a.Title, &a.Content, &a.ContentHash,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
		&p.StemmedContent, &p.WordFrequenciesJSON, &p.StemmingStatus, &p.WordcloudStatus,
		&p.ProcessingStartedAt, &p.ProcessingFinishedAt,
	)
	if err != nil {
		return nil, err
	}

	p.ArticleID = a.ID
	a.Processing = &p
	return &a, nil
}
