# Stress Test ArticleSwap

Folder ini berisi script stress test k6 untuk menguji performa dan ketahanan sistem ArticleSwap.

## Target Pengujian

Stress testing digunakan untuk memenuhi bagian optimasi pada PDF project:

- mengetahui batas sistem,
- mengukur latency,
- mengukur throughput,
- melihat error rate,
- membandingkan performa sebelum dan sesudah optimasi.

## Skenario

| Script | Deskripsi | VU | Durasi |
| --- | --- | --- | --- |
| `baseline.js` | Beban ringan, semua endpoint utama | 10 | 1 menit |
| `stress.js` | Beban naik bertahap sampai 100 VU | 10→100 | 5 menit |
| `cache-test.js` | Submit konten sama berulang untuk membuktikan Redis cache | 10 | 1 menit |
| `idempotency-test.js` | Submit idempotency key sama untuk cek tidak ada duplikasi | 5 | 30 detik |
| `degraded-test.js` | Test saat wordcloud worker dimatikan | 5 | 30 detik |
| `metrics-test.js` | Test endpoint `GET /metrics/summary` di bawah beban | 10 | 1 menit |
| `ratelimit-test.js` | Test rate limiting Redis (memaksa HTTP 429) | 1 | 30 detik |
| `pipeline-test.js` | End-to-end: submit → polling sampai status `completed` | 5 | 2 menit |
| `pool-saturation-test.js` | Membanjiri connection pool PostgreSQL (80 VU) | 20→80 | 2.5 menit |

## Cara Menjalankan

Pastikan aplikasi sudah berjalan:

```bash
docker compose up --build
```

Jalankan salah satu test:

```bash
k6 run stress/baseline.js
```

Atau menggunakan Makefile (Linux/Mac):

```bash
make stress-baseline
make stress-stress
make stress-cache
make stress-idempotency
make stress-degraded
make stress-metrics
make stress-ratelimit
make stress-pipeline
make stress-pool
```

### Catatan Khusus

**Degraded test** — matikan wordcloud worker dulu sebelum menjalankan:

```bash
docker compose stop wordcloud-worker
k6 run stress/degraded-test.js
```

**Rate limit test** — pastikan `API_RATE_LIMIT_PER_MINUTE=60` di `.env` (default).

**Pipeline test** — pastikan semua worker (stemmer, wordcloud, aggregator) sudah berjalan.

**Pool saturation test** — default `DB_POOL_MAX_CONNS=20`. Test ini akan membanjiri sampai 80 VU.

## Data Yang Dicatat

- average latency,
- P95 latency,
- P99 latency,
- throughput (req/s),
- error rate,
- cache hit (pada cache-test),
- rate limit block rate (pada ratelimit-test),
- pipeline completion time (pada pipeline-test),
- connection pool errors (pada pool-saturation-test),
- jumlah artikel sukses/gagal.

Data ini dipakai untuk poster halaman 1.
