.PHONY: help env up down logs ps scale-workers stress-baseline stress-stress stress-cache stress-idempotency stress-degraded stress-metrics stress-ratelimit stress-pipeline stress-pool

help:
	@echo "ArticleSwap commands:"
	@echo "  make env                 Salin .env.example ke .env jika belum ada"
	@echo "  make up                  Jalankan semua service dengan Docker Compose"
	@echo "  make down                Matikan semua service"
	@echo "  make logs                Lihat log semua service"
	@echo "  make ps                  Lihat status container"
	@echo "  make scale-workers       Jalankan worker paralel untuk demo scaling"
	@echo "  make stress-baseline     Jalankan stress test baseline"
	@echo "  make stress-stress       Jalankan stress test bertahap"
	@echo "  make stress-cache        Jalankan test cache Redis"
	@echo "  make stress-idempotency  Jalankan test idempotency"
	@echo "  make stress-degraded     Jalankan test degraded mode"
	@echo "  make stress-metrics      Jalankan test endpoint metrics"
	@echo "  make stress-ratelimit    Jalankan test rate limiting"
	@echo "  make stress-pipeline     Jalankan test pipeline end-to-end"
	@echo "  make stress-pool         Jalankan test connection pool saturation"

env:
	@if not exist .env copy .env.example .env

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

scale-workers:
	docker compose up --build --scale stemmer-worker=3 --scale wordcloud-worker=3

stress-baseline:
	k6 run stress/baseline.js

stress-stress:
	k6 run stress/stress.js

stress-cache:
	k6 run stress/cache-test.js

stress-idempotency:
	k6 run stress/idempotency-test.js

stress-degraded:
	k6 run stress/degraded-test.js

stress-metrics:
	k6 run stress/metrics-test.js

stress-ratelimit:
	k6 run stress/ratelimit-test.js

stress-pipeline:
	k6 run stress/pipeline-test.js

stress-pool:
	k6 run stress/pool-saturation-test.js

