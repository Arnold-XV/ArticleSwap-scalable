# Stress Test ArticleSwap

Folder ini berisi script stress test k6 yang akan ditambahkan saat implementasi endpoint selesai.

## Target Pengujian

Stress testing digunakan untuk memenuhi bagian optimasi pada PDF project:

- mengetahui batas sistem,
- mengukur latency,
- mengukur throughput,
- melihat error rate,
- membandingkan performa sebelum dan sesudah optimasi.

## Skenario

Script yang akan tersedia:

- `baseline.js`: 10 virtual users selama 1 menit.
- `stress.js`: beban naik bertahap sampai 100 virtual users.
- `cache-test.js`: submit konten sama berulang untuk membuktikan Redis cache.
- `idempotency-test.js`: submit idempotency key sama untuk memastikan tidak ada duplikasi.
- `degraded-test.js`: test saat wordcloud worker dimatikan.

## Cara Menjalankan

Pastikan aplikasi sudah berjalan:

```bash
docker compose up --build
```

Jalankan salah satu test:

```bash
k6 run stress/baseline.js
```

## Data Yang Dicatat

- average latency,
- P95 latency,
- P99 latency,
- throughput,
- error rate,
- cache hit,
- jumlah artikel sukses,
- jumlah artikel gagal.

Data ini dipakai untuk poster halaman 1.

