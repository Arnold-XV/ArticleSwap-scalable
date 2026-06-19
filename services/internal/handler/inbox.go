package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"articleswap-scalable/services/internal/models"
)

// InboxHandler handles GET /users/:id/inbox.
type InboxHandler struct {
	Pool *pgxpool.Pool
}

// ServeHTTP handles GET /users/:id/inbox.
func (h *InboxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse user ID from path: /users/:id/inbox
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "inbox" {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}
	userID := parts[0]
	if userID == "" {
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}

	// Verify user exists.
	var exists bool
	err := h.Pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID,
	).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	// Query inbox articles (received by this user), ordered by newest first.
	rows, err := h.Pool.Query(r.Context(),
		`SELECT a.id, a.sender_id, u.name AS sender_name, a.title, a.status, a.created_at
		 FROM articles a
		 JOIN users u ON a.sender_id = u.id
		 WHERE a.receiver_id = $1
		 ORDER BY a.created_at DESC
		 LIMIT 100`, userID)
	if err != nil {
		http.Error(w, `{"error":"failed to query inbox"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	articles := make([]models.InboxArticle, 0)
	for rows.Next() {
		var a models.InboxArticle
		if err := rows.Scan(&a.ID, &a.SenderID, &a.SenderName, &a.Title, &a.Status, &a.CreatedAt); err != nil {
			http.Error(w, `{"error":"failed to scan article"}`, http.StatusInternalServerError)
			return
		}
		articles = append(articles, a)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error":"error iterating inbox"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articles)
}
