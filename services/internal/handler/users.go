package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"articleswap-scalable/services/internal/models"
)

// UsersHandler returns the list of users from the database.
type UsersHandler struct {
	Pool *pgxpool.Pool
}

// ServeHTTP handles GET /users.
func (h *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.Pool.Query(r.Context(),
		`SELECT id, name, email, created_at FROM users ORDER BY name`)
	if err != nil {
		http.Error(w, `{"error":"failed to query users"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			http.Error(w, `{"error":"failed to scan user"}`, http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error":"error iterating users"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
