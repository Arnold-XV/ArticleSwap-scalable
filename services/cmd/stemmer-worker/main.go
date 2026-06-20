package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"articleswap-scalable/services/internal/broker"
	"articleswap-scalable/services/internal/cache"
	"articleswap-scalable/services/internal/config"
	"articleswap-scalable/services/internal/database"
	"articleswap-scalable/services/internal/models"
)

func doStemming(content string) string {
	// Stemming rule-based sangat sederhana untuk demo
	words := strings.Fields(strings.ToLower(content))
	var stemmed []string
	for _, w := range words {
		w = strings.TrimSuffix(w, "lah")
		w = strings.TrimSuffix(w, "nya")
		w = strings.TrimSuffix(w, "ku")
		w = strings.TrimSuffix(w, "mu")
		stemmed = append(stemmed, w)
	}
	return strings.Join(stemmed, " ")
}

func main() {
	log.Println("[stemmer-worker] starting...")
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("[stemmer-worker] gagal konek database: %v", err)
	}
	defer pool.Close()

	rc, err := cache.NewRedisCache(ctx, cfg)
	if err != nil {
		log.Fatalf("[stemmer-worker] gagal konek redis: %v", err)
	}
	defer rc.Close()

	mq, err := broker.NewBroker(cfg)
	if err != nil {
		log.Fatalf("[stemmer-worker] gagal konek rabbitmq: %v", err)
	}
	defer mq.Close()

	msgs, err := mq.Consume(cfg.QueueStemming)
	if err != nil {
		log.Fatalf("[stemmer-worker] gagal consume queue: %v", err)
	}

	// logEvent mencatat satu baris ke tabel pipeline_events.
	// Non-critical: kalau gagal insert, cukup di-log, tidak menghentikan proses utama.
	logEvent := func(articleID, eventType, message string) {
		_, err := pool.Exec(ctx,
			`INSERT INTO pipeline_events (article_id, service_name, event_type, message)
			 VALUES ($1, 'stemmer-worker', $2, $3)`,
			articleID, eventType, message)
		if err != nil {
			log.Printf("[stemmer] gagal insert pipeline event: %v", err)
		}
	}

	go func() {
		for msg := range msgs {
			var job models.ArticleJob
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("[stemmer] payload rusak, dibuang: %v", err)
				msg.Ack(false) // payload korup, ack saja supaya tidak stuck di queue
				continue
			}

			logEvent(job.ArticleID, "stemming_started", "Stemming dimulai")

			cacheKey := "stem:" + job.ContentHash
			cached, _ := rc.Get(ctx, cacheKey)

			var stemmedContent string
			fromCache := false
			if cached != "" {
				stemmedContent = cached
				fromCache = true
				log.Printf("[stemmer] Cache HIT untuk hash: %s", job.ContentHash)
			} else {
				stemmedContent = doStemming(job.Content)
				if err := rc.Set(ctx, cacheKey, stemmedContent); err != nil {
					log.Printf("[stemmer] gagal set cache (lanjut tanpa cache): %v", err)
				}
				log.Printf("[stemmer] Diproses untuk hash: %s", job.ContentHash)
			}

			// Update DB
			_, dbErr := pool.Exec(ctx, `UPDATE article_processing_results
				SET stemmed_content = $1, stemming_status = 'done',
				    processing_started_at = COALESCE(processing_started_at, NOW())
				WHERE article_id = $2`, stemmedContent, job.ArticleID)

			if dbErr == nil {
				doneMsg := "Stemming selesai"
				if fromCache {
					doneMsg = "Stemming selesai (dari cache)"
				}
				logEvent(job.ArticleID, "stemming_done", doneMsg)

				// Forward ke aggregator
				if err := mq.Publish(ctx, broker.RoutingKeyAggregator, job); err != nil {
					log.Printf("[stemmer] gagal publish ke aggregator: %v", err)
				}
				msg.Ack(false)
				continue
			}

			// --- Retry sederhana dengan batas (max retries dari config) ---
			log.Printf("[stemmer] gagal update DB untuk artikel %s: %v", job.ArticleID, dbErr)

			if job.RetryCount < cfg.WorkerMaxRetries {
				job.RetryCount++
				log.Printf("[stemmer] retry ke-%d/%d untuk artikel %s",
					job.RetryCount, cfg.WorkerMaxRetries, job.ArticleID)
				logEvent(job.ArticleID, "stemming_retry",
					fmt.Sprintf("Retry stemming, percobaan ke-%d", job.RetryCount))

				// Republish job dengan retry_count yang sudah dinaikkan,
				// lalu ack pesan lama (hindari duplikasi requeue otomatis RabbitMQ).
				if err := mq.Publish(ctx, broker.RoutingKeyStemming, job); err != nil {
					log.Printf("[stemmer] gagal republish, fallback ke nack/requeue: %v", err)
					msg.Nack(false, true)
					continue
				}
				msg.Ack(false)
			} else {
				// Jatah retry habis: tandai gagal di DB, buang ke queue "failed", lalu ack.
				log.Printf("[stemmer] retry habis untuk artikel %s, dipindah ke queue failed", job.ArticleID)
				logEvent(job.ArticleID, "stemming_failed", "Stemming gagal setelah retry habis")
				pool.Exec(ctx, `UPDATE article_processing_results
					SET stemming_status = 'failed' WHERE article_id = $1`, job.ArticleID)
				pool.Exec(ctx, `UPDATE articles SET status = 'failed', updated_at = NOW() WHERE id = $1`, job.ArticleID)

				if err := mq.PublishFailed(ctx, job); err != nil {
					log.Printf("[stemmer] gagal publish ke queue failed: %v", err)
				}
				msg.Ack(false)
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("[stemmer-worker] shutting down...")
}
