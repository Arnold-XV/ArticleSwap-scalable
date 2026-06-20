package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"articleswap-scalable/services/internal/broker"
	"articleswap-scalable/services/internal/cache"
	"articleswap-scalable/services/internal/config"
	"articleswap-scalable/services/internal/database"
	"articleswap-scalable/services/internal/models"
)

func doWordcloud(content string) map[string]int {
	freq := make(map[string]int)
	re := regexp.MustCompile(`[a-z0-9]+`)
	words := re.FindAllString(strings.ToLower(content), -1)

	for _, w := range words {
		if len(w) > 3 { // Abaikan kata hubung/pendek
			freq[w]++
		}
	}
	return freq
}

func main() {
	log.Println("[wordcloud-worker] starting...")
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("[wordcloud-worker] gagal konek database: %v", err)
	}
	defer pool.Close()

	rc, err := cache.NewRedisCache(ctx, cfg)
	if err != nil {
		log.Fatalf("[wordcloud-worker] gagal konek redis: %v", err)
	}
	defer rc.Close()

	mq, err := broker.NewBroker(cfg)
	if err != nil {
		log.Fatalf("[wordcloud-worker] gagal konek rabbitmq: %v", err)
	}
	defer mq.Close()

	msgs, err := mq.Consume(cfg.QueueWordcloud)
	if err != nil {
		log.Fatalf("[wordcloud-worker] gagal consume queue: %v", err)
	}

	logEvent := func(articleID, eventType, message string) {
		_, err := pool.Exec(ctx,
			`INSERT INTO pipeline_events (article_id, service_name, event_type, message)
			 VALUES ($1, 'wordcloud-worker', $2, $3)`,
			articleID, eventType, message)
		if err != nil {
			log.Printf("[wordcloud] gagal insert pipeline event: %v", err)
		}
	}

	go func() {
		for msg := range msgs {
			var job models.ArticleJob
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("[wordcloud] payload rusak, dibuang: %v", err)
				msg.Ack(false)
				continue
			}

			logEvent(job.ArticleID, "wordcloud_started", "Pembuatan word cloud dimulai")

			cacheKey := "wc:" + job.ContentHash
			cached, _ := rc.Get(ctx, cacheKey)

			var wcJSON string
			fromCache := false
			if cached != "" {
				wcJSON = cached
				fromCache = true
				log.Printf("[wordcloud] Cache HIT untuk hash: %s", job.ContentHash)
			} else {
				freq := doWordcloud(job.Content)
				b, _ := json.Marshal(freq)
				wcJSON = string(b)
				if err := rc.Set(ctx, cacheKey, wcJSON); err != nil {
					log.Printf("[wordcloud] gagal set cache (lanjut tanpa cache): %v", err)
				}
				log.Printf("[wordcloud] Diproses untuk hash: %s", job.ContentHash)
			}

			// Update DB
			_, dbErr := pool.Exec(ctx, `UPDATE article_processing_results
				SET word_frequencies_json = $1, wordcloud_status = 'done',
				    processing_started_at = COALESCE(processing_started_at, NOW())
				WHERE article_id = $2`, wcJSON, job.ArticleID)

			if dbErr == nil {
				doneMsg := "Word cloud selesai"
				if fromCache {
					doneMsg = "Word cloud selesai (dari cache)"
				}
				logEvent(job.ArticleID, "wordcloud_done", doneMsg)

				if err := mq.Publish(ctx, broker.RoutingKeyAggregator, job); err != nil {
					log.Printf("[wordcloud] gagal publish ke aggregator: %v", err)
				}
				msg.Ack(false)
				continue
			}

			// --- Retry sederhana dengan batas (max retries dari config) ---
			log.Printf("[wordcloud] gagal update DB untuk artikel %s: %v", job.ArticleID, dbErr)

			if job.RetryCount < cfg.WorkerMaxRetries {
				job.RetryCount++
				log.Printf("[wordcloud] retry ke-%d/%d untuk artikel %s",
					job.RetryCount, cfg.WorkerMaxRetries, job.ArticleID)
				logEvent(job.ArticleID, "wordcloud_retry",
					fmt.Sprintf("Retry word cloud, percobaan ke-%d", job.RetryCount))

				if err := mq.Publish(ctx, broker.RoutingKeyWordcloud, job); err != nil {
					log.Printf("[wordcloud] gagal republish, fallback ke nack/requeue: %v", err)
					msg.Nack(false, true)
					continue
				}
				msg.Ack(false)
			} else {
				log.Printf("[wordcloud] retry habis untuk artikel %s, dipindah ke queue failed", job.ArticleID)
				logEvent(job.ArticleID, "wordcloud_failed", "Word cloud gagal setelah retry habis")
				pool.Exec(ctx, `UPDATE article_processing_results
					SET wordcloud_status = 'failed' WHERE article_id = $1`, job.ArticleID)
				pool.Exec(ctx, `UPDATE articles SET status = 'failed', updated_at = NOW() WHERE id = $1`, job.ArticleID)

				if err := mq.PublishFailed(ctx, job); err != nil {
					log.Printf("[wordcloud] gagal publish ke queue failed: %v", err)
				}
				msg.Ack(false)
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("[wordcloud-worker] shutting down...")
}
