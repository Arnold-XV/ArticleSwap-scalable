package models

import (
	"encoding/json"
	"time"
)

// User represents a platform user seeded in the database.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Article represents an article submitted through the platform.
type Article struct {
	ID          string    `json:"id"`
	SenderID    string    `json:"sender_id"`
	ReceiverID  string    `json:"receiver_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	ContentHash string    `json:"content_hash"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Joined from article_processing_results (optional, populated on detail view).
	Processing *ArticleProcessingResult `json:"processing,omitempty"`
}

// ArticleProcessingResult holds the output from stemmer and wordcloud workers.
type ArticleProcessingResult struct {
	ArticleID           string           `json:"article_id"`
	StemmedContent      *string          `json:"stemmed_content"`
	WordFrequenciesJSON *json.RawMessage `json:"word_frequencies"`
	StemmingStatus      string           `json:"stemming_status"`
	WordcloudStatus     string           `json:"wordcloud_status"`
	ProcessingStartedAt *time.Time       `json:"processing_started_at"`
	ProcessingFinishedAt *time.Time      `json:"processing_finished_at"`
}

// IdempotencyKey prevents duplicate article submissions.
type IdempotencyKey struct {
	Key         string    `json:"key"`
	RequestHash string    `json:"request_hash"`
	ArticleID   string    `json:"article_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// PipelineEvent tracks processing events for observability.
type PipelineEvent struct {
	ID          string    `json:"id"`
	ArticleID   string    `json:"article_id"`
	ServiceName string    `json:"service_name"`
	EventType   string    `json:"event_type"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// ArticleJob is the message payload published to RabbitMQ queues.
type ArticleJob struct {
	ArticleID   string `json:"article_id"`
	ContentHash string `json:"content_hash"`
	Content     string `json:"content"`
	RetryCount  int    `json:"retry_count"`
}

// SubmitArticleRequest is the expected JSON body for POST /articles.
type SubmitArticleRequest struct {
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	IdempotencyKey string `json:"idempotency_key"`
}

// InboxArticle is a simplified view for inbox listing.
type InboxArticle struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
