# Panduan Project ArticleSwap

Panduan ini ditulis dalam Bahasa Indonesia agar mudah dipakai untuk pengerjaan, demo, dan laporan akhir.

## Ringkasan Tugas Dari PDF

Project meminta tim membangun **ArticleSwap**, platform pertukaran artikel real-time. Artikel yang dikirim tidak hanya disimpan, tetapi juga melewati pipeline pengolahan teks.

Kebutuhan utama:

- merancang arsitektur minimal dalam bentuk diagram visual,
- membangun fitur submit artikel,
- membuat pipeline pengolahan,
- menerapkan stemming,
- membuat word cloud,
- meneruskan artikel ke penerima,
- menjalankan aplikasi dengan Docker,
- memakai message broker,
- memakai database dengan connection pooling,
- melakukan stress testing,
- menerapkan optimasi berbasis hasil pengujian,
- membuat poster/infografis 2 halaman.

## Mapping Bab 10 Ke Implementasi

| Konsep Bab 10 | Implementasi Di Project |
| --- | --- |
| Profiling dan bottleneck | Stress test dengan k6 dan metrics summary |
| Connection pooling | `pgxpool` di Go API Gateway dan worker |
| Cold start vs hot start | Docker service tetap hidup, health check, dan worker scaling |
| Konkurensi dan paralelisme | Worker RabbitMQ berjalan paralel |
| Stress testing | Script k6 di folder `stress/` |
| Caching | Redis cache berdasarkan `content_hash` |
| Asynchronous processing | RabbitMQ queue untuk pipeline |
| Loosely coupled | API, stemmer, wordcloud, aggregator dipisah |

## Mapping Bab 11 Ke Implementasi

| Konsep Bab 11 | Implementasi Di Project |
| --- | --- |
| Redundansi | Scale worker dengan Docker Compose |
| Circuit breaker | Redis menyimpan circuit state per service |
| Retry dan exponential backoff | Worker retry job gagal maksimal 3 kali |
| Failover sederhana | Artikel tetap terkirim walau fitur tambahan gagal |
| Load balancer/gateway | API Gateway sebagai pintu masuk sistem |
| Idempotensi | `idempotencyKey` mencegah artikel duplikat |
| SLI/SLO/SLA | P95/P99 latency dan error rate dari stress test |

## Pembagian Tugas Tim

### Maulana Faris Al Ghifari (24/544029/PA/23119)

Peran: Frontend Developer.

Tugas:

- halaman submit artikel,
- inbox penerima,
- detail artikel,
- integrasi API,
- tampilan status pipeline.

### Raditya Nathaniel Nugroho (24/543188/PA/23069)

Peran: Frontend Developer.

Tugas:

- visual word cloud,
- tampilan hasil stemming,
- state loading/error,
- integrasi polling data artikel.

### Ajie Armansyah Sunaryo (24/545286/PA/23170)

Peran: UI/UX Developer dan QA.

Tugas:

- desain alur penggunaan,
- layout UI Bahasa Indonesia,
- testing manual,
- checklist demo,
- validasi pengalaman pengguna.

### Arnoldus Dharma Wasesa M. (24/545535/PA/23182)

Peran: Backend Developer.

Tugas:

- API Gateway Go,
- PostgreSQL schema,
- connection pooling,
- endpoint artikel, inbox, dan metrics summary.

### Aliya Khairun Nisa (24/543832/PA/23111)

Peran: Backend Developer.

Tugas:

- RabbitMQ pipeline,
- stemmer worker,
- wordcloud worker,
- Redis cache,
- retry/backoff,
- circuit breaker sederhana.

### Semua Anggota

Tugas bersama:

- finalisasi laporan,
- finalisasi poster,
- quality check,
- stress testing,
- latihan demo dan presentasi.

## Urutan Pengerjaan Yang Disarankan

1. Siapkan Docker Compose, `.env`, PostgreSQL, Redis, dan RabbitMQ.
2. Buat schema database dan seed user demo.
3. Implementasi API Gateway.
4. Implementasi worker stemming.
5. Implementasi worker word cloud.
6. Implementasi aggregator worker.
7. Implementasi frontend.
8. Jalankan functional test.
9. Jalankan stress test baseline.
10. Scale worker dan aktifkan cache.
11. Jalankan stress test ulang.
12. Catat hasil untuk poster.
13. Finalisasi dokumentasi dan poster.

## Checklist Demo

- Docker Compose berhasil menjalankan semua service.
- Frontend dapat dibuka di `http://localhost:3000`.
- Artikel bisa dikirim dari satu user ke user lain.
- Artikel muncul di inbox penerima.
- Artikel mentah bisa dibaca sebelum proses selesai.
- Hasil stemming muncul setelah worker selesai.
- Word cloud muncul setelah worker selesai.
- Submit ulang dengan idempotency key sama tidak membuat duplikat.
- Konten artikel sama memakai cache Redis.
- Worker dapat di-scale menjadi 3 replika.
- Stress test menghasilkan angka latency, throughput, dan error rate.

## Narasi Singkat Presentasi

ArticleSwap dirancang sebagai sistem loosely coupled. API Gateway hanya menerima dan menyimpan artikel, sedangkan proses berat seperti stemming dan word cloud dipindahkan ke worker asynchronous melalui RabbitMQ. Dengan desain ini, pengguna tetap mendapat respons cepat walaupun pipeline masih berjalan di belakang layar. PostgreSQL digunakan sebagai penyimpanan utama dengan connection pooling, Redis digunakan untuk caching dan circuit breaker state, lalu Docker Compose digunakan untuk menjalankan dan menskalakan service.

