package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"articleswap-scalable/services/internal/broker"
	"articleswap-scalable/services/internal/config"
	"articleswap-scalable/services/internal/database"
	"articleswap-scalable/services/internal/models"
)

func main() {
	log.Println("[aggregator-worker] starting...")
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("[aggregator-worker] gagal konek database: %v", err)
	}
	defer pool.Close()

	mq, err := broker.NewBroker(cfg)
	if err != nil {
		log.Fatalf("[aggregator-worker] gagal konek rabbitmq: %v", err)
	}
	defer mq.Close()

	msgs, err := mq.Consume(cfg.QueueAggregator)
	if err != nil {
		log.Fatalf("[aggregator-worker] gagal consume queue: %v", err)
	}

	logEvent := func(articleID, eventType, message string) {
		_, err := pool.Exec(ctx,
			`INSERT INTO pipeline_events (article_id, service_name, event_type, message)
			 VALUES ($1, 'aggregator-worker', $2, $3)`,
			articleID, eventType, message)
		if err != nil {
			log.Printf("[aggregator] gagal insert pipeline event: %v", err)
		}
	}

	go func() {
		for msg := range msgs {
			var job models.ArticleJob
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("[aggregator] payload rusak, dibuang: %v", err)
				msg.Ack(false)
				continue
			}

			var sStatus, wStatus, currentArticleStatus string
			err := pool.QueryRow(ctx,
				`SELECT p.stemming_status, p.wordcloud_status, a.status
				 FROM article_processing_results p
				 JOIN articles a ON a.id = p.article_id
				 WHERE p.article_id=$1`,
				job.ArticleID).Scan(&sStatus, &wStatus, &currentArticleStatus)

			if err == nil {
				// Cek apakah KEDUA proses (stem & wordcloud) sudah bukan 'queued' lagi,
				// DAN status artikel belum final (hindari double-update kalau stemmer
				// dan wordcloud sama-sama publish ke aggregator untuk artikel yang sama).
				alreadyFinal := currentArticleStatus == "processed" ||
					currentArticleStatus == "degraded" ||
					currentArticleStatus == "failed"

				if sStatus != "queued" && wStatus != "queued" && !alreadyFinal {
					finalStatus := "processed"
					if sStatus == "failed" || wStatus == "failed" {
						finalStatus = "degraded"
						if sStatus == "failed" && wStatus == "failed" {
							finalStatus = "failed"
						}
					}

					// Update status akhir artikel
					pool.Exec(ctx, `UPDATE articles SET status=$1, updated_at=NOW() WHERE id=$2`, finalStatus, job.ArticleID)
					pool.Exec(ctx, `UPDATE article_processing_results SET processing_finished_at=NOW() WHERE article_id=$1`, job.ArticleID)
					logEvent(job.ArticleID, "pipeline_finished", "Pipeline selesai dengan status: "+finalStatus)
					log.Printf("[aggregator] Artikel %s selesai dengan status: %s", job.ArticleID, finalStatus)
				}
			} else {
				log.Printf("[aggregator] gagal query status artikel %s: %v", job.ArticleID, err)
			}
			// Aggregator bersifat memantau, selalu Ack agar message hilang setelah dievaluasi
			msg.Ack(false)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("[aggregator-worker] shutting down...")
}
